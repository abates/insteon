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
	"math/rand"
	"sort"
	"strings"
	"text/template"

	"github.com/abates/insteon"
	"github.com/abates/insteon/devices"
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
func LinksToText(links []insteon.LinkRecord, printHeading bool) string {
	builder := &strings.Builder{}
	textlinksTmpl.Execute(builder, struct {
		PrintHeading bool
		Links        []insteon.LinkRecord
	}{
		PrintHeading: printHeading,
		Links:        links,
	})
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
// LinkDB and return any links that are duplicates. Comparison
// is done with the provided equal function.  If the equal
// function returns true, then the two links are considered
// duplicated
func FindDuplicateLinks(linkable devices.Linkable, equal func(l1, l2 insteon.LinkRecord) bool) ([]insteon.LinkRecord, error) {
	duplicates := make([]insteon.LinkRecord, 0)
	links, err := linkable.Links()
	if err == nil {
		for i, l1 := range links {
			for _, l2 := range links[i+1:] {
				// Available links cannot be duplicates
				if equal(l1, l2) {
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
func FindLinkRecord(device devices.Linkable, controller bool, address insteon.Address, group insteon.Group) (found insteon.LinkRecord, err error) {
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

// Anonymize returns a copy of the links where each link
// Address field has been replaced, in order, by the given
// replacement addresses.  If fewer replacement addresses are
// supplied than required, a random address is generated using
// rand.Int31()
//
// Links that have the same addresses in their Address field
// will all have the same substituted Address
func Anonymize(links []insteon.LinkRecord, replacements ...insteon.Address) []insteon.LinkRecord {
	newLinks := []insteon.LinkRecord{}
	index := make(map[insteon.Address]insteon.Address)
	for _, link := range links {
		newAddr, found := index[link.Address]
		if !found {
			if len(replacements) > 0 {
				newAddr = replacements[0]
				replacements = replacements[1:]
			} else {
				newAddr = insteon.Address(0x00ffffff & rand.Int31())
			}
			index[link.Address] = newAddr
		}
		link.Address = newAddr
		newLinks = append(newLinks, link)
	}
	return newLinks
}

// MissingCrosslinks loops through the links and looks for
// controller and responder records for each of the addresses
// if either a controller or responder link is missing in the
// input list, then it is added to the slice of "missing"
// records returned
func MissingCrosslinks(links []insteon.LinkRecord, forAddresses ...insteon.Address) []insteon.LinkRecord {
	missed := []insteon.LinkRecord{}
	for _, forAddr := range forAddresses {
		processed := make(map[insteon.Address]bool)
		for i, l1 := range links {
			found := false
			if l1.Flags.Available() || l1.Address != forAddr || processed[l1.Address] {
				continue
			}
			processed[l1.Address] = true

			for _, l2 := range links[i+1:] {
				if l2.Address == forAddr && l1.Group == l2.Group {
					if l1.Flags.Controller() == l2.Flags.Responder() {
						found = true
						break
					}
				}
			}

			// if we get to here, then no matching record was found
			if !found {
				l2 := l1
				if l2.Flags.Controller() {
					l2.Flags.SetResponder()
				} else {
					l2.Flags.SetController()
				}
				missed = append(missed, l2)
			}
		}
	}
	return missed
}

func AddLinks(device devices.Linkable, addLinks ...insteon.LinkRecord) (err error) {
	newLinks := []insteon.LinkRecord{}
	existingLinks, err := device.Links()
	for i := 0; i < len(existingLinks) && err == nil; i++ {
		existing := existingLinks[i]
		if wr, ok := device.(devices.WriteLink); ok {
			if len(addLinks) > 0 && existing.Flags.Available() {
				err = wr.WriteLink(i, addLinks[0])
				addLinks = addLinks[1:]
			}
		} else {
			if len(addLinks) > 0 && existing.Flags.Available() {
				newLinks = append(newLinks, addLinks[0])
				addLinks = addLinks[1:]
			} else {
				newLinks = append(newLinks, existing)
			}
		}
	}

	if err == nil {
		if len(addLinks) > 0 {
			if wr, ok := device.(devices.WriteLink); ok {
				for i, link := range addLinks {
					err = wr.WriteLink(len(existingLinks)+i, link)
				}
			} else {
				newLinks = append(newLinks, addLinks...)
			}
		}

		if len(newLinks) > 0 {
			err = device.WriteLinks(newLinks...)
		}
	}
	return err
}

func fixCrosslinks(links []insteon.LinkRecord, forAddresses ...insteon.Address) []insteon.LinkRecord {
	newLinks := make([]insteon.LinkRecord, len(links))
	copy(newLinks, links)
	for _, m := range MissingCrosslinks(links, forAddresses...) {
		added := false
		for i, l := range newLinks {
			if l.Flags.Available() {
				newLinks[i] = m
				added = true
				break
			}
		}
		if !added {
			newLinks = append(links, m)
		}
	}
	return newLinks
}

// CrossLinkAll will create bi-directional links among all the devices
// listed. This is useful for creating virtual N-Way connections
func CrossLinkAll(group insteon.Group, devices ...devices.Linkable) (err error) {
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
func CrossLink(group insteon.Group, d1, d2 devices.Linkable) error {
	err := Link(group, d1, d2)
	if err == nil || errors.Is(err, ErrAlreadyLinked) {
		err = Link(group, d2, d1)
		if errors.Is(err, ErrAlreadyLinked) {
			err = nil
		}
	}
	return err
}

// ForceLink will create links in the controller and responder All-Link
// databases without first checking if the links exist. The links are
// created by simulating set button presses (using EnterLinkingMode)
func ForceLink(group insteon.Group, controller, responder devices.Linkable) error {
	// The sequence to create a link between two devices follows:
	// 1) Controller enters linking mode (same as holding down the set button for 10 seconds)
	// 2) Controller sends a "Set-Button Pressed Controller" broadcast message
	// 3) Responder enters linking mode (just like holding down the set button)
	// 4) Responder sends a "Set-Button Pressed Responder" broadcast message
	//
	// At this point the two devices will exchange direct messages that won't necessarily
	// be seen by the initiator (such as a PLM), so as soon as the responder broadcast
	// is received, we assume the linking is complete
	devices.LogDebug.Printf("Putting controller %s into linking mode", controller)

	// controller enters all-linking mode
	// and waits for set-button message.  If not
	// set-button message is received, err will
	// be ErrReadTimeout
	err := controller.EnterLinkingMode(group)

	if err == nil {
		// responder pushes the set button responder and
		// waits for the set-button message
		devices.LogDebug.Printf("Assigning responder to group")
		err = responder.EnterLinkingMode(group)

		controller.ExitLinkingMode()
		responder.ExitLinkingMode()
	}
	return err
}

// Link will add appropriate entries to the controller's and responder's All-Link
// database. Each devices' ALDB will be searched for existing links, if both entries
// exist (a controller link and a responder link) then nothing is done. If only one
// entry exists than the other is deleted and new links are created. Once the link
// check/cleanup has taken place the new links are created using ForceLink
func Link(group insteon.Group, controller, responder devices.Linkable) (err error) {
	devices.LogDebug.Printf("Looking for existing links")
	var controllerLink, responderLink insteon.LinkRecord
	controllerLink, err = FindLinkRecord(controller, true, responder.Address(), group)

	if err == ErrLinkNotFound {
		responderLink, err = FindLinkRecord(responder, false, controller.Address(), group)

		if err == nil {
			// the controller did not have a link to the responder, but
			// the responder had a link to the controller so we want to
			// remove it before re-linking the devices
			devices.LogDebug.Printf("Responder link already exists, deleting it")
			err = RemoveLinks(responder, responderLink)
		}

		if err == nil || err == ErrLinkNotFound {
			err = ForceLink(group, controller, responder)
		}
	} else if err == nil {
		_, err = FindLinkRecord(responder, false, controller.Address(), group)
		if err == ErrLinkNotFound {
			// The controller link exists, but no matching responder link
			// exists, so we want to remove the controller link before
			// re-linking
			devices.LogDebug.Printf("Responder link already exists, deleting it")
			err = RemoveLinks(controller, controllerLink)
			if err == nil {
				err = ForceLink(group, controller, responder)
			}
		}
	}
	return err
}

// UnlinkAll will unlink all groups between a controller and
// a responder device
func UnlinkAll(controller, responder devices.Linkable) error {
	links, err := controller.Links()
	if err == nil {
		for _, link := range links {
			if link.Address == responder.Address() {
				if link.Flags.Controller() {
					err = Unlink(link.Group, controller, responder)
				} else {
					err = Unlink(link.Group, responder, controller)
				}
			}
		}
	}
	return err
}

// Unlink will unlink a controller from a responder for a given Group. The
// controller is put into UnlinkingMode (analogous to unlinking mode via
// the set button) and then the responder is put into unlinking mode (also
// analogous to the set button pressed)
func Unlink(group insteon.Group, controller, responder devices.Linkable) error {
	// controller enters all-linking mode
	devices.LogDebug.Printf("Putting controller %v into unlinking mode", controller)
	err := controller.EnterUnlinkingMode(group)

	// responder pushes the set button responder
	if err == nil {
		devices.LogDebug.Printf("Instructing responder %v to unlink", responder)
		err = responder.EnterLinkingMode(group)
		controller.ExitLinkingMode()
		responder.ExitLinkingMode()
	}

	return err
}

func RemoveLinks(device devices.Linkable, remove ...insteon.LinkRecord) error {
	links, err := device.Links()
	if err == nil {
		removeLinks := []insteon.LinkRecord{}
		for i, link := range links {
			for _, r := range remove {
				if link.Equal(&r) {
					link.Flags.SetAvailable()
					if wl, ok := device.(devices.WriteLink); ok {
						wl.WriteLink(i, link)
					} else {
						removeLinks = append(removeLinks, link)
					}
					break
				}
			}
		}

		if len(removeLinks) > 0 {
			err = device.UpdateLinks(removeLinks...)
		}
	}
	return err
}

func DumpLinkDatabase(out io.Writer, linkable devices.Linkable) error {
	links, err := linkable.Links()
	if err == nil {
		err = DumpLinks(out, links)
	}
	return err
}

func DumpLinks(out io.Writer, links []insteon.LinkRecord) error {
	return dumplinksTmpl.Execute(out, links)
}

func PrintLinkDatabase(out io.Writer, linkable devices.Linkable) error {
	links, err := linkable.Links()
	if err == nil {
		err = PrintLinks(out, links)
	}
	return err
}

func PrintLinks(out io.Writer, links []insteon.LinkRecord) error {
	sort.Sort(Links(links))
	return printlinksTmpl.Execute(out, links)
}
