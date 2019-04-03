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

type linkingMode byte

type allLinkReq struct {
	Mode  linkingMode
	Group insteon.Group
}

func (alr *allLinkReq) MarshalBinary() ([]byte, error) {
	return []byte{byte(alr.Mode), byte(alr.Group)}, nil
}

func (alr *allLinkReq) UnmarshalBinary(buf []byte) error {
	if len(buf) < 2 {
		return fmt.Errorf("Needed 2 bytes to unmarshal all link request.  Got %d", len(buf))
	}
	alr.Mode = linkingMode(buf[0])
	alr.Group = insteon.Group(buf[1])
	return nil
}

func (alr *allLinkReq) String() string {
	return fmt.Sprintf("%02x %d", alr.Mode, alr.Group)
}

func (plm *PLM) Links() ([]*insteon.LinkRecord, error) {
	plm.Lock()
	defer plm.Unlock()
	links := make([]*insteon.LinkRecord, 0)
	insteon.Log.Debugf("Retrieving PLM link database")
	_, err := plm.tx(&Packet{Command: CmdGetFirstAllLink})
	// TODO: does this loop need a timeout as well?  If, for
	// some reason, we are receiving packets on the plmCh, but
	// none of them are AllLinkRecordResponses, then this will
	// loop forever until a read timeout occurs...
	for err == nil {
		select {
		case pkt := <-plm.plmCh:
			if pkt.Command == CmdAllLinkRecordResp {
				link := &insteon.LinkRecord{}
				err = link.UnmarshalBinary(pkt.Payload)
				if err == nil {
					insteon.Log.Debugf("Received PLM record response %v", link)
					links = append(links, link)
					_, err = plm.tx(&Packet{Command: CmdGetNextAllLink})
				} else {
					insteon.Log.Infof("Failed to unmarshal link record: %v", err)
					break
				}
			}
		case <-time.After(plm.timeout):
			err = ErrReadTimeout
		}
	}

	if err == ErrNak {
		err = nil
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
				_, err = plm.send(&Packet{Command: CmdManageAllLinkRecord, Payload: payload})
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

func (plm *PLM) AddLink(newLink *insteon.LinkRecord) error {
	var command recordRequestCommand
	if newLink.Flags.Controller() {
		command = LinkCmdModFirstCtrl
	} else {
		command = LinkCmdModFirstResp
	}
	rr := &manageRecordRequest{command: command, link: newLink}
	payload, _ := rr.MarshalBinary()
	resp, err := plm.send(&Packet{Command: CmdManageAllLinkRecord, Payload: payload})

	if resp.NAK() {
		err = fmt.Errorf("Failed to add link back to ALDB")
	}

	return err
}

func (plm *PLM) WriteLink(*insteon.LinkRecord) error {
	return insteon.ErrNotImplemented
}

func (plm *PLM) AppendLink(*insteon.LinkRecord) error {
	return insteon.ErrNotImplemented
}

func (plm *PLM) Cleanup() (err error) {
	removeable := make([]*insteon.LinkRecord, 0)
	links, err := plm.Links()
	if err == nil {
		for i, l1 := range links {
			for _, l2 := range links[i+1:] {
				if l1.Equal(l2) {
					removeable = append(removeable, l2)
				}
			}
		}

		err = plm.RemoveLinks(removeable...)
	}
	return err
}

func (plm *PLM) AddManualLink(group insteon.Group) error {
	return plm.EnterLinkingMode(group)
}

func (plm *PLM) EnterLinkingMode(group insteon.Group) error {
	lr := &allLinkReq{Mode: linkingMode(0x03), Group: group}
	payload, _ := lr.MarshalBinary()
	ack, err := plm.Retry(&Packet{
		Command: CmdStartAllLink,
		Payload: payload,
	}, 3)

	if err == nil && ack.NAK() {
		err = ErrNak
	}
	return err
}

func (plm *PLM) ExitLinkingMode() error {
	ack, err := plm.Retry(&Packet{
		Command: CmdCancelAllLink,
	}, 3)

	if err == nil && ack.NAK() {
		err = ErrNak
	}
	return err
}

func (plm *PLM) EnterUnlinkingMode(group insteon.Group) error {
	lr := &allLinkReq{Mode: linkingMode(0xff), Group: group}
	payload, _ := lr.MarshalBinary()
	ack, err := plm.Retry(&Packet{
		Command: CmdStartAllLink,
		Payload: payload,
	}, 3)

	if err == nil && ack.NAK() {
		err = ErrNak
	}
	return err
}

func (plm *PLM) AssignToAllLinkGroup(insteon.Group) error { return ErrNotImplemented }

func (plm *PLM) DeleteFromAllLinkGroup(insteon.Group) error { return ErrNotImplemented }
