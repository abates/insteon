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

package devices

import (
	"fmt"
	"time"

	"github.com/abates/insteon"
	"github.com/abates/insteon/commands"
)

const (
	// BaseLinkDBAddress is the base address of devices All-Link database
	BaseLinkDBAddress = MemAddress(0x0fff)

	// LinkRecordSize is the size, in bytes, of a single All-Link record
	LinkRecordSize = MemAddress(8)
)

var (
	// MaxLinkDbAge is the amount of time to wait until the local link database
	// is considered old
	MaxLinkDbAge = time.Hour
)

// MemAddress is an integer representing a specific location in a device's memory
type MemAddress int

func (ma MemAddress) String() string {
	return fmt.Sprintf("%02x.%02x", byte(ma>>8), byte(ma&0xff))
}

// There are two link request types, one used to read the link database and
// one used to write links
const (
	readLink     LinkRequestType = 0x00
	linkResponse LinkRequestType = 0x01
	writeLink    LinkRequestType = 0x02
)

// LinkRequestType is used to indicate whether an ALDB request is for reading
// or writing the database
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

// LinkRequest is the message sent to a device to request reading or writing
// all-link database records
type LinkRequest struct {
	Type       LinkRequestType
	MemAddress MemAddress
	NumRecords int
	Link       *insteon.LinkRecord
}

func (lr *LinkRequest) String() string {
	if lr.Link == nil {
		return fmt.Sprintf("%s %s %d", lr.Type, lr.MemAddress, lr.NumRecords)
	}
	return fmt.Sprintf("%s %s %d %s", lr.Type, lr.MemAddress, lr.NumRecords, lr.Link)
}

// UnmarshalBinary will take the byte slice and convert it to a LinkRequest object
func (lr *LinkRequest) UnmarshalBinary(buf []byte) (err error) {
	if len(buf) < 5 {
		return fmt.Errorf("%w: wanted 6 bytes got %d", insteon.ErrBufferTooShort, len(buf))
	}

	lr.Type = LinkRequestType(buf[1])
	lr.MemAddress = MemAddress(buf[2]) << 8
	lr.MemAddress |= MemAddress(buf[3])

	switch lr.Type {
	case 0x00:
		lr.NumRecords = int(buf[4])
	case 0x01:
		lr.Link = &insteon.LinkRecord{}
	case 0x02:
		lr.NumRecords = int(buf[4])
		lr.Link = &insteon.LinkRecord{}
	default:
		err = ErrInvalidResponse
	}

	if err == nil && lr.Link != nil {
		err = lr.Link.UnmarshalBinary(buf[5:])
	}

	return err
}

// MarshalBinary will convert the LinkRequest to a byte slice appropriate for
// sending out to the insteon network
func (lr *LinkRequest) MarshalBinary() (buf []byte, err error) {
	var linkData []byte
	buf = make([]byte, 14)
	buf[1] = byte(lr.Type)
	buf[2] = byte(lr.MemAddress >> 8)
	buf[3] = byte(lr.MemAddress & 0xff)
	switch lr.Type {
	case readLink:
		buf[4] = byte(lr.NumRecords)
	case linkResponse:
		buf[4] = 0x00
		linkData, err = lr.Link.MarshalBinary()
		copy(buf[5:], linkData)
	case writeLink:
		buf[4] = 0x08
		linkData, err = lr.Link.MarshalBinary()
		copy(buf[5:], linkData)
	}
	return buf, err
}

const (
	maxAge = time.Second * 10
)

type linkdb struct {
	MessageWriter
	age   time.Time
	links []insteon.LinkRecord
	index map[insteon.LinkID]int
}

func (ldb *linkdb) old() bool {
	return ldb.age.Add(MaxLinkDbAge).Before(time.Now())
}

func (ldb *linkdb) refresh() error {
	if !ldb.old() {
		return nil
	}
	ldb.index = make(map[insteon.LinkID]int)

	ldb.links = nil
	LogDebug.Printf("Retrieving Device link database")
	lastAddress := MemAddress(0)

	buf, _ := (&LinkRequest{Type: readLink, NumRecords: 0}).MarshalBinary()
	_, err := ldb.Write(&insteon.Message{Command: commands.ReadWriteALDB, Payload: buf})
	var msg *insteon.Message
	for err == nil {
		msg, err = Read(ldb, CmdMatcher(commands.ReadWriteALDB))
		if err == nil {
			lr := &LinkRequest{}
			err = lr.UnmarshalBinary(msg.Payload)
			// make sure there was no error unmarshalling, also make sure
			// that it's a new memory address.  Since insteon messages
			// are retransmitted, it is possible that the same ALDB response
			// will be received more than once
			if err == nil {
				if lr.MemAddress != lastAddress {
					lastAddress = lr.MemAddress
					if lr.Link.Flags.LastRecord() {
						break
					} else {
						ldb.links = append(ldb.links, *lr.Link)
						ldb.index[lr.Link.ID()] = len(ldb.links) - 1
					}
				}
			}
		}
	}

	if err == nil {
		ldb.age = time.Now()
	}
	return err
}

