package insteon

import (
	"fmt"
)

type LinkRequestType byte

const (
	ReadLink  LinkRequestType = 0x00
	WriteLink LinkRequestType = 0x02
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

type MemAddress int

func (ma MemAddress) String() string {
	return fmt.Sprintf("%02x.%02x", byte(ma>>8), byte(ma&0xff))
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

type I2LinkDB struct {
	BaseLinkDB
	conn Connection
}

func NewI2LinkDB(conn Connection) LinkDB {
	db := &I2LinkDB{conn: conn}
	db.BaseLinkDB.LinkWriter = db
	return db
}

func (db *I2LinkDB) Refresh() error {
	db.BaseLinkDB.links = make([]*Link, 0)
	request := &LinkRequest{Type: ReadLink, NumRecords: 0}
	_, err := SendExtendedCommand(db.conn, CmdReadWriteALDB, request)
	if err != nil {
		return err
	}

	var msg *Message
	for {
		msg, err = db.conn.Receive()
		if err != nil {
			break
		}

		if lr, ok := msg.Payload.(*LinkRequest); ok {
			if lr.Link.Flags == 0x00 {
				break
			}
			db.links = append(db.BaseLinkDB.links, lr.Link)
		}
	}
	return err
}

func (db *I2LinkDB) WriteLink(memAddress MemAddress, link *Link) error {
	request := &LinkRequest{Type: WriteLink, Link: link}
	_, err := SendExtendedCommand(db.conn, CmdReadWriteALDB, request)
	return err
}
