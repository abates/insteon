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
	"time"
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
			&linkRequest{0x01, 0x0fff, 0, &LinkRecord{Flags: 0xd0, Group: Group(1), Address: Address{1, 2, 3}, Data: [3]byte{4, 5, 6}}},
			nil,
		},
		{"Write Link",
			mkPayload(0x00, 0x02, 0x0f, 0xff, 1, 0xd0, 0x01, 1, 2, 3, 4, 5, 6),
			&linkRequest{0x02, 0x0fff, 1, &LinkRecord{Flags: 0xd0, Group: Group(1), Address: Address{1, 2, 3}, Data: [3]byte{4, 5, 6}}},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := &linkRequest{}
			gotErr := got.UnmarshalBinary(test.input)
			if !IsError(gotErr, test.wantErr) {
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
		want    []*LinkRecord
		wantErr error
	}{
		{"not old", time.Now().Add(time.Hour), nil, nil},
		{"old 1", time.Now().Add(-1 * time.Hour), []*LinkRecord{ControllerLink(1, Address{1, 2, 3})}, nil},
		{"old 2", time.Now().Add(-1 * time.Hour), []*LinkRecord{ControllerLink(1, Address{4, 5, 6}), ControllerLink(1, Address{1, 2, 3})}, nil},
		{"timeout", time.Now().Add(-1 * time.Hour), nil, ErrReadTimeout},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Add high water mark
			links := append(test.want, &LinkRecord{})
			conn := &testConnection{acks: []*Message{{Command: CmdReadWriteALDB, Flags: StandardDirectAck}}}
			memAddress := BaseLinkDBAddress
			for _, link := range links {
				lr := &linkRequest{Type: linkResponse, MemAddress: memAddress, Link: link}
				msg := &Message{Command: CmdReadWriteALDB, Flags: ExtendedDirectMessage, Payload: make([]byte, 14)}
				buf, _ := lr.MarshalBinary()
				copy(msg.Payload, buf)
				conn.recv = append(conn.recv, msg)
				memAddress -= LinkRecordSize
			}

			MaxLinkDbAge = time.Millisecond
			linkdb := linkdb{age: test.age, dialer: testDeviceDialer{conn}}
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
		links          []*LinkRecord
		inputIndex     int
		inputRecord    *LinkRecord
		wantLinksSize  int
		wantMemAddress MemAddress
		wantErr        error
	}{
		{"Invalid Index", nil, 1, nil, 0, BaseLinkDBAddress, ErrLinkIndexOutOfRange},
		{"Base Address", nil, 0, ControllerLink(1, Address{1, 2, 3}), 1, BaseLinkDBAddress, nil},
		{"Truncate existing links", []*LinkRecord{ControllerLink(1, Address{1, 2, 3}), ResponderLink(1, Address{1, 2, 3}), ControllerLink(1, Address{4, 5, 6})}, 2, &LinkRecord{Flags: 0xfc}, 2, BaseLinkDBAddress - LinkRecordSize*2, nil},
		{"Replace existing link", []*LinkRecord{ControllerLink(1, Address{1, 2, 3}), ResponderLink(1, Address{1, 2, 3}), ControllerLink(1, Address{4, 5, 6})}, 1, ResponderLink(43, Address{11, 12, 13}), 3, BaseLinkDBAddress - LinkRecordSize, nil},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			conn := &testConnection{acks: []*Message{TestAck}}
			linkdb := linkdb{dialer: testDeviceDialer{conn}, links: test.links}
			gotErr := linkdb.writeLink(test.inputIndex, test.inputRecord)
			if test.wantErr != gotErr {
				t.Errorf("Want err %v got %v", test.wantErr, gotErr)
			} else if gotErr == nil {
				lr := &linkRequest{}
				lr.UnmarshalBinary(conn.sent[0].Payload)
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
					if *linkdb.links[test.inputIndex] != *test.inputRecord {
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
		input          []*LinkRecord
		wantMemAddress []MemAddress
	}{
		{"nothing", []*LinkRecord{}, []MemAddress{BaseLinkDBAddress}},
		{"one link", []*LinkRecord{ControllerLink(1, Address{1, 2, 3})}, []MemAddress{BaseLinkDBAddress, BaseLinkDBAddress - LinkRecordSize}},
		{"two links", []*LinkRecord{ControllerLink(1, Address{1, 2, 3}), ControllerLink(1, Address{4, 5, 6})}, []MemAddress{BaseLinkDBAddress, BaseLinkDBAddress - LinkRecordSize, BaseLinkDBAddress - LinkRecordSize*2}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			conn := &testConnection{}
			for i := 0; i < len(test.wantMemAddress); i++ {
				conn.acks = append(conn.acks, TestAck)
			}
			linkdb := linkdb{dialer: testDeviceDialer{conn}}
			linkdb.WriteLinks(test.input...)
			gotMemAddress := []MemAddress{}
			gotLinks := []*LinkRecord{}
			for _, msg := range conn.sent {
				lr := &linkRequest{}
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
					if !test.input[i].Equal(link) {
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
		existingLinks []*LinkRecord
		input         []*LinkRecord
		want          []MemAddress
	}{
		{
			"no existing links",
			nil,
			[]*LinkRecord{ControllerLink(1, Address{1, 2, 3})},
			[]MemAddress{BaseLinkDBAddress, BaseLinkDBAddress - LinkRecordSize},
		},
		{
			"duplicate links",
			[]*LinkRecord{ControllerLink(1, Address{1, 2, 3})},
			[]*LinkRecord{ControllerLink(1, Address{1, 2, 3})},
			nil,
		},
		{
			"duplicate link (update flags)",
			[]*LinkRecord{{AvailableController, 1, Address{1, 2, 3}, [3]byte{}}},
			[]*LinkRecord{ControllerLink(1, Address{1, 2, 3})},
			[]MemAddress{BaseLinkDBAddress},
		},
		{
			"available and append links",
			[]*LinkRecord{{AvailableController, 1, Address{1, 2, 3}, [3]byte{}}, ControllerLink(1, Address{4, 5, 6})},
			[]*LinkRecord{ControllerLink(1, Address{6, 7, 8}), ResponderLink(1, Address{5, 6, 7})},
			[]MemAddress{BaseLinkDBAddress, BaseLinkDBAddress - 2*LinkRecordSize, BaseLinkDBAddress - 3*LinkRecordSize},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			conn := &testConnection{}
			for i := 0; i < len(test.want); i++ {
				conn.acks = append(conn.acks, TestAck)
			}
			index := make(map[LinkID]int)
			for i, link := range test.existingLinks {
				index[link.id()] = i
			}
			MaxLinkDbAge = time.Second
			ldb := &linkdb{age: time.Now().Add(time.Hour), links: test.existingLinks, index: index, dialer: testDeviceDialer{conn}}
			ldb.UpdateLinks(test.input...)

			for i, msg := range conn.sent {
				want := test.want[i]
				lr := &linkRequest{}
				lr.UnmarshalBinary(msg.Payload)
				got := lr.MemAddress
				if want != got {
					t.Errorf("Want link mem address %v got %v", want, got)
				}
				i++
			}

			if len(conn.sent) < len(test.want) {
				t.Errorf("Wanted %d addresses got %d", len(test.want), len(conn.sent))
			}
		})
	}
}
