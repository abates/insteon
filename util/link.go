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
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/abates/insteon"
)

var (
	// ErrAlreadyLinked is returned when creating a link and an existing matching link is found
	ErrAlreadyLinked = errors.New("Responder already linked to controller")

	// ErrLinkNotFound is returned by the Find function when no matching record was found
	ErrLinkNotFound = errors.New("Link was not found in the database")
)

type Linkable interface {
	LinkDatabase() (insteon.Linkable, error)
}

func IfLinkable(thing Linkable, cb func(linkable insteon.Linkable) error) error {
	linkdb, err := thing.LinkDatabase()
	if err == nil {
		return cb(linkdb)
	}
	return insteon.ErrNotLinkable
}

// FindDuplicateLinks will perform a linear search of the
// LinkDB and return any links that are duplicates. Duplicate
// links are those that are equivalent as reported by LinkRecord.Equal
func FindDuplicateLinks(linkable insteon.Linkable) ([]*insteon.LinkRecord, error) {
	duplicates := make([]*insteon.LinkRecord, 0)
	links, err := linkable.Links()
	if err == nil {
		for i, l1 := range links {
			for _, l2 := range links[i+1:] {
				// Available links cannot be duplicates
				if !l1.Flags.Available() && l1.Equal(l2) {
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
func FindLinkRecord(linkable insteon.Linkable, controller bool, address insteon.Address, group insteon.Group) (found *insteon.LinkRecord, err error) {
	links, err := linkable.Links()
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
	return
}

// CrossLinkAll will create bi-directional links among all the devices
// listed. This is useful for creating virtual N-Way connections
func CrossLinkAll(group insteon.Group, linkable ...insteon.Linkable) error {
	for i, l1 := range linkable {
		for _, l2 := range linkable[i:] {
			if l1 != l2 {
				err := CrossLink(group, l1, l2)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// CrossLink will create bi-directional links between the two linkable
// devices. Each device will get both a controller and responder
// link for the given group. When using lighting control devices, this
// will effectively create a 3-Way light switch configuration
func CrossLink(group insteon.Group, l1, l2 insteon.Linkable) error {
	err := Link(group, l1, l2)
	if err == nil || err == ErrAlreadyLinked {
		err = Link(group, l2, l1)
		if err == ErrAlreadyLinked {
			err = nil
		}
	}

	return err
}

// ForceLink will create links in the controller and responder All-Link
// databases without first checking if the links exist. The links are
// created by simulating set button presses (using EnterLinkingMode)
func ForceLink(group insteon.Group, controller, responder insteon.Linkable) (err error) {
	// The sequence to create a link between two devices follows:
	// 1) Controller enters linking mode (same as holding down the set button for 10 seconds)
	// 2) Controller sends a "Set-Button Pressed Controller" broadcast message
	// 3) Responder enters linking mode (just like holding down the set button)
	// 4) Responder sends a "Set-Button Pressed Responder" broadcast message
	//
	// At this point the two devices will exchange direct messages that won't necessarily
	// be seen by the initiator (such as a PLM), so as soon as the responder broadcast
	// is received, we assume the linking is complete
	insteon.Log.Debugf("Putting controller %s into linking mode", controller)

	// controller enters all-linking mode
	// and waits for set-button message.  If not
	// set-button message is received, err will
	// be ErrReadTimeout
	err = controller.EnterLinkingMode(group)

	if err == nil {
		defer controller.ExitLinkingMode()

		// responder pushes the set button responder and
		// waits for the set-button message
		insteon.Log.Debugf("Assigning responder to group")
		err = responder.EnterLinkingMode(group)
		defer responder.ExitLinkingMode()
	}
	return
}

// UnlinkAll will unlink all groups between a controller and
// a responder device
func UnlinkAll(controller, responder insteon.Linkable) (err error) {
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
func Unlink(group insteon.Group, controller, responder insteon.Linkable) (err error) {
	// controller enters all-linking mode
	insteon.Log.Debugf("Putting controller %v into unlinking mode", controller)
	err = controller.EnterUnlinkingMode(group)
	defer controller.ExitLinkingMode()

	// responder pushes the set button responder
	if err == nil {
		insteon.Log.Debugf("Instructing responder %v to unlink", responder)
		err = responder.EnterLinkingMode(group)
		time.Sleep(2 * time.Second)
		defer responder.ExitLinkingMode()
	}

	return
}

func RemoveLinks(linkable insteon.Linkable, remove ...*insteon.LinkRecord) error {
	links, err := linkable.Links()
	if err == nil {
		removeLinks := []*insteon.LinkRecord{}
		for _, link := range links {
			for _, r := range remove {
				if link.Equal(r) {
					link.Flags.SetAvailable()
					removeLinks = append(removeLinks, link)
					break
				}
			}
		}
		err = linkable.UpdateLinks(removeLinks...)
	}
	return err
}

// Link will add appropriate entries to the controller's and responder's All-Link
// database. Each devices' ALDB will be searched for existing links, if both entries
// exist (a controller link and a responder link) then nothing is done. If only one
// entry exists than the other is deleted and new links are created. Once the link
// check/cleanup has taken place the new links are created using ForceLink
func Link(group insteon.Group, controller, responder insteon.Linkable) (err error) {
	insteon.Log.Debugf("Looking for existing links")
	var controllerLink, responderLink *insteon.LinkRecord
	controllerLink, err = FindLinkRecord(controller, true, responder.Address(), group)

	if err == ErrLinkNotFound {
		responderLink, err = FindLinkRecord(responder, false, controller.Address(), group)
		if err == nil {
			// found a responder link, but not a controller link
			insteon.Log.Debugf("Controller link already exists, deleting it")
			err = RemoveLinks(responder, responderLink)
		}

		if err == nil || err == ErrLinkNotFound {
			err = ForceLink(group, controller, responder)
		}
	} else if err == nil {
		_, err = FindLinkRecord(responder, false, controller.Address(), group)
		if err == ErrLinkNotFound {
			// found a controller link, but not a responder link
			insteon.Log.Debugf("Responder link already exists, deleting it")
			err = RemoveLinks(controller, controllerLink)
			err = ForceLink(group, controller, responder)
		}
	}
	return err
}

func DumpLinkDatabase(out io.Writer, linkable Linkable) error {
	return IfLinkable(linkable, func(linkable insteon.Linkable) error {
		links, err := linkable.Links()
		if err == nil {
			fmt.Fprintf(out, "links:\n")
			for _, link := range links {
				buf, _ := link.MarshalBinary()
				s := make([]string, len(buf))
				for i, b := range buf {
					s[i] = fmt.Sprintf("0x%02x", b)
				}
				fmt.Fprintf(out, "- [ %s ]\n", strings.Join(s, ", "))
			}
		}
		return err
	})
}

func PrintLinks(out io.Writer, linkable Linkable) error {
	return IfLinkable(linkable, func(linkable insteon.Linkable) error {
		dbLinks, err := linkable.Links()
		fmt.Fprintf(out, "Link Database:\n")
		if len(dbLinks) > 0 {
			fmt.Fprintf(out, "    Flags Group Address    Data\n")

			links := make(map[string][]*insteon.LinkRecord)
			for _, link := range dbLinks {
				links[link.Address.String()] = append(links[link.Address.String()], link)
			}

			linkAddresses := []string{}
			for linkAddress := range links {
				linkAddresses = append(linkAddresses, linkAddress)
			}
			sort.Strings(linkAddresses)

			for _, linkAddress := range linkAddresses {
				for _, link := range links[linkAddress] {
					fmt.Fprintf(out, "    %-5s %5s %8s   %02x %02x %02x\n", link.Flags, link.Group, link.Address, link.Data[0], link.Data[1], link.Data[2])
				}
			}
		} else {
			fmt.Fprintf(out, "    No links defined\n")
		}
		return err
	})
}
