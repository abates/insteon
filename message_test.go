package insteon

import (
	"bytes"
	"reflect"
	"testing"
)

func TestMessageMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		input           []byte
		version         EngineVersion
		expectedSrc     Address
		expectedDst     Address
		expectedFlags   Flags
		expectedCommand *Command
		expectedPayload Payload
		err             error
	}{
		{
			input:           []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x0a, 0x10, 0x00},
			version:         VerI1,
			expectedSrc:     Address{0x01, 0x02, 0x03},
			expectedDst:     Address{0x04, 0x05, 0x06},
			expectedFlags:   StandardDirectMessage,
			expectedCommand: CmdIDReq,
		},
		{
			input:           []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x0a, 0x10},
			version:         VerI1,
			expectedSrc:     Address{0x01, 0x02, 0x03},
			expectedDst:     Address{0x04, 0x05, 0x06},
			expectedFlags:   StandardDirectMessage,
			expectedCommand: CmdIDReq,
			err:             ErrBufferTooShort,
		},
		{
			input:           []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x1a, 0x09, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe},
			version:         VerI1,
			expectedSrc:     Address{0x01, 0x02, 0x03},
			expectedDst:     Address{0x04, 0x05, 0x06},
			expectedFlags:   ExtendedDirectMessage,
			expectedCommand: CmdEnterLinkingModeExtended,
			expectedPayload: &BufPayload{[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe}},
		},
		{
			input:           []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x1a, 0x09, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			version:         VerI1,
			expectedSrc:     Address{0x01, 0x02, 0x03},
			expectedDst:     Address{0x04, 0x05, 0x06},
			expectedFlags:   ExtendedDirectMessage,
			expectedCommand: CmdEnterLinkingModeExtended,
			err:             ErrExtendedBufferTooShort,
		},
		{
			input:           []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x1a, 0x2f, 0x00, 0x00, 0x02, 0x0f, 0xff, 0x08, 0xe2, 0x01, 0x08, 0xb6, 0xea, 0x00, 0x1b, 0x01, 0x12},
			version:         VerI2Cs,
			expectedSrc:     Address{0x01, 0x02, 0x03},
			expectedDst:     Address{0x04, 0x05, 0x06},
			expectedFlags:   ExtendedDirectMessage,
			expectedCommand: CmdReadWriteALDB,
			expectedPayload: &LinkRequest{LinkRequestType(0x02), MemAddress(0x0fff), 8, &Link{RecordControlFlags(0xe2), Group(0x01), Address{0x08, 0xb6, 0xea}, [3]byte{0x00, 0x1b, 0x01}}},
		},
	}

	for i, test := range tests {
		message := &Message{version: test.version}
		err := message.UnmarshalBinary(test.input)
		if err != test.err {
			t.Errorf("tests[%d] expected %v got %v", i, test.err, err)
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

		if !reflect.DeepEqual(test.expectedPayload, message.Payload) {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedPayload, message.Payload)
		}

		buf, _ := message.MarshalBinary()

		if !bytes.Equal(buf, test.input) {
			t.Errorf("tests[%d] expected %v got %v", i, test.input, buf)
		}
	}
}
