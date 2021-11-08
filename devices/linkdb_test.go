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

package devices

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/abates/insteon"
	"github.com/abates/insteon/commands"
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
			lrt := LinkRequestType(test.input)
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
		want    *LinkRequest
		wantErr error
	}{
		{"Short Buffer", nil, nil, insteon.ErrBufferTooShort},
		{"Read Link",
			mkPayload(0x00, 0x00, 0x0f, 0xff, 0x01),
			&LinkRequest{0x00, 0x0fff, 1, nil},
			nil,
		},
		{"Link Response",
			mkPayload(0x00, 0x01, 0x0f, 0xff, 0x00, 0xd0, 0x01, 1, 2, 3, 4, 5, 6),
			&LinkRequest{0x01, 0x0fff, 0, &insteon.LinkRecord{Flags: 0xd0, Group: insteon.Group(1), Address: insteon.Address{1, 2, 3}, Data: [3]byte{4, 5, 6}}},
			nil,
		},
		{"Write Link",
			mkPayload(0x00, 0x02, 0x0f, 0xff, 1, 0xd0, 0x01, 1, 2, 3, 4, 5, 6),
			&LinkRequest{0x02, 0x0fff, 1, &insteon.LinkRecord{Flags: 0xd0, Group: insteon.Group(1), Address: insteon.Address{1, 2, 3}, Data: [3]byte{4, 5, 6}}},
			nil,
		},
		// this message caused an error, so I wrote a test for it
		// 1:0 46.d1.b2 -> 49.f7.51 Read/Write ALDB(7) [ff f9 e7 8b 20 aa 01 b6 d5 1a 07 e3 fd b8]
		{
			desc:    "",
			input:   []byte{0xff, 0xf9, 0xe7, 0x8b, 0x20, 0xaa, 0x01, 0xb6, 0xd5, 0x1a, 0x07, 0xe3, 0xfd, 0xb8},
			want:    &LinkRequest{},
			wantErr: ErrInvalidResponse,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := &LinkRequest{}
			gotErr := got.UnmarshalBinary(test.input)
			if !errors.Is(gotErr, test.wantErr) {
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
		input *LinkRequest
		want  []byte
	}{
		{"Read Link",
			&LinkRequest{0x00, 0x0fff, 1, nil},
			mkPayload(0x00, 0x00, 0x0f, 0xff, 0x01),
		},
		{"Link Response",
			&LinkRequest{0x01, 0x0fff, 1, &insteon.LinkRecord{Flags: 0xd0, Group: insteon.Group(1), Address: insteon.Address{1, 2, 3}, Data: [3]byte{4, 5, 6}}},
			mkPayload(0x00, 0x01, 0x0f, 0xff, 0x00, 0xd0, 0x01, 1, 2, 3, 4, 5, 6),
		},
		{"Write Link",
			&LinkRequest{0x02, 0x0fff, 1, &insteon.LinkRecord{Flags: 0xd0, Group: insteon.Group(1), Address: insteon.Address{1, 2, 3}, Data: [3]byte{4, 5, 6}}},
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

func TestLinkDbOld(t *testing.T) {
	MaxLinkDbAge = time.Second
	ldb := linkdb{}
	if !ldb.old() {
		t.Errorf("Expected ldb to report old")
	}

	ldb.age = time.Now().Add(10 * time.Second)
	if ldb.old() {
		t.Errorf("Expected ldb to report not old")
	}
}

func TestLinkdbLinks(t *testing.T) {
	tests := []struct {
		desc    string
		age     time.Time
		want    []*insteon.LinkRecord
		wantErr error
	}{
		{"not old", time.Now().Add(time.Hour), nil, nil},
		{"old 1", time.Now().Add(-1 * time.Hour), []*insteon.LinkRecord{insteon.ControllerLink(1, insteon.Address{1, 2, 3})}, nil},
		{"old 2", time.Now().Add(-1 * time.Hour), []*insteon.LinkRecord{insteon.ControllerLink(1, insteon.Address{4, 5, 6}), insteon.ControllerLink(1, insteon.Address{1, 2, 3})}, nil},
		{"timeout", time.Now().Add(-1 * time.Hour), nil, ErrReadTimeout},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Add high water mark
			links := append(test.want, &insteon.LinkRecord{})
			tw := &testWriter{}

			// Test ignoring acks on receive channel
			tw.acks = append(tw.acks, &insteon.Message{Command: commands.ReadWriteALDB, Flags: insteon.StandardDirectAck})
			tw.acks = append(tw.acks, &insteon.Message{Command: commands.ReadWriteALDB, Flags: insteon.StandardDirectNak})
			memAddress := BaseLinkDBAddress
			for _, link := range links {
				lr := &LinkRequest{Type: linkResponse, MemAddress: memAddress, Link: link}
				msg := &insteon.Message{Command: commands.ReadWriteALDB, Flags: insteon.ExtendedDirectMessage, Payload: make([]byte, 14)}
				buf, _ := lr.MarshalBinary()
				copy(msg.Payload, buf)
				tw.read = append(tw.read, msg)
				memAddress -= LinkRecordSize
			}

			MaxLinkDbAge = time.Millisecond
			linkdb := linkdb{age: test.age, MessageWriter: tw}
			got, err := linkdb.Links()
			if err == nil {
				if len(got) == len(test.want) {
				} else {
					t.Errorf("wanted %d links got %d", len(test.want), len(got))
				}
			} else {
				t.Errorf("Unexpected error %v", err)
			}
		})
	}
}

func TestLinkdbWriteLink(t *testing.T) {
	tests := []struct {
		desc           string
		links          []insteon.LinkRecord
		inputIndex     int
		inputRecord    *insteon.LinkRecord
		wantLinksSize  int
		wantMemAddress MemAddress
		wantErr        error
	}{
		{"Invalid Index", nil, 1, nil, 0, BaseLinkDBAddress, ErrLinkIndexOutOfRange},
		{"Base Address", nil, 0, insteon.ControllerLink(1, insteon.Address{1, 2, 3}), 1, BaseLinkDBAddress, nil},
		{"Truncate existing links", []insteon.LinkRecord{*insteon.ControllerLink(1, insteon.Address{1, 2, 3}), *insteon.ResponderLink(1, insteon.Address{1, 2, 3}), *insteon.ControllerLink(1, insteon.Address{4, 5, 6})}, 2, &insteon.LinkRecord{Flags: 0xfc}, 2, BaseLinkDBAddress - LinkRecordSize*2, nil},
		{"Replace existing link", []insteon.LinkRecord{*insteon.ControllerLink(1, insteon.Address{1, 2, 3}), *insteon.ResponderLink(1, insteon.Address{1, 2, 3}), *insteon.ControllerLink(1, insteon.Address{4, 5, 6})}, 1, insteon.ResponderLink(43, insteon.Address{11, 12, 13}), 3, BaseLinkDBAddress - LinkRecordSize, nil},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			tw := &testWriter{}
			linkdb := linkdb{MessageWriter: tw, links: test.links}
			gotErr := linkdb.writeLink(test.inputIndex, test.inputRecord)
			if test.wantErr != gotErr {
				t.Errorf("Want err %v got %v", test.wantErr, gotErr)
			} else if gotErr == nil {
				lr := &LinkRequest{}
				lr.UnmarshalBinary(tw.written[0].Payload)
				if lr.Type != writeLink {
					t.Errorf("Expected %v got %v", writeLink, lr.Type)
				}
				gotMemAddress := lr.MemAddress
				if test.wantMemAddress != gotMemAddress {
					t.Errorf("Want memory address %v got %v", test.wantMemAddress, gotMemAddress)
				}

				if test.wantLinksSize != len(linkdb.links) {
					t.Errorf("Wanted %d links got %d", test.wantLinksSize, len(linkdb.links))
				}

				if test.inputIndex < test.wantLinksSize {
					if linkdb.links[test.inputIndex] != *test.inputRecord {
						t.Errorf("Wanted link %+v got %+v", test.inputRecord, linkdb.links[test.inputIndex])
					}
				}
			}
		})
	}
}

