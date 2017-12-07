package insteon

import (
	"errors"
	"time"
)

var (
	ErrLinkNotFound  = errors.New("Link not found in database")
	ErrAlreadyLinked = errors.New("Responder already linked to controller")
)

type Linkable interface {
	Address() Address
	AssignToAllLinkGroup(Group) error
	DeleteFromAllLinkGroup(Group) error
	EnterLinkingMode(Group) error
	EnterUnlinkingMode(Group) error
	ExitLinkingMode() error
	LinkDB() (LinkDB, error)
}

type LinkDB interface {
	AddLink(*Link) error
	RemoveLink(*Link) error
	Refresh() error
	Links() []*Link
	Cleanup() error
}

type LinkWriter interface {
	WriteLink(MemAddress, *Link) error
}

type BaseLinkDB struct {
	LinkWriter
	links []*Link
}

func (db *BaseLinkDB) AddLink(newLink *Link) error {
	linkPos := -1
	memAddress := MemAddress(0x0fff)
	for i, link := range db.links {
		if link.Flags.Available() {
			linkPos = i
			break
		}
		memAddress -= 8
	}

	memAddress -= 8
	if linkPos >= 0 {
		db.links[linkPos] = newLink
	} else {
		db.links = append(db.links, newLink)
	}

	// if this fails, then the local link database
	// could be different from the remove database
	return db.WriteLink(memAddress, newLink)
}

func (db *BaseLinkDB) Links() []*Link {
	return db.links
}

func (db *BaseLinkDB) RemoveLink(oldLink *Link) error {
	Log.Debugf("Attempting to remove %v", oldLink)
	memAddress := MemAddress(0x0fff)
	for _, link := range db.links {
		memAddress -= 8
		if link.Equal(oldLink) {
			link.Flags.setAvailable()
			Log.Debugf("Marking link available")
			return db.WriteLink(memAddress, link)
		}
	}
	Log.Debugf("Link not found in database")
	return nil
}

func (db *BaseLinkDB) Cleanup() (err error) {
	removeable := make([]*Link, 0)
	for i, l1 := range db.links {
		for _, l2 := range db.links[i+1:] {
			if l1.Equal(l2) {
				Log.Debugf("Link to be removed: %v", l2)
				removeable = append(removeable, l2)
			}
		}
	}

	for _, link := range removeable {
		err = db.RemoveLink(link)
		if err != nil {
			Log.Infof("Failed to remove link %s: %v", link, err)
			break
		}
	}
	return err
}

func FindLink(db LinkDB, controller bool, address Address, group Group) *Link {
	for _, link := range db.Links() {
		if link.Flags.Controller() == controller && link.Address == address && link.Group == group {
			return link
		}
	}
	return nil
}

func CrossLink(l1, l2 Linkable, group Group) error {
	err := CreateLink(l1, l2, group)
	if err == nil || err == ErrAlreadyLinked {
		err = nil
		err = CreateLink(l2, l1, group)
		if err == ErrAlreadyLinked {
			err = nil
		}
	}

	return err
}

// Don't look through the database first
func ForceCreateLink(controller, responder Linkable, group Group) (err error) {
	Log.Debugf("Putting controller %s into linking mode", controller)
	// controller enters all-linking mode
	err = controller.EnterLinkingMode(group)
	defer controller.ExitLinkingMode()

	// wait a moment for messages to propogate
	time.Sleep(1 * time.Second)

	// responder pushes the set button responder
	if err == nil {
		Log.Debugf("Assigning responder to group")
		err = responder.EnterLinkingMode(group)
		defer responder.ExitLinkingMode()
	}

	// wait a moment for messages to propogate
	time.Sleep(1 * time.Second)

	return
}

func Unlink(controller, responder Linkable) error {
	err := DeleteLinks(controller, responder)
	if err == nil {
		err = DeleteLinks(responder, controller)
	}
	return err
}

func DeleteLinks(controller, responder Linkable) (err error) {
	controllerDB, err := controller.LinkDB()

	if err == nil {
		for _, link := range controllerDB.Links() {
			if link.Address == responder.Address() {
				err = DeleteLink(responder, controller, link.Group)
			}
		}
	}
	return err
}

func DeleteLink(controller, responder Linkable, group Group) (err error) {
	// controller enters all-linking mode
	err = controller.EnterUnlinkingMode(group)
	defer responder.ExitLinkingMode()

	// wait a moment for messages to propogate
	time.Sleep(time.Second)

	// responder pushes the set button responder
	if err == nil {
		Log.Debugf("Unlinking responder from group")
		err = responder.EnterUnlinkingMode(group)
		defer controller.ExitLinkingMode()
	}

	// wait a moment for messages to propogate
	time.Sleep(time.Second)

	return
}

func CreateLink(controller, responder Linkable, group Group) error {
	// check for existing link
	Log.Debugf("Retrieving link databases...")
	var responderDB LinkDB
	controllerDB, err := controller.LinkDB()
	if err == nil || err == ErrNotLinked {
		responderDB, err = responder.LinkDB()
	}

	if err == nil || err == ErrNotLinked {
		Log.Debugf("Looking for existing links")
		controllerLink := FindLink(controllerDB, true, responder.Address(), group)
		responderLink := FindLink(responderDB, false, controller.Address(), group)

		if controllerLink != nil && responderLink != nil {
			err = ErrAlreadyLinked
		} else {
			// correct a mismatch by deleting the one link found
			// and recreating both
			if controllerLink != nil {
				Log.Debugf("Controller link already exists, deleting it")
				err = controllerDB.RemoveLink(controllerLink)
			}

			if err == nil && responderLink != nil {
				Log.Debugf("Responder link already exists, deleting it")
				err = responderDB.RemoveLink(controllerLink)
			}

			ForceCreateLink(controller, responder, group)
		}
	}
	return err
}
