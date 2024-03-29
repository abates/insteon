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
	"fmt"
	"testing"
)

func TestAddressMarshalUnmarshalText(t *testing.T) {
	tests := []struct {
		input       string
		want        Address
		wantMarshal string
		wantErr     bool
	}{
		{"01.02.03", Address(0x010203), "01.02.03", false},
		{"a1.b2.c3", Address(0xA1B2C3), "a1.b2.c3", false},
		{"D1.E2.F3", Address(0xD1E2F3), "d1.e2.f3", false},
		{"a1b2c3", Address(0xA1B2C3), "a1.b2.c3", false},
		{"D1E2F3", Address(0xD1E2F3), "d1.e2.f3", false},
		{"abcd", Address(0), "00.00.00", true},
		{"abcdefg", Address(0), "00.00.00", true},
		{"01b.02.03", Address(0), "00.00.00", true},
		{"vx.02.03", Address(0), "00.00.00", true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			address := Address(0)
			err := address.Set(test.input)
			if test.wantErr && err == nil {
				t.Errorf("expected failure for UnmarshalText(%q)", test.input)
			} else if !test.wantErr && err != nil {
				t.Errorf("UnmarshalText(%q) failed: %v", test.input, err)
			}
			if test.want != address {
				t.Errorf("UnmarshalText(%q) got %q, want %q", test.input, address, test.want)
			}
			if err == nil {
				got, _ := address.MarshalText()
				if test.wantMarshal != string(got) {
					t.Errorf("Wanted marshaled text %v got %v", test.wantMarshal, string(got))
				}
			}
		})
	}
}

func TestAddressString(t *testing.T) {
	tests := []struct {
		input Address
		want  string
	}{
		{Address(0x000102), "00.01.02"},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.input), func(t *testing.T) {
			got := test.input.String()
			if got != test.want {
				t.Errorf("%q.String(): got %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestAddressMarshaling(t *testing.T) {
	tests := []struct {
		input           string
		expectedAddress Address
		expectedJSON    string
		expectedError   string
	}{
		{"\"01.02.03\"", Address(0x010203), "\"01.02.03\"", ""},
		{"\"01.02\"", Address(0), "", "address format is xx.xx.xx (digits in hex)"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			var address Address
			err := address.UnmarshalJSON([]byte(test.input))
			if err == nil {
				if address != test.expectedAddress {
					t.Errorf("address.UnmarshalJSON(%q) got %+v, want %+v", test.input, address, test.expectedAddress)
				} else {
					data, _ := address.MarshalJSON()
					if string(data) != test.expectedJSON {
						t.Errorf("address.UnmarshalJSON(%q) got %+v, want %+v", test.input, string(data), test.expectedJSON)
					}
				}
			} else if err.Error() != test.expectedError {
				t.Errorf("address.UnmarshalJSON(%q) error got %q expected %q", test.input, err, test.expectedError)
			}
		})
	}
}
