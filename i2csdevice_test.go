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

func TestChecksum(t *testing.T) {
	tests := []struct {
		desc     string
		input    []byte
		expected byte
	}{
		{"1", []byte{0x2E, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xd1},
		{"2", []byte{0x2F, 0x00, 0x00, 0x00, 0x0F, 0xFF, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xC2},
		{"3", []byte{0x2F, 0x00, 0x01, 0x01, 0x0F, 0xFF, 0x00, 0xA2, 0x00, 0x19, 0x70, 0x1A, 0xFF, 0x1F, 0x01}, 0x5D},
		{"4", []byte{0x2F, 0x00, 0x00, 0x00, 0x0F, 0xF7, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xCA},
		{"5", []byte{0x2F, 0x00, 0x01, 0x01, 0x0F, 0xF7, 0x00, 0xE2, 0x01, 0x19, 0x70, 0x1A, 0xFF, 0x1F, 0x01}, 0x24},
		{"6", []byte{0x2F, 0x00, 0x00, 0x00, 0x0F, 0xEF, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xD2},
		{"7", []byte{0x2F, 0x00, 0x01, 0x01, 0x0F, 0xEF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xD1},
		{"8", []byte{0x2F, 0x00, 0x01, 0x02, 0x0F, 0xFF, 0x08, 0xE2, 0x01, 0x08, 0xB6, 0xEA, 0x00, 0x1B, 0x01}, 0x11},
		{"9", []byte{0x09, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xF6},
		{"A", []byte{0x2f, 0x00, 0x00, 0x02, 0x0f, 0xff, 0x08, 0xe2, 0x01, 0x08, 0xb6, 0xea, 0x00, 0x1b, 0x01}, 0x12},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := checksum(Command{0x00, test.input[0], test.input[1]}, test.input[2:])
			if got != test.expected {
				t.Errorf("got checksum %02x, want %02x", got, test.expected)
			}
		})
	}
}

func TestI2CsErrLookup(t *testing.T) {
	tests := []struct {
		desc  string
		input *Message
		err   error
		want  error
	}{
		{"nil error", &Message{}, nil, nil},
		{"ErrIllegalValue", &Message{Command: Command{0, 0, 0xfb}, Flags: StandardDirectNak}, nil, ErrIllegalValue},
		{"ErrPreNak", &Message{Command: Command{0, 0, 0xfc}, Flags: StandardDirectNak}, nil, ErrPreNak},
		{"ErrIncorrectChecksum", &Message{Command: Command{0, 0, 0xfd}, Flags: StandardDirectNak}, nil, ErrIncorrectChecksum},
		{"ErrNoLoadDetected", &Message{Command: Command{0, 0, 0xfe}, Flags: StandardDirectNak}, nil, ErrNoLoadDetected},
		{"ErrNotLinked", &Message{Command: Command{0, 0, 0xff}, Flags: StandardDirectNak}, nil, ErrNotLinked},
		{"ErrUnexpectedResponse", &Message{Command: Command{0, 0, 0xfa}, Flags: StandardDirectNak}, nil, ErrUnexpectedResponse},
		{"ErrReadTimeout", &Message{Command: Command{0, 0, 0xfa}, Flags: StandardDirectNak}, ErrReadTimeout, ErrReadTimeout},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			_, got := i2csErrLookup(test.input, test.err)
			if !isError(got, test.want) {
				t.Errorf("want %v got %v", test.want, got)
			}
		})
	}
}

func TestI2CsDeviceSendCommand(t *testing.T) {
	tests := []struct {
		desc    string
		command Command
		payload []byte
		flags   Flags
	}{
		{"SD", Command{byte(StandardDirectMessage), 1, 2}, nil, StandardDirectMessage},
		{"ED", Command{byte(ExtendedDirectMessage), 2, 3}, []byte{1, 2, 3, 4}, ExtendedDirectMessage},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			conn := &testConnection{sendCh: make(chan *Message, 1), ackCh: make(chan *Message, 1)}
			device := NewI2CsDevice(conn, time.Millisecond)
			ackFlags := StandardDirectAck
			if len(test.payload) > 0 {
				ackFlags = ExtendedDirectAck
			}
			conn.ackCh <- &Message{Flags: ackFlags}

			device.SendCommand(test.command, test.payload)

			msg := <-conn.sendCh

			if test.command[0] == byte(ExtendedDirectMessage) && msg.Payload[len(msg.Payload)-1] == 0 {
				t.Errorf("Expected checksum to be set")
			}
		})
	}
}

func TestI2CsDeviceCommands(t *testing.T) {
	tests := []*commandTest{
		{"EnterLinkingMode", func(d Device) error { return d.(*I2CsDevice).EnterLinkingMode(15) }, CmdEnterLinkingModeExt.SubCommand(15), nil, nil},
	}

	testDeviceCommands(t, func(conn *testConnection) Device { return NewI2CsDevice(conn, time.Millisecond) }, tests)
}

func TestI2CsDeviceReceive(t *testing.T) {
	tests := []struct {
		desc    string
		input   *Message
		wantErr error
	}{
		{"ErrIllegalValue", &Message{Command: Command{0, 0, 0xfb}, Flags: StandardDirectNak}, ErrIllegalValue},
		{"ErrPreNak", &Message{Command: Command{0, 0, 0xfc}, Flags: StandardDirectNak}, ErrPreNak},
		{"ErrIncorrectChecksum", &Message{Command: Command{0, 0, 0xfd}, Flags: StandardDirectNak}, ErrIncorrectChecksum},
		{"ErrNoLoadDetected", &Message{Command: Command{0, 0, 0xfe}, Flags: StandardDirectNak}, ErrNoLoadDetected},
		{"ErrNotLinked", &Message{Command: Command{0, 0, 0xff}, Flags: StandardDirectNak}, ErrNotLinked},
		{"ErrUnexpectedResponse", &Message{Command: Command{0, 0, 0xfa}, Flags: StandardDirectNak}, ErrUnexpectedResponse},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			conn := &testConnection{recvCh: make(chan *Message, 1)}
			conn.recvCh <- test.input
			device := NewI2CsDevice(conn, time.Millisecond)
			_, err := device.Receive()
			if !isError(err, test.wantErr) {
				t.Errorf("want error %v got %v", test.wantErr, err)
			}
		})
	}
}
