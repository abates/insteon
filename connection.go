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

type MessageSender interface {
	Send(*Message) error
}

type Connection interface {
	Send(*Message) (ack *Message, err error)
	Receive() (*Message, error)
}

type connection struct {
	sync.Mutex
	addr    Address
	match   []Command
	timeout time.Duration
	sender  MessageSender

	rxCh    <-chan *Message
	msgCh   chan *Message
	closeCh chan chan error
}

func NewConnection(sender MessageSender, rxCh <-chan *Message, addr Address, timeout time.Duration, match ...Command) Connection {
	conn := &connection{
		addr:    addr,
		match:   match,
		timeout: timeout,
		sender:  sender,

		rxCh:    rxCh,
		msgCh:   make(chan *Message, 10),
		closeCh: make(chan chan error),
	}

	go conn.readLoop()
	return conn
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
	conn.Lock()
	defer conn.Unlock()

	msg.Dst = conn.addr
	err = conn.sender.Send(msg)

	// wait for ack
	timeout := time.Now().Add(conn.timeout)
	for err == nil {
		ack, err = conn.receive()
		if err == nil && (ack.Ack() || ack.Nak()) {
			break
		} else if timeout.Before(time.Now()) {
			err = ErrReadTimeout
		}
	}

	return ack, err
}

func (conn *connection) Receive() (msg *Message, err error) {
	conn.Lock()
	defer conn.Unlock()
	return conn.receive()
}

func (conn *connection) receive() (msg *Message, err error) {
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