func TestLinkdbWriteLinks(t *testing.T) {
	tests := []struct {
		desc           string
		input          []insteon.LinkRecord
		wantMemAddress []MemAddress
	}{
		{"nothing", []insteon.LinkRecord{}, []MemAddress{BaseLinkDBAddress}},
		{"one link", []insteon.LinkRecord{*insteon.ControllerLink(1, insteon.Address{1, 2, 3})}, []MemAddress{BaseLinkDBAddress, BaseLinkDBAddress - LinkRecordSize}},
		{"two links", []insteon.LinkRecord{*insteon.ControllerLink(1, insteon.Address{1, 2, 3}), *insteon.ControllerLink(1, insteon.Address{4, 5, 6})}, []MemAddress{BaseLinkDBAddress, BaseLinkDBAddress - LinkRecordSize, BaseLinkDBAddress - LinkRecordSize*2}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			tw := &testWriter{}
			linkdb := linkdb{MessageWriter: tw}
			linkdb.WriteLinks(test.input...)
			gotMemAddress := []MemAddress{}
			gotLinks := []*insteon.LinkRecord{}
			for _, msg := range tw.written {
				lr := &LinkRequest{}
				lr.UnmarshalBinary(msg.Payload)
				gotLinks = append(gotLinks, lr.Link)
				gotMemAddress = append(gotMemAddress, lr.MemAddress)
			}

			if !reflect.DeepEqual(test.wantMemAddress, gotMemAddress) {
				t.Errorf("Want memory addresses %v got %v", test.wantMemAddress, gotMemAddress)
			}

			if time.Now().After(linkdb.age.Add(10 * time.Second)) {
				t.Errorf("Wanted age to be updated")
			}

			if len(linkdb.links) != len(test.input) {
				t.Errorf("Expected links to be set")
			} else {
				for i, link := range linkdb.links {
					if !test.input[i].Equal(&link) {
						t.Errorf("Expected %+v got %+v", test.input[i], link)
					}
				}

				link := gotLinks[len(gotLinks)-1]
				if !link.Flags.LastRecord() {
					t.Errorf("Expected last link request to include a last link record")
				}
			}
		})
	}
}

