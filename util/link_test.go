package util

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/abates/insteon"
	"github.com/abates/insteon/devices"
)

type testLinkable struct {
	links    []insteon.LinkRecord
	name     string
	commands chan<- string
}

type testWritableLinkable struct {
	*testLinkable
}

func (twl *testWritableLinkable) WriteLink(i int, link insteon.LinkRecord) error {
	if i < len(twl.links) {
		twl.links[i] = link
	} else if i == len(twl.links) {
		twl.links = append(twl.links, link)
	} else {
		return fmt.Errorf("Index out of range have %d need <= %d", i, len(twl.links))
	}
	return nil
}

func (tl *testLinkable) Address() insteon.Address {
	return insteon.Address(0x010203)
}

func (tl *testLinkable) Links() ([]insteon.LinkRecord, error) {
	return tl.links, nil
}

func (tl *testLinkable) WriteLinks(links ...insteon.LinkRecord) error {
	tl.links = make([]insteon.LinkRecord, len(links))
	copy(tl.links, links)
	return nil
}

func (tl *testLinkable) UpdateLinks(...insteon.LinkRecord) error { return nil }

func (tl *testLinkable) EnterLinkingMode(insteon.Group) error {
	tl.commands <- fmt.Sprintf("%s EnterLinkingMode", tl.name)
	return nil
}

func (tl *testLinkable) EnterUnlinkingMode(insteon.Group) error {
	tl.commands <- fmt.Sprintf("%s EnterUnLinkingMode", tl.name)
	return nil
}

func (tl *testLinkable) ExitLinkingMode() error {
	tl.commands <- fmt.Sprintf("%s ExitLinkingMode", tl.name)
	return nil
}

func TestLinkUtils(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		test   func(input []insteon.LinkRecord, meta []string, want string, t *testing.T)
	}{
		{"Fix Cross Links", "testdata/fixcrosslink_", testFixCrosslinks},
		{"Add Links", "testdata/addlinks_", testAddLinks},
		{"Print Links", "testdata/printlinks", testPrintLinks},
		{"Dump Links", "testdata/dumplinks", testDumpLinks},
		{"Find Links", "testdata/findlinkrecord", testFindLinkRecord},
		{"Anonymize", "testdata/anonymize", testAnonymize},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inputs, _ := filepath.Glob(fmt.Sprintf("%s*.input", test.prefix))
			for _, inputFile := range inputs {
				inputBytes, err := ioutil.ReadFile(inputFile)
				if err != nil {
					t.Fatalf("Failed to read input file %v", err)
				}

				input, err := TextToLinks(string(inputBytes))
				if err != nil {
					t.Fatalf("Failed to convert %s to links %v", inputFile, err)
				}

				meta := []string{}
				for _, line := range strings.Split(string(inputBytes), "\n") {
					if strings.HasPrefix(line, "#") {
						meta = append(meta, strings.TrimSpace(strings.TrimPrefix(line, "#")))
					}
				}

				ext := filepath.Ext(inputFile)
				wantFile := inputFile[0:len(inputFile)-len(ext)] + ".want"
				want, err := ioutil.ReadFile(wantFile)
				if err != nil && !errors.Is(err, fs.ErrNotExist) {
					t.Fatalf("Failed to read want file %v", err)
				}

				test.test(input, meta, string(want), t)
			}
		})
	}
}

func testFixCrosslinks(input []insteon.LinkRecord, meta []string, wantStr string, t *testing.T) {
	tl := &testLinkable{links: input}
	want, _ := TextToLinks(wantStr)
	addresses := Addresses{}
	addresses.Set(meta)

	got := fixCrosslinks(tl.links, addresses...)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Wanted\n%s got\n%s", LinksToText(want, false), LinksToText(got, false))
	}
}

func testAddLinks(input []insteon.LinkRecord, meta []string, wantStr string, t *testing.T) {
	tl := &testLinkable{links: input}
	twl := &testWritableLinkable{&testLinkable{}}
	twl.links = make([]insteon.LinkRecord, len(tl.links))
	copy(twl.links, tl.links)

	want := []insteon.LinkRecord{}
	addlinks := []insteon.LinkRecord{}
	for i, m := range meta {
		l := insteon.LinkRecord{}
		err := l.UnmarshalText([]byte(m))
		if err != nil {
			t.Fatalf("Failed to unmarshal meta record %d: %v", i, err)
		}
		addlinks = append(addlinks, l)
	}

	want, err := TextToLinks(wantStr)
	if err != nil {
		t.Fatalf("Failed to convert want string to links: %v", err)
	}

	for _, l := range []devices.Linkable{tl, twl} {
		err := AddLinks(l, addlinks...)
		if err == nil {
			got, _ := l.Links()
			if !reflect.DeepEqual(want, got) {
				t.Errorf("%T: Wanted\n%s got\n%s", l, LinksToText(want, false), LinksToText(got, false))
			}
		} else {
			t.Errorf("Unexpected error %v", err)
		}
	}
}

