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
	"fmt"
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

	for _, test := range tests {
		t.Run(fmt.Sprintf("%02x", test.input), func(t *testing.T) {
			flags := RecordControlFlags(test.input)
			if flags.InUse() != test.expectedInUse {
				t.Errorf("got InUse %v, want %v", flags.InUse(), test.expectedInUse)
			}

			if flags.Available() == test.expectedInUse {
				t.Errorf("got Available %v, want %v", flags.Available(), !test.expectedInUse)
			}

			if flags.Controller() != test.expectedController {
				t.Errorf("got Controller %v, want %v", flags.Controller(), !test.expectedController)
			}

			if flags.Responder() == test.expectedController {
				t.Errorf("got Responder %v, want %v", flags.Responder(), !test.expectedController)
			}

			if flags.String() != test.expectedString {
				t.Errorf("got String %q, want %q", flags.String(), test.expectedString)
			}
		})
	}
}

func TestRecordControlFlagsUnmarshalText(t *testing.T) {
	tests := []struct {
		input       string
		expectedErr string
		expected    RecordControlFlags
	}{
		{"A", "Expected 2 characters got 1", RecordControlFlags(0x00)},
		{"AR", "", RecordControlFlags(0x00)},
		{"UR", "", RecordControlFlags(0x80)},
		{"AC", "", RecordControlFlags(0x40)},
		{"UC", "", RecordControlFlags(0xc0)},
		{"FR", "Invalid value for Available flag", RecordControlFlags(0x00)},
		{"AZ", "Invalid value for Controller flag", RecordControlFlags(0x00)},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			var flags RecordControlFlags
			err := flags.UnmarshalText([]byte(test.input))
			if err == nil {
				if test.expectedErr != "" {
					t.Errorf("got nil error, want %q", test.expectedErr)
				} else if flags != test.expected {
					t.Errorf("got flags 0x%02x, want 0x%02x", flags, test.expected)
				}
			} else if err.Error() != test.expectedErr {
				t.Errorf("got error %q, want %q", err, test.expectedErr)
			}
		})
	}
}

func TestSettingRecordControlFlags(t *testing.T) {
	flags := RecordControlFlags(0xff)
	tests := []struct {
		desc     string
		set      func()
		expected byte
	}{
		{"setAvailable", flags.setAvailable, 0x7f},
		{"setInUse", flags.setInUse, 0xff},
		{"setResponder", flags.setResponder, 0xbf},
		{"setController", flags.setController, 0xff},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			test.set()
			if byte(flags) != test.expected {
				t.Errorf("got flags 0x%02x, want 0x%02x", byte(flags), test.expected)
			}
		})
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
		desc     string
		link1    *LinkRecord
		link2    *LinkRecord
		expected bool
	}{
		{"not equal 1", newLink(usedController, Group(0x01), Address{0x01, 0x02, 0x03}), newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), false},
		{"not equal 2", newLink(usedResponder, Group(0x01), Address{0x01, 0x02, 0x03}), newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), false},
		{"not equal 3", newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), true},
		{"not equal 4", newLink(availableResponder, Group(0x01), Address{0x01, 0x02, 0x03}), newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), false},
		{"not equal 5", newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x04}), false},
		{"not equal 6", newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), nil, false},
		{"equal", l1, l2, true},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			if test.link1.Equal(test.link2) != test.expected {
				t.Errorf("got %v, want %v", !test.expected, test.expected)
			}
		})
	}
}

func TestLinkMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		desc            string
		input           []byte
		expectedFlags   RecordControlFlags
		expectedGroup   Group
		expectedAddress Address
		expectedData    [3]byte
		expectedString  string
		expectedError   error
	}{
		{
			desc:            "success",
			input:           []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			expectedFlags:   RecordControlFlags(0x01),
			expectedGroup:   Group(0x02),
			expectedAddress: Address{0x03, 0x04, 0x05},
			expectedData:    [3]byte{0x06, 0x07, 0x08},
			expectedString:  "AR 2 03.04.05 0x06 0x07 0x08",
		},
		{
			desc:          "ErrBufferTooShort",
			input:         []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
			expectedError: ErrBufferTooShort,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			link := &LinkRecord{}
			err := link.UnmarshalBinary(test.input)
			if !isError(err, test.expectedError) {
				t.Errorf("got error %v, want %v", err, test.expectedError)
				return
			} else if err != nil {
				return
			}

			if link.Flags != test.expectedFlags {
				t.Errorf("got Flags %v, want %v", link.Flags, test.expectedFlags)
			}

			if link.Group != test.expectedGroup {
				t.Errorf("got Group %v, want %v", link.Group, test.expectedGroup)
			}

			if link.Address != test.expectedAddress {
				t.Errorf("got Address %v, want %v", link.Address, test.expectedAddress)
			}

			if link.Data != test.expectedData {
				t.Errorf("got Data %v, want %v", link.Data, test.expectedData)
			}

			if link.String() != test.expectedString {
				t.Errorf("got String %q, want %q", link.String(), test.expectedString)
			}

			buf, _ := link.MarshalBinary()

			if !bytes.Equal(buf, test.input) {
				t.Errorf("expected %v got %v", test.input, buf)
			}
		})
	}
}

func TestLinkRecordMarshalText(t *testing.T) {
	tests := []struct {
		expectedString string
		expected       LinkRecord
		expectedErr    string
	}{
		{"UC        1 01.01.01   00 00 00", LinkRecord{0, RecordControlFlags(0xc0), Group(1), Address{1, 1, 1}, [3]byte{0, 0, 0}}, ""},
		{"UC        1 01.01.01   00 00", LinkRecord{}, "Expected 6 fields got 5"},
	}

	for _, test := range tests {
		t.Run(test.expectedString, func(t *testing.T) {
			if test.expectedErr == "" {
				buf, _ := test.expected.MarshalText()
				if !bytes.Equal([]byte(test.expectedString), buf) {
					t.Errorf("got %q, want %q", string(buf), test.expectedString)
				}
			}

			var linkRecord LinkRecord
			err := linkRecord.UnmarshalText([]byte(test.expectedString))
			if err == nil {
				if test.expectedErr != "" {
					t.Errorf("got error nil, want %q", test.expectedErr)
				} else if test.expected != linkRecord {
					t.Errorf("got LinkRecord %v, want %v", linkRecord, test.expected)
				}
			} else if test.expectedErr != err.Error() {
				t.Errorf("got error %q, want %q", err.Error(), test.expectedErr)
			}
		})
	}
}

func TestGroupUnmarshalText(t *testing.T) {
	tests := []struct {
		input       string
		expectedErr string
		expected    Group
	}{
		{"1", "", Group(1)},
		{"wxyz", "invalid number format", Group(0)},
		{"256", "valid groups are between 1 and 255 (inclusive)", Group(0)},
		{"-1", "valid groups are between 1 and 255 (inclusive)", Group(0)},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			var group Group
			err := group.UnmarshalText([]byte(test.input))
			if err == nil {
				if test.expectedErr != "" {
					t.Errorf("got error %q, want %q", err, test.expectedErr)
				} else if group != test.expected {
					t.Errorf("got Group %d, want %d", group, test.expected)
				}
			} else if test.expectedErr != err.Error() {
				t.Errorf("got error %q, want %q", err.Error(), test.expectedErr)
			}
		})
	}
}
