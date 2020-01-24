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

func TestAddressSetUnmarshalText(t *testing.T) {
	tests := []struct {
		input   string
		want    Address
		wantErr bool
	}{
		{"01.02.03", Address{1, 2, 3}, false},
		{"a1.b2.c3", Address{0xA1, 0xB2, 0xC3}, false},
		{"D1.E2.F3", Address{0xD1, 0xE2, 0xF3}, false},
		{"a1b2c3", Address{0xA1, 0xB2, 0xC3}, false},
		{"D1E2F3", Address{0xD1, 0xE2, 0xF3}, false},
		{"abcd", Address{}, true},
		{"abcdefg", Address{}, true},
		{"01b.02.03", Address{}, true},
		{"vx.02.03", Address{}, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			address := Address{}
			err := address.Set(test.input)
			if test.wantErr && err == nil {
				t.Errorf("expected failure for UnmarshalText(%q)", test.input)
			} else if !test.wantErr && err != nil {
				t.Errorf("UnmarshalText(%q) failed: %v", test.input, err)
			}
			if test.want != address {
				t.Errorf("UnmarshalText(%q) got %q, want %q", test.input, address, test.want)
			}
		})
	}
}

func TestAddressString(t *testing.T) {
	tests := []struct {
		input Address
		want  string
	}{
		{Address{0, 1, 2}, "00.01.02"},
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
		{"\"01.02.03\"", Address{1, 2, 3}, "\"01.02.03\"", ""},
		{"\"01.02\"", Address{0, 0, 0}, "", "Expected Scanf to parse 3 digits, got 2"},
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

func TestAddressables(t *testing.T) {
	tests := []struct {
		desc   string
		device Addressable
		want   Address
	}{
		{"I1Device", newI1Device(&testConnection{addr: Address{1, 2, 3}}, 0), Address{1, 2, 3}},
		{"I2Device", newI2Device(&testConnection{addr: Address{1, 2, 3}}, 0), Address{1, 2, 3}},
		{"I2CsDevice", newI2CsDevice(&testConnection{addr: Address{1, 2, 3}}, 0), Address{1, 2, 3}},
		{"Switch", NewSwitch(DeviceInfo{}, &testConnection{addr: Address{1, 2, 3}}, 0), Address{1, 2, 3}},
		{"Dimmer", NewDimmer(DeviceInfo{}, &testConnection{addr: Address{1, 2, 3}}, 0), Address{1, 2, 3}},
		{"Connection", &connection{addr: Address{1, 2, 3}}, Address{1, 2, 3}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := test.device.Address()
			if test.want != got {
				t.Errorf("want %q got %q", test.want, got)
			}
		})
	}
}
