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

const (
	NumRetries = 5

	Timeout = time.Second * 5
)

// Sender will deliver an Insteon message to the network
type Sender interface {
	// Send will send a message to the device
	Send(*Message) error
}

type MessageReader interface {
	ReadMessage() (*Message, error)
}

type MessageWriter interface {
	WriteMessage(*Message) error
}

type Bridge interface {
	MessageReader
	MessageWriter
}

type Dialer interface {
	Dial(dst Address, cmds ...Command) (conn Connection, err error)
}

// Demux provides a way to demultiplex incoming Insteon messages to
// individual devices.  This abstracts the complexity away from things
// like the PLM so that the PLM can focus solely on bridging between the
// Insteon network and the software interface.  The Demux itself is
// not thread safe and any locking should be provided by the caller.  In
// most cases, if the demux is only used within an event loop (such as the
// readLoop in the PLM implementation) then no locking is required.  The
// underlying Connections use channel to pass messages so they are inherently
// thread safe
type Demux interface {
	Dialer
	MessageWriter
	Dispatch(msg *Message)
}

type demux struct {
	// Sender is the upstream device (PLM, for instance)
	MessageWriter
	reader      MessageReader
	connections map[Address][]*connection
	lock        sync.Mutex
	options     []ConnectionOption
}

func NewDemux(writer MessageWriter, options ...ConnectionOption) Demux {
	d := &demux{
		MessageWriter: writer,
		connections:   make(map[Address][]*connection),
	}

	return d
}

func (d *demux) Dispatch(msg *Message) {
	d.lock.Lock()
	defer d.lock.Unlock()
	Log.Debugf("RX %s", msg)
	for _, addr := range []Address{msg.Src, Wildcard} {
		for _, conn := range d.connections[addr] {
			conn.dispatch(msg)
		}
	}
}

func (d *demux) Dial(addr Address, cmds ...Command) (Connection, error) {
	conn, err := newConnection(d, addr, cmds, d.options...)

	if err == nil {
		d.lock.Lock()
		defer d.lock.Unlock()
		d.connections[addr] = append(d.connections[addr], conn)
	}
	return conn, err
}

func (d *demux) close(conn *connection) {
	d.lock.Lock()
	defer d.lock.Unlock()
	for i, lconn := range d.connections[conn.addr] {
		if conn == lconn {
			d.connections[conn.addr] = append(d.connections[conn.addr][0:i], d.connections[conn.addr][i+1:]...)
			break
		}
	}
}

type MessageSender interface {
	// Send will send the message to the device and wait for the
	// device to respond with an Ack/Nak.  Send will always return
	// but may return with a read timeout or other communication error
	Send(*Message) (ack *Message, err error)
}

type MessageReceiver interface {
	// Receive waits for the next message from the device.  Receive
	// always returns, but may return with an error (such as ErrReadTimeout)
	Receive() (*Message, error)
}

// Connection is a very basic communication mechanism to send
// and receive messages with individual Insteon devices.
type Connection interface {
	MessageReceiver
	io.Closer

	Send(msg *Message) (ack *Message, err error)
}

type connection struct {
	MessageWriter
	recvCh  chan *Message
	addr    Address
	match   []Command
	timeout time.Duration
	ttl     uint8
	retries int
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

func newConnection(writer MessageWriter, addr Address, cmds []Command, options ...ConnectionOption) (*connection, error) {
	conn := &connection{
		MessageWriter: writer,
		recvCh:        make(chan *Message, 1),
		addr:          addr,
		match:         cmds,
		timeout:       Timeout,
		retries:       NumRetries,
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

func (conn *connection) Close() error {
	if demux, ok := conn.MessageWriter.(*demux); ok {
		demux.close(conn)
	}
	return nil
}

func (conn *connection) Send(msg *Message) (ack *Message, err error) {
	msg.Dst = conn.addr
	msg.Flags = StandardDirectMessage
	if len(msg.Payload) > 0 {
		msg.Flags = ExtendedDirectMessage
		if len(msg.Payload) < 14 {
			tmp := make([]byte, 14)
			copy(tmp, msg.Payload)
			msg.Payload = tmp
		}
	}
	retries := 0
	for err == nil && retries <= conn.retries {
		err = conn.WriteMessage(msg)

		if err == nil {
			Log.Debugf("TX %s", msg)
			// wait for ack
			ack, err = conn.Receive()
			if err == nil {
				if ack.Nak() {
					err = ErrNak
				}
				break
			} else if err == ErrReadTimeout {
				err = nil
				retries++
				if retries > conn.retries {
					err = ErrAckTimeout
					break
				}
			}
		}
	}
	return ack, err
}

func (conn *connection) Receive() (*Message, error) {
	select {
	case msg := <-conn.recvCh:
		return msg, nil
	case <-time.After(conn.timeout):
	}
	return nil, ErrReadTimeout
}

func (conn *connection) dispatch(msg *Message) {
	if len(conn.match) == 0 {
		// match everything
		conn.recvCh <- msg
	} else {
		for _, cmd := range conn.match {
			if cmd[1] == msg.Command[1] {
				conn.recvCh <- msg
				break
			}
		}
	}
}

func IDRequest(dialer Dialer, dst Address) (version FirmwareVersion, devCat DevCat, err error) {
	conn, err := dialer.Dial(dst, CmdIDRequest, CmdSetButtonPressedResponder, CmdSetButtonPressedController)
	defer conn.Close()
	if err == nil {
		_, err := conn.Send(&Message{Command: CmdIDRequest})
		if err == nil {
			var msg *Message
			msg, err = conn.Receive()
			if err == nil {
				version = FirmwareVersion(msg.Dst[2])
				devCat = DevCat{msg.Dst[0], msg.Dst[1]}
			}
		}
	}
	return
}

func GetEngineVersion(dialer Dialer, dst Address) (version EngineVersion, err error) {
	conn, err := dialer.Dial(dst, CmdGetEngineVersion)
	defer conn.Close()
	if err == nil {
		var ack *Message
		ack, err = conn.Send(&Message{Command: CmdGetEngineVersion})
		if err == nil {
			Log.Debugf("Device %v responded with an engine version %d", conn, ack.Command[2])
			version = EngineVersion(ack.Command[2])
		} else if err == ErrNak {
			// This only happens if the device is an I2Cs device and
			// is not linked to the queryier
			if ack.Command[2] == 0xff {
				Log.Debugf("Device %v is an unlinked I2Cs device", conn)
				version = VerI2Cs
				err = ErrNotLinked
			} else {
				err = ErrNak
			}
		}
	}
	return
}
