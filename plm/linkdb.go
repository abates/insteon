package plm

import (
	"fmt"

	"github.com/abates/insteon"
)

type recordRequestCommand byte

const (
	LinkCmdFindFirst    recordRequestCommand = 0x00
	LinkCmdFindNext     recordRequestCommand = 0x01
	LinkCmdModFirst     recordRequestCommand = 0x20
	LinkCmdModFirstCtrl recordRequestCommand = 0x40
	LinkCmdModFirstResp recordRequestCommand = 0x41
	LinkCmdDeleteFirst  recordRequestCommand = 0x80
)

type manageRecordRequest struct {
	command recordRequestCommand
	link    *insteon.Link
}

func (mrr *manageRecordRequest) String() string {
	return fmt.Sprintf("%02x %s", mrr.command, mrr.link)
}

func (mrr *manageRecordRequest) MarshalBinary() ([]byte, error) {
	payload, err := mrr.link.MarshalBinary()
	buf := make([]byte, len(payload)+1)
	buf[0] = byte(mrr.command)
	copy(buf[1:], payload)
	return buf, err
}

func (mrr *manageRecordRequest) UnmarshalBinary(buf []byte) error {
	mrr.command = recordRequestCommand(buf[0])
	mrr.link = &insteon.Link{}
	return mrr.link.UnmarshalBinary(buf[1:])
}

type PLMLinkDB struct {
	plm   *PLM
	links []*insteon.Link
}

func (db *PLMLinkDB) Links() []*insteon.Link {
	if db.links == nil {
		db.Refresh()
	}
	return db.links
}

func (db *PLMLinkDB) Refresh() error {
	db.links = make([]*insteon.Link, 0)

	resp, err := db.plm.Send(&Packet{Command: CmdGetFirstAllLink})
	if err == nil && resp.ACK() {
		for {
			packet, err := db.plm.Receive()
			if err == nil {
				link := packet.Payload.(*insteon.Link)
				db.links = append(db.links, link)
				resp, err = db.plm.Send(&Packet{Command: CmdGetNextAllLink})
				if resp.NAK() {
					break
				}
			} else {
				break
			}
		}
	}

	return err
}

func (db *PLMLinkDB) RemoveLink(oldLink *insteon.Link) (err error) {
	var resp *Packet
	deletedLinks := make([]*insteon.Link, 0)
	for {
		resp, err = db.plm.Send(&Packet{Command: CmdManageAllLinkRecord, Payload: &manageRecordRequest{command: LinkCmdFindFirst, link: oldLink}})
		if resp.NAK() {
			break
		} else {
			resp, err = db.plm.Receive()
			if err == nil {
				if !oldLink.Equal(resp.Payload.(*insteon.Link)) {
					deletedLinks = append(deletedLinks, resp.Payload.(*insteon.Link))
				}
				_, err = db.plm.Send(&Packet{Command: CmdManageAllLinkRecord, Payload: &manageRecordRequest{command: LinkCmdDeleteFirst, link: oldLink}})
				if err != nil {
					break
				}
			} else {
				break
			}
		}
	}

	// add back links that we didn't want deleted
	for _, link := range deletedLinks {
		db.AddLink(link)
	}
	return err
}

func (db *PLMLinkDB) AddLink(newLink *insteon.Link) error {
	var command recordRequestCommand
	if newLink.Flags.Controller() {
		command = LinkCmdModFirstCtrl
	} else {
		command = LinkCmdModFirstResp
	}
	resp, err := db.plm.Send(&Packet{Command: CmdManageAllLinkRecord, Payload: &manageRecordRequest{command: command, link: newLink}})

	if resp.NAK() {
		err = fmt.Errorf("Failed to add link back to ALDB")
	}

	return err
}

func (db *PLMLinkDB) Cleanup() (err error) {
	removeable := make([]*insteon.Link, 0)
	for i, l1 := range db.links {
		for _, l2 := range db.links[i+1:] {
			if l1.Equal(l2) {
				removeable = append(removeable, l2)
			}
		}
	}

	for _, link := range removeable {
		err = db.RemoveLink(link)
		if err != nil {
			break
		}
	}
	return err
}