func TestFindDuplicateLinks(t *testing.T) {
	links := []insteon.LinkRecord{
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address(0x010203)},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address(0x010203)},
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address(0x040506)},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address(0x040506)},
	}
	want := []insteon.LinkRecord{}

	tl := &testLinkable{links: links}
	got, _ := FindDuplicateLinks(tl, func(l1, l2 insteon.LinkRecord) bool { return l1.Equal(&l2) })

	if !reflect.DeepEqual(want, got) {
		t.Errorf("want duplicate links %v got %v", want, got)
	}

	// create a duplicate
	dup := links[0]
	want = append(want, dup)
	tl.links = append(tl.links, dup)

	got, _ = FindDuplicateLinks(tl, func(l1, l2 insteon.LinkRecord) bool { return l1.Equal(&l2) })

	if !reflect.DeepEqual(want, got) {
		t.Errorf("want duplicate links %v got %v", want, got)
	}
}

func TestLinksToText(t *testing.T) {
	links := []insteon.LinkRecord{
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address(0x010203)},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address(0x010203)},
		{Flags: insteon.UnavailableController, Group: 1, Address: insteon.Address(0x040506)},
		{Flags: insteon.UnavailableResponder, Group: 1, Address: insteon.Address(0x040506)},
	}

	want := `UC        1 01.02.03   00 00 00
UR        1 01.02.03   00 00 00
UC        1 04.05.06   00 00 00
UR        1 04.05.06   00 00 00
`
	got := LinksToText(links, false)
	if want != got {
		t.Errorf("Wanted %q\n got %q", want, got)
	}

	want = `#
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

	got = LinksToText(links, true)
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

func testAnonymize(input []insteon.LinkRecord, meta []string, wantStr string, t *testing.T) {
	addresses := Addresses{}
	addresses.Set(meta)
	want, err := TextToLinks(wantStr)
	if err == nil {
		got := Anonymize(input, addresses...)
		if !reflect.DeepEqual(want, got) {
			t.Errorf("Wanted\n%v\ngot\n%v\n", LinksToText(want, false), LinksToText(got, false))
		}
	} else {
		t.Errorf("Unexpected error %v", err)
	}
}

func testFindLinkRecord(input []insteon.LinkRecord, meta []string, _ string, t *testing.T) {
	tl := &testLinkable{links: input}
	for i, str := range meta {
		data := strings.Split(str, ":")
		wantFound := true
		if data[0] == "found" {
			wantFound = true
		} else if data[0] == "not_found" {
			wantFound = false
		} else {
			t.Fatalf("Unknown operation %q", data[0])
		}

		want := insteon.LinkRecord{}
		err := want.UnmarshalText([]byte(data[1]))
		if err == nil {
			got, err := FindLinkRecord(tl, want.Flags.Controller(), want.Address, want.Group)
			if wantFound && err != nil {
				t.Errorf("Wanted record to be found got %v", err)
			} else if !wantFound && err == nil {
				t.Errorf("Wanted record to not be found, but %v was found", got)
			} else if err == nil && want != got {
				t.Errorf("want link %v got %v", want, got)
			}
		} else {
			t.Fatalf("Failed to unmarshal link record %d: %v", i, err)
		}
	}
}

func testPrintLinks(input []insteon.LinkRecord, meta []string, want string, t *testing.T) {
	tl := &testLinkable{links: input}
	out := &bytes.Buffer{}
	err := PrintLinkDatabase(out, tl)
	if err == nil {
		got := string(out.Bytes())
		if want != got {
			t.Errorf("Wanted:\n%v\nGot:\n%v", want, got)
		}
	} else {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func testDumpLinks(input []insteon.LinkRecord, meta []string, want string, t *testing.T) {
	tl := &testLinkable{links: input}
	out := &bytes.Buffer{}
	err := DumpLinkDatabase(out, tl)
	if err == nil {
		got := string(out.Bytes())
		if want != got {
			t.Errorf("Wanted:%q\nGot:%q", want, got)
		}
	} else {
		t.Fatalf("Unexpected error: %v", err)
	}
}
