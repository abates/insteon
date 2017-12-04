package insteon

import (
	"errors"
	"fmt"
)

type LinkRequestType byte

const (
	ReadLink  LinkRequestType = 0x00
	WriteLink LinkRequestType = 0x02
)

var (
	ErrLinkNotFound  = errors.New("Link not found in database")
	ErrAlreadyLinked = errors.New("Responder already linked to controller")
)

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

type RecordControlFlags byte

func (rcf *RecordControlFlags) setBit(pos uint) {
	*rcf |= (1 << pos)
}

func (rcf *RecordControlFlags) clearBit(pos uint) {
	*rcf &= ^(1 << pos)
}

func (rcf RecordControlFlags) InUse() bool      { return rcf&0x80 == 0x80 }
func (rcf *RecordControlFlags) setInUse()       { rcf.setBit(7) }
func (rcf RecordControlFlags) Available() bool  { return rcf&0x80 == 0x00 }
func (rcf *RecordControlFlags) setAvailable()   { rcf.clearBit(7) }
func (rcf RecordControlFlags) Controller() bool { return rcf&0x40 == 0x40 }
func (rcf *RecordControlFlags) setController()  { rcf.setBit(6) }
func (rcf RecordControlFlags) Responder() bool  { return rcf&0x40 == 0x00 }
func (rcf *RecordControlFlags) setResponder()   { rcf.clearBit(6) }

func (rcf RecordControlFlags) String() string {
	str := "A"
	if rcf.InUse() {
		str = "U"
	}

	if rcf.Controller() {
		str += "C"
	} else {
		str += "R"
	}
	return str
}

type Group byte

func (g Group) String() string { return fmt.Sprintf("%d", byte(g)) }

type MemAddress int

func (ma MemAddress) String() string {
	return fmt.Sprintf("%02x.%02x", byte(ma>>8), byte(ma&0xff))
}

type Link struct {
	Flags   RecordControlFlags
	Group   Group
	Address Address
	Data    [3]byte
}

func (l *Link) String() string {
	return fmt.Sprintf("%s %s %s 0x%02x 0x%02x 0x%02x", l.Flags, l.Group, l.Address, l.Data[0], l.Data[1], l.Data[2])
}

func (l *Link) Equal(other *Link) bool {
	if l == other {
		return true
	}

	if l == nil || other == nil {
		return false
	}

	return l.Flags.InUse() == other.Flags.InUse() && l.Flags.Controller() == other.Flags.Controller() && l.Group == other.Group && l.Address == other.Address
}

func (l *Link) MarshalBinary() ([]byte, error) {
	data := make([]byte, 8)
	data[0] = byte(l.Flags)
	data[1] = byte(l.Group)
	copy(data[2:5], l.Address[:])
	copy(data[5:8], l.Data[:])
	return data, nil
}

func (l *Link) UnmarshalBinary(buf []byte) error {
	if len(buf) < 8 {
		return fmt.Errorf("link is 8 bytes, got %d", len(buf))
	}
	l.Flags = RecordControlFlags(buf[0])
	l.Group = Group(buf[1])
	copy(l.Address[:], buf[2:5])
	copy(l.Data[:], buf[5:8])
	return nil
}

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

type LinearLinkDB struct {
	conn  Connection
	links []*Link
}

func (ldb *LinearLinkDB) Links() []*Link {
	return ldb.links
}

func (ldb *LinearLinkDB) Refresh() error {
	ldb.links = make([]*Link, 0)
	request := &LinkRequest{Type: ReadLink, NumRecords: 0}
	err := ldb.conn.SendExtendedCommand(CmdReadWriteALDB, request)
	if err != nil {
		return err
	}

	var msg *Message
	for {
		msg, err = ldb.conn.Receive()
		if err != nil {
			break
		}

		if lr, ok := msg.Payload.(*LinkRequest); ok {
			if lr.Link.Flags == 0x00 {
				break
			}
			ldb.links = append(ldb.links, lr.Link)
		}
	}
	return err
}

func (ldb *LinearLinkDB) WriteLink(memAddress MemAddress, link *Link) error {
	request := &LinkRequest{Type: WriteLink, Link: link}
	return ldb.conn.SendExtendedCommand(CmdReadWriteALDB, request)
}

func (ldb *LinearLinkDB) RemoveLink(oldLink *Link) error {
	memAddress := MemAddress(0x0fff)
	for _, link := range ldb.links {
		memAddress -= 8
		if link.Equal(oldLink) {
			link.Flags.setAvailable()
			return ldb.WriteLink(memAddress, link)
		}
	}
	return nil
}

func (ldb *LinearLinkDB) AddLink(newLink *Link) error {
	linkPos := -1
	memAddress := MemAddress(0x0fff)
	for i, link := range ldb.links {
		if link.Flags.Available() {
			linkPos = i
			break
		}
		memAddress -= 8
	}

	memAddress -= 8
	if linkPos >= 0 {
		ldb.links[linkPos] = newLink
	} else {
		ldb.links = append(ldb.links, newLink)
	}

	// if this fails, then the local link database
	// could be different from the remove database
	return ldb.WriteLink(memAddress, newLink)
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

func CreateLink(controller Linkable, responder Linkable, group Group) (err error) {
	// check for existing link
	Log.Tracef("Retrieving link databases...")
	var controllerDB, responderDB LinkDB
	controllerDB, err = controller.LinkDB()
	if err == nil || err == ErrNotLinked {
		responderDB, err = responder.LinkDB()
	}

	if err == nil || err == ErrNotLinked {
		Log.Tracef("Looking for existing links")
		controllerLink := FindLink(controllerDB, true, responder.Address(), group)
		responderLink := FindLink(responderDB, false, controller.Address(), group)

		if controllerLink != nil && responderLink != nil {
			err = ErrAlreadyLinked
		} else {
			// correct a mismatch by deleting the one link found
			// and recreating both
			if controllerLink != nil {
				Log.Tracef("Controller link already exists, deleting it")
				err = controllerDB.RemoveLink(controllerLink)
			}

			if err == nil && responderLink != nil {
				Log.Tracef("Responder link already exists, deleting it")
				err = responderDB.RemoveLink(controllerLink)
			}

			// controller enters all-linking mode
			Log.Tracef("Putting controller into linking mode")
			controller.EnterLinkingMode(group)

			// responder pushes the set button responder
			if err == nil {
				Log.Tracef("Assigning responder to group")
				err = responder.EnterLinkingMode(group)
			}
		}
	}
	return err
}
