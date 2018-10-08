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
	"errors"
	"time"
)

var (
	// ErrAlreadyLinked is returned when creating a link and an existing matching link is found
	ErrAlreadyLinked = errors.New("Responder already linked to controller")
)

const (
	// BaseLinkDBAddress is the base address of devices All-Link database
	BaseLinkDBAddress = MemAddress(0x0fff)
)

// MemAddress is an integer representing a specific location in a device's memory
type MemAddress int

func (ma MemAddress) String() string {
	return sprintf("%02x.%02x", byte(ma>>8), byte(ma&0xff))
}

// There are two link request types, one used to read the link database and
// one used to write links
const (
	ReadLink  LinkRequestType = 0x00
	WriteLink LinkRequestType = 0x02
)

// LinkRequestType is used to indicate whether an ALDB request is for reading
// or writing the database
type LinkRequestType byte

func (lrt LinkRequestType) String() string {
	switch lrt {
	case 0x00:
		return "Link Read"
	case 0x01:
		return "Link Resp"
	case 0x02:
		return "Link Write"
	}
	return "Unknown"
}

// LinkRequest is the message sent to a device to request reading or writing
// all-link database records
type LinkRequest struct {
	Type       LinkRequestType
	MemAddress MemAddress
	NumRecords int
	Link       *LinkRecord
}

func (lr *LinkRequest) String() string {
	if lr.Link == nil {
		return sprintf("%s %s %d", lr.Type, lr.MemAddress, lr.NumRecords)
	}
	return sprintf("%s %s %d %s", lr.Type, lr.MemAddress, lr.NumRecords, lr.Link)
}

// UnmarshalBinary will take the byte slice and convert it to a LinkRequest object
func (lr *LinkRequest) UnmarshalBinary(buf []byte) (err error) {
	if len(buf) < 5 {
		return newBufError(ErrBufferTooShort, 6, len(buf))
	}
	lr.Type = LinkRequestType(buf[1])
	lr.MemAddress = MemAddress(buf[2]) << 8
	lr.MemAddress |= MemAddress(buf[3])

	switch lr.Type {
	case 0x00:
		lr.NumRecords = int(buf[4])
	case 0x01:
		lr.Link = &LinkRecord{}
	case 0x02:
		lr.NumRecords = int(buf[4])
		lr.Link = &LinkRecord{}
	}

	if lr.Link != nil {
		err = lr.Link.UnmarshalBinary(buf[5:])
		lr.Link.memAddress = lr.MemAddress
	}
	return err
}

// MarshalBinary will convert the LinkRequest to a byte slice appropriate for
// sending out to the insteon network
func (lr *LinkRequest) MarshalBinary() (buf []byte, err error) {
	var linkData []byte
	buf = make([]byte, 14)
	buf[1] = byte(lr.Type)
	buf[2] = byte(lr.MemAddress >> 8)
	buf[3] = byte(lr.MemAddress & 0xff)
	switch lr.Type {
	case 0x00:
		buf[4] = byte(lr.NumRecords)
	case 0x01:
		buf[4] = 0x00
		linkData, err = lr.Link.MarshalBinary()
		copy(buf[5:], linkData)
	case 0x02:
		buf[4] = 0x08
		linkData, err = lr.Link.MarshalBinary()
		copy(buf[5:], linkData)
	}
	return buf, err
}

// FindDuplicateLinks will perform a linear search of the
// LinkDB and return any links that are duplicates. Duplicate
// links are those that are equivalent as reported by LinkRecord.Equal
func FindDuplicateLinks(linkable LinkableDevice) ([]*LinkRecord, error) {
	duplicates := make([]*LinkRecord, 0)
	links, err := linkable.Links()
	if err == nil {
		for i, l1 := range links {
			for _, l2 := range links[i+1:] {
				if l1.Equal(l2) {
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
func FindLinkRecord(linkable LinkableDevice, controller bool, address Address, group Group) (*LinkRecord, error) {
	links, err := linkable.Links()
	if err == nil {
		for _, link := range links {
			if link.Flags.Controller() == controller && link.Address == address && link.Group == group {
				return link, nil
			}
		}
	}
	return nil, err
}

// CrossLinkAll will create bi-directional links among all the devices
// listed. This is useful for creating virtual N-Way connections
func CrossLinkAll(group Group, linkable ...LinkableDevice) error {
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
func CrossLink(group Group, l1, l2 LinkableDevice) error {
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
func ForceLink(group Group, controller, responder LinkableDevice) (err error) {
	Log.Debugf("Putting controller %s into linking mode", controller)
	// controller enters all-linking mode
	err = controller.EnterLinkingMode(group)
	defer controller.ExitLinkingMode()
	time.Sleep(2 * time.Second)

	// responder pushes the set button responder
	if err == nil {
		Log.Debugf("Assigning responder to group")
		err = responder.EnterLinkingMode(group)
		defer responder.ExitLinkingMode()
	}
	time.Sleep(time.Second)
	return
}

// UnlinkAll will unlink all groups between a controller and
// a responder device
func UnlinkAll(controller, responder LinkableDevice) (err error) {
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
func Unlink(group Group, controller, responder LinkableDevice) (err error) {
	// controller enters all-linking mode
	err = controller.EnterUnlinkingMode(group)
	defer controller.ExitLinkingMode()

	// wait a moment for messages to propagate
	time.Sleep(2 * time.Second)

	// responder pushes the set button responder
	if err == nil {
		Log.Debugf("Unlinking responder from group")
		err = responder.EnterLinkingMode(group)
		defer responder.ExitLinkingMode()
	}

	// wait a moment for messages to propagate
	time.Sleep(time.Second)

	return
}

// Link will add appropriate entries to the controller's and responder's All-Link
// database. Each devices' ALDB will be searched for existing links, if both entries
// exist (a controller link and a responder link) then nothing is done. If only one
// entry exists than the other is deleted and new links are created. Once the link
// check/cleanup has taken place the new links are created using ForceLink
func Link(group Group, controller, responder LinkableDevice) (err error) {
	Log.Debugf("Looking for existing links")
	var controllerLink *LinkRecord
	controllerLink, err = FindLinkRecord(controller, true, responder.Address(), group)

	if err == nil {
		var responderLink *LinkRecord
		responderLink, err = FindLinkRecord(responder, false, controller.Address(), group)

		if err == nil {
			if controllerLink != nil && responderLink != nil {
				err = ErrAlreadyLinked
			} else {
				// correct a mismatch by deleting the one link found
				// and recreating both
				if controllerLink != nil {
					Log.Debugf("Controller link already exists, deleting it")
					err = controller.RemoveLinks(controllerLink)
				}

				if err == nil && responderLink != nil {
					Log.Debugf("Responder link already exists, deleting it")
					err = responder.RemoveLinks(controllerLink)
				}

				ForceLink(group, controller, responder)
			}
		}
	}
	return err
}
