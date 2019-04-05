// Copyright 2018 Andrew Bates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package insteon

import (
	"sync"
	"time"
)

// Connection is a very basic communication mechanism to send
// and receive messages with individual Insteon devices.
type Connection interface {
	// Address returns the unique Insteon address of the device
	Address() Address

	// Send will send the message to the device and wait for the
	// device to respond with an Ack/Nak.  Send will always return
	// but may return with a read timeout or other communication error
	Send(*Message) (ack *Message, err error)

	// Receive waits for the next message from the device.  Receive
	// always returns, but may return with an error (such as ErrReadTimeout)
	Receive() (*Message, error)

	// IDRequest sends an IDRequest command to the device and waits for
	// the corresponding Set Button Pressed Controller/Responder message.
	// The response is parsed and the Firmware version and DevCat are
	// returned.  A ReadTimeout may occur if the device doesn't respond
	// with the appropriate broadcast message, or if the local system
	// doesn't receive it
	IDRequest() (FirmwareVersion, DevCat, error)

	// EngineVersion will query the device for its Insteon Engine Version
	// and returns the response.  If the device never responds, then ErrReadTimeout
	// is the returned error.  If the device responds with a Nak and Command 2
	// is 0xff then that means the engine version is I2Cs and the device does
	// not have a corresponding link entry.  In this case, VerI2Cs is returned
	// as well as ErrNotLinked
	EngineVersion() (EngineVersion, error)

	// AddListener will return a channel that receives any messages matching
	// the flags an the cmd1 flag of a Command.
	AddListener(t MessageType, cmds ...Command) <-chan *Message

	// RemoveListener will remove a previously allocated listener channel to be
	// closed and removed from the connection
	RemoveListener(<-chan *Message)

	// Lock the connection so that it not usable by other go routines.  This is
	// implemented by an underlying sync.Mutex object
	Lock()

	// Unlock is the complement to the Lock function effectively unlocking the Mutex
	Unlock()
}

type connection struct {
	*sync.Mutex
	*msgListeners

	addr    Address
	match   []Command
	timeout time.Duration

	txCh    chan<- *Message
	rxCh    <-chan *Message
	msgCh   chan *Message
	closeCh chan chan error
}

// ConnectionOption provides a means to customize the connection config
type ConnectionOption func(*connection)

// ConnectionTimeout is a ConnectionOption that will set the connection's read
// timeout
func ConnectionTimeout(timeout time.Duration) ConnectionOption {
	return func(conn *connection) {
		conn.timeout = timeout
	}
}

// ConnectionFilter will configure the connection to filter all traffic
// except messages with matching commands
func ConnectionFilter(match ...Command) ConnectionOption {
	return func(conn *connection) {
		conn.match = match
	}
}

// ConnectionMutex provides a way to set the underlying Mutex.  This allows a global
// mutex to be used (as in the case of a PLM)
func ConnectionMutex(mu *sync.Mutex) ConnectionOption {
	return func(conn *connection) {
		if conn != nil {
			conn.Mutex = mu
		}
	}
}

type msgListener struct {
	ch   chan<- *Message
	t    MessageType
	cmds []Command
}

type msgListeners struct {
	mu        sync.Mutex
	listeners map[<-chan *Message]*msgListener
}

func (ml *msgListeners) deliver(msg *Message) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	for _, listener := range ml.listeners {
		if msg.Flags.Type() == listener.t {
			for _, cmd := range listener.cmds {
				if msg.Command[1] == cmd[1] {
					listener.ch <- msg
					return
				}
			}
		}
	}
}

func (ml *msgListeners) RemoveListener(ch <-chan *Message) {
	ml.mu.Lock()
	if listener, found := ml.listeners[ch]; found {
		close(listener.ch)
		delete(ml.listeners, ch)
	}
	ml.mu.Unlock()
}

func (ml *msgListeners) AddListener(t MessageType, cmds ...Command) <-chan *Message {
	ch := make(chan *Message, 1)
	ml.mu.Lock()
	ml.listeners[ch] = &msgListener{ch, t, cmds}
	ml.mu.Unlock()
	return ch
}

