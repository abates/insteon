package insteon

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

// Group is the Insteon group to which the Link Record corresponds
type Group byte

// String representation of the group number
func (g Group) String() string { return sprintf("%d", byte(g)) }

// LinkRecord is a single All-Link record in an All-Link database
type LinkRecord struct {
	Flags   RecordControlFlags
	Group   Group
	Address Address
	Data    [3]byte
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
