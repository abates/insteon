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
		correct bool
	}{
		{"01.02.03", true},
		{"abcd", false},
		{"01b.02.03", false},
	}

	for i, test := range tests {
		address := Address{}
		err := address.UnmarshalText([]byte(test.input))
		if err == nil && !test.correct {
			t.Errorf("tests[%d] expected failure", i)
		} else if err != nil && test.correct {
			t.Errorf("tests[%d] failed: %v", i, err)
		}

		if test.correct && test.input != address.String() {
			t.Errorf("tests[%d] expected %q got %q", i, test.input, address.String())
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
