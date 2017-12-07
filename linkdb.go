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
	LinkDB() (LinkDB, error)
}

type LinkDB interface {
	AddLink(*Link) error
	RemoveLink(*Link) error
	Refresh() error
	Links() []*Link
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
	memAddress := MemAddress(0x0fff)
	for _, link := range db.links {
		memAddress -= 8
		if link.Equal(oldLink) {
			link.Flags.setAvailable()
			return db.WriteLink(memAddress, link)
		}
	}
	return nil
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
	// controller enters all-linking mode
	err = controller.EnterLinkingMode(group)

	// wait a moment for messages to propogate
	time.Sleep(time.Second)

	// responder pushes the set button responder
	if err == nil {
		Log.Debugf("Assigning responder to group")
		err = responder.EnterLinkingMode(group)
	}
	return
}

func Unlink(controller, responder Linkable) error {
	controllerDB, err := controller.LinkDB()

	if err == nil {
		for _, link := range controllerDB.Links() {
			if link.Address == responder.Address() {
				err = DeleteLink(responder, controller, link.Group)
				if err == nil {
					err = DeleteLink(controller, responder, link.Group)
				}
			}
		}
	}
	return err
}

func DeleteLink(controller, responder Linkable, group Group) (err error) {
	// controller enters all-linking mode
	err = controller.EnterUnlinkingMode(group)

	// wait a moment for messages to propogate
	time.Sleep(time.Second)

	// responder pushes the set button responder
	if err == nil {
		Log.Debugf("Unlinking responder from group")
		err = responder.EnterUnlinkingMode(group)
	}
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
