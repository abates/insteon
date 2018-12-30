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

import "testing"

func TestAddressUnmarshalText(t *testing.T) {
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
	}

	for i, test := range tests {
		address := Address{}
		err := address.UnmarshalText([]byte(test.input))
		if test.wantErr && err == nil {
			t.Errorf("tests[%d] expected failure", i)
		} else if !test.wantErr && err != nil {
			t.Errorf("tests[%d] failed: %v", i, err)
		}
		if test.want != address {
			t.Errorf("tests[%d] expected %q got %q", i, test.input, test.want)
		}
	}
}

func TestAddressString(t *testing.T) {
	tests := []struct {
		input Address
		want  string
	}{
		{Address{0, 1, 2}, "00.01.02"},
	}
	for i, test := range tests {
		got := test.input.String()
		if got != test.want {
			t.Errorf("tests[%d] %q.String(): expected %q got %q", i, test.input, test.want, got)
		}
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

	for i, test := range tests {
		var address Address
		err := address.UnmarshalJSON([]byte(test.input))
		if err == nil {
			if address != test.expectedAddress {
				t.Errorf("tests[%d] expected %q got %q", i, test.expectedAddress, address)
			} else {
				data, _ := address.MarshalJSON()
				if string(data) != test.expectedJSON {
					t.Errorf("tests[%d] expected %q got %q", i, test.expectedJSON, string(data))
				}
			}
		} else if err.Error() != test.expectedError {
			t.Errorf("tests[%d] expected error %v got %v", i, test.expectedError, err)
		}
	}
}
