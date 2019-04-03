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

package insteon

const (
	// BaseLinkDBAddress is the base address of devices All-Link database
	BaseLinkDBAddress = MemAddress(0x0fff)

	// LinkRecordSize is the syze, in bytes, of a single All-Link record
	LinkRecordSize = MemAddress(8)
)

// MemAddress is an integer representing a specific location in a device's memory
type MemAddress int

func (ma MemAddress) String() string {
	return sprintf("%02x.%02x", byte(ma>>8), byte(ma&0xff))
}

// There are two link request types, one used to read the link database and
// one used to write links
const (
	readLink     linkRequestType = 0x00
	linkResponse linkRequestType = 0x01
	writeLink    linkRequestType = 0x02
)

// linkRequestType is used to indicate whether an ALDB request is for reading
// or writing the database
type linkRequestType byte

func (lrt linkRequestType) String() string {
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

// linkRequest is the message sent to a device to request reading or writing
// all-link database records
type linkRequest struct {
	Type       linkRequestType
	MemAddress MemAddress
	NumRecords int
	Link       *LinkRecord
}

func (lr *linkRequest) String() string {
	if lr.Link == nil {
		return sprintf("%s %s %d", lr.Type, lr.MemAddress, lr.NumRecords)
	}
	return sprintf("%s %s %d %s", lr.Type, lr.MemAddress, lr.NumRecords, lr.Link)
}

// UnmarshalBinary will take the byte slice and convert it to a LinkRequest object
func (lr *linkRequest) UnmarshalBinary(buf []byte) (err error) {
	if len(buf) < 5 {
		return newBufError(ErrBufferTooShort, 6, len(buf))
	}
	lr.Type = linkRequestType(buf[1])
	lr.MemAddress = MemAddress(buf[2]) << 8
	lr.MemAddress |= MemAddress(buf[3])

	switch lr.Type {
	case 0x00:
		lr.NumRecords = int(buf[4])
	case 0x01:
		lr.Link = &LinkRecord{}
	case 0x02:
		lr.NumRecords = int(buf[4])
		lr.Link = &LinkRecord{}
	}

	if lr.Link != nil {
		err = lr.Link.UnmarshalBinary(buf[5:])
		lr.Link.memAddress = lr.MemAddress
	}
	return err
}

// MarshalBinary will convert the LinkRequest to a byte slice appropriate for
// sending out to the insteon network
func (lr *linkRequest) MarshalBinary() (buf []byte, err error) {
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
