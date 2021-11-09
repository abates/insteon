package util

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/abates/insteon"
)

type testLinkable struct {
	links []insteon.LinkRecord
}

func (tl *testLinkable) Address() insteon.Address {
	return insteon.Address{1, 2, 3}
}

func (tl *testLinkable) Links() ([]insteon.LinkRecord, error) {
	return tl.links, nil
}
func (tl *testLinkable) WriteLink(int, *insteon.LinkRecord) error { return nil }
func (tl *testLinkable) WriteLinks(...insteon.LinkRecord) error   { return nil }
func (tl *testLinkable) UpdateLinks(...insteon.LinkRecord) error  { return nil }
func (tl *testLinkable) EnterLinkingMode(insteon.Group) error     { return nil }
func (tl *testLinkable) EnterUnlinkingMode(insteon.Group) error   { return nil }
func (tl *testLinkable) ExitLinkingMode() error                   { return nil }

func TestFindDuplicateLinks(t *testing.T) {
	links := []insteon.LinkRecord{
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address{1, 2, 3}},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address{1, 2, 3}},
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address{4, 5, 6}},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address{4, 5, 6}},
	}
	want := []insteon.LinkRecord{}

	tl := &testLinkable{links: links}
	got, _ := FindDuplicateLinks(tl)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("want duplicate links %v got %v", want, got)
	}

	// create a duplicate
	dup := links[0]
	want = append(want, dup)
	tl.links = append(tl.links, dup)

	got, _ = FindDuplicateLinks(tl)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("want duplicate links %v got %v", want, got)
	}
}

func TestLinksToText(t *testing.T) {
	links := []insteon.LinkRecord{
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address{1, 2, 3}},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address{1, 2, 3}},
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address{4, 5, 6}},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address{4, 5, 6}},
	}

	want := `#
# Lines beginning with a # are ignored
# DO NOT delete lines, this will cause the entries to
# shift up and then the last entry will be in the database twice
# To delete a record simply mark it 'Available' by changing the
# first letter of the Flags to 'A'
#
# Flags Group Address    Data
UC        1 01.02.03   00 00 00
UR        1 01.02.03   00 00 00
UC        1 04.05.06   00 00 00
UR        1 04.05.06   00 00 00
`

	got := LinksToText(links)
	if want != got {
		t.Errorf("Wanted %q\n got %q", want, got)
	}

	gotLinks, err := TextToLinks(got)
	if err == nil {
		if len(gotLinks) == len(links) {
			for i, wantLink := range links {
				if !wantLink.Equal(&gotLinks[i]) {
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
	links := []insteon.LinkRecord{
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
		want            insteon.LinkRecord
		wantErr         error
	}{
		{"found", true, insteon.Address{1, 2, 3}, 1, links[0], nil},
		{"not found", true, insteon.Address{7, 8, 9}, 1, insteon.LinkRecord{}, ErrLinkNotFound},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := FindLinkRecord(tl, test.inputController, test.inputAddress, test.inputGroup)
			if err == test.wantErr {
				if test.want != got {
					t.Errorf("want link %v got %v", test.want, got)
				}
			} else {
				t.Errorf("Wanted error %v got %v", test.wantErr, err)
			}
		})
	}
}

func TestPrintLinks(t *testing.T) {
	links := []insteon.LinkRecord{
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address{1, 2, 3}},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address{1, 2, 3}},
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address{4, 5, 6}},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address{4, 5, 6}},
	}

	want := `Link Database:
    No links defined
`
	out := &bytes.Buffer{}
	l := &testLinkable{}

	err := PrintLinkDatabase(out, l)
	if err == nil {
		got := string(out.Bytes())
		if want != got {
			t.Errorf("Wanted:\n%v\nGot:\n%v", want, got)
		}
	} else {
		t.Fatalf("Unexpected error: %v", err)
	}
	out.Reset()
	want = `Link Database:
    Flags Group Address    Data
    UC        1 01.02.03   00 00 00
    UR        1 01.02.03   00 00 00
    UC        1 04.05.06   00 00 00
    UR        1 04.05.06   00 00 00
`

	l.links = links
	err = PrintLinkDatabase(out, l)
	if err == nil {
		got := string(out.Bytes())
		if want != got {
			t.Errorf("Wanted:\n%v\nGot:\n%v", want, got)
		}
	} else {
		t.Fatalf("Unexpected error: %v", err)
	}
}
