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

package plm

import "github.com/abates/insteon"

type connection struct {
	sendCmd        Command
	matches        []Command
	sendCh         chan *insteon.PacketRequest
	upstreamSendCh chan<- *CommandRequest
	recvCh         chan []byte
	upstreamRecvCh <-chan *Packet
}

func newConnection(sendCh chan<- *CommandRequest, recvCh <-chan *Packet, sendCmd Command, recvCmds ...Command) *connection {
	conn := &connection{
		sendCmd:        sendCmd,
		matches:        recvCmds,
		sendCh:         make(chan *insteon.PacketRequest, 1),
		upstreamSendCh: sendCh,
		recvCh:         make(chan []byte, 1),
		upstreamRecvCh: recvCh,
	}

	go conn.process()

	return conn
}

func (conn *connection) process() {
	for {
		select {
		case request, open := <-conn.sendCh:
			if !open {
				close(conn.upstreamSendCh)
				close(conn.recvCh)
				return
			}
			conn.send(request)
		case packet, open := <-conn.upstreamRecvCh:
			if !open {
				close(conn.upstreamSendCh)
				close(conn.recvCh)
				return
			}
			conn.receive(packet)
		}
	}
}

func (conn *connection) send(request *insteon.PacketRequest) {
	doneCh := make(chan *CommandRequest)
	payload := request.Payload
	// PLM expects that the payload begins with the
	// destinations address so we have to slice off
	// the src address
	if conn.sendCmd == CmdSendInsteonMsg && len(payload) > 3 {
		payload = payload[3:]
	}

	conn.upstreamSendCh <- &CommandRequest{Command: conn.sendCmd, Payload: payload, DoneCh: doneCh}
	upstreamRequest := <-doneCh
	// ignore read timeout for testing
	if upstreamRequest.Err != ErrReadTimeout {
		request.Err = upstreamRequest.Err
	}
	request.DoneCh <- request
}

func (conn *connection) receive(packet *Packet) {
	if len(conn.matches) > 0 {
		for _, match := range conn.matches {
			if match == packet.Command {
				conn.recvCh <- packet.Payload
				return
			}
		}
	} else {
		conn.recvCh <- packet.Payload
	}
}
