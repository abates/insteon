package util

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/abates/insteon"
)

const (
	Links              = "Links() ([]*insteon.LinkRecord, error)"
	RemoveLinks        = "RemoveLinks(...*insteon.LinkRecord) error"
	WriteLink          = "WriteLink(*insteon.LinkRecord) error"
	AddLink            = "AddLink(*insteon.LinkRecord) error"
	AppendLink         = "AppendLink(*insteon.LinkRecord) error"
	EnterLinkingMode   = "EnterLinkingMode(insteon.Group) error"
	EnterUnlinkingMode = "EnterUnlinkingMode(insteon.Group) error"
	ExitLinkingMode    = "ExitLinkingMode() error"
)

func msc(f string) string { return fmt.Sprintf("controller %v", f) }
func msr(f string) string { return fmt.Sprintf("responder %v", f) }

type testLinkable struct {
	name     string
	addr     insteon.Address
	links    []*insteon.LinkRecord
	sequence chan string
}

func (tl *testLinkable) run(f string) {
	if cap(tl.sequence) > 1 && cap(tl.sequence) == len(tl.sequence) {
		<-tl.sequence
	}

	select {
	case tl.sequence <- fmt.Sprintf("%s %s", tl.name, f):
	default:
	}
}

func (tl *testLinkable) Address() insteon.Address { return tl.addr }
func (tl *testLinkable) Links() ([]*insteon.LinkRecord, error) {
	tl.run(Links)
	return tl.links, nil
}

func (tl *testLinkable) RemoveLinks(...*insteon.LinkRecord) error {
	tl.run(RemoveLinks)
	return nil
}

func (tl *testLinkable) WriteLink(*insteon.LinkRecord) error {
	tl.run(WriteLink)
	return nil
}

func (tl *testLinkable) AddLink(*insteon.LinkRecord) error {
	tl.run(AddLink)
	return nil
}

func (tl *testLinkable) AppendLink(*insteon.LinkRecord) error {
	tl.run(AppendLink)
	return nil
}

func (tl *testLinkable) EnterLinkingMode(insteon.Group) error {
	tl.run(EnterLinkingMode)
	return nil
}

func (tl *testLinkable) EnterUnlinkingMode(insteon.Group) error {
	tl.run(EnterUnlinkingMode)
	return nil
}

func (tl *testLinkable) ExitLinkingMode() error {
	tl.run(ExitLinkingMode)
	return nil
}

func TestFindDuplicateLinks(t *testing.T) {
	links := []*insteon.LinkRecord{
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address{1, 2, 3}},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address{1, 2, 3}},
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address{4, 5, 6}},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address{4, 5, 6}},
	}
	want := []*insteon.LinkRecord{}

	tl := &testLinkable{links: links, sequence: make(chan string, 2)}
	got, _ := FindDuplicateLinks(tl)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("want duplicate links %v got %v", want, got)
	}

	// create a duplicate
	dup := &insteon.LinkRecord{}
	*dup = *links[0]
	want = append(want, dup)
	tl.links = append(tl.links, dup)

	got, _ = FindDuplicateLinks(tl)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("want duplicate links %v got %v", want, got)
	}
}

func TestFindLinkRecord(t *testing.T) {
	links := []*insteon.LinkRecord{
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address{1, 2, 3}},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address{1, 2, 3}},
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address{4, 5, 6}},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address{4, 5, 6}},
	}

	tests := []struct {
		desc            string
		inputController bool
		inputAddress    insteon.Address
		inputGroup      insteon.Group
		want            *insteon.LinkRecord
	}{
		{"found", true, insteon.Address{1, 2, 3}, 1, links[0]},
		{"not found", true, insteon.Address{7, 8, 9}, 1, nil},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			tl := &testLinkable{links: links, sequence: make(chan string, 1)}
			got, _ := FindLinkRecord(tl, test.inputController, test.inputAddress, test.inputGroup)
			if test.want != got {
				t.Errorf("want link %v got %v", test.want, got)
			}
		})
	}
}

func TestForceLink(t *testing.T) {
	want := []string{
		msc(EnterLinkingMode),
		msr(EnterLinkingMode),
		msr(ExitLinkingMode),
		msc(ExitLinkingMode),
	}

	sequence := make(chan string, len(want))
	controller := &testLinkable{name: "controller", sequence: sequence}
	responder := &testLinkable{name: "responder", sequence: sequence}
	ForceLink(1, controller, responder)
	close(sequence)
	got := []string{}
	for s := range sequence {
		got = append(got, s)
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("want sequence %v got %v", want, got)
	}
}

func TestLink(t *testing.T) {
	tests := []struct {
		desc            string
		controllerLinks []*insteon.LinkRecord
		responderLinks  []*insteon.LinkRecord
		wantSequence    []string
		wantErr         error
	}{
		{"no links", nil, nil, []string{msc(EnterLinkingMode), msr(EnterLinkingMode), msr(ExitLinkingMode), msc(ExitLinkingMode)}, nil},
		{
			"responder link",
			nil,
			[]*insteon.LinkRecord{{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address{1, 2, 3}}},
			[]string{msr(RemoveLinks), msc(EnterLinkingMode), msr(EnterLinkingMode), msr(ExitLinkingMode), msc(ExitLinkingMode)},
			nil,
		},
		{
			"controller link",
			[]*insteon.LinkRecord{{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address{4, 5, 6}}},
			nil,
			[]string{msc(RemoveLinks), msc(EnterLinkingMode), msr(EnterLinkingMode), msr(ExitLinkingMode), msc(ExitLinkingMode)},
			nil,
		},
		{
			"both links",
			[]*insteon.LinkRecord{{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address{4, 5, 6}}},
			[]*insteon.LinkRecord{{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address{1, 2, 3}}},
			[]string{msc(Links), msr(Links)},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			sequence := make(chan string, len(test.wantSequence))
			controller := &testLinkable{addr: insteon.Address{1, 2, 3}, name: "controller", sequence: sequence, links: test.controllerLinks}
			responder := &testLinkable{addr: insteon.Address{4, 5, 6}, name: "responder", sequence: sequence, links: test.responderLinks}
			err := Link(1, controller, responder)
			if err != test.wantErr {
				t.Errorf("want error %v got %v", test.wantErr, err)
			} else if err == nil {
				close(sequence)
				gotSequence := []string{}
				for s := range sequence {
					gotSequence = append(gotSequence, s)
				}

				if !reflect.DeepEqual(test.wantSequence, gotSequence) {
					t.Errorf("want sequence %v got %v", test.wantSequence, gotSequence)
				}
			}
		})
	}
}

func TestPrintLinks(t *testing.T) {
	tests := []struct {
		desc  string
		links []*insteon.LinkRecord
		want  string
	}{
		{"no links", nil, "Link Database:\n    No links defined\n"},
		{
			"one link",
			[]*insteon.LinkRecord{{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address{1, 2, 3}, Data: [3]byte{0xa, 0xb, 0xc}}},
			"Link Database:\n    Flags Group Address    Data\n    UC        1 01.02.03   0a 0b 0c\n",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			tl := &testLinkable{links: test.links}
			str := &strings.Builder{}
			PrintLinks(str, tl)
			got := str.String()
			if test.want != got {
				t.Errorf("want string %q got %q", test.want, got)
			}
		})
	}
}
