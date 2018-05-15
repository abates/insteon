package plm

import (
	"fmt"
	"time"

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
	link    *insteon.LinkRecord
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
	mrr.link = &insteon.LinkRecord{}
	return mrr.link.UnmarshalBinary(buf[1:])
}

type LinkingMode byte

type AllLinkReq struct {
	Mode  LinkingMode
	Group insteon.Group
}

func (alr *AllLinkReq) MarshalBinary() ([]byte, error) {
	return []byte{byte(alr.Mode), byte(alr.Group)}, nil
}

func (alr *AllLinkReq) UnmarshalBinary(buf []byte) error {
	if len(buf) < 2 {
		return fmt.Errorf("Needed 2 bytes to unmarshal all link request.  Got %d", len(buf))
	}
	alr.Mode = LinkingMode(buf[0])
	alr.Group = insteon.Group(buf[1])
	return nil
}

func (alr *AllLinkReq) String() string {
	return fmt.Sprintf("%02x %d", alr.Mode, alr.Group)
}

func (db *PLM) Links() ([]*insteon.LinkRecord, error) {
	links := make([]*insteon.LinkRecord, 0)
	// receive all-link record responses

	insteon.Log.Debugf("Retrieving PLM link database")
	resp, err := db.send(&Packet{Command: CmdGetFirstAllLink})
	if resp.NAK() {
		err = nil
	} else if err == nil {
	loop:
		for {
			select {
			case packet := <-db.linkCh:
				link := &insteon.LinkRecord{}
				err := link.UnmarshalBinary(packet.payload)
				if err == nil {
					insteon.Log.Debugf("Received PLM record response %v", link)
					links = append(links, link)
					var resp *Packet
					resp, err = db.send(&Packet{Command: CmdGetNextAllLink})
					if resp.NAK() || err != nil {
						break loop
					}
				} else {
					insteon.Log.Infof("Failed to unmarshal link record: %v", err)
					break loop
				}
			case <-time.After(insteon.Timeout):
				err = insteon.ErrReadTimeout
				break loop
			}
		}
	}
	return links, err
}

func (plm *PLM) RemoveLinks(oldLinks ...*insteon.LinkRecord) (err error) {
	deletedLinks := make([]*insteon.LinkRecord, 0)
	for _, oldLink := range oldLinks {
		numDelete := 0
		var links []*insteon.LinkRecord
		links, err = plm.Links()
		if err == nil {
			for _, link := range links {
				if link.Group == oldLink.Group && link.Address == oldLink.Address {
					numDelete++
					if !oldLink.Equal(link) {
						deletedLinks = append(deletedLinks, link)
					}
				}
			}

			for i := 0; i < numDelete; i++ {
				rr := &manageRecordRequest{command: LinkCmdDeleteFirst, link: oldLink}
				payload, _ := rr.MarshalBinary()
				_, err = plm.send(&Packet{Command: CmdManageAllLinkRecord, payload: payload})
				if err != nil {
					insteon.Log.Infof("Failed to remove link: %v", err)
					break
				}
			}
		} else {
			insteon.Log.Infof("Failed to retrieve links: %v", err)
			break
		}
	}

	if err == nil {
		// add back links that we didn't want deleted
		for _, link := range deletedLinks {
			plm.AddLink(link)
		}
	}
	return err
}

func (db *PLM) AddLink(newLink *insteon.LinkRecord) error {
	var command recordRequestCommand
	if newLink.Flags.Controller() {
		command = LinkCmdModFirstCtrl
	} else {
		command = LinkCmdModFirstResp
	}
	rr := &manageRecordRequest{command: command, link: newLink}
	payload, _ := rr.MarshalBinary()
	resp, err := db.send(&Packet{Command: CmdManageAllLinkRecord, payload: payload})

	if resp.NAK() {
		err = fmt.Errorf("Failed to add link back to ALDB")
	}

	return err
}

func (db *PLM) WriteLink(*insteon.LinkRecord) error {
	return insteon.ErrNotImplemented
}

func (db *PLM) Cleanup() (err error) {
	removeable := make([]*insteon.LinkRecord, 0)
	links, err := db.Links()
	if err == nil {
		for i, l1 := range links {
			for _, l2 := range links[i+1:] {
				if l1.Equal(l2) {
					removeable = append(removeable, l2)
				}
			}
		}

		err = db.RemoveLinks(removeable...)
	}
	return err
}

func (db *PLM) AddManualLink(group insteon.Group) error {
	return db.EnterLinkingMode(group)
}

func (db *PLM) EnterLinkingMode(group insteon.Group) error {
	lr := &AllLinkReq{Mode: LinkingMode(0x03), Group: group}
	payload, _ := lr.MarshalBinary()
	ack, err := db.Retry(&Packet{
		Command: CmdStartAllLink,
		payload: payload,
	}, 3)

	if err == nil && ack.NAK() {
		err = ErrNak
	}
	return err
}

func (db *PLM) ExitLinkingMode() error {
	ack, err := db.Retry(&Packet{
		Command: CmdCancelAllLink,
	}, 3)

	if err == nil && ack.NAK() {
		err = ErrNak
	}
	return err
}

func (db *PLM) EnterUnlinkingMode(group insteon.Group) error {
	lr := &AllLinkReq{Mode: LinkingMode(0xff), Group: group}
	payload, _ := lr.MarshalBinary()
	ack, err := db.Retry(&Packet{
		Command: CmdStartAllLink,
		payload: payload,
	}, 3)

	if err == nil && ack.NAK() {
		err = ErrNak
	}
	return err
}

func (db *PLM) AssignToAllLinkGroup(insteon.Group) error   { return ErrNotImplemented }
func (db *PLM) DeleteFromAllLinkGroup(insteon.Group) error { return ErrNotImplemented }
