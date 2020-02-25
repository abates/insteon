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
	"time"
)

// Sender will deliver an Insteon message to the network
type Sender interface {
	// Send will send a message to the device
	Send(*Message) error
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
type Demux struct {
	// Sender is the upstream device (PLM, for instance)
	Sender
	connections map[Address][]Connection
}

// Dispatch a message to the proper listening connection
func (d *Demux) Dispatch(msg *Message) {
	for _, addr := range []Address{msg.Src, Wildcard} {
		if connections, found := d.connections[addr]; found {
			for _, conn := range connections {
				conn.Dispatch(msg)
			}
		}
	}
}

// New creates a new connection that will receive messages dispatched by
// this demux.  The returned connection will use the upstream sender
// directly.
func (d *Demux) New(addr Address, options ...ConnectionOption) (Connection, error) {
	var err error
	conn, err := newConnection(d, addr, options...)
	if err == nil {
		if d.connections == nil {
			d.connections = make(map[Address][]Connection)
		}
		d.connections[addr] = append(d.connections[addr], conn)
	}
	return conn, err
}

// Connection is a very basic communication mechanism to send
// and receive messages with individual Insteon devices.
type Connection interface {
	// Address returns the unique Insteon address of the device
	Address() Address

	// Dispatch will deliver a message to this connection.  This should
	// only be called by upstream interfaces, such as Demux or PLM
	Dispatch(*Message)

	// Send will send the message to the device and wait for the
	// device to respond with an Ack/Nak.  Send will always return
	// but may return with a read timeout or other communication error
	Send(*Message) (ack *Message, err error)

	// Receive waits for the next message from the device.  Receive
	// always returns, but may return with an error (such as ErrReadTimeout)
	//Receive() (*Message, error)

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

	// AddHandler adds an event handler that is called when a message containing the
	// first byte of the given command is received.  This is useful to setup
	// event handlers to update internal state, which may arrive in asynchronous
	// messages
	AddHandler(chan<- *Message, ...Command)

	// RemoveHandler will remove a previously set message handler
	RemoveHandler(chan<- *Message, ...Command)
}

type connection struct {
	addr Address
	//match    []Command
	timeout  time.Duration
	ttl      uint8
	upstream Sender
	handlers map[byte]map[chan<- *Message]interface{}
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

// NewConnection will return a connection that is setup and ready to be used.  The txCh and
// rxCh will be used to send and receive insteon messages.  Any supplied options will be
// used to customize the connection's config
func NewConnection(upstream Sender, addr Address, options ...ConnectionOption) (Connection, error) {
	return newConnection(upstream, addr, options...)
}

func newConnection(upstream Sender, addr Address, options ...ConnectionOption) (*connection, error) {
	conn := &connection{
		addr:    addr,
		timeout: 3 * time.Second,

		handlers: make(map[byte]map[chan<- *Message]interface{}),
		upstream: upstream,
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

func (conn *connection) Dispatch(msg *Message) {
	handlers, found := conn.handlers[msg.Command[1]]
	// look for wildcard handlers if any exist
	if !found {
		handlers, found = conn.handlers[0xff]
	}

	if found {
		for handler := range handlers {
			handler <- msg
		}
	}
}

func (conn *connection) AddHandler(ch chan<- *Message, cmds ...Command) {
	for _, cmd := range cmds {
		handlers, found := conn.handlers[cmd[1]]
		if !found {
			handlers = make(map[chan<- *Message]interface{})
			conn.handlers[cmd[1]] = handlers
		}
		handlers[ch] = struct{}{}
	}
}

func (conn *connection) RemoveHandler(ch chan<- *Message, cmds ...Command) {
	for _, cmd := range cmds {
		if handlers, found := conn.handlers[cmd[1]]; found {
			delete(handlers, ch)
		}
	}
}

func (conn *connection) Send(msg *Message) (ack *Message, err error) {
	return conn.send(msg)
}

func (conn *connection) addHandler(buflen int, cmds ...Command) chan *Message {
	ch := make(chan *Message, buflen)
	conn.AddHandler(ch, cmds...)
	return ch
}

func readFromCh(ch <-chan *Message, timeout time.Duration) (*Message, error) {
	select {
	case msg := <-ch:
		return msg, nil
	case <-time.After(timeout):
	}
	return nil, ErrReadTimeout
}

func (conn *connection) send(msg *Message) (ack *Message, err error) {
	ch := conn.addHandler(1, msg.Command)
	defer conn.RemoveHandler(ch, msg.Command)

	msg.Dst = conn.addr
	msg.Flags = Flag(MsgTypeDirect, len(msg.Payload) > 0, conn.ttl, conn.ttl)
	Log.Tracef("Connection %v TX %v", conn.addr, msg)
	err = conn.upstream.Send(msg)

	if err == nil {
		ack, err = readFromCh(ch, conn.timeout)
		if err == ErrReadTimeout {
			err = ErrAckTimeout
		}
	}
	return ack, err
}

func (conn *connection) IDRequest() (version FirmwareVersion, devCat DevCat, err error) {
	ch := conn.addHandler(1, CmdSetButtonPressedResponder, CmdSetButtonPressedController)
	defer conn.RemoveHandler(ch, CmdSetButtonPressedResponder, CmdSetButtonPressedController)
	_, err = conn.send(&Message{Command: CmdIDRequest, Flags: StandardDirectMessage})
	msg, err := readFromCh(ch, conn.timeout)
	if err == nil {
		version = FirmwareVersion(msg.Dst[2])
		devCat = DevCat{msg.Dst[0], msg.Dst[1]}
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
				Log.Debugf("Device %v is an unlinked I2Cs device", conn.Address())
				version = VerI2Cs
				err = ErrNotLinked
			} else {
				err = ErrNak
			}
		} else {
			Log.Debugf("Device %v responded with an engine version %d", conn.Address(), ack.Command[2])
			version = EngineVersion(ack.Command[2])
		}
	}
	return
}
