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

package util

import (
	_ "embed"
	"errors"
	"io"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/abates/insteon"
)

var (
	// ErrAlreadyLinked is returned when creating a link and an existing matching link is found
	ErrAlreadyLinked = errors.New("Responder already linked to controller")

	// ErrLinkNotFound is returned by the Find function when no matching record was found
	ErrLinkNotFound = errors.New("Link was not found in the database")
)

//go:embed printlinks.tmpl
var printlinksStr string
var printlinksTmpl *template.Template

//go:embed dumplinks.tmpl
var dumplinksStr string
var dumplinksTmpl *template.Template

//go:embed textlinks.tmpl
var textlinksStr string
var textlinksTmpl *template.Template

func init() {
	printlinksTmpl = template.Must(template.New("printlinks").Parse(printlinksStr))
	dumplinksTmpl = template.Must(template.New("dumplinks").Parse(dumplinksStr))
	textlinksTmpl = template.Must(template.New("textlinks").Parse(textlinksStr))
}

type Links []insteon.LinkRecord

func (l Links) Len() int      { return len(l) }
func (l Links) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l Links) Less(i, j int) bool {
	if l[i].Address == l[j].Address {
		return l[i].Group < l[j].Group
	}
	return l[i].Address.String() < l[j].Address.String()
}

// LinksToText will take a list of links and marshal them
// to text for editing
func LinksToText(links []insteon.LinkRecord) string {
	builder := &strings.Builder{}
	textlinksTmpl.Execute(builder, links)
	return builder.String()
}

// TextToLinks will take an input string and parse it into a list
// of link records.  This is useful for manually editing link databases
func TextToLinks(input string) (links []insteon.LinkRecord, err error) {
	lines := strings.Split(input, "\n")
	for i := 0; i < len(lines) && err == nil; i++ {
		line := lines[i]
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.Index(line, "#") == 0 {
			continue
		}
		link := insteon.LinkRecord{}
		err = link.UnmarshalText([]byte(line))
		if err == nil {
			links = append(links, link)
		}
	}
	return
}

// FindDuplicateLinks will perform a linear search of the
// LinkDB and return any links that are duplicates. Duplicate
// links are those that are equivalent as reported by LinkRecord.Equal
func FindDuplicateLinks(linkable insteon.Linkable) ([]insteon.LinkRecord, error) {
	duplicates := make([]insteon.LinkRecord, 0)
	links, err := linkable.Links()
	if err == nil {
		for i, l1 := range links {
			for _, l2 := range links[i+1:] {
				// Available links cannot be duplicates
				if !l1.Flags.Available() && l1.Equal(&l2) {
					duplicates = append(duplicates, l2)
				}
			}
		}
	}
	return duplicates, err
}

// FindLinkRecord will perform a linear search of the database and return
// a LinkRecord that matches the group, address and controller/responder
// indicator
func FindLinkRecord(device insteon.Linkable, controller bool, address insteon.Address, group insteon.Group) (found insteon.LinkRecord, err error) {
	links, err := device.Links()
	if err == nil {
		err = ErrLinkNotFound
		for _, link := range links {
			if !link.Flags.Available() && link.Flags.Controller() == controller && link.Address == address && link.Group == group {
				found = link
				err = nil
				break
			}
		}
	}
	return found, err
}

// CrossLinkAll will create bi-directional links among all the devices
// listed. This is useful for creating virtual N-Way connections
func CrossLinkAll(group insteon.Group, devices ...insteon.Linkable) (err error) {
	for i, d1 := range devices {
		for _, d2 := range devices[i:] {
			if d1 != d2 {
				err = CrossLink(group, d1, d2)
				if err != nil {
					return err
				}
			}
		}
	}
	return err
}

// CrossLink will create bi-directional links between the two linkable
// devices. Each device will get both a controller and responder
// link for the given group. When using lighting control devices, this
// will effectively create a 3-Way light switch configuration
func CrossLink(group insteon.Group, d1, d2 insteon.Linkable) error {
	err := Link(group, d1, d2)
	if err == nil || err == ErrAlreadyLinked {
		err = Link(group, d2, d1)
		if err == ErrAlreadyLinked {
			err = nil
		}
	}
	return err
}

