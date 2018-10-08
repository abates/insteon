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

type connection struct {
	addr    Address
	match   []Command
	version EngineVersion
	timeout time.Duration

	sendCh         chan *MessageRequest
	upstreamSendCh chan<- *MessageRequest
	recvCh         chan *Message
	upstreamRecvCh <-chan *Message

	queue []*MessageRequest
}

func newConnection(upstreamSendCh chan<- *MessageRequest, upstreamRecvCh <-chan *Message, addr Address, version EngineVersion, timeout time.Duration, match ...Command) *connection {
	conn := &connection{
		addr:    addr,
		match:   match,
		version: version,
		timeout: timeout,

		sendCh:         make(chan *MessageRequest, 1),
		upstreamSendCh: upstreamSendCh,
		recvCh:         make(chan *Message, 1),
		upstreamRecvCh: upstreamRecvCh,
	}

	go conn.process()
	return conn
}

func (conn *connection) process() {
	for {
		select {
		case msg, open := <-conn.upstreamRecvCh:
			if !open {
				close(conn.recvCh)
				close(conn.upstreamSendCh)
				return
			}
			conn.receive(msg)
		case request, open := <-conn.sendCh:
			if !open {
				close(conn.recvCh)
				close(conn.upstreamSendCh)
				return
			}
			conn.queue = append(conn.queue, request)
			if len(conn.queue) == 1 {
				conn.send()
			}
		case <-time.After(conn.timeout):
			// prevent head of line blocking for a lost/nonexistant Ack
			if len(conn.queue) > 0 && conn.queue[0].timeout.Before(time.Now()) {
				conn.queue[0].Err = ErrReadTimeout
				conn.queue[0].DoneCh <- conn.queue[0]
				conn.queue = conn.queue[1:]
				conn.send()
			}
		}
	}
}

func errLookup(command Command) (err error) {
	switch command[2] & 0xff {
	case 0xfd:
		err = ErrUnknownCommand
	case 0xfe:
		err = ErrNoLoadDetected
	case 0xff:
		err = ErrNotLinked
	default:
		err = NewTraceError(ErrUnexpectedResponse)
	}
	return
}

func i2csErrLookup(command Command) (err error) {
	switch command[2] & 0xff {
	case 0xfb:
		err = ErrIllegalValue
	case 0xfc:
		err = ErrPreNak
	case 0xfd:
		err = ErrIncorrectChecksum
	case 0xfe:
		err = ErrNoLoadDetected
	case 0xff:
		err = ErrNotLinked
	default:
		err = NewTraceError(ErrUnexpectedResponse)
	}
	return
}

func (conn *connection) receiveAck(msg *Message) {
	if len(conn.queue) > 0 {
		request := conn.queue[0]
		if msg.Command[1]&0xff == request.Message.Command[1]&0xff {
			conn.queue[0].Ack = msg
			if msg.Flags.Type() == MsgTypeDirectNak {
				if VerI1 <= conn.version && conn.version <= VerI2 {
					request.Err = errLookup(msg.Command)
				} else if conn.version == VerI2Cs {
					request.Err = i2csErrLookup(msg.Command)
				}
			}

			conn.queue[0].DoneCh <- conn.queue[0]

			conn.queue = conn.queue[1:]
			conn.send()
		}
	}
}

func (conn *connection) receiveMatch(msg *Message) {
	for _, m := range conn.match {
		if (msg.Command == m) || (msg.Command[1] == m[1] && m[2] == 0x00) {
			conn.recvCh <- msg
		}
	}
}

func (conn *connection) receive(msg *Message) {
	if msg.Src == conn.addr {
		if msg.Ack() || msg.Nak() {
			conn.receiveAck(msg)
		} else if len(conn.match) > 0 {
			conn.receiveMatch(msg)
		} else {
			conn.recvCh <- msg
		}
	}
}

func (conn *connection) send() {
	if len(conn.queue) > 0 {
		request := conn.queue[0]
		request.timeout = time.Now().Add(conn.timeout)
		request.Message.Dst = conn.addr

		oldCh := request.DoneCh
		doneCh := make(chan *MessageRequest, 1)
		request.DoneCh = doneCh
		conn.upstreamSendCh <- request
		<-doneCh
		request.DoneCh = oldCh

		if request.Err != nil {
			conn.queue = conn.queue[1:]
			request.DoneCh <- request
		}
	}
}

/*func (conn *connection) SendMessage(msg *Message) (*Message, error) {
	doneCh := make(chan bool, 1)
	request := &MessageRequest{Message: msg, DoneCh: doneCh}
	conn.sendCh <- request
	<-doneCh
	return request.Ack, request.Err
}

func (conn *connection) ReceiveMessage() (msg *Message, err error) {
	select {
	case msg = <-conn.recvCh:
	case <-time.After(conn.timeout):
		err = ErrReadTimeout
	}
	return
}

func (conn *connection) Close() error {
	close(conn.sendCh)
	return nil
}*/
