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
	"bytes"
	"testing"
)

func testMsg(typ MessageType, cmd Command, payload ...byte) *Message {
	return &Message{Address{6, 7, 8}, Address{9, 10, 11}, Flag(typ, len(payload) > 0, 2, 2), cmd, payload}
}

var (
	SetButtonPressed = func(controller bool, domain, category, firmware byte) *Message {
		sbp := CmdSetButtonPressedResponder
		if controller {
			sbp = CmdSetButtonPressedController
		}
		return &Message{testDstAddr, Address{domain, category, firmware}, StandardBroadcast, sbp, nil}
	}

	Ack = func(ack bool, cmd1, cmd2 byte) *Message {
		flags := StandardDirectNak
		if ack {
			flags = StandardDirectAck
		}

		return &Message{testSrcAddr, testDstAddr, flags, Command(int(cmd1)<<7 | int(cmd2)), nil}
	}

	testSrcAddr = Address{1, 2, 3}
	testDstAddr = Address{3, 4, 5}

	TestMessageEngineVersion1   = &Message{testSrcAddr, testDstAddr, StandardDirectAck, Command(0x000d00), nil}
	TestMessageEngineVersion2   = &Message{testSrcAddr, testDstAddr, StandardDirectAck, Command(0x000d01), nil}
	TestMessageEngineVersion2cs = &Message{testSrcAddr, testDstAddr, StandardDirectAck, Command(0x000d02), nil}
	TestMessageEngineVersion3   = &Message{testSrcAddr, testDstAddr, StandardDirectAck, Command(0x000d03), nil}

	TestMessagePing    = &Message{testSrcAddr, testDstAddr, StandardDirectMessage, Command(0x000f00), nil}
	TestMessagePingAck = &Message{testDstAddr, testSrcAddr, StandardDirectAck, Command(0x000f00), nil}
	TestPingNak        = &Message{testSrcAddr, testDstAddr, StandardDirectNak, Command(0x000f00), nil}
	TestAck            = &Message{testSrcAddr, testDstAddr, StandardDirectAck, Command(0x000000), nil}

	TestProductDataResponse = &Message{testDstAddr, testSrcAddr, ExtendedDirectMessage, CmdProductDataResp, []byte{0, 1, 2, 3, 4, 5, 0xff, 0xff, 0, 0, 0, 0, 0, 0}}
	TestDeviceLink1         = &Message{testSrcAddr, testDstAddr, ExtendedDirectMessage, CmdReadWriteALDB, []byte{0, 1, 0x0f, 0xff, 0, 0xc0, 1, 7, 8, 9, 0, 0, 0, 0}}
	TestDeviceLink2         = &Message{testSrcAddr, testDstAddr, ExtendedDirectMessage, CmdReadWriteALDB, []byte{0, 1, 0x0f, 0xf7, 0, 0xc0, 1, 10, 11, 12, 0, 0, 0, 0}}
	TestDeviceLink3         = &Message{testSrcAddr, testDstAddr, ExtendedDirectMessage, CmdReadWriteALDB, []byte{0, 1, 0x0f, 0xf7, 0, 0x00, 0, 0, 0, 0, 0, 0, 0, 0}}

	TestMessageUnknownCommandNak  = &Message{testDstAddr, testSrcAddr, StandardDirectNak, Command(0x0000fd), nil}
	TestMessageNoLoadDetected     = &Message{testDstAddr, testSrcAddr, StandardDirectNak, Command(0x0000fe), nil}
	TestMessageNotLinked          = &Message{testDstAddr, testSrcAddr, StandardDirectNak, Command(0x0000ff), nil}
	TestMessageIllegalValue       = &Message{testDstAddr, testSrcAddr, StandardDirectNak, Command(0x0000fb), nil}
	TestMessagePreNak             = &Message{testDstAddr, testSrcAddr, StandardDirectNak, Command(0x0000fc), nil}
	TestMessageIncorrectChecksum  = &Message{testDstAddr, testSrcAddr, StandardDirectNak, Command(0x0000fd), nil}
	TestMessageNoLoadDetectedI2Cs = &Message{testDstAddr, testSrcAddr, StandardDirectNak, Command(0x0000fe), nil}
	TestMessageNotLinkedI2Cs      = &Message{testDstAddr, testSrcAddr, StandardDirectNak, Command(0x0000ff), nil}

	TestPing    = &Message{testSrcAddr, testDstAddr, StandardDirectMessage, CmdPing, nil}
	TestPingAck = &Message{testDstAddr, testSrcAddr, StandardDirectAck, CmdPing, nil}
)