// Links will retrieve the link-database from the device and
// return a list of LinkRecords
func (ldb *linkdb) Links() (links []insteon.LinkRecord, err error) {
	err = ldb.refresh()
	if err == nil {
		links = make([]insteon.LinkRecord, len(ldb.links))
		copy(links, ldb.links)
	}
	return links, err
}

func (ldb *linkdb) AddLinks(addLinks ...insteon.LinkRecord) (err error) {
	if len(addLinks) == 0 {
		return nil
	}

	err = ldb.refresh()
	if err == nil {
		for i, link := range ldb.links {
			if link.Flags.Available() {
				err = ldb.WriteLink(i, link)
				if err == nil {
					addLinks[0].Flags.SetInUse()
					addLinks[0].Flags.ClearLastRecord()
					ldb.links[i] = addLinks[0]
				} else {
					break
				}
			}
			addLinks = addLinks[1:]
			if len(addLinks) == 0 {
				break
			}
		}

		if err == nil && len(addLinks) > 0 {
			for _, link := range addLinks {
				link.Flags.ClearLastRecord()
				err = ldb.WriteLink(len(ldb.links), link)
				if err != nil {
					break
				}
			}
		}
	}
	return err
}

func (ldb *linkdb) WriteLink(index int, link insteon.LinkRecord) (err error) {
	if index > len(ldb.links) {
		return ErrLinkIndexOutOfRange
	}
	memAddress := BaseLinkDBAddress - (MemAddress(index) * LinkRecordSize)
	buf, _ := (&LinkRequest{MemAddress: memAddress, Type: writeLink, Link: &link}).MarshalBinary()
	_, err = ldb.Write(&insteon.Message{Command: commands.ReadWriteALDB, Payload: buf})
	if err == nil {
		if link.Flags.LastRecord() {
			// if the last record comes before the end of the cached links then
			// slice the local list at the index
			if index < len(ldb.links) {
				ldb.links = ldb.links[0:index]
			}
		} else {
			// copy the link so it can't be modified outside of the database
			if index == len(ldb.links) {
				ldb.links = append(ldb.links, link)
			} else {
				ldb.links[index] = link
			}
		}
	}
	return err
}

func (ldb *linkdb) WriteLinks(links ...insteon.LinkRecord) (err error) {
	err = ldb.writeLinks(links...)
	return err
}

func (ldb *linkdb) writeLinks(links ...insteon.LinkRecord) (err error) {
	for i := 0; i < len(links) && err == nil; i++ {
		links[i].Flags.ClearLastRecord()
		err = ldb.WriteLink(i, links[i])
	}

	if err == nil {
		link := insteon.LinkRecord{}
		link.Flags.SetLastRecord()
		err = ldb.WriteLink(len(ldb.links), link)
		if err == nil {
			ldb.age = time.Now()
		}
	}
	return
}

func (ldb *linkdb) UpdateLinks(links ...insteon.LinkRecord) (err error) {
	err = ldb.refresh()

	if err == nil {
		for i := 0; err == nil && i < len(links); i++ {
			if j, found := ldb.index[links[i].ID()]; found {
				if ldb.links[j].Flags != links[i].Flags {
					err = ldb.WriteLink(i, links[i])
				}
				links = append(links[0:i], links[i+1:]...)
				i--
			}
		}

		for i := 0; err == nil && i < len(ldb.links); i++ {
			if ldb.links[i].Flags.Available() && len(links) > 0 {
				links[0].Flags.ClearLastRecord()
				err = ldb.WriteLink(i, links[0])
				if err == nil {
					links = links[1:]
				}
			}
		}

		if err == nil && len(links) > 0 {
			i := len(ldb.links)
			for _, link := range links {
				link.Flags.ClearLastRecord()
				err = ldb.WriteLink(i, link)
				i++
			}

			if err == nil {
				link := insteon.LinkRecord{}
				link.Flags.SetLastRecord()
				err = ldb.WriteLink(i, link)
			}
		}
	}

	return err
}
