package insteon

import (
	"fmt"
)

type RecordControlFlags byte

func (rcf *RecordControlFlags) setBit(pos uint) {
	*rcf |= (1 << pos)
}

func (rcf *RecordControlFlags) clearBit(pos uint) {
	*rcf &= ^(1 << pos)
}

func (rcf RecordControlFlags) InUse() bool      { return rcf&0x80 == 0x80 }
func (rcf *RecordControlFlags) setInUse()       { rcf.setBit(7) }
func (rcf RecordControlFlags) Available() bool  { return rcf&0x80 == 0x00 }
func (rcf *RecordControlFlags) setAvailable()   { rcf.clearBit(7) }
func (rcf RecordControlFlags) Controller() bool { return rcf&0x40 == 0x40 }
func (rcf *RecordControlFlags) setController()  { rcf.setBit(6) }
func (rcf RecordControlFlags) Responder() bool  { return rcf&0x40 == 0x00 }
func (rcf *RecordControlFlags) setResponder()   { rcf.clearBit(6) }

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

type Group byte

func (g Group) String() string { return fmt.Sprintf("%d", byte(g)) }

type LinkRecord struct {
	Flags   RecordControlFlags
	Group   Group
	Address Address
	Data    [3]byte
}

func (l *LinkRecord) String() string {
	return fmt.Sprintf("%s %s %s 0x%02x 0x%02x 0x%02x", l.Flags, l.Group, l.Address, l.Data[0], l.Data[1], l.Data[2])
}

func (l *LinkRecord) Equal(other *LinkRecord) bool {
	if l == other {
		return true
	}

	if l == nil || other == nil {
		return false
	}

	return l.Flags.InUse() == other.Flags.InUse() && l.Flags.Controller() == other.Flags.Controller() && l.Group == other.Group && l.Address == other.Address
}

func (l *LinkRecord) MarshalBinary() ([]byte, error) {
	data := make([]byte, 8)
	data[0] = byte(l.Flags)
	data[1] = byte(l.Group)
	copy(data[2:5], l.Address[:])
	copy(data[5:8], l.Data[:])
	return data, nil
}

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
