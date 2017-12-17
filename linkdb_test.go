package insteon

import (
	"bytes"
	"testing"
)

func TestMemAddress(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0xffff, "ff.ff"},
		{0x0fff, "0f.ff"},
		{0x0f00, "0f.00"},
	}

	for i, test := range tests {
		addr := MemAddress(test.input)
		if addr.String() != test.expected {
			t.Errorf("tests[%d] expected %v got %v", i, test.expected, addr.String())
		}
	}
}

func TestLinkRequestType(t *testing.T) {
	tests := []struct {
		input    byte
		expected string
	}{
		{0x00, "Link Read"},
		{0x01, "Link Resp"},
		{0x02, "Link Write"},
		{0x03, "Unknown"},
	}

	for i, test := range tests {
		lrt := LinkRequestType(test.input)
		if test.expected != lrt.String() {
			t.Errorf("tests[%d] expected %v got %v", i, test.expected, lrt.String())
		}
	}
}

func TestLinkRequest(t *testing.T) {
	tests := []struct {
		input           []byte
		marshal         []byte
		expectedType    LinkRequestType
		expectedAddress MemAddress
		expectedRecords int
		expectedString  string
		expectedError   error
	}{
		{
			input:         []byte{},
			expectedError: ErrBufferTooShort,
		},
		{
			input:           []byte{0xff, 0x00, 0x0f, 0xff, 0x08},
			marshal:         []byte{0x0, 0x00, 0x0f, 0xff, 0x08, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedType:    LinkRequestType(0x00),
			expectedAddress: MemAddress(0x0fff),
			expectedRecords: 8,
			expectedString:  "Link Read 0f.ff 8",
		},
		{
			input:           []byte{0xff, 0x01, 0x0f, 0xff, 0x08, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			marshal:         []byte{0x00, 0x01, 0x0f, 0xff, 0x00, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedType:    LinkRequestType(0x01),
			expectedAddress: MemAddress(0x0fff),
			expectedRecords: 0,
			expectedString:  "Link Resp 0f.ff 0",
		},
		{
			input:           []byte{0xff, 0x02, 0x0f, 0xff, 0x08, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			marshal:         []byte{0x00, 0x02, 0x0f, 0xff, 0x08, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedType:    LinkRequestType(0x02),
			expectedAddress: MemAddress(0x0fff),
			expectedRecords: 8,
			expectedString:  "Link Write 0f.ff 8",
		},
	}

	for i, test := range tests {
		linkRequest := &LinkRequest{}
		err := linkRequest.UnmarshalBinary(test.input)
		if !IsError(test.expectedError, err) {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedError, err)
			continue
		} else if err != nil {
			continue
		}

		if linkRequest.Type != test.expectedType {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedType, linkRequest.Type)
		}

		if linkRequest.MemAddress != test.expectedAddress {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedAddress, linkRequest.MemAddress)
		}

		if linkRequest.NumRecords != test.expectedRecords {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedRecords, linkRequest.NumRecords)
		}

		if linkRequest.String()[0:len(test.expectedString)] != test.expectedString {
			t.Errorf("tests[%d] expected %q got %q", i, test.expectedString, linkRequest.String()[0:len(test.expectedString)])
		}

		buf, _ := linkRequest.MarshalBinary()
		if !bytes.Equal(test.marshal, buf) {
			t.Errorf("tests[%d] expected %v got %v", i, test.marshal, buf)
		}
	}
}

func TestCleanup(t *testing.T) {
	/*var flags RecordControlFlags
	flags.setController()

	link1 := func(available bool) *Link {
		if available {
			flags.setAvailable()
		}
		return &Link{Flags: flags, Group: Group(0x01), Address: Address{0x01, 0x02, 0x03}, Data: [3]byte{0x04, 0x05, 0x06}}
	}

	link2 := func(available bool) *Link {
		if available {
			flags.setAvailable()
		}
		return &Link{Flags: flags, Group: Group(0x01), Address: Address{0x07, 0x08, 0x09}, Data: [3]byte{0x0a, 0x0b, 0x0c}}
	}

	tests := []struct {
		input    []*Link
		expected []*Link
	}{
		{
			input:    []*Link{link1(false), link1(false), link1(false), link2(false), link2(false), link2(false)},
			expected: []*Link{link1(false), link1(true), link1(true), link2(false), link2(true), link2(true)},
		},
	}

	for i, test := range tests {
		linkdb := DeviceLinkDB{
			conn: &testConnection{},
		}
		oldLinkFunc := linkFunc
		linkFunc = func(conn Connection) ([]*Link, error) {
			return test.input, nil
		}

		linkdb.Cleanup()
		if !reflect.DeepEqual(test.expected, linkdb.links) {
			t.Errorf("tests[%d] expected:\n%s\ngot\n%s", i, test.expected, linkdb.links)
		}
		linkFunc = oldLinkFunc
	}*/
}
