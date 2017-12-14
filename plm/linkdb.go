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

type LinkDB struct {
	plm       *PLM
	rrCh      chan *Packet
	refreshCh chan chan bool
	linkCh    chan chan []*insteon.Link
}

func NewLinkDB(plm *PLM) *LinkDB {
	db := &LinkDB{
		plm:       plm,
		rrCh:      make(chan *Packet),
		refreshCh: make(chan chan bool),
		linkCh:    make(chan chan []*insteon.Link),
	}

	go db.controlLoop()
	return db
}

func (db *LinkDB) controlLoop() {
	links := make([]*insteon.Link, 0)
	var refreshCh chan bool
	for {
		select {
		case refreshCh = <-db.refreshCh:
			links = make([]*insteon.Link, 0)
			resp, err := db.plm.Send(&Packet{Command: CmdGetFirstAllLink})
			if resp.NAK() {
				refreshCh <- true
				close(refreshCh)
				refreshCh = nil
			} else if err != nil {
				insteon.Log.Infof("Error sending GetNextAllLink command: %v", err)
			}
		case packet := <-db.rrCh:
			// only do something if we are in the process of refreshing
			if refreshCh != nil {
				link := packet.Payload.(*insteon.Link)
				links = append(links, link)
				resp, err := db.plm.Send(&Packet{Command: CmdGetNextAllLink})
				if resp.NAK() {
					refreshCh <- true
					close(refreshCh)
					refreshCh = nil
				} else if err != nil {
					insteon.Log.Infof("Error sending GetNextAllLink command: %v", err)
				}
			}
		case linkCh := <-db.linkCh:
			newLinks := make([]*insteon.Link, len(links))
			for i, link := range links {
				newLink := *link
				newLinks[i] = &newLink
			}
			linkCh <- newLinks
			close(linkCh)
		}
	}
}

func (db *LinkDB) Links() []*insteon.Link {
	ch := make(chan []*insteon.Link)
	db.linkCh <- ch
	return <-ch
}

func (db *LinkDB) Refresh() error {
	ch := make(chan bool)
	db.refreshCh <- ch
	<-ch
	return nil
}

func (db *LinkDB) RemoveLink(oldLink *insteon.Link) (err error) {
	numDelete := 0
	deletedLinks := make([]*insteon.Link, 0)
	for _, link := range db.Links() {
		if link.Group == oldLink.Group && link.Address == oldLink.Address {
			numDelete++
			if !oldLink.Equal(link) {
				deletedLinks = append(deletedLinks, link)
			}
		}
	}

	for i := 0; i < numDelete; i++ {
		_, err = db.plm.Send(&Packet{Command: CmdManageAllLinkRecord, Payload: &manageRecordRequest{command: LinkCmdDeleteFirst, link: oldLink}})
		if err != nil {
			insteon.Log.Infof("Failed to remove link: %v", err)
		}
	}

	// add back links that we didn't want deleted
	for _, link := range deletedLinks {
		db.AddLink(link)
	}
	return err
}

func (db *LinkDB) AddLink(newLink *insteon.Link) error {
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

func (db *LinkDB) Cleanup() (err error) {
	/*removeable := make([]*insteon.Link, 0)
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
	return err*/
	return ErrNotImplemented
}
