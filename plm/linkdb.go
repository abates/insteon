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
	cmd  recordRequestCommand
	link *insteon.LinkRecord
}

func (mrr *manageRecordRequest) String() string {
	return fmt.Sprintf("%02x %s", mrr.cmd, mrr.link)
}

func (mrr *manageRecordRequest) MarshalBinary() ([]byte, error) {
	payload, err := mrr.link.MarshalBinary()
	buf := make([]byte, len(payload)+1)
	buf[0] = byte(mrr.cmd)
	copy(buf[1:], payload)
	return buf, err
}

func (mrr *manageRecordRequest) UnmarshalBinary(buf []byte) error {
	mrr.cmd = recordRequestCommand(buf[0])
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

func (alr *allLinkReq) UnmarshalBinary(buf []byte) (err error) {
	if len(buf) < 2 {
		err = fmt.Errorf("%w wanted 2 got %d", insteon.ErrBufferTooShort, len(buf))
	} else {
		alr.Mode = linkingMode(buf[0])
		alr.Group = insteon.Group(buf[1])
	}
	return err
}

func (alr *allLinkReq) String() string {
	return fmt.Sprintf("%02x %d", alr.Mode, alr.Group)
}

type linkdb struct {
	age     time.Time
	links   []insteon.LinkRecord
	plm     PacketWriter
	retries int
	timeout time.Duration
}

func (ldb *linkdb) old() bool {
	return ldb.age.Add(ldb.timeout).Before(time.Now())
}

func (ldb *linkdb) refresh() error {
	if !ldb.old() {
		return nil
	}
	links := make([]insteon.LinkRecord, 0)
	_, err := RetryWriter(ldb.plm, ldb.retries, true).WritePacket(&Packet{Command: CmdGetFirstAllLink})
	for err == nil {
		var pkt *Packet
		pkt, err = ldb.plm.ReadPacket()
		if err == nil {
			if pkt.Command == CmdAllLinkRecordResp {
				link := insteon.LinkRecord{}
				err = link.UnmarshalBinary(pkt.Payload)
				if err == nil {
					links = append(links, link)
					_, err = RetryWriter(ldb.plm, ldb.retries, false).WritePacket(&Packet{Command: CmdGetNextAllLink})
				}
			}
		}
	}

	if err == ErrNak {
		err = nil
		ldb.links = links
	}
	return err
}

func (ldb *linkdb) Links() ([]insteon.LinkRecord, error) {
	err := ldb.refresh()
	links := make([]insteon.LinkRecord, len(ldb.links))
	copy(links, ldb.links)
	return links, err
}

func (ldb *linkdb) deleteLink(link *insteon.LinkRecord) (*Packet, error) {
	mrr := &manageRecordRequest{
		cmd:  LinkCmdDeleteFirst,
		link: link,
	}
	payload, _ := mrr.MarshalBinary()
	return RetryWriter(ldb.plm, 3, true).WritePacket(&Packet{Command: CmdManageAllLinkRecord, Payload: payload})
}

func (ldb *linkdb) writeLink(link *insteon.LinkRecord) (*Packet, error) {
	mrr := &manageRecordRequest{
		cmd:  LinkCmdModFirstResp,
		link: link,
	}
	if link.Flags.Controller() {
		mrr.cmd = LinkCmdModFirstCtrl
	}
	payload, _ := mrr.MarshalBinary()
	return RetryWriter(ldb.plm, 3, true).WritePacket(&Packet{Command: CmdManageAllLinkRecord, Payload: payload})
}

func (ldb *linkdb) WriteLinks(newLinks ...insteon.LinkRecord) (err error) {
	if err == nil {
		// blow away the database and pray something doesn't go wrong, it's really
		// too bad that the PLM doesn't support some form of transactions or modifying
		// ALDB records by address
		for i := 0; i < len(ldb.links) && err == nil; i++ {
			_, err = ldb.deleteLink(&ldb.links[i])
		}
	}

	if err == nil {
		ldb.links = nil

		for _, link := range newLinks {
			var ack *Packet
			ack, err = ldb.writeLink(&link)
			//_, err = ldb.plm.Write(&Packet{Command: CmdManageAllLinkRecord, Payload: payload})
			if err == nil {
				ldb.links = append(ldb.links, link)
			} else if err == ErrWrongPayload {
				// For some reason (at least with my PLM) it is common for
				// a link record to be shifted over a few bytes after sending
				// it to the PLM.  This can be detected because the ACK payload
				// won't match the transmitted packet.  We try to fix this by
				// deleting the corrupted record and re-adding it
				delLink := &insteon.LinkRecord{}
				delLink.UnmarshalBinary(ack.Payload)
				_, err = ldb.deleteLink(delLink)
				if err == nil {
					// try again
					_, err = ldb.writeLink(&link)
					if err == nil {
						ldb.links = append(ldb.links, link)
					}
				}
			}

			if err != nil {
				break
			}
		}
	}

	return err
}

func (ldb *linkdb) UpdateLinks(...insteon.LinkRecord) error {
	return insteon.ErrNotImplemented
}

func (ldb *linkdb) EnterLinkingMode(group insteon.Group) error {
	lr := &allLinkReq{Mode: linkingMode(0x03), Group: group}
	payload, _ := lr.MarshalBinary()
	_, err := ldb.plm.WritePacket(&Packet{Command: CmdStartAllLink, Payload: payload})
	return err
}

func (ldb *linkdb) ExitLinkingMode() error {
	_, err := ldb.plm.WritePacket(&Packet{Command: CmdCancelAllLink})
	return err
}

func (ldb *linkdb) EnterUnlinkingMode(group insteon.Group) error {
	lr := &allLinkReq{Mode: linkingMode(0xff), Group: group}
	payload, _ := lr.MarshalBinary()
	_, err := ldb.plm.WritePacket(&Packet{Command: CmdStartAllLink, Payload: payload})
	return err
}
