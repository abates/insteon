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

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// RecordControlFlags indicate whether a link record is a
// controller or responder and whether it is available or in
// use
type RecordControlFlags byte

// RecordControlFlags indicating the different availability/type of
// link records
const (
	AvailableController   = RecordControlFlags(0x40)
	UnavailableController = RecordControlFlags(0xc0)
	AvailableResponder    = RecordControlFlags(0x00)
	UnavailableResponder  = RecordControlFlags(0x80)
)

func (rcf *RecordControlFlags) setBit(pos uint) {
	*rcf |= (1 << pos)
}

func (rcf *RecordControlFlags) clearBit(pos uint) {
	*rcf &= ^(1 << pos)
}

// InUse indicates if a link record is currently in use by the device
func (rcf RecordControlFlags) InUse() bool { return rcf&0x80 == 0x80 }
func (rcf *RecordControlFlags) setInUse()  { rcf.setBit(7) }

// Available indicates if a link record is available and can be
// overwritten by a new record. Available is synonmous with "deleted"
func (rcf RecordControlFlags) Available() bool { return rcf&0x80 == 0x00 }
func (rcf *RecordControlFlags) setAvailable()  { rcf.clearBit(7) }

// Controller indicates that the device is a controller for the device in the
// link record
func (rcf RecordControlFlags) Controller() bool { return rcf&0x40 == 0x40 }
func (rcf *RecordControlFlags) setController()  { rcf.setBit(6) }

// Responder indicates that the device is a reponder to the device listed in
// the link record
func (rcf RecordControlFlags) Responder() bool { return rcf&0x40 == 0x00 }
func (rcf *RecordControlFlags) setResponder()  { rcf.clearBit(6) }

// String will be "A" or "U" (available or in use) followed by "C" or
// "R" (controller or responder). This string will always be two
// characters wide
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

// UnmarshalText takes a two character input string and converts it
// to the correct RecordControlFlags.  The first character can be
// either "A" for available or "U" for unavailable (in use) and the
// second character is either "C" for controller or "R" for responder
func (rcf *RecordControlFlags) UnmarshalText(text []byte) (err error) {
	str := strings.Split(string(text), "")
	if len(str) != 2 {
		return fmt.Errorf("Expected 2 characters got %d", len(str))
	}

	if str[0] == "A" {
		rcf.setAvailable()
	} else if str[0] == "U" {
		rcf.setInUse()
	} else {
		err = errors.New("Invalid value for Available flag")
	}

	if str[1] == "C" {
		rcf.setController()
	} else if str[1] == "R" {
		rcf.setResponder()
	} else {
		err = errors.New("Invalid value for Controller flag")
	}
	return err
}

// Group is the Insteon group to which the Link Record corresponds
type Group byte

// String representation of the group number
func (g Group) String() string { return sprintf("%d", byte(g)) }

// UnmarshalText takes an input string and converts
// it to its Group equivalent.  The decimal input value
// must be positive and less than 256
func (g *Group) UnmarshalText(text []byte) error {
	value, err := strconv.Atoi(string(text))
	if err == nil {
		if 0 < value && value < 256 {
			*g = Group(byte(value))
		} else {
			err = errors.New("valid groups are between 1 and 255 (inclusive)")
		}
	} else {
		err = errors.New("invalid number format")
	}
	return err
}

// LinkRecord is a single All-Link record in an All-Link database
type LinkRecord struct {
	memAddress MemAddress
	Flags      RecordControlFlags
	Group      Group
	Address    Address
	Data       [3]byte
}

// String converts the LinkRecord to a human readable string that looks similar to:
//    UR        1 01.02.03   00 1c 01
func (l *LinkRecord) String() string {
	return sprintf("%s %s %s 0x%02x 0x%02x 0x%02x", l.Flags, l.Group, l.Address, l.Data[0], l.Data[1], l.Data[2])
}

// Equal will determine if another LinkRecord is equivalent. The records are
// equivalent if they both have the same availability, type (controller/responder)
// and address
func (l *LinkRecord) Equal(other *LinkRecord) bool {
	if l == other {
		return true
	}

	if l == nil || other == nil {
		return false
	}

	return l.Flags.InUse() == other.Flags.InUse() && l.Flags.Controller() == other.Flags.Controller() && l.Group == other.Group && l.Address == other.Address
}

// MarshalBinary converts the link-record to a byte string that can be
// used in a record request
func (l *LinkRecord) MarshalBinary() ([]byte, error) {
	data := make([]byte, 8)
	data[0] = byte(l.Flags)
	data[1] = byte(l.Group)
	copy(data[2:5], l.Address[:])
	copy(data[5:8], l.Data[:])
	return data, nil
}

// MarshalText will convert the LinkRecord to a text string that can be
// used as input to the UnmarshalText. This is useful in allowing a user
// to manuall edit link records
func (l *LinkRecord) MarshalText() ([]byte, error) {
	str := sprintf("%-5s %5s %8s   %02x %02x %02x", l.Flags, l.Group, l.Address, l.Data[0], l.Data[1], l.Data[2])
	return []byte(str), nil
}

// UnmarshalBinary will convert the byte string received in a message
// request to a LinkRecord
func (l *LinkRecord) UnmarshalBinary(buf []byte) (err error) {
	if len(buf) < 8 {
		return newBufError(ErrBufferTooShort, 8, len(buf))
	}
	l.Flags = RecordControlFlags(buf[0])
	if buf[0] == 0x00 {
		err = ErrEndOfLinks
	}

	l.Group = Group(buf[1])
	copy(l.Address[:], buf[2:5])
	copy(l.Data[:], buf[5:8])
	return err
}

// UnmarshalText takes an input text string and assigns the values
// to the RecordControlFlags receiver.  The input text string
// should be in the following form:
//    Flags Group Address    Data
//    UR        1 01.02.03   00 1c 01
// Each field is unmarshaled using the corresponding type's
// UnmarshalText functions
func (l *LinkRecord) UnmarshalText(buf []byte) (err error) {
	fields := bytes.Fields(buf)
	if len(fields) != 6 {
		err = fmt.Errorf("Expected 6 fields got %d", len(fields))
	}

	if err == nil {
		err = l.Flags.UnmarshalText(fields[0])
	}

	if err == nil {
		err = l.Group.UnmarshalText(fields[1])
	}

	if err == nil {
		err = l.Address.UnmarshalText(fields[2])
	}

	for i := 0; i < 3 && err == nil; i++ {
		_, err = fmt.Sscanf(string(fields[3+i]), "%x", &l.Data[i])
	}
	return
}
