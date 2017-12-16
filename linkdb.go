package insteon

import (
	"errors"
	"fmt"
	"time"
)

const (
	BaseLinkDBAddress = MemAddress(0x0fff)

	ReadLink  LinkRequestType = 0x00
	WriteLink LinkRequestType = 0x02
)

var (
	ErrLinkNotFound  = errors.New("Link not found in database")
	ErrAlreadyLinked = errors.New("Responder already linked to controller")
)

type MemAddress int

func (ma MemAddress) String() string {
	return fmt.Sprintf("%02x.%02x", byte(ma>>8), byte(ma&0xff))
}

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
	Refresh() error
	Links() []*Link
	RemoveLink(oldLink *Link) error
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

	refreshCh chan chan bool
	linkCh    chan chan []*Link
	closeCh   chan chan error
}

func NewDeviceLinkDB(conn Connection) *DeviceLinkDB {
	db := &DeviceLinkDB{
		conn: conn,

		refreshCh: make(chan chan bool),
		linkCh:    make(chan chan []*Link),
		closeCh:   make(chan chan error),
	}
	go db.readWriteLoop()
	return db
}

func (db *DeviceLinkDB) readWriteLoop() {
	var lastAddress MemAddress
	var refreshCh chan bool
	var closeCh chan error

	links := make([]*Link, 0)
	rrCh := db.conn.Subscribe(CmdReadWriteALDB)

loop:
	for {
		select {
		case refreshCh = <-db.refreshCh:
			Log.Debugf("Refreshing Device link database")
			links = make([]*Link, 0)
			lastAddress = MemAddress(0)
			request := &LinkRequest{Type: ReadLink, NumRecords: 0}
			_, err := SendExtendedCommand(db.conn, CmdReadWriteALDB, request)
			if err != nil {
				Log.Infof("Failed to send message: %v", err)
			}
		case msg := <-rrCh:
			// only do something if we are in the process of refreshing
			if refreshCh != nil {
				lr := msg.Payload.(*LinkRequest)
				if lr.MemAddress != lastAddress {
					lastAddress = lr.MemAddress
					if lr.Link.Flags != 0x00 {
						links = append(links, lr.Link)
					} else {
						refreshCh <- true
						close(refreshCh)
						refreshCh = nil
					}
				}
			}
		case linkCh := <-db.linkCh:
			Log.Debugf("Returning Device link database")
			newLinks := make([]*Link, len(links))
			for i, link := range links {
				newLink := *link
				newLinks[i] = &newLink
			}
			linkCh <- newLinks
			close(linkCh)
		case closeCh = <-db.closeCh:
			break loop
		}
	}

	if refreshCh != nil {
		close(refreshCh)
	}
	closeCh <- nil
}

func (db *DeviceLinkDB) Close() error {
	ch := make(chan error)
	db.closeCh <- ch
	return <-ch
}

func (db *DeviceLinkDB) AddLink(newLink *Link) error {
	/*linkPos := -1
	memAddress := BaseLinkDBAddress
	for i, link := range db.Links() {
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
	return db.WriteLink(memAddress, newLink)*/
	return ErrNotImplemented
}

func (db *DeviceLinkDB) RemoveLink(oldLink *Link) error {
	/*Log.Debugf("Attempting to remove %v", oldLink)
	memAddress := BaseLinkDBAddress
	for _, link := range db.links {
		memAddress -= 8
		if link.Equal(oldLink) {
			link.Flags.setAvailable()
			Log.Debugf("Marking link available")
			return db.WriteLink(memAddress, link)
		}
	}
	Log.Debugf("Link not found in database")
	return nil*/
	return ErrNotImplemented
}

func (db *DeviceLinkDB) Cleanup() (err error) {
	/*addresses := make([]MemAddress, 0)
	removeable := make([]*Link, 0)
	for i, l1 := range db.links {
		memAddress := BaseLinkDBAddress - (8 * (MemAddress(i) + 1))
		for _, l2 := range db.links[i+1:] {
			memAddress -= 8
			if l1.Equal(l2) {
				Log.Debugf("Link to be removed: %v", l2)
				addresses = append(addresses, memAddress)
				removeable = append(removeable, l2)
			}
		}
	}

	for i, link := range removeable {
		link.Flags.setAvailable()
		Log.Debugf("Marking link available")
		err = db.WriteLink(addresses[i], link)
		if err != nil {
			Log.Infof("Failed to remove link %s: %v", link, err)
			break
		}
	}
	return err
	*/
	return ErrNotImplemented
}

func (db *DeviceLinkDB) Links() []*Link {
	ch := make(chan []*Link)
	db.linkCh <- ch
	return <-ch
}

func (db *DeviceLinkDB) Refresh() error {
	ch := make(chan bool)
	db.refreshCh <- ch
	<-ch
	return nil
}

func (db *DeviceLinkDB) WriteLink(memAddress MemAddress, link *Link) error {
	request := &LinkRequest{Type: WriteLink, Link: link}
	_, err := SendExtendedCommand(db.conn, CmdReadWriteALDB, request)
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
	//defer responder.ExitLinkingMode()

	// wait a moment for messages to propogate
	//time.Sleep(time.Second)

	// responder pushes the set button responder
	if err == nil {
		Log.Debugf("Unlinking responder from group")
		err = responder.EnterLinkingMode(group)
		//defer controller.ExitLinkingMode()
	}

	// wait a moment for messages to propogate
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
