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

func init() {
	// turn off logging for tests
	Log.Level = LevelNone
}

func TestAddress(t *testing.T) {
	tests := []struct {
		input [3]byte
		str   string
	}{
		{[3]byte{0x47, 0x2d, 0x10}, "47.2d.10"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.input), func(t *testing.T) {
			address := Address(test.input)

			if address.String() != test.str {
				t.Errorf("got %q, want %q", address.String(), test.str)
			}
		})
	}
}

func TestProductKey(t *testing.T) {
	tests := []struct {
		input          [3]byte
		expectedString string
	}{
		{[3]byte{0x01, 0x02, 0x03}, "0x010203"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.input), func(t *testing.T) {
			key := ProductKey(test.input)
			if key.String() != test.expectedString {
				t.Errorf("got ProductKey %q, want %q", key.String(), test.expectedString)
			}
		})
	}
}

func TestDevCat(t *testing.T) {
	tests := []struct {
		input            [2]byte
		expectedDomain   Domain
		expectedCategory Category
		expectedString   string
	}{
		{[2]byte{0x01, 0x02}, Domain(0x01), Category(0x02), "01.02"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.input), func(t *testing.T) {
			devCat := DevCat(test.input)
			if devCat.Domain() != test.expectedDomain {
				t.Errorf("got Domain 0x%02x, want 0x%02x", devCat.Domain(), test.expectedDomain)
			}

			if devCat.Category() != test.expectedCategory {
				t.Errorf("got Category 0x%02x, want 0x%02x", devCat.Category(), test.expectedCategory)
			}

			if devCat.String() != test.expectedString {
				t.Errorf("got String %q, want %q", devCat.String(), test.expectedString)
			}
		})
	}
}

func TestDevCatMarshaling(t *testing.T) {
	tests := []struct {
		input          string
		expectedDevCat DevCat
		expectedJSON   string
		expectedError  string
	}{
		{"\"01.02\"", DevCat{1, 2}, "\"01.02\"", ""},
		{"\"01\"", DevCat{0, 0}, "", "Expected Scanf to parse 2 digits, got 1"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.input), func(t *testing.T) {
			var devCat DevCat
			err := devCat.UnmarshalJSON([]byte(test.input))
			if err == nil {
				if devCat != test.expectedDevCat {
					t.Errorf("got %q, want %q", devCat, test.expectedDevCat)
				} else {
					data, _ := devCat.MarshalJSON()
					if string(data) != test.expectedJSON {
						t.Errorf("got JSON %q, want %q", string(data), test.expectedJSON)
					}
				}
			} else if err.Error() != test.expectedError {
				t.Errorf("got error %v, want %v", err, test.expectedError)
			}
		})
	}
}

func TestProductDataMarshaling(t *testing.T) {
	tests := []struct {
		desc           string
		input          []byte
		expectedDevCat [2]byte
		expectedKey    [3]byte
		expectedError  error
	}{
		{"0 1 2 3...", []byte{0, 1, 2, 3, 4, 5, 255, 0, 0, 0, 0, 0, 0, 0}, [2]byte{4, 5}, [3]byte{1, 2, 3}, nil},
		{"too short", []byte{}, [2]byte{0, 0}, [3]byte{0, 0, 0}, ErrBufferTooShort},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			pd := &ProductData{}
			err := pd.UnmarshalBinary(test.input)
			if !IsError(err, test.expectedError) {
				t.Errorf("got error %v, want %v", err, test.expectedError)
			}

			if err == nil {
				if pd.Key != ProductKey(test.expectedKey) {
					t.Errorf("got ProductKey %x, want %x", pd.Key, test.expectedKey)
				}

				if pd.DevCat != DevCat(test.expectedDevCat) {
					t.Errorf("got DevCat %x, want %x", pd.DevCat, test.expectedDevCat)
				}

				buf, _ := pd.MarshalBinary()
				if !bytes.Equal(buf, test.input[0:7]) {
					t.Errorf("got MarshalBinary %x, want %x", buf, test.input[0:7])
				}
			}
		})
	}
}
