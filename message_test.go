package insteon

import (
	"bytes"
	"testing"
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

	for i, test := range tests {
		if test.input.Direct() != test.expectedDirect {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedDirect, test.input.Direct())
		}

		if test.input.Broadcast() != test.expectedBroadcast {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedBroadcast, test.input.Broadcast())
		}

		if test.input.String() != test.expectedString {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedString, test.input.String())
		}
	}
}

func TestFlags(t *testing.T) {
	tests := []struct {
		input            Flags
		expectedType     MessageType
		expectedExtended bool
		expectedStandard bool
		expectedTTL      int
		expectedMaxTTL   int
		expectedString   string
	}{
		{0x0f, MsgTypeDirect, false, true, 3, 3, "SD     3:3"},
		{0x2f, MsgTypeDirectAck, false, true, 3, 3, "SD Ack 3:3"},
		{0x4f, MsgTypeAllLinkCleanup, false, true, 3, 3, "SC     3:3"},
		{0x6f, MsgTypeAllLinkCleanupAck, false, true, 3, 3, "SC Ack 3:3"},
		{0x8f, MsgTypeBroadcast, false, true, 3, 3, "SB     3:3"},
		{0xaf, MsgTypeDirectNak, false, true, 3, 3, "SD NAK 3:3"},
		{0xcf, MsgTypeAllLinkBroadcast, false, true, 3, 3, "SA     3:3"},
		{0xef, MsgTypeAllLinkCleanupNak, false, true, 3, 3, "SC NAK 3:3"},
		{0x1f, MsgTypeDirect, true, false, 3, 3, "ED     3:3"},
		{0x3f, MsgTypeDirectAck, true, false, 3, 3, "ED Ack 3:3"},
		{0x5f, MsgTypeAllLinkCleanup, true, false, 3, 3, "EC     3:3"},
		{0x7f, MsgTypeAllLinkCleanupAck, true, false, 3, 3, "EC Ack 3:3"},
		{0x9f, MsgTypeBroadcast, true, false, 3, 3, "EB     3:3"},
		{0xbf, MsgTypeDirectNak, true, false, 3, 3, "ED NAK 3:3"},
		{0xdf, MsgTypeAllLinkBroadcast, true, false, 3, 3, "EA     3:3"},
		{0xff, MsgTypeAllLinkCleanupNak, true, false, 3, 3, "EC NAK 3:3"},
	}

	for i, test := range tests {
		if test.input.Type() != test.expectedType {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedType, test.input.Type())
		}

		if test.input.Standard() != test.expectedStandard {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedStandard, test.input.Standard())
		}

		if test.input.Extended() != test.expectedExtended {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedExtended, test.input.Extended())
		}

		if test.input.TTL() != test.expectedTTL {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedTTL, test.input.TTL())
		}

		if test.input.MaxTTL() != test.expectedMaxTTL {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedMaxTTL, test.input.MaxTTL())
		}

		if test.input.String() != test.expectedString {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedString, test.input.String())
		}
	}
}

func TestMessageMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		input           []byte
		version         EngineVersion
		expectedSrc     Address
		expectedDst     Address
		expectedFlags   Flags
		expectedCommand CommandBytes
		expectedError   error
	}{
		// Test 0
		{
			input:           []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x0a, 0x10, 0x00},
			version:         VerI1,
			expectedSrc:     Address{0x01, 0x02, 0x03},
			expectedDst:     Address{0x04, 0x05, 0x06},
			expectedFlags:   StandardDirectMessage,
			expectedCommand: CommandBytes{Command1: 0x10, Command2: 0x0},
		},
		// Test 1
		{
			input:           []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x8a, 0x01, 0x00},
			version:         VerI1,
			expectedSrc:     Address{0x01, 0x02, 0x03},
			expectedDst:     Address{0x04, 0x05, 0x06},
			expectedFlags:   Flags(0x8a),
			expectedCommand: CommandBytes{Command1: 0x01, Command2: 0x00},
		},
		// Test 2
		{
			input:         []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x0a, 0x10},
			expectedError: ErrBufferTooShort,
		},
		// Test 3
		{
			input:           []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x1a, 0x09, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe},
			version:         VerI1,
			expectedSrc:     Address{0x01, 0x02, 0x03},
			expectedDst:     Address{0x04, 0x05, 0x06},
			expectedFlags:   ExtendedDirectMessage,
			expectedCommand: CommandBytes{Command1: 0x09, Command2: 0x00},
		},
		// Test 4
		{
			input:         []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x1a, 0x09, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			expectedError: ErrBufferTooShort,
		},
		// Test 5
		{
			input:           []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x1a, 0x2f, 0x00, 0x00, 0x02, 0x0f, 0xff, 0x08, 0xe2, 0x01, 0x08, 0xb6, 0xea, 0x00, 0x1b, 0x01, 0x12},
			version:         VerI2Cs,
			expectedSrc:     Address{0x01, 0x02, 0x03},
			expectedDst:     Address{0x04, 0x05, 0x06},
			expectedFlags:   ExtendedDirectMessage,
			expectedCommand: CommandBytes{Command1: 0x2f, Command2: 0x00},
		},
	}

	for i, test := range tests {
		message := &Message{version: test.version}
		err := message.UnmarshalBinary(test.input)
		if !IsError(test.expectedError, err) {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedError, err)
			continue
		} else if err != nil {
			continue
		}

		if test.expectedSrc != message.Src {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedSrc, message.Src)
		}

		if test.expectedDst != message.Dst {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedDst, message.Dst)
		}

		if test.expectedFlags != message.Flags {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedFlags, message.Flags)
		}

		if test.expectedCommand != message.Command {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedCommand, message.Command)
		}

		buf, _ := message.MarshalBinary()

		if !bytes.Equal(buf, test.input) {
			t.Errorf("tests[%d] expected %v got %v", i, test.input, buf)
		}
	}
}

func TestChecksum(t *testing.T) {
	tests := []struct {
		input    []byte
		expected byte
	}{
		{[]byte{0x2E, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xd1},
		{[]byte{0x2F, 0x00, 0x00, 0x00, 0x0F, 0xFF, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xC2},
		{[]byte{0x2F, 0x00, 0x01, 0x01, 0x0F, 0xFF, 0x00, 0xA2, 0x00, 0x19, 0x70, 0x1A, 0xFF, 0x1F, 0x01}, 0x5D},
		{[]byte{0x2F, 0x00, 0x00, 0x00, 0x0F, 0xF7, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xCA},
		{[]byte{0x2F, 0x00, 0x01, 0x01, 0x0F, 0xF7, 0x00, 0xE2, 0x01, 0x19, 0x70, 0x1A, 0xFF, 0x1F, 0x01}, 0x24},
		{[]byte{0x2F, 0x00, 0x00, 0x00, 0x0F, 0xEF, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xD2},
		{[]byte{0x2F, 0x00, 0x01, 0x01, 0x0F, 0xEF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xD1},
		{[]byte{0x2F, 0x00, 0x01, 0x02, 0x0F, 0xFF, 0x08, 0xE2, 0x01, 0x08, 0xB6, 0xEA, 0x00, 0x1B, 0x01}, 0x11},
		{[]byte{0x09, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xF6},
		{[]byte{0x2f, 0x00, 0x00, 0x02, 0x0f, 0xff, 0x08, 0xe2, 0x01, 0x08, 0xb6, 0xea, 0x00, 0x1b, 0x01}, 0x12},
	}

	for i, test := range tests {
		got := checksum(test.input)
		if got != test.expected {
			t.Errorf("tests[%d] expected %02x got %02d", i, test.expected, got)
		}
	}
}
