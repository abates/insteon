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

func TestMemAddress(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    int
		expected string
	}{
		{0xffff, "ff.ff"},
		{0x0fff, "0f.ff"},
		{0x0f00, "0f.00"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%04x", test.input), func(t *testing.T) {
			t.Parallel()
			addr := MemAddress(test.input)
			if addr.String() != test.expected {
				t.Errorf("got %v, want %v", addr.String(), test.expected)
			}
		})
	}
}

func TestLinkRequestType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    byte
		expected string
	}{
		{0x00, "Link Read"},
		{0x01, "Link Resp"},
		{0x02, "Link Write"},
		{0x03, "Unknown"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%02x", test.input), func(t *testing.T) {
			t.Parallel()
			lrt := LinkRequestType(test.input)
			if test.expected != lrt.String() {
				t.Errorf("got %v, want %v", lrt.String(), test.expected)
			}
		})
	}
}

func TestLinkRequest(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc            string
		input           []byte
		marshal         []byte
		expectedType    LinkRequestType
		expectedAddress MemAddress
		expectedRecords int
		expectedString  string
		expectedError   error
	}{
		{
			desc:          "error buffer too short",
			input:         []byte{},
			expectedError: ErrBufferTooShort,
		},
		{
			desc:            "success",
			input:           []byte{0xff, 0x00, 0x0f, 0xff, 0x08},
			marshal:         []byte{0x0, 0x00, 0x0f, 0xff, 0x08, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedType:    LinkRequestType(0x00),
			expectedAddress: MemAddress(0x0fff),
			expectedRecords: 8,
			expectedString:  "Link Read 0f.ff 8",
		},
		{
			desc:            "error end of links 1",
			input:           []byte{0xff, 0x01, 0x0f, 0xff, 0x08, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			marshal:         []byte{0x00, 0x01, 0x0f, 0xff, 0x00, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedType:    LinkRequestType(0x01),
			expectedAddress: MemAddress(0x0fff),
			expectedRecords: 0,
			expectedString:  "Link Resp 0f.ff 0",
			expectedError:   ErrEndOfLinks,
		},
		{
			desc:            "error end of links 2",
			input:           []byte{0xff, 0x02, 0x0f, 0xff, 0x08, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			marshal:         []byte{0x00, 0x02, 0x0f, 0xff, 0x08, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedType:    LinkRequestType(0x02),
			expectedAddress: MemAddress(0x0fff),
			expectedRecords: 8,
			expectedString:  "Link Write 0f.ff 8",
			expectedError:   ErrEndOfLinks,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			linkRequest := &LinkRequest{}
			err := linkRequest.UnmarshalBinary(test.input)
			if !isError(err, test.expectedError) {
				t.Errorf("got error %v, want %v", err, test.expectedError)
				return
			} else if err != nil {
				return
			}

			if linkRequest.Type != test.expectedType {
				t.Errorf("got Type %v, want %v", linkRequest.Type, test.expectedType)
			}

			if linkRequest.MemAddress != test.expectedAddress {
				t.Errorf("got MemAddress %v, want %v", linkRequest.MemAddress, test.expectedAddress)
			}

			if linkRequest.NumRecords != test.expectedRecords {
				t.Errorf("got NumRecords %v, want %v", linkRequest.NumRecords, test.expectedRecords)
			}

			if linkRequest.String()[0:len(test.expectedString)] != test.expectedString {
				t.Errorf("got String %q, want %q", linkRequest.String()[0:len(test.expectedString)], test.expectedString)
			}

			buf, _ := linkRequest.MarshalBinary()
			if !bytes.Equal(test.marshal, buf) {
				t.Errorf("got MarshalBinary %v, want %v", buf, test.marshal)
			}
		})
	}
}

func TestAddLink(t *testing.T) {
	t.Parallel()
}

func TestRemoveLink(t *testing.T) {
	t.Parallel()
}

func TestCleanup(t *testing.T) {
	t.Parallel()
}