func TestMessageType(t *testing.T) {
	tests := []struct {
		input             MessageType
		expectedDirect    bool
		expectedBroadcast bool
		expectedString    string
	}{
		{MsgTypeDirect, true, false, "D"},
		{MsgTypeDirectAck, true, false, "D Ack"},
		{MsgTypeAllLinkCleanup, true, false, "C"},
		{MsgTypeAllLinkCleanupAck, true, false, "C Ack"},
		{MsgTypeBroadcast, false, true, "B"},
		{MsgTypeDirectNak, true, false, "D NAK"},
		{MsgTypeAllLinkBroadcast, false, true, "A"},
		{MsgTypeAllLinkCleanupNak, true, false, "C NAK"},
	}

	for _, test := range tests {
		t.Run(test.expectedString, func(t *testing.T) {
			if test.input.Direct() != test.expectedDirect {
				t.Errorf("got Direct %v, want %v", test.input.Direct(), test.expectedDirect)
			}

			if test.input.Broadcast() != test.expectedBroadcast {
				t.Errorf("got Broadcast %v, want %v", test.input.Broadcast(), test.expectedBroadcast)
			}

			if test.input.String() != test.expectedString {
				t.Errorf("got String %q, want %q", test.input.String(), test.expectedString)
			}
		})
	}
}

func TestFlags(t *testing.T) {
	tests := []struct {
		desc             string
		input            Flags
		expectedType     MessageType
		expectedExtended bool
		expectedStandard bool
		expectedTTL      uint8
		expectedMaxTTL   uint8
	}{
		{"MsgTypeDirect", 0x0f, MsgTypeDirect, false, true, 3, 3},
		{"MsgTypeDirectAck", 0x2f, MsgTypeDirectAck, false, true, 3, 3},
		{"MsgTypeAllLinkCleanup", 0x4f, MsgTypeAllLinkCleanup, false, true, 3, 3},
		{"MsgTypeAllLinkCleanupAck", 0x6f, MsgTypeAllLinkCleanupAck, false, true, 3, 3},
		{"MsgTypeBroadcast", 0x8f, MsgTypeBroadcast, false, true, 3, 3},
		{"MsgTypeDirectNak", 0xaf, MsgTypeDirectNak, false, true, 3, 3},
		{"MsgTypeAllLinkBroadcast", 0xcf, MsgTypeAllLinkBroadcast, false, true, 3, 3},
		{"MsgTypeAllLinkCleanupNak", 0xef, MsgTypeAllLinkCleanupNak, false, true, 3, 3},
		{"MsgTypeDirect", 0x1f, MsgTypeDirect, true, false, 3, 3},
		{"MsgTypeDirectAck", 0x3f, MsgTypeDirectAck, true, false, 3, 3},
		{"MsgTypeAllLinkCleanup", 0x5f, MsgTypeAllLinkCleanup, true, false, 3, 3},
		{"MsgTypeAllLinkCleanupAck", 0x7f, MsgTypeAllLinkCleanupAck, true, false, 3, 3},
		{"MsgTypeBroadcast", 0x9f, MsgTypeBroadcast, true, false, 3, 3},
		{"MsgTypeDirectNak", 0xbf, MsgTypeDirectNak, true, false, 3, 3},
		{"MsgTypeAllLinkBroadcast", 0xdf, MsgTypeAllLinkBroadcast, true, false, 3, 3},
		{"MsgTypeAllLinkCleanupNak", 0xff, MsgTypeAllLinkCleanupNak, true, false, 3, 3},
		{"Flag 1", Flag(MsgTypeDirect, false, 2, 2), MsgTypeDirect, false, true, 2, 2},
		{"Flag 2", Flag(MsgTypeDirect, true, 3, 3), MsgTypeDirect, true, false, 3, 3},
		{"Flag 2", Flag(MsgTypeDirect, false, 4, 4), MsgTypeDirect, false, true, 0, 0},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			if test.input.Type() != test.expectedType {
				t.Errorf("got Type %v, want %v", test.input.Type(), test.expectedType)
			}

			if test.input.Standard() != test.expectedStandard {
				t.Errorf("got Standard %v, want %v", test.input.Standard(), test.expectedStandard)
			}

			if test.input.Extended() != test.expectedExtended {
				t.Errorf("got Extended %v, want %v", test.input.Extended(), test.expectedExtended)
			}

			if test.input.TTL() != test.expectedTTL {
				t.Errorf("got TTL %v, want %v", test.input.TTL(), test.expectedTTL)
			}

			if test.input.MaxTTL() != test.expectedMaxTTL {
				t.Errorf("got MaxTTL %v, want %v", test.input.MaxTTL(), test.expectedMaxTTL)
			}
		})
	}
}

func TestMessageMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		desc          string
		input         []byte
		version       EngineVersion
		wantSrc       Address
		wantDst       Address
		wantFlags     Flags
		wantAck       bool
		wantNak       bool
		wantCmd       Command
		expectedError error
	}{
		// Test 0
		{
			desc:      "0",
			input:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x0a, 0x10, 0x00},
			version:   VerI1,
			wantSrc:   Address{0x01, 0x02, 0x03},
			wantDst:   Address{0x04, 0x05, 0x06},
			wantFlags: StandardDirectMessage,
			wantCmd:   Command(0x001000),
		},
		// Test 1
		{
			desc:      "1",
			input:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x8a, 0x01, 0x00},
			version:   VerI1,
			wantSrc:   Address{0x01, 0x02, 0x03},
			wantDst:   Address{0x04, 0x05, 0x06},
			wantFlags: Flags(0x8a),
			wantCmd:   Command(0x080100),
		},
		// Test 2
		{
			desc:          "2",
			input:         []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x0a, 0x10},
			expectedError: ErrBufferTooShort,
		},
		// Test 3
		{
			desc:      "3",
			input:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x1a, 0x09, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe},
			version:   VerI1,
			wantSrc:   Address{0x01, 0x02, 0x03},
			wantDst:   Address{0x04, 0x05, 0x06},
			wantFlags: ExtendedDirectMessage,
			wantCmd:   Command(0x010900),
		},
		// Test 4
		{
			desc:          "4",
			input:         []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x1a, 0x09, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			expectedError: ErrBufferTooShort,
		},
		// Test 5
		{
			desc:      "5",
			input:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x1a, 0x2f, 0x00, 0x00, 0x02, 0x0f, 0xff, 0x08, 0xe2, 0x01, 0x08, 0xb6, 0xea, 0x00, 0x1b, 0x01, 0x12},
			version:   VerI2Cs,
			wantSrc:   Address{0x01, 0x02, 0x03},
			wantDst:   Address{0x04, 0x05, 0x06},
			wantFlags: ExtendedDirectMessage,
			wantCmd:   Command(0x012f00),
		},
		// Test 6
		{
			desc:      "6",
			input:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0xaa, 0x01, 0x00},
			version:   VerI1,
			wantSrc:   Address{0x01, 0x02, 0x03},
			wantDst:   Address{0x04, 0x05, 0x06},
			wantFlags: Flags(0xaa),
			wantNak:   true,
			wantCmd:   Command(0x020100),
		},
		// Test 7
		{
			desc:      "7",
			input:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x2a, 0x01, 0x00},
			version:   VerI1,
			wantSrc:   Address{0x01, 0x02, 0x03},
			wantDst:   Address{0x04, 0x05, 0x06},
			wantFlags: Flags(0x2a),
			wantAck:   true,
			wantCmd:   Command(0x020100),
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			message := &Message{}
			err := message.UnmarshalBinary(test.input)
			if !IsError(err, test.expectedError) {
				t.Errorf("expected %v got %v", test.expectedError, err)
				return
			} else if err != nil {
				return
			}

			if test.wantSrc != message.Src {
				t.Errorf("got Src %v, want %v", message.Src, test.wantSrc)
			}

			if test.wantDst != message.Dst {
				t.Errorf("got Dst %v, want %v", message.Dst, test.wantDst)
			}

			if test.wantFlags != message.Flags {
				t.Errorf("got Flags %v, want %v", message.Flags, test.wantFlags)
			}

			if test.wantAck != message.Ack() {
				t.Errorf("got Ack %v, want %v", message.Ack(), test.wantAck)
			}

			if test.wantNak != message.Nak() {
				t.Errorf("got Nak %v, want %v", message.Nak(), test.wantNak)
			}

			if test.wantCmd != message.Command {
				t.Errorf("want Command 0x%06x got 0x%06x", int(test.wantCmd), int(message.Command))
			}

			buf, _ := message.MarshalBinary()

			if !bytes.Equal(buf, test.input) {
				t.Errorf("got bytes %v, want %v", buf, test.input)
			}
		})
	}
}

func TestCommonTypeConsts(t *testing.T) {
	tests := []struct {
		want Flags
		MessageType
		Extended bool
		MaxHops  uint8
		HopsLeft uint8
	}{
		{StandardBroadcast, MsgTypeBroadcast, false, 2, 2},
		{StandardAllLinkBroadcast, MsgTypeAllLinkBroadcast, false, 2, 2},
		{StandardDirectMessage, MsgTypeDirect, false, 2, 2},
		{StandardDirectAck, MsgTypeDirectAck, false, 2, 2},
		{StandardDirectNak, MsgTypeDirectNak, false, 2, 2},
		{ExtendedDirectMessage, MsgTypeDirect, true, 2, 2},
		{ExtendedDirectAck, MsgTypeDirectAck, true, 2, 2},
		{ExtendedDirectNak, MsgTypeDirectNak, true, 2, 2},
	}

	for _, test := range tests {
		got := Flag(test.MessageType, test.Extended, test.MaxHops, test.HopsLeft)
		if got != test.want {
			t.Errorf("Got %v, wanted %v", got, test.want)
		}
	}
}
