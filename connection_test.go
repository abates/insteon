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
	"testing"
	"time"
)

func newTestConnection(dst Address) (*connection, chan *MessageRequest, chan *Message) {
	sendCh := make(chan *MessageRequest, 10)
	recvCh := make(chan *Message, 10)
	return newConnection(sendCh, recvCh, dst, 1, time.Millisecond), sendCh, recvCh
}

// TODO need to rewrite this test because it sucks and is full
// of race conditions
func TestConnectionProcess(t *testing.T) {
	t.Parallel()
	doneCh := make(chan *MessageRequest, 1)
	recvCh := make(chan *Message, 1)
	upstreamRecvCh := make(chan *Message, 1)
	upstreamSendCh := make(chan *MessageRequest, 1)

	conn := &connection{
		recvCh:         recvCh,
		upstreamRecvCh: upstreamRecvCh,
		upstreamSendCh: upstreamSendCh,
		queue:          []*MessageRequest{{DoneCh: doneCh}},
		timeout:        time.Millisecond,
	}

	go func() {
		request := <-doneCh
		if request.Err != ErrReadTimeout {
			t.Errorf("Expected %v got %v", ErrReadTimeout, request.Err)
		}
		close(upstreamRecvCh)
	}()

	conn.process()

	if len(conn.queue) > 0 {
		t.Error("Expected empty queue")
	}
}

func TestConnectionReceiveAck(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc        string
		version     EngineVersion
		returnedAck *Message
		expectedErr error
	}{
		{"I1 MessageUnknownCommandNak", VerI1, TestMessageUnknownCommandNak, ErrUnknownCommand},
		{"I1 MessageNoLoadDetected", VerI1, TestMessageNoLoadDetected, ErrNoLoadDetected},
		{"I1 MessageNotLinked", VerI1, TestMessageNotLinked, ErrNotLinked},
		{"I1 UnexpectedResponse", VerI1, &Message{Src: testDstAddr, Flags: StandardDirectNak, Command: Command{0x00, 0x00, 0x01}}, ErrUnexpectedResponse},
		{"I2Cs MessageIllegalValue", VerI2Cs, TestMessageIllegalValue, ErrIllegalValue},
		{"I2Cs MessagePreNak", VerI2Cs, TestMessagePreNak, ErrPreNak},
		{"I2Cs MessageIncorrectChecksum", VerI2Cs, TestMessageIncorrectChecksum, ErrIncorrectChecksum},
		{"I2Cs MessageNoLoadDetectedI2Cs", VerI2Cs, TestMessageNoLoadDetectedI2Cs, ErrNoLoadDetected},
		{"I2Cs MessageNotLinkedI2Cs", VerI2Cs, TestMessageNotLinkedI2Cs, ErrNotLinked},
		{"I2Cs UnexpectedResponse", VerI2Cs, &Message{Src: testDstAddr, Flags: StandardDirectNak, Command: Command{0x00, 0x00, 0x01}}, ErrUnexpectedResponse},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			conn := &connection{
				addr:    testDstAddr,
				version: test.version,
			}
			doneCh := make(chan *MessageRequest, 1)
			request := &MessageRequest{Message: &Message{Command: Command{test.returnedAck.Command[0], test.returnedAck.Command[1], 0x00}}, DoneCh: doneCh}
			conn.queue = append(conn.queue, request)
			conn.receive(test.returnedAck)

			if !isError(request.Err, test.expectedErr) {
				t.Errorf("expected %v got %v", test.expectedErr, request.Err)
			} else if request.Ack != test.returnedAck {
				t.Errorf("expected %v got %v", test.returnedAck, request.Ack)
			}
		})
	}
}

func TestConnectionReceiveMatch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    *Message
		match    Command
		expected int
	}{
		{&Message{Src: testDstAddr, Command: Command{0x00, 0x00, 0x00}}, Command{0x00, 0x01, 0x01}, 0},
		{&Message{Src: testDstAddr, Command: Command{0x00, 0x01, 0xff}}, Command{0x00, 0x01, 0x01}, 0},
		{&Message{Src: testDstAddr, Command: Command{0x00, 0x01, 0x01}}, Command{0x00, 0x01, 0x01}, 1},
		{&Message{Src: testDstAddr, Command: Command{0x00, 0x01, 0x01}}, Command{0x00, 0x01, 0x00}, 1},
	}

	for i, test := range tests {
		conn := &connection{
			addr:   testDstAddr,
			match:  []Command{test.match},
			recvCh: make(chan *Message, 1),
		}

		conn.receive(test.input)

		if test.expected != len(conn.recvCh) {
			t.Errorf("tests[%d] Expected %d packets to be received. Got %d", i, test.expected, len(conn.recvCh))
		}
	}
}

func TestConnectionReceive(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    *Message
		expected int
	}{
		{&Message{Src: testSrcAddr}, 0},
		{&Message{Src: testDstAddr}, 1},
	}

	for i, test := range tests {
		conn := &connection{
			addr:   testDstAddr,
			recvCh: make(chan *Message, 1),
		}
		conn.receive(test.input)

		if test.expected != len(conn.recvCh) {
			t.Errorf("tests[%d] Expected %d packets to be received. Got %d", i, test.expected, len(conn.recvCh))
		}
	}
}

func TestConnectionSend(t *testing.T) {
	t.Parallel()
	upstreamSendCh := make(chan *MessageRequest, 1)
	conn := &connection{
		addr:           testDstAddr,
		upstreamSendCh: upstreamSendCh,
	}

	doneCh := make(chan *MessageRequest, 1)
	request := &MessageRequest{Message: &Message{}, DoneCh: doneCh}
	conn.queue = append(conn.queue, request)
	go func() {
		request := <-upstreamSendCh
		request.Err = ErrReadTimeout
		request.DoneCh <- request
	}()

	conn.send()

	<-doneCh
	if request.Message.Dst != testDstAddr {
		t.Errorf("Expected destination to be %v got %v", testSrcAddr, request.Message.Dst)
	}

	if request.Err != ErrReadTimeout {
		t.Errorf("Expected %v got %v", ErrReadTimeout, request.Err)
	}
}
