package util

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/abates/insteon"
)

type testLinkable struct {
	links []*insteon.LinkRecord
}

func (tl *testLinkable) Address() insteon.Address {
	return insteon.Address{1, 2, 3}
}

func (tl *testLinkable) Links() ([]*insteon.LinkRecord, error) {
	return tl.links, nil
}
func (tl *testLinkable) WriteLink(int, *insteon.LinkRecord) error { return nil }
func (tl *testLinkable) WriteLinks(...*insteon.LinkRecord) error  { return nil }
func (tl *testLinkable) UpdateLinks(...*insteon.LinkRecord) error { return nil }
func (tl *testLinkable) EnterLinkingMode(insteon.Group) error     { return nil }
func (tl *testLinkable) EnterUnlinkingMode(insteon.Group) error   { return nil }
func (tl *testLinkable) ExitLinkingMode() error                   { return nil }

func TestFindDuplicateLinks(t *testing.T) {
	links := []*insteon.LinkRecord{
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address{1, 2, 3}},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address{1, 2, 3}},
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address{4, 5, 6}},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address{4, 5, 6}},
	}
	want := []*insteon.LinkRecord{}

	tl := &testLinkable{links: links}
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

func TestLinksToText(t *testing.T) {
	links := []*insteon.LinkRecord{
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address{1, 2, 3}},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address{1, 2, 3}},
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address{4, 5, 6}},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address{4, 5, 6}},
	}

	want := `UC        1 01.02.03   00 00 00
UR        1 01.02.03   00 00 00
UC        1 04.05.06   00 00 00
UR        1 04.05.06   00 00 00
`

	want = fmt.Sprintf("%s%s", linkTextHeader, want)
	got := LinksToText(links)
	if want != got {
		t.Errorf("Wanted %q got %q", want, got)
	}

	gotLinks, err := TextToLinks(got)
	if err == nil {
		if len(gotLinks) == len(links) {
			for i, wantLink := range links {
				if !wantLink.Equal(gotLinks[i]) {
					t.Errorf("Wanted link: %v got %v", wantLink, gotLinks[i])
				}
			}
		} else {
			t.Errorf("Expected %d links got %d", len(links), len(gotLinks))
		}
	} else {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestFindLinkRecord(t *testing.T) {
	links := []*insteon.LinkRecord{
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address{1, 2, 3}},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address{1, 2, 3}},
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address{4, 5, 6}},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address{4, 5, 6}},
	}
	tl := &testLinkable{links: links}

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
			got, _ := FindLinkRecord(tl, test.inputController, test.inputAddress, test.inputGroup)
			if test.want != got {
				t.Errorf("want link %v got %v", test.want, got)
			}
		})
	}
}
