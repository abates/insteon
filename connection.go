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
	"fmt"
	"io"
	"sync"
	"time"
)

// Sender will deliver an Insteon message to the network
type Sender interface {
	// Send will send a message to the device
	Send(*Message) error
}

type Demux interface {
	Dispatch(*Message)
	New(addr Address, options ...ConnectionOption) (Connection, error)
}

func NewDemux(sender Sender) Demux {
	return &demux{
		Sender:      sender,
		connections: make(map[Address]*connection),
	}
}

type demux struct {
	Sender
	mu          sync.Mutex
	connections map[Address]*connection
}

func (d *demux) Dispatch(msg *Message) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for addr, conn := range d.connections {
		if addr == Wildcard || addr == msg.Src {
			conn.dispatch(msg)
			break
		}
	}
}

func (d *demux) New(addr Address, options ...ConnectionOption) (Connection, error) {
	var err error
	d.mu.Lock()
	defer d.mu.Unlock()
	conn, found := d.connections[addr]
	if !found {
		conn, err = newConnection(d, addr, options...)
		d.connections[addr] = conn
	}
	return conn, err
}

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

	// Lock the connection so that it not usable by other go routines.  This is
	// implemented by an underlying sync.Mutex object
	Lock()

	// Unlock is the complement to the Lock function effectively unlocking the Mutex
	Unlock()
}

type connection struct {
	*sync.Mutex

	addr     Address
	match    []Command
	timeout  time.Duration
	ttl      uint8
	upstream Sender

	msgCh   chan *Message
	closeCh chan chan error
}

// ConnectionOption provides a means to customize the connection config
type ConnectionOption func(*connection) error

// ConnectionTTL will set the connection's time to live flag
func ConnectionTTL(ttl uint8) ConnectionOption {
	return func(conn *connection) error {
		if ttl < 0 || ttl > 3 {
			return fmt.Errorf("invalid ttl %d, must be in range 0-3", ttl)
		}
		conn.ttl = ttl
		return nil
	}
}

// ConnectionTimeout is a ConnectionOption that will set the connection's read
// timeout
func ConnectionTimeout(timeout time.Duration) ConnectionOption {
	return func(conn *connection) error {
		conn.timeout = timeout
		return nil
	}
}

// ConnectionFilter will configure the connection to filter all traffic
// except messages with matching commands
func ConnectionFilter(match ...Command) ConnectionOption {
	return func(conn *connection) error {
		conn.match = match
		return nil
	}
}

// ConnectionMutex provides a way to set the underlying Mutex.  This allows a global
// mutex to be used (as in the case of a PLM)
func ConnectionMutex(mu *sync.Mutex) ConnectionOption {
	return func(conn *connection) error {
		if conn != nil {
			conn.Mutex = mu
		}
		return nil
	}
}

// NewConnection will return a connection that is setup and ready to be used.  The txCh and
// rxCh will be used to send and receive insteon messages.  Any supplied options will be
// used to customize the connection's config
func NewConnection(upstream Sender, addr Address, options ...ConnectionOption) (Connection, error) {
	return newConnection(upstream, addr, options...)
}

func newConnection(upstream Sender, addr Address, options ...ConnectionOption) (*connection, error) {
	conn := &connection{
		Mutex: &sync.Mutex{},

		addr:    addr,
		timeout: 3 * time.Second,

		upstream: upstream,
		msgCh:    make(chan *Message, 1),
		closeCh:  make(chan chan error),
	}

	for _, option := range options {
		err := option(conn)
		if err != nil {
			Log.Infof("error setting connection option: %v", err)
			return nil, err
		}
	}

	return conn, nil
}

func (conn *connection) Address() Address {
	return conn.addr
}

func (conn *connection) dispatch(msg *Message) {
	if len(conn.match) > 0 {
		for _, m := range conn.match {
			if (msg.Command == m) || (msg.Command[1] == m[1] && m[2] == 0x00) {
				Log.Tracef("Connection %v RX %v", conn.addr, msg)
				conn.msgCh <- msg
			}
		}
	} else {
		Log.Tracef("Connection %v RX %v", conn.addr, msg)
		conn.msgCh <- msg
	}
}

func (conn *connection) Send(msg *Message) (ack *Message, err error) {
	msg.Dst = conn.addr
	msg.Flags = Flag(MsgTypeDirect, len(msg.Payload) > 0, conn.ttl, conn.ttl)
	Log.Tracef("Connection %v TX %v", conn.addr, msg)
	err = conn.upstream.Send(msg)

	if err == nil {
		// wait for ack
		err = Receive(conn, conn.timeout, func(msg *Message) (err error) {
			if msg.Ack() || msg.Nak() {
				ack = msg
				err = ErrReceiveComplete
			}
			return err
		})
	}
	return ack, err
}

func (conn *connection) Receive() (msg *Message, err error) {
	return readFromCh(conn.msgCh, conn.timeout)
}

func (conn *connection) IDRequest() (version FirmwareVersion, devCat DevCat, err error) {
	_, err = conn.Send(&Message{Command: CmdIDRequest, Flags: StandardDirectMessage})
	err = Receive(conn, conn.timeout, func(msg *Message) error {
		if msg.Broadcast() && (msg.Command[1] == 0x01 || msg.Command[1] == 0x02) {
			version = FirmwareVersion(msg.Dst[2])
			devCat = DevCat{msg.Dst[0], msg.Dst[1]}
			err = ErrReceiveComplete
		}
		return err
	})
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

// Receive is a utility function that wraps up receiving with timeout functionality allowing
// callers to only deal with the received messages and not dealing with timeout circumstances.
// When the callback is done receiving an ErrReceiveComplete should be returned causing Receive
// to return with no error.  If the callback returns ErrReadContinue then the timeout is updated
// for an additional read.  If the callback returns any other error then that error will
// be returned
func Receive(conn Connection, timeout time.Duration, cb func(*Message) error) (err error) {
	readTimeout := time.Now().Add(timeout)
	for err == nil {
		var msg *Message
		msg, err = conn.Receive()
		if err == nil {
			if readTimeout.Before(time.Now()) {
				err = ErrReadTimeout
			} else {
				err = cb(msg)
				if err == ErrReceiveContinue {
					readTimeout = time.Now().Add(timeout)
					err = nil
				}
			}
		}
	}
	if err == ErrReceiveComplete {
		err = nil
	}
	return err
}

func readFromCh(ch <-chan *Message, timeout time.Duration) (msg *Message, err error) {
	var open bool
	select {
	case msg, open = <-ch:
		if !open {
			err = io.EOF
		}
	case <-time.After(timeout):
		err = ErrReadTimeout
	}
	return
}
