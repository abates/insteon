package insteon

import (
	"bytes"
	"testing"
)

func TestRecordControlFlags(t *testing.T) {
	tests := []struct {
		input              byte
		expectedInUse      bool
		expectedController bool
		expectedString     string
	}{
		{0x40, false, true, "AC"},
		{0x00, false, false, "AR"},
		{0xc0, true, true, "UC"},
		{0x80, true, false, "UR"},
	}

	for i, test := range tests {
		flags := RecordControlFlags(test.input)
		if flags.InUse() != test.expectedInUse {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedInUse, flags.InUse())
		}

		if flags.Available() == test.expectedInUse {
			t.Errorf("tests[%d] expected %v got %v", i, !test.expectedInUse, flags.Available())
		}

		if flags.Controller() != test.expectedController {
			t.Errorf("tests[%d] expected %v got %v", i, !test.expectedController, flags.Controller())
		}

		if flags.Responder() == test.expectedController {
			t.Errorf("tests[%d] expected %v got %v", i, !test.expectedController, flags.Responder())
		}

		if flags.String() != test.expectedString {
			t.Errorf("tests[%d] expected %q got %q", i, test.expectedString, flags.String())
		}
	}
}

func TestSettingRecordControlFlags(t *testing.T) {
	flags := RecordControlFlags(0xff)
	tests := []struct {
		set      func()
		expected byte
	}{
		{flags.setAvailable, 0x7f},
		{flags.setInUse, 0xff},
		{flags.setResponder, 0xbf},
		{flags.setController, 0xff},
	}

	for i, test := range tests {
		test.set()
		if byte(flags) != test.expected {
			t.Errorf("tests[%d] expected 0x%02x got 0x%02x", i, test.expected, byte(flags))
		}
	}
}

func TestLinkEqual(t *testing.T) {
	availableController := RecordControlFlags(0x40)
	availableResponder := RecordControlFlags(0x00)
	usedController := RecordControlFlags(0xc0)
	usedResponder := RecordControlFlags(0x80)

	newLink := func(flags RecordControlFlags, group Group, address Address) *LinkRecord {
		buffer := []byte{byte(flags), byte(group), address[0], address[1], address[2], 0x00, 0x00, 0x00}
		link := &LinkRecord{}
		link.UnmarshalBinary(buffer)
		return link
	}

	l1 := newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03})
	l2 := l1

	tests := []struct {
		link1    *LinkRecord
		link2    *LinkRecord
		expected bool
	}{
		{newLink(usedController, Group(0x01), Address{0x01, 0x02, 0x03}), newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), false},
		{newLink(usedResponder, Group(0x01), Address{0x01, 0x02, 0x03}), newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), false},
		{newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), true},
		{newLink(availableResponder, Group(0x01), Address{0x01, 0x02, 0x03}), newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), false},
		{newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x04}), false},
		{newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), nil, false},
		{l1, l2, true},
	}

	for i, test := range tests {
		if test.link1.Equal(test.link2) != test.expected {
			t.Errorf("tests[%d] expected %v got %v", i, test.expected, !test.expected)
		}
	}
}

func TestLinkMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		input           []byte
		expectedFlags   RecordControlFlags
		expectedGroup   Group
		expectedAddress Address
		expectedData    [3]byte
		expectedString  string
		expectedError   error
	}{
		{
			input:           []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			expectedFlags:   RecordControlFlags(0x01),
			expectedGroup:   Group(0x02),
			expectedAddress: Address{0x03, 0x04, 0x05},
			expectedData:    [3]byte{0x06, 0x07, 0x08},
			expectedString:  "AR 2 03.04.05 0x06 0x07 0x08",
		},
		{
			input:         []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
			expectedError: ErrBufferTooShort,
		},
	}

	for i, test := range tests {
		link := &LinkRecord{}
		err := link.UnmarshalBinary(test.input)
		if !IsError(test.expectedError, err) {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedError, err)
			continue
		} else if err != nil {
			continue
		}

		if link.Flags != test.expectedFlags {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedFlags, link.Flags)
		}

		if link.Group != test.expectedGroup {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedGroup, link.Group)
		}

		if link.Address != test.expectedAddress {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedAddress, link.Address)
		}

		if link.Data != test.expectedData {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedData, link.Data)
		}

		if link.String() != test.expectedString {
			t.Errorf("tests[%d] expected %q got %q", i, test.expectedString, link.String())
		}

		buf, _ := link.MarshalBinary()

		if !bytes.Equal(buf, test.input) {
			t.Errorf("tests[%d] expected %v got %v", i, test.input, buf)
		}
	}
}
