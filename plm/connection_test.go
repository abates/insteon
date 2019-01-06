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

import (
	"testing"
	"time"

	"github.com/abates/insteon"
)

func testClosedChannels(sendCh chan *CommandRequest, conn *connection) (string, bool) {
	select {
	case _, open := <-sendCh:
		if open {
			return "Expected upstreamSendCh to be closed", false
		}
	case <-time.After(time.Second):
		return "Timed out waiting for upstreamSendCh to be closed", false
	}

	select {
	case _, open := <-conn.recvCh:
		if open {
			return "Expected recvCh to be closed", false
		}
	case <-time.After(time.Second):
		return "Timed out waiting for recvCh to be closed", false
	}
	return "", true
}

func TestConnectionSend(t *testing.T) {
	sendCh := make(chan *CommandRequest, 1)
	conn := newConnection(sendCh, nil, CmdSendInsteonMsg)

	doneCh := make(chan *insteon.PacketRequest, 1)
	conn.sendCh <- &insteon.PacketRequest{DoneCh: doneCh}
	request := <-sendCh
	if request.Command != CmdSendInsteonMsg {
		t.Errorf("sent %v, want %v", request.Command, CmdSendInsteonMsg)
		request.DoneCh <- request
	} else {
		request.Err = ErrReadTimeout
		request.DoneCh <- request
		packetRequest := <-doneCh
		if packetRequest.Err != ErrReadTimeout {
			t.Errorf("got error %v, want %v", packetRequest.Err, ErrReadTimeout)
		}
	}

	close(conn.sendCh)
	if msg, passed := testClosedChannels(sendCh, conn); !passed {
		t.Errorf("%v", msg)
	}
}

func TestConnectionReceive(t *testing.T) {
	tests := []struct {
		desc     string
		input    *Packet
		match    []Command
		expected bool
	}{
		{"Received-Received", &Packet{Command: CmdStdMsgReceived}, []Command{CmdStdMsgReceived}, true},
		{"Nak-Received", &Packet{Command: CmdNak}, []Command{CmdStdMsgReceived}, false},
		{"Nak", &Packet{Command: CmdNak}, nil, true},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			sendCh := make(chan *CommandRequest, 1)
			recvCh := make(chan *Packet, 1)
			conn := &connection{
				upstreamRecvCh: recvCh,
				matches:        test.match,
				upstreamSendCh: sendCh,
				recvCh:         make(chan []byte, 1),
			}

			recvCh <- test.input
			close(recvCh)
			conn.process()

			if test.expected && len(conn.recvCh) == 0 {
				t.Errorf("expected packet to be delivered")
			} else {
				<-conn.recvCh
			}

			if msg, passed := testClosedChannels(sendCh, conn); !passed {
				t.Errorf("%v", msg)
			}
		})
	}
}
