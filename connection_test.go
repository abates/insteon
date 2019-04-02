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
	"io"
	"testing"
	"time"
)

type testConnection struct {
	addr            Address
	devCat          DevCat
	firmwareVersion FirmwareVersion

	sendCh  chan *Message
	ackCh   chan *Message
	sendErr error

	recvCh  chan *Message
	recvErr error
}

func (tc *testConnection) Address() Address { return tc.addr }
func (tc *testConnection) IDRequest() (FirmwareVersion, DevCat, error) {
	return tc.firmwareVersion, tc.devCat, nil
}

func (tc *testConnection) SendCommand(cmd Command, payload []byte) (Command, error) {
	msg, err := tc.Send(&Message{Command: cmd, Payload: payload})
	return msg.Command, err
}

func (tc *testConnection) Send(msg *Message) (*Message, error) {
	tc.sendCh <- msg
	if tc.sendErr != nil {
		return nil, tc.sendErr
	}
	return <-tc.ackCh, nil
}

func (tc *testConnection) Receive() (*Message, error) {
	if tc.recvErr == nil {
		return <-tc.recvCh, nil
	}
	return nil, tc.recvErr
}

func TestConnectionSend(t *testing.T) {
	tests := []struct {
		name        string
		input       *Message
		expectedErr error
	}{
		{"I1 Send", TestProductDataResponse, nil},
		{"I2 Send", TestProductDataResponse, nil},
		{"I2Cs Send", TestProductDataResponse, nil},
		{"I2Cs Send", TestProductDataResponse, ErrReadTimeout},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			txCh := make(chan *Message, 1)
			rxCh := make(chan *Message, 1)
			conn := NewConnection(txCh, rxCh, Address{}, time.Millisecond)
			go func() {
				<-txCh
				if test.expectedErr == nil {
					ack := *test.input
					src := ack.Src
					ack.Src = ack.Dst
					ack.Dst = src
					ack.Flags = StandardDirectAck
					if test.input.Flags.Extended() {
						ack.Flags = ExtendedDirectAck
					}
					rxCh <- &ack
				}
			}()

			_, err := conn.Send(test.input)
			if err != test.expectedErr {
				t.Errorf("Want %v got %v", test.expectedErr, err)
			}
			if closer, ok := conn.(io.Closer); ok {
				closer.Close()
			}
		})
	}
}

func TestConnectionReceive(t *testing.T) {
	tests := []struct {
		name        string
		input       *Message
		match       Command
		expectedErr error
	}{
		{"ReadTimeout 1", &Message{Command: Command{0x00, 0x00, 0x00}}, Command{0x00, 0x01, 0x01}, ErrReadTimeout},
		{"ReadTimeout 2", &Message{Command: Command{0x00, 0x01, 0xff}}, Command{0x00, 0x01, 0x01}, ErrReadTimeout},
		{"Match 1", &Message{Command: Command{0x00, 0x01, 0x01}}, Command{0x00, 0x01, 0x01}, nil},
		{"Match 2", &Message{Command: Command{0x00, 0x01, 0x01}}, Command{0x00, 0x01, 0x00}, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			txCh := make(chan *Message, 1)
			rxCh := make(chan *Message, 1)
			rxCh <- test.input
			conn := NewConnection(txCh, rxCh, Address{}, time.Millisecond, test.match)
			_, err := conn.Receive()

			if test.expectedErr != err {
				t.Errorf("want %v got %v", test.expectedErr, err)
			}
			if closer, ok := conn.(io.Closer); ok {
				closer.Close()
			}
		})
	}
}
