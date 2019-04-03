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
	"reflect"
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

	for _, test := range tests {
		t.Run(fmt.Sprintf("%04x", test.input), func(t *testing.T) {
			addr := MemAddress(test.input)
			if addr.String() != test.expected {
				t.Errorf("got %v, want %v", addr.String(), test.expected)
			}
		})
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

	for _, test := range tests {
		t.Run(fmt.Sprintf("%02x", test.input), func(t *testing.T) {
			lrt := linkRequestType(test.input)
			if test.expected != lrt.String() {
				t.Errorf("got %v, want %v", lrt.String(), test.expected)
			}
		})
	}
}

func TestLinkRequestUnmarshalBinary(t *testing.T) {
	tests := []struct {
		desc    string
		input   []byte
		want    *linkRequest
		wantErr error
	}{
		{"Short Buffer", nil, nil, ErrBufferTooShort},
		{"Read Link",
			mkPayload(0x00, 0x00, 0x0f, 0xff, 0x01),
			&linkRequest{0x00, 0x0fff, 1, nil},
			nil,
		},
		{"Link Response",
			mkPayload(0x00, 0x01, 0x0f, 0xff, 0x00, 0xd0, 0x01, 1, 2, 3, 4, 5, 6),
			&linkRequest{0x01, 0x0fff, 0, &LinkRecord{memAddress: 0x0fff, Flags: 0xd0, Group: Group(1), Address: Address{1, 2, 3}, Data: [3]byte{4, 5, 6}}},
			nil,
		},
		{"Write Link",
			mkPayload(0x00, 0x02, 0x0f, 0xff, 1, 0xd0, 0x01, 1, 2, 3, 4, 5, 6),
			&linkRequest{0x02, 0x0fff, 1, &LinkRecord{memAddress: 0x0fff, Flags: 0xd0, Group: Group(1), Address: Address{1, 2, 3}, Data: [3]byte{4, 5, 6}}},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := &linkRequest{}
			gotErr := got.UnmarshalBinary(test.input)
			if !isError(gotErr, test.wantErr) {
				t.Errorf("want error %v got %v", test.wantErr, gotErr)
			} else if gotErr == nil {
				if !reflect.DeepEqual(got, test.want) {
					t.Errorf("want link %#v got %#v", test.want, got)
				}
			}
		})
	}
}

func TestLinkRequestMarshalBinary(t *testing.T) {
	tests := []struct {
		desc  string
		input *linkRequest
		want  []byte
	}{
		{"Read Link",
			&linkRequest{0x00, 0x0fff, 1, nil},
			mkPayload(0x00, 0x00, 0x0f, 0xff, 0x01),
		},
		{"Link Response",
			&linkRequest{0x01, 0x0fff, 1, &LinkRecord{Flags: 0xd0, Group: Group(1), Address: Address{1, 2, 3}, Data: [3]byte{4, 5, 6}}},
			mkPayload(0x00, 0x01, 0x0f, 0xff, 0x00, 0xd0, 0x01, 1, 2, 3, 4, 5, 6),
		},
		{"Write Link",
			&linkRequest{0x02, 0x0fff, 1, &LinkRecord{Flags: 0xd0, Group: Group(1), Address: Address{1, 2, 3}, Data: [3]byte{4, 5, 6}}},
			mkPayload(0x00, 0x02, 0x0f, 0xff, 0x08, 0xd0, 0x01, 1, 2, 3, 4, 5, 6),
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, _ := test.input.MarshalBinary()
			if !bytes.Equal(test.want, got) {
				t.Errorf("want %v bytes got %v", test.want, got)
			}
		})
	}
}
