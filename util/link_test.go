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

func TestForceLinkUnlink(t *testing.T) {
	want := []string{
		"controller EnterLinkingMode 2",
		"responder EnterLinkingMode 2",
		"controller ExitLinkingMode",
		"responder ExitLinkingMode",
		"controller EnterUnlinkingMode 2",
		"responder EnterLinkingMode 2",
		"controller ExitLinkingMode",
		"responder ExitLinkingMode",
	}

	got := &cmdLogger{}
	controller := &testLinkable{name: "controller", commands: got}
	responder := &testLinkable{name: "responder", commands: got}
	err := ForceLink(2, controller, responder)
	if err == nil {
		err = Unlink(2, controller, responder)
	}

	if err == nil {
		wantStr := strings.Join(want, " ")
		gotStr := strings.Join(got.commands, " ")
		if wantStr != gotStr {
			t.Errorf("Wanted commands %q got %q", wantStr, gotStr)
		}
	} else {
		t.Errorf("unexpected error %v", err)
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

func getControllerResponder(input []insteon.LinkRecord, wantStr string) (*testWritableLinkable, *testWritableLinkable, []string, *cmdLogger) {
	want := strings.Split(wantStr, "\n")
	if want[len(want)-1] == "" {
		want = want[0 : len(want)-1]
	}

	got := &cmdLogger{}
	controllerLinks := []insteon.LinkRecord{}
	responderLinks := []insteon.LinkRecord{}
	if len(input) > 0 {
		controllerLinks = input[0 : len(input)/2]
		responderLinks = input[len(input)/2:]
	}
	controller := &testWritableLinkable{&testLinkable{address: insteon.Address(1), name: "controller", links: controllerLinks, commands: got}}
	responder := &testWritableLinkable{&testLinkable{address: insteon.Address(2), name: "responder", links: responderLinks, commands: got}}

	return controller, responder, want, got
}

func testUnlinkAll(input []insteon.LinkRecord, meta []string, wantStr string, t *testing.T) {
	controller, responder, want, got := getControllerResponder(input, wantStr)
	UnlinkAll(controller, responder)

	if strings.Join(want, " ") != strings.Join(got.commands, " ") {
		t.Errorf("Wanted commands %q got %q", strings.Join(want, " "), strings.Join(got.commands, " "))
	}
}

func testLink(input []insteon.LinkRecord, meta []string, wantStr string, t *testing.T) {
	controller, responder, want, got := getControllerResponder(input, wantStr)
	group := insteon.Group(0)
	if len(meta) > 0 {
		err := group.Set(meta[0])
		if err != nil {
			t.Fatalf("Failed to set group %v", err)
		}
	} else {
		t.Fatalf("Expected meta to have group id")
	}

	Link(group, controller, responder)

	if strings.Join(want, " ") != strings.Join(got.commands, " ") {
		t.Errorf("Wanted commands %q got %q", strings.Join(want, " "), strings.Join(got.commands, " "))
	}
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
		{"Link", "testdata/link_", testLink},
		{"UnlinkAll", "testdata/unlinkall_", testUnlinkAll},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inputs, _ := filepath.Glob(fmt.Sprintf("%s*.input", test.prefix))
			for _, inputFile := range inputs {
				t.Run(filepath.Base(inputFile), func(t *testing.T) {
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
				})
			}
		})
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

type cmdLogger struct {
	commands []string
}

func (c *cmdLogger) push(cmd string) {
	c.commands = append(c.commands, cmd)
}

type testLinkable struct {
	address  insteon.Address
	links    []insteon.LinkRecord
	name     string
	commands *cmdLogger
}

func (tl *testLinkable) logCmd(cmd string) {
	if tl.commands != nil {
		tl.commands.push(cmd)
	}
}

func (tl *testLinkable) Address() insteon.Address {
	return tl.address
}

func (tl *testLinkable) EnterLinkingMode(group insteon.Group) error {
	tl.logCmd(fmt.Sprintf("%s EnterLinkingMode %d", tl.name, group))
	return nil
}

func (tl *testLinkable) EnterUnlinkingMode(group insteon.Group) error {
	tl.logCmd(fmt.Sprintf("%s EnterUnlinkingMode %d", tl.name, group))
	return nil
}

func (tl *testLinkable) ExitLinkingMode() error {
	tl.logCmd(fmt.Sprintf("%s ExitLinkingMode", tl.name))
	return nil
}

func (tl *testLinkable) Links() ([]insteon.LinkRecord, error) {
	return tl.links, nil
}

func (tl *testLinkable) UpdateLinks(...insteon.LinkRecord) error { return nil }

func (tl *testLinkable) WriteLinks(links ...insteon.LinkRecord) error {
	tl.links = make([]insteon.LinkRecord, len(links))
	copy(tl.links, links)
	return nil
}

type testWritableLinkable struct {
	*testLinkable
}

func (twl *testWritableLinkable) WriteLink(i int, link insteon.LinkRecord) error {
	twl.logCmd(fmt.Sprintf("%s WriteLink %d %s", twl.name, i, link))
	if i < len(twl.links) {
		twl.links[i] = link
	} else if i == len(twl.links) {
		twl.links = append(twl.links, link)
	} else {
		return fmt.Errorf("Index out of range have %d need <= %d", i, len(twl.links))
	}
	return nil
}
