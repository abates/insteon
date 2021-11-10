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
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestPacketAckNak(t *testing.T) {
	tests := []struct {
		cmd   Command
		input byte
		ack   bool
		nak   bool
	}{
		{0x60, 0x06, true, false},
		{0x60, 0x15, false, true},
		{0x01, 0x06, false, false},
		{0x01, 0x15, false, false},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("0x%02x 0x%02x", test.cmd, test.input), func(t *testing.T) {
			p := &Packet{Command: test.cmd, Ack: test.input}
			if p.ACK() != test.ack {
				t.Errorf("got ack %v, want %v ", p.ACK(), test.ack)
			}

			if p.NAK() != test.nak {
				t.Errorf("got nak %v, want %v", p.NAK(), test.nak)
			}
		})
	}
}

func TestPacketMarshalUnmarshalBinary(t *testing.T) {
	tests := []struct {
		desc        string
		input       []byte
		expected    *Packet
		expectedErr error
	}{
		{"error", []byte{0x00}, &Packet{}, ErrNoSync},
		{
			desc:        "std message ack",
			input:       []byte{0x02, 0x62, 0x3d, 0x96, 0xe1, 0x0f, 0x0d, 0x00, 0x06},
			expected:    &Packet{Ack: 0x06, Command: CmdSendInsteonMsg, Payload: []byte{0x00, 0x00, 0x00, 0x3d, 0x96, 0xe1, 0x0f, 0x0d, 0x00}},
			expectedErr: nil,
		},
		{
			desc:        "ext message ack",
			input:       []byte{0x02, 0x62, 0x57, 0x2c, 0x5c, 0x1f, 0x2f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xd1, 0x06},
			expected:    &Packet{Ack: 0x06, Command: CmdSendInsteonMsg, Payload: []byte{0, 0, 0, 0x57, 0x2c, 0x5c, 0x1f, 0x2f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xd1}},
			expectedErr: nil,
		},
		{"std message no ack", []byte{0x02, byte(CmdSendInsteonMsg), 0x01, 0x02, 0x03, 0x04}, &Packet{Ack: 0x00, Command: CmdSendInsteonMsg, Payload: []byte{0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04}}, nil},
		{"ext message no ack", []byte{0x02, byte(CmdSendInsteonMsg), 0x01, 0x02, 0x03, 0x04, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, &Packet{Ack: 0x00, Command: CmdSendInsteonMsg, Payload: []byte{0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}}, nil},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			packet := &Packet{}
			err := packet.UnmarshalBinary(test.input)
			if err == test.expectedErr {
				if err == nil {
					if !reflect.DeepEqual(packet, test.expected) {
						t.Errorf("got %v, want %v", packet, test.expected)
					}
				}
			} else {
				t.Errorf("got error %v, want %v", err, test.expectedErr)
			}
		})
	}
}

func TestPacketMarshalBinary(t *testing.T) {
	tests := []struct {
		desc     string
		input    *Packet
		expected []byte
	}{
		{"1234", &Packet{Ack: 0x06, Command: CmdSendInsteonMsg, Payload: []byte{0x01, 0x02, 0x03, 0x04}}, []byte{0x02, byte(CmdSendInsteonMsg), 0x01, 0x02, 0x03, 0x04}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			buf, _ := test.input.MarshalBinary()
			if !bytes.Equal(test.expected, buf) {
				t.Errorf("got %v, want %v", test.expected, buf)
			}
		})
	}
}

func TestPacketFormat(t *testing.T) {
	tests := []struct {
		format   string
		input    *Packet
		expected string
	}{
		{"%x", &Packet{Command: 0x01, Ack: 0x06}, "0106"},
		{"%x", &Packet{Command: 0x01, Payload: []byte{2, 3, 4, 5, 15}, Ack: 0x06}, "01020304050f06"},
		{"%X", &Packet{Command: 0x01, Payload: []byte{2, 3, 4, 5, 15}, Ack: 0x06}, "01020304050F06"},
		{"%s", &Packet{Command: CmdGetInfo, Ack: 0x06}, fmt.Sprintf("%v ACK", CmdGetInfo)},
		{"%v", &Packet{Command: CmdGetInfo, Payload: []byte{2, 3, 4, 5, 15}, Ack: 0x06}, fmt.Sprintf("%v 02 03 04 05 0f ACK", CmdGetInfo)},
		{"%v", &Packet{Command: CmdGetInfo, Payload: []byte{2, 3, 4, 5, 15}, Ack: 0x15}, fmt.Sprintf("%v 02 03 04 05 0f NAK", CmdGetInfo)},
		{"%q", &Packet{Command: CmdGetInfo, Payload: []byte{2, 3, 4, 5, 15}, Ack: 0x15}, fmt.Sprintf("%q", fmt.Sprintf("%v NAK", CmdGetInfo))},
		{"%d", &Packet{Command: CmdGetInfo, Ack: 0x06}, fmt.Sprintf("%%!d(packet=%v ACK)", CmdGetInfo)},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			str := fmt.Sprintf(test.format, test.input)
			if str != test.expected {
				t.Errorf("got %q, want %q", str, test.expected)
			}
		})
	}
}
