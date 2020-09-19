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
			cmd := Command(int(test.input[0])<<8 | int(test.input[1]))
			got := byte(cmd.Command1()) + byte(cmd.Command2())
			want := test.input[0] + test.input[1]
			if want != got {
				t.Errorf("Wanted byte 0x%02x got 0x%02x", want, got)
			}

			got = checksum(cmd, test.input[2:])
			if got != test.expected {
				t.Errorf("got checksum %02x, want %02x", got, test.expected)
			}
		})
	}
}

func TestI2CsErrLookup(t *testing.T) {
	tests := []struct {
		desc     string
		input    *Message
		inputErr error
		want     error
	}{
		{"nil error", &Message{}, nil, nil},
		{"ErrIllegalValue", &Message{Command: Command(0x0000fb), Flags: StandardDirectNak}, ErrNak, ErrIllegalValue},
		{"ErrPreNak", &Message{Command: Command(0x0000fc), Flags: StandardDirectNak}, ErrNak, ErrPreNak},
		{"ErrIncorrectChecksum", &Message{Command: Command(0x0000fd), Flags: StandardDirectNak}, ErrNak, ErrIncorrectChecksum},
		{"ErrNoLoadDetected", &Message{Command: Command(0x0000fe), Flags: StandardDirectNak}, ErrNak, ErrNoLoadDetected},
		{"ErrNotLinked", &Message{Command: Command(0x0000ff), Flags: StandardDirectNak}, ErrNak, ErrNotLinked},
		{"ErrUnexpectedResponse", &Message{Command: Command(0x0000fa), Flags: StandardDirectNak}, ErrNak, ErrUnexpectedResponse},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			_, got := i2csErrLookup(test.input, test.inputErr)
			if !IsError(got, test.want) {
				t.Errorf("want %v got %v", test.want, got)
			}
		})
	}
}

func TestI2CsDeviceSendCommand(t *testing.T) {
	tests := []struct {
		desc       string
		sndCmd     Command
		sndPayload []byte
		wantCmd    Command
	}{
		{"SD", Command((int(StandardDirectMessage) & 0xff << 16) | 0x0102), nil, Command((int(StandardDirectMessage) & 0xff << 16) | 0x0102)},
		{"ED", Command((int(ExtendedDirectMessage) & 0xff << 16) | 0x0203), []byte{1, 2, 3, 4}, Command((int(ExtendedDirectMessage) & 0xff << 16) | 0x0203)},
		{"Enter Linking Mode", CmdEnterLinkingMode.SubCommand(42), nil, CmdEnterLinkingModeExt.SubCommand(42)},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ackFlags := StandardDirectAck
			if len(test.sndPayload) > 0 {
				ackFlags = ExtendedDirectAck
			}
			b := &testBus{publishResp: []*Message{{Flags: ackFlags}}}
			device := newI2CsDevice(b, DeviceInfo{})
			device.SendCommand(test.sndCmd, test.sndPayload)

			gotMsg := b.published

			if test.sndCmd.Command0() == int(ExtendedDirectMessage) && gotMsg.Payload[len(gotMsg.Payload)-1] == 0 {
				t.Errorf("Expected checksum to be set")
			}

			gotCmd := gotMsg.Command
			if test.wantCmd != gotCmd {
				t.Errorf("Wanted command %v got %v", test.wantCmd, gotCmd)
			}
		})
	}
}
