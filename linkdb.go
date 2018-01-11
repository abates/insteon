package insteon

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrLinkNotFound  = errors.New("Link not found in database")
	ErrAlreadyLinked = errors.New("Responder already linked to controller")
)

const (
	BaseLinkDBAddress = MemAddress(0x0fff)
)

type MemAddress int

func (ma MemAddress) String() string {
	return fmt.Sprintf("%02x.%02x", byte(ma>>8), byte(ma&0xff))
}

const (
	ReadLink  LinkRequestType = 0x00
	WriteLink LinkRequestType = 0x02
)

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

type LinkRequest struct {
	Type       LinkRequestType
	MemAddress MemAddress
	NumRecords int
	Link       *Link
}

func (lr *LinkRequest) String() string {
	if lr.Link == nil {
		return fmt.Sprintf("%s %s %d", lr.Type, lr.MemAddress, lr.NumRecords)
	}
	return fmt.Sprintf("%s %s %d %s", lr.Type, lr.MemAddress, lr.NumRecords, lr.Link)
}

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
		lr.Link = &Link{}
	case 0x02:
		lr.NumRecords = int(buf[4])
		lr.Link = &Link{}
	}

	if lr.Link != nil {
		err = lr.Link.UnmarshalBinary(buf[5:])
	}
	return err
}

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

type LinkDB interface {
	Links() ([]*Link, error)
	AddLink(newLink *Link) error
	RemoveLinks(oldLinks ...*Link) error
}

type Linkable interface {
	Address() Address
	AssignToAllLinkGroup(Group) error
	DeleteFromAllLinkGroup(Group) error
	EnterLinkingMode(Group) error
	EnterUnlinkingMode(Group) error
	ExitLinkingMode() error
	LinkDB() (LinkDB, error)
}

type DeviceLinkDB struct {
	conn Connection
}

func NewDeviceLinkDB(conn Connection) *DeviceLinkDB {
	return &DeviceLinkDB{conn}
}

func (db *DeviceLinkDB) AddLink(newLink *Link) error {
	memAddress := BaseLinkDBAddress - 7
	links, err := db.Links()
	if err == nil {
		for _, link := range links {
			if link.Flags.Available() {
				break
			}
			memAddress -= 8
		}

		err = linkWriter(db.conn, memAddress, newLink)
	}
	return err
}

func (db *DeviceLinkDB) RemoveLinks(oldLinks ...*Link) error {
	links, err := db.Links()
	if err == nil {
	top:
		for _, oldLink := range oldLinks {
			Log.Debugf("Attempting to remove %v", oldLink)
			memAddress := BaseLinkDBAddress - 7
			for i, link := range links {
				if link.Equal(oldLink) {
					link.Flags.setAvailable()
					Log.Debugf("Marking link available")
					err = linkWriter(db.conn, memAddress, link)
					if err != nil {
						break top
					}
					links = append(links[0:i], links[i+1:]...)
				}
				memAddress -= 8
			}
		}
	}
	return err
}

func readLinks(conn Connection) ([]*Link, error) {
	rrCh := conn.Subscribe(CmdReadWriteALDB)
	defer conn.Unsubscribe(rrCh)
	Log.Debugf("Retrieving Device link database")
	links := make([]*Link, 0)
	lastAddress := MemAddress(0)
	_, err := SendExtendedCommand(conn, CmdReadWriteALDB, &LinkRequest{Type: ReadLink, NumRecords: 0})
	if err != nil {
		return nil, err
	}

loop:
	for {
		select {
		case msg := <-rrCh:
			lr := msg.Payload.(*LinkRequest)
			if lr.MemAddress != lastAddress {
				lastAddress = lr.MemAddress
				if lr.Link.Flags == 0x00 {
					break loop
				} else {
					links = append(links, lr.Link)
				}
			}
		case <-time.After(Timeout):
			err = ErrReadTimeout
			break loop
		}
	}
	return links, err
}

var linkReader = readLinks

func (db *DeviceLinkDB) Links() ([]*Link, error) {
	return linkReader(db.conn)
}

var linkWriter = writeLink

func writeLink(conn Connection, memAddress MemAddress, link *Link) error {
	request := &LinkRequest{Type: WriteLink, Link: link}
	_, err := SendExtendedCommand(conn, CmdReadWriteALDB, request)
	return err
}

func DuplicateLinks(db LinkDB) ([]*Link, error) {
	duplicates := make([]*Link, 0)
	links, err := db.Links()
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

func FindLink(db LinkDB, controller bool, address Address, group Group) (*Link, error) {
	links, err := db.Links()
	if err == nil {
		for _, link := range links {
			if link.Flags.Controller() == controller && link.Address == address && link.Group == group {
				return link, nil
			}
		}
	}
	return nil, err
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

func Unlink(controller, responder Linkable, group Group) error {
	err := controller.EnterUnlinkingMode(group)
	defer controller.ExitLinkingMode()

	if err == nil {
		err = responder.EnterUnlinkingMode(group)
		defer responder.ExitLinkingMode()
	}
	time.Sleep(time.Second)
	return err
}

func DeleteLinks(controller, responder Linkable) (err error) {
	controllerDB, err := controller.LinkDB()

	if err == nil {
		var links []*Link
		links, err = controllerDB.Links()
		if err == nil {
			for _, link := range links {
				if link.Address == responder.Address() {
					err = DeleteLink(responder, controller, link.Group)
				}
			}
		}
	}
	return err
}

func DeleteLink(controller, responder Linkable, group Group) (err error) {
	// controller enters all-linking mode
	err = controller.EnterUnlinkingMode(group)
	//defer responder.ExitLinkingMode()

	// wait a moment for messages to propagate
	//time.Sleep(time.Second)

	// responder pushes the set button responder
	if err == nil {
		Log.Debugf("Unlinking responder from group")
		err = responder.EnterLinkingMode(group)
		//defer controller.ExitLinkingMode()
	}

	// wait a moment for messages to propagate
	//time.Sleep(time.Second)

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
		var controllerLink *Link
		controllerLink, err = FindLink(controllerDB, true, responder.Address(), group)

		if err == nil {
			var responderLink *Link
			responderLink, err = FindLink(responderDB, false, controller.Address(), group)

			if err == nil {
				if controllerLink != nil && responderLink != nil {
					err = ErrAlreadyLinked
				} else {
					// correct a mismatch by deleting the one link found
					// and recreating both
					if controllerLink != nil {
						Log.Debugf("Controller link already exists, deleting it")
						err = controllerDB.RemoveLinks(controllerLink)
					}

					if err == nil && responderLink != nil {
						Log.Debugf("Responder link already exists, deleting it")
						err = responderDB.RemoveLinks(controllerLink)
					}

					ForceCreateLink(controller, responder, group)
				}
			}
		}
	}
	return err
}
