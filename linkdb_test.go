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
		// test 0
		{
			input:         []byte{},
			expectedError: ErrBufferTooShort,
		},
		// test 1
		{
			input:           []byte{0xff, 0x00, 0x0f, 0xff, 0x08},
			marshal:         []byte{0x0, 0x00, 0x0f, 0xff, 0x08, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedType:    LinkRequestType(0x00),
			expectedAddress: MemAddress(0x0fff),
			expectedRecords: 8,
			expectedString:  "Link Read 0f.ff 8",
		},
		// test 2
		{
			input:           []byte{0xff, 0x01, 0x0f, 0xff, 0x08, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			marshal:         []byte{0x00, 0x01, 0x0f, 0xff, 0x00, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedType:    LinkRequestType(0x01),
			expectedAddress: MemAddress(0x0fff),
			expectedRecords: 0,
			expectedString:  "Link Resp 0f.ff 0",
			expectedError:   ErrEndOfLinks,
		},
		// test 3
		{
			input:           []byte{0xff, 0x02, 0x0f, 0xff, 0x08, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			marshal:         []byte{0x00, 0x02, 0x0f, 0xff, 0x08, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedType:    LinkRequestType(0x02),
			expectedAddress: MemAddress(0x0fff),
			expectedRecords: 8,
			expectedString:  "Link Write 0f.ff 8",
			expectedError:   ErrEndOfLinks,
		},
	}

	for i, test := range tests {
		linkRequest := &LinkRequest{}
		err := linkRequest.UnmarshalBinary(test.input)
		if !IsError(err, test.expectedError) {
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

func TestAddLink(t *testing.T) {
}

func TestRemoveLink(t *testing.T) {
}

func TestCleanup(t *testing.T) {
}
