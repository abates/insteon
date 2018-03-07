package insteon

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// RecordControlFlags indicate whether a link record is a
// controller or responder and whether it is available or in
// use
type RecordControlFlags byte

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

func (rcf *RecordControlFlags) UnmarshalText(text []byte) error {
	str := strings.Split(string(text), "")
	if len(str) != 2 {
		return fmt.Errorf("Expected 2 characters got %d", len(str))
	}

	if str[0] == "A" {
		rcf.setAvailable()
	} else {
		rcf.setInUse()
	}

	if str[1] == "C" {
		rcf.setController()
	} else {
		rcf.setResponder()
	}
	return nil
}

// Group is the Insteon group to which the Link Record corresponds
type Group byte

// String representation of the group number
func (g Group) String() string { return sprintf("%d", byte(g)) }

func (g *Group) UnmarshalText(text []byte) error {
	value, err := strconv.Atoi(string(text))
	if err == nil {
		*g = Group(byte(value))
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

func (l *LinkRecord) MarshalText() ([]byte, error) {
	str := fmt.Sprintf("%-5s %5s %8s   %02x %02x %02x", l.Flags, l.Group, l.Address, l.Data[0], l.Data[1], l.Data[2])
	return []byte(str), nil
}

// UnmarshalBinary will convert the byte string received in a message
// request to a LinkRecord
func (l *LinkRecord) UnmarshalBinary(buf []byte) error {
	if len(buf) < 8 {
		return newBufError(ErrBufferTooShort, 8, len(buf))
	}
	l.Flags = RecordControlFlags(buf[0])
	l.Group = Group(buf[1])
	copy(l.Address[:], buf[2:5])
	copy(l.Data[:], buf[5:8])
	return nil
}

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
