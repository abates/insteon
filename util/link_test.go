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
	links []insteon.LinkRecord
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

func (tl *testLinkable) Load(filename string) (meta []string, want string, err error) {
	inputStr, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(inputStr), "\n") {
		if strings.HasPrefix(line, "#") {
			meta = append(meta, strings.TrimSpace(strings.TrimPrefix(line, "#")))
		}
	}

	tl.links, err = TextToLinks(string(inputStr))
	if err == nil {
		ext := filepath.Ext(filename)
		wantFile := filename[0:len(filename)-len(ext)] + ".want"
		var wantBytes []byte
		wantBytes, err = ioutil.ReadFile(wantFile)
		want = string(wantBytes)
		if errors.Is(err, fs.ErrNotExist) {
			err = nil
		} else if err != nil {
			err = fmt.Errorf("Failed reading %s %w", wantFile, err)
		}
	}
	return
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
func (tl *testLinkable) EnterLinkingMode(insteon.Group) error    { return nil }
func (tl *testLinkable) EnterUnlinkingMode(insteon.Group) error  { return nil }
func (tl *testLinkable) ExitLinkingMode() error                  { return nil }

func TestFixCrosslinks(t *testing.T) {
	inputs, _ := filepath.Glob("testdata/fixcrosslink_*.input")
	for _, inputFile := range inputs {
		t.Run(filepath.Base(inputFile), func(t *testing.T) {
			tl := &testLinkable{}
			meta, wantStr, err := tl.Load(inputFile)
			want, _ := TextToLinks(wantStr)
			if err == nil {
				addresses := []insteon.Address{}
				for _, str := range meta {
					address := insteon.Address(0)
					address.Set(str)
					addresses = append(addresses, address)
				}

				got := fixCrosslinks(tl.links, addresses...)
				if !reflect.DeepEqual(want, got) {
					t.Errorf("Wanted\n%s got\n%s", LinksToText(want, false), LinksToText(got, false))
				}
			} else {
				t.Errorf("Unexpected error %v", err)
			}
		})
	}
}

func TestAddLinks(t *testing.T) {
	inputs, _ := filepath.Glob("testdata/addlinks_*.input")
	for _, inputFile := range inputs {
		t.Run(filepath.Base(inputFile), func(t *testing.T) {
			tl := &testLinkable{}
			meta, wantStr, err := tl.Load(inputFile)
			twl := &testWritableLinkable{&testLinkable{}}
			twl.links = make([]insteon.LinkRecord, len(tl.links))
			copy(twl.links, tl.links)

			want := []insteon.LinkRecord{}
			addlinks := []insteon.LinkRecord{}
			if err == nil {
				for _, m := range meta {
					l := insteon.LinkRecord{}
					err = l.UnmarshalText([]byte(m))
					if err != nil {
						break
					}
					addlinks = append(addlinks, l)
				}
			}

			if err == nil {
				want, err = TextToLinks(wantStr)
			}

			if err == nil {
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
			} else {
				t.Errorf("Unexpected error %v", err)
			}
		})
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

func TestFindLinkRecord(t *testing.T) {
	tests := []struct {
		desc    string
		input   insteon.LinkRecord
		want    insteon.LinkRecord
		wantErr error
	}{
		{"found", insteon.ControllerLink(1, insteon.Address(0x010203)), insteon.ControllerLink(1, insteon.Address(0x010203)), nil},
		{"not found", insteon.ControllerLink(1, insteon.Address(0x070809)), insteon.LinkRecord{}, ErrLinkNotFound},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			tl := &testLinkable{}
			_, _, err := tl.Load("testdata/findlinkrecord.input")
			if err == nil {
				input := test.input
				got, err := FindLinkRecord(tl, input.Flags.Controller(), input.Address, input.Group)
				if err == test.wantErr {
					if test.want != got {
						t.Errorf("want link %x got %x", byte(test.want.Flags), byte(got.Flags))
					}
				} else {
					t.Errorf("Wanted error %v got %v", test.wantErr, err)
				}
			} else {
				t.Errorf("Unexpected error %v", err)
			}
		})
	}
}

func TestPrintLinks(t *testing.T) {
	inputs, _ := filepath.Glob("testdata/printlinks*.input")
	for _, input := range inputs {
		t.Run(filepath.Base(input), func(t *testing.T) {

			tl := &testLinkable{}
			_, want, err := tl.Load(input)
			if err == nil {
				out := &bytes.Buffer{}
				err = PrintLinkDatabase(out, tl)
				if err == nil {
					got := string(out.Bytes())
					if want != got {
						t.Errorf("Wanted:\n%v\nGot:\n%v", want, got)
					}
				} else {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestDumpLinks(t *testing.T) {
	inputs, _ := filepath.Glob("testdata/dumplinks*.input")
	for _, input := range inputs {
		t.Run(filepath.Base(input), func(t *testing.T) {

			tl := &testLinkable{}
			_, want, err := tl.Load(input)
			if err == nil {
				out := &bytes.Buffer{}
				err = DumpLinkDatabase(out, tl)
				if err == nil {
					got := string(out.Bytes())
					if want != got {
						t.Errorf("Wanted:%q\nGot:%q", want, got)
					}
				} else {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}