func TestLinkdbUpdateLinks(t *testing.T) {
	tests := []struct {
		desc          string
		existingLinks []insteon.LinkRecord
		input         []insteon.LinkRecord
		want          []MemAddress
	}{
		{
			"no existing links",
			nil,
			[]insteon.LinkRecord{*insteon.ControllerLink(1, insteon.Address{1, 2, 3})},
			[]MemAddress{BaseLinkDBAddress, BaseLinkDBAddress - LinkRecordSize},
		},
		{
			"duplicate links",
			[]insteon.LinkRecord{*insteon.ControllerLink(1, insteon.Address{1, 2, 3})},
			[]insteon.LinkRecord{*insteon.ControllerLink(1, insteon.Address{1, 2, 3})},
			nil,
		},
		{
			"duplicate link (update flags)",
			[]insteon.LinkRecord{{insteon.AvailableController, 1, insteon.Address{1, 2, 3}, [3]byte{}}},
			[]insteon.LinkRecord{*insteon.ControllerLink(1, insteon.Address{1, 2, 3})},
			[]MemAddress{BaseLinkDBAddress},
		},
		{
			"available and append links",
			[]insteon.LinkRecord{{insteon.AvailableController, 1, insteon.Address{1, 2, 3}, [3]byte{}}, *insteon.ControllerLink(1, insteon.Address{4, 5, 6})},
			[]insteon.LinkRecord{*insteon.ControllerLink(1, insteon.Address{6, 7, 8}), *insteon.ResponderLink(1, insteon.Address{5, 6, 7})},
			[]MemAddress{BaseLinkDBAddress, BaseLinkDBAddress - 2*LinkRecordSize, BaseLinkDBAddress - 3*LinkRecordSize},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			tw := &testWriter{}
			index := make(map[insteon.LinkID]int)
			for i, link := range test.existingLinks {
				index[link.ID()] = i
			}
			MaxLinkDbAge = time.Second
			ldb := &linkdb{age: time.Now().Add(time.Hour), links: test.existingLinks, index: index, MessageWriter: tw}
			ldb.UpdateLinks(test.input...)

			for i, msg := range tw.written {
				want := test.want[i]
				lr := &LinkRequest{}
				lr.UnmarshalBinary(msg.Payload)
				got := lr.MemAddress
				if want != got {
					t.Errorf("Want link mem address %v got %v", want, got)
				}
				i++
			}

			if len(tw.written) < len(test.want) {
				t.Errorf("Wanted %d addresses got %d", len(test.want), len(tw.written))
			}
		})
	}
}
