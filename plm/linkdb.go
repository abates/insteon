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

func (alr *allLinkReq) UnmarshalBinary(buf []byte) (err error) {
	if len(buf) < 2 {
		err = &insteon.BufError{insteon.ErrBufferTooShort, 2, len(buf)}
	} else {
		alr.Mode = linkingMode(buf[0])
		alr.Group = insteon.Group(buf[1])
	}
	return err
}

func (alr *allLinkReq) String() string {
	return fmt.Sprintf("%02x %d", alr.Mode, alr.Group)
}

type senderReceiver interface {
	send(*Packet) (*Packet, error)
	ReadPacket() (*Packet, error)
}

type linkdb struct {
	age     time.Time
	links   []*insteon.LinkRecord
	plm     senderReceiver
	timeout time.Duration
}

func (ldb *linkdb) old() bool {
	return ldb.age.Add(ldb.timeout).Before(time.Now())
}

func (ldb *linkdb) refresh() error {
	if !ldb.old() {
		return nil
	}
	links := make([]*insteon.LinkRecord, 0)
	_, err := ldb.plm.send(&Packet{Command: CmdGetFirstAllLink})
	for err == nil {
		var pkt *Packet
		pkt, err = ldb.plm.ReadPacket()
		if err == nil {
			if pkt.Command == CmdAllLinkRecordResp {
				link := &insteon.LinkRecord{}
				err = link.UnmarshalBinary(pkt.Payload)
				if err == nil {
					links = append(links, link)
					_, err = ldb.plm.send(&Packet{Command: CmdGetNextAllLink})
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

func (ldb *linkdb) Links() ([]*insteon.LinkRecord, error) {
	err := ldb.refresh()
	return ldb.links, err
}

func (ldb *linkdb) WriteLinks(...*insteon.LinkRecord) error {
	return insteon.ErrNotImplemented
}

func (ldb *linkdb) UpdateLinks(...*insteon.LinkRecord) error {
	return insteon.ErrNotImplemented
}

func (ldb *linkdb) EnterLinkingMode(group insteon.Group) error {
	lr := &allLinkReq{Mode: linkingMode(0x03), Group: group}
	payload, _ := lr.MarshalBinary()
	_, err := ldb.plm.send(&Packet{Command: CmdStartAllLink, Payload: payload})
	if err == nil {
		time.Sleep(insteon.PropagationDelay(3, true))
	}
	return err
}

func (ldb *linkdb) ExitLinkingMode() error {
	_, err := ldb.plm.send(&Packet{Command: CmdCancelAllLink})
	return err
}

func (ldb *linkdb) EnterUnlinkingMode(group insteon.Group) error {
	lr := &allLinkReq{Mode: linkingMode(0xff), Group: group}
	payload, _ := lr.MarshalBinary()
	_, err := ldb.plm.send(&Packet{Command: CmdStartAllLink, Payload: payload})
	if err == nil {
		time.Sleep(insteon.PropagationDelay(3, true))
	}
	return err
}