// ForceLink will create links in the controller and responder All-Link
// databases without first checking if the links exist. The links are
// created by simulating set button presses (using EnterLinkingMode)
func ForceLink(group insteon.Group, controller, responder insteon.Linkable) error {
	// The sequence to create a link between two devices follows:
	// 1) Controller enters linking mode (same as holding down the set button for 10 seconds)
	// 2) Controller sends a "Set-Button Pressed Controller" broadcast message
	// 3) Responder enters linking mode (just like holding down the set button)
	// 4) Responder sends a "Set-Button Pressed Responder" broadcast message
	//
	// At this point the two devices will exchange direct messages that won't necessarily
	// be seen by the initiator (such as a PLM), so as soon as the responder broadcast
	// is received, we assume the linking is complete
	insteon.LogDebug.Printf("Putting controller %s into linking mode", controller)

	// controller enters all-linking mode
	// and waits for set-button message.  If not
	// set-button message is received, err will
	// be ErrReadTimeout
	err := controller.EnterLinkingMode(group)

	if err == nil {
		// responder pushes the set button responder and
		// waits for the set-button message
		insteon.LogDebug.Printf("Assigning responder to group")
		err = responder.EnterLinkingMode(group)

		controller.ExitLinkingMode()
		responder.ExitLinkingMode()
	}
	return err
}

// UnlinkAll will unlink all groups between a controller and
// a responder device
func UnlinkAll(controller, responder insteon.Linkable) error {
	links, err := controller.Links()
	if err == nil {
		for _, link := range links {
			if link.Address == responder.Address() {
				err = Unlink(link.Group, responder, controller)
			}
		}
	}
	return err
}

// Unlink will unlink a controller from a responder for a given Group. The
// controller is put into UnlinkingMode (analogous to unlinking mode via
// the set button) and then the responder is put into unlinking mode (also
// analogous to the set button pressed)
func Unlink(group insteon.Group, controller, responder insteon.Linkable) error {
	// controller enters all-linking mode
	insteon.LogDebug.Printf("Putting controller %v into unlinking mode", controller)
	err := controller.EnterUnlinkingMode(group)

	// responder pushes the set button responder
	if err == nil {
		insteon.LogDebug.Printf("Instructing responder %v to unlink", responder)
		err = responder.EnterLinkingMode(group)
		time.Sleep(2 * time.Second)

		controller.ExitLinkingMode()
		responder.ExitLinkingMode()
	}

	return err
}

func RemoveLinks(device insteon.Linkable, remove ...insteon.LinkRecord) error {
	links, err := device.Links()
	if err == nil {
		removeLinks := []insteon.LinkRecord{}
		for _, link := range links {
			for _, r := range remove {
				if link.Equal(&r) {
					link.Flags.SetAvailable()
					removeLinks = append(removeLinks, link)
					break
				}
			}
		}
		err = device.UpdateLinks(removeLinks...)
	}
	return err
}

// Link will add appropriate entries to the controller's and responder's All-Link
// database. Each devices' ALDB will be searched for existing links, if both entries
// exist (a controller link and a responder link) then nothing is done. If only one
// entry exists than the other is deleted and new links are created. Once the link
// check/cleanup has taken place the new links are created using ForceLink
func Link(group insteon.Group, controller, responder insteon.Linkable) (err error) {
	insteon.LogDebug.Printf("Looking for existing links")
	var controllerLink, responderLink insteon.LinkRecord
	controllerLink, err = FindLinkRecord(controller, true, controller.Address(), group)

	if err == ErrLinkNotFound {
		responderLink, err = FindLinkRecord(responder, false, responder.Address(), group)

		if err == nil {
			// found a responder link, but not a controller link
			insteon.LogDebug.Printf("Controller link already exists, deleting it")
			err = RemoveLinks(responder, responderLink)
		}

		if err == nil || err == ErrLinkNotFound {
			err = ForceLink(group, controller, responder)
		}
	} else if err == nil {
		_, err = FindLinkRecord(responder, false, controller.Address(), group)
		if err == ErrLinkNotFound {
			// found a controller link, but not a responder link
			insteon.LogDebug.Printf("Responder link already exists, deleting it")
			err = RemoveLinks(controller, controllerLink)
			err = ForceLink(group, controller, responder)
		}
	}
	return err
}

func DumpLinkDatabase(out io.Writer, linkable insteon.Linkable) error {
	links, err := linkable.Links()
	if err == nil {
		DumpLinks(out, links)
	}
	return err
}

func DumpLinks(out io.Writer, links []insteon.LinkRecord) {
	dumplinksTmpl.Execute(out, links)
}

func PrintLinkDatabase(out io.Writer, linkable insteon.Linkable) error {
	links, err := linkable.Links()
	if err == nil {
		PrintLinks(out, links)
	}
	return err
}

func PrintLinks(out io.Writer, links []insteon.LinkRecord) {
	sort.Sort(Links(links))
	printlinksTmpl.Execute(out, links)
}
