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
	"time"
)

type MessageSender interface {
	Send(*Message) error
}

type Connection interface {
	Address() Address
	Send(*Message) (ack *Message, err error)
	Receive() (*Message, error)
	IDRequest() (FirmwareVersion, DevCat, error)
	EngineVersion() (EngineVersion, error)
}

type connection struct {
	addr    Address
	match   []Command
	timeout time.Duration

	txCh    chan<- *Message
	rxCh    <-chan *Message
	msgCh   chan *Message
	closeCh chan chan error
}

func NewConnection(txCh chan<- *Message, rxCh <-chan *Message, addr Address, timeout time.Duration, match ...Command) Connection {
	conn := &connection{
		addr:    addr,
		match:   match,
		timeout: timeout,

		txCh:    txCh,
		rxCh:    rxCh,
		msgCh:   make(chan *Message, 10),
		closeCh: make(chan chan error),
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
							conn.msgCh <- msg
						}
					}
				} else {
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
	select {
	case msg = <-conn.msgCh:
	case <-time.After(conn.timeout):
		err = ErrReadTimeout
	}
	return
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
			if ack.Command[2] == 0xff {
				version = VerI2Cs
			} else {
				err = ErrNak
			}
		} else {
			version = EngineVersion(ack.Command[2])
		}
	}
	return
}
