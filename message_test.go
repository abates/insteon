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
		expectedCommand *Command
		expectedError   error
		expectedString  string
	}{
		{
			input:           []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x0a, 0x10, 0x00},
			version:         VerI1,
			expectedSrc:     Address{0x01, 0x02, 0x03},
			expectedDst:     Address{0x04, 0x05, 0x06},
			expectedFlags:   StandardDirectMessage,
			expectedCommand: CmdIDReq,
			expectedString:  "01.02.03 -> 04.05.06 SD     2:2",
		},
		{
			input:         []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x0a, 0x10},
			expectedError: ErrBufferTooShort,
		},
		{
			input:           []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x1a, 0x09, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe},
			version:         VerI1,
			expectedSrc:     Address{0x01, 0x02, 0x03},
			expectedDst:     Address{0x04, 0x05, 0x06},
			expectedFlags:   ExtendedDirectMessage,
			expectedCommand: CmdEnterLinkingModeExtended,
			expectedString:  "01.02.03 -> 04.05.06 ED     2:2",
		},
		{
			input:         []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x1a, 0x09, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			expectedError: ErrBufferTooShort,
		},
		{
			input:           []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x1a, 0x2f, 0x00, 0x00, 0x02, 0x0f, 0xff, 0x08, 0xe2, 0x01, 0x08, 0xb6, 0xea, 0x00, 0x1b, 0x01, 0x12},
			version:         VerI2Cs,
			expectedSrc:     Address{0x01, 0x02, 0x03},
			expectedDst:     Address{0x04, 0x05, 0x06},
			expectedFlags:   ExtendedDirectMessage,
			expectedCommand: CmdReadWriteALDB,
			expectedString:  "01.02.03 -> 04.05.06 ED     2:2",
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
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedFlags, message.Flags)
		}

		if test.expectedString != message.String()[0:len(test.expectedString)] {
			t.Errorf("tests[%d] expected %q got %q", i, test.expectedString, message.String()[0:len(test.expectedString)])
		}

		buf, _ := message.MarshalBinary()

		if !bytes.Equal(buf, test.input) {
			t.Errorf("tests[%d] expected %v got %v", i, test.input, buf)
		}
	}
}