// NewConnection will return a connection that is setup and ready to be used.  The txCh and
// rxCh will be used to send and receive insteon messages.  Any supplied options will be
// used to customize the connection's config
func NewConnection(txCh chan<- *Message, rxCh <-chan *Message, addr Address, options ...ConnectionOption) Connection {
	conn := &connection{
		Mutex:        &sync.Mutex{},
		msgListeners: &msgListeners{listeners: make(map[<-chan *Message]*msgListener)},

		addr:    addr,
		timeout: 3 * time.Second,

		txCh:    txCh,
		rxCh:    rxCh,
		msgCh:   make(chan *Message, 10),
		closeCh: make(chan chan error),
	}

	for _, option := range options {
		option(conn)
	}

	go conn.readLoop()
	return conn
}

func (conn *connection) Address() Address {
	return conn.addr
}

func (conn *connection) readLoop() {
	for {
		select {
		case msg := <-conn.rxCh:
			if msg.Src == conn.addr {
				if len(conn.match) > 0 {
					for _, m := range conn.match {
						if (msg.Command == m) || (msg.Command[1] == m[1] && m[2] == 0x00) {
							Log.Tracef("Connection %v RX %v", conn.addr, msg)
							conn.msgListeners.deliver(msg)
							conn.msgCh <- msg
						}
					}
				} else {
					Log.Tracef("Connection %v RX %v", conn.addr, msg)
					conn.msgListeners.deliver(msg)
					conn.msgCh <- msg
				}
			}
		case ch := <-conn.closeCh:
			close(conn.msgCh)
			ch <- nil
			return
		}
	}
}

func (conn *connection) Send(msg *Message) (ack *Message, err error) {
	msg.Dst = conn.addr
	Log.Tracef("Connection %v TX %v", conn.addr, msg)
	conn.txCh <- msg

	// wait for ack
	timeout := time.Now().Add(conn.timeout)
	for err == nil {
		ack, err = conn.Receive()
		if err == nil && (ack.Ack() || ack.Nak()) {
			break
		} else if timeout.Before(time.Now()) {
			err = ErrReadTimeout
		}
	}

	return ack, err
}

func (conn *connection) Receive() (msg *Message, err error) {
	return readFromCh(conn.msgCh, conn.timeout)
}

func (conn *connection) Close() error {
	ch := make(chan error)
	conn.closeCh <- ch
	return <-ch
}

func (conn *connection) IDRequest() (version FirmwareVersion, devCat DevCat, err error) {
	_, err = conn.Send(&Message{Command: CmdIDRequest, Flags: StandardDirectMessage})
	timeout := time.Now().Add(conn.timeout)
	for err == nil {
		var msg *Message
		msg, err = conn.Receive()
		if err == nil {
			if msg.Broadcast() && (msg.Command[1] == 0x01 || msg.Command[1] == 0x02) {
				version = FirmwareVersion(msg.Dst[2])
				devCat = DevCat{msg.Dst[0], msg.Dst[1]}
				break
			} else if timeout.Before(time.Now()) {
				err = ErrReadTimeout
			}
		}
	}
	return
}

func (conn *connection) EngineVersion() (version EngineVersion, err error) {
	ack, err := conn.Send(&Message{Command: CmdGetEngineVersion, Flags: StandardDirectMessage})
	if err == nil {
		if ack.Nak() {
			// This only happens if the device is an I2Cs device and
			// is not linked to the queryier
			if ack.Command[2] == 0xff {
				version = VerI2Cs
				err = ErrNotLinked
			} else {
				err = ErrNak
			}
		} else {
			version = EngineVersion(ack.Command[2])
		}
	}
	return
}

func readFromCh(ch <-chan *Message, timeout time.Duration) (msg *Message, err error) {
	if err == nil {
		select {
		case msg = <-ch:
		case <-time.After(timeout):
			err = ErrReadTimeout
		}
	}
	return
}
