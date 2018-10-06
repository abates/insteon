package insteon

const (
	// StandardMsgLen is the length of an insteon standard message minus one byte (the crc byte)
	StandardMsgLen = 9

	// ExtendedMsgLen is the length of an insteon extended message minus one byte (the crc byte)
	ExtendedMsgLen = 23
)

// MessageType is an integer representing one of a list of the following types
type MessageType int

// All of the valid message types
const (
	MsgTypeDirect = iota
	MsgTypeDirectAck
	MsgTypeAllLinkCleanup
	MsgTypeAllLinkCleanupAck
	MsgTypeBroadcast
	MsgTypeDirectNak
	MsgTypeAllLinkBroadcast
	MsgTypeAllLinkCleanupNak
)

func (m MessageType) String() string {
	str := "unknown"
	switch m {
	case MsgTypeDirect:
		str = "D"
	case MsgTypeDirectAck:
		str = "D Ack"
	case MsgTypeAllLinkCleanup:
		str = "C"
	case MsgTypeAllLinkCleanupAck:
		str = "C Ack"
	case MsgTypeBroadcast:
		str = "B"
	case MsgTypeDirectNak:
		str = "D NAK"
	case MsgTypeAllLinkBroadcast:
		str = "A"
	case MsgTypeAllLinkCleanupNak:
		str = "C NAK"
	}

	return str
}

// Direct will indicate whether the MessageType represents a direct message
func (m MessageType) Direct() bool {
	return !m.Broadcast()
}

// Broadcast will indicate whether the MessageType represents a broadcast message
func (m MessageType) Broadcast() bool {
	return m&0x0f == 0x04 || m&0x0f == 0x06
}

// Flags for common message types
const (
	StandardBroadcast        = Flags(0x8a)
	StandardAllLinkBroadcast = Flags(0xca)
	StandardDirectMessage    = Flags(0x0a)
	StandardDirectAck        = Flags(0x2a)
	StandardDirectNak        = Flags(0xaa)
	ExtendedDirectMessage    = Flags(0x1a)
	ExtendedDirectAck        = Flags(0x3a)
	ExtendedDirectNak        = Flags(0xba)
)

// Flags is the flags byte in an insteon message
type Flags byte

// Type will return the MessageType of the flags
func (f Flags) Type() MessageType { return MessageType((f & 0xf0) >> 5) }

// Standard will indicate if the insteon message is standard length
func (f Flags) Standard() bool { return f&0x10 == 0x00 }

// Extended will indicate if the insteon message is extended length
func (f Flags) Extended() bool { return f&0x10 == 0x10 }

// TTL is the remaining number of times an insteon message will be
// retransmitted. This is decremented each time a message is repeated
func (f Flags) TTL() int { return int((f & 0x0f) >> 2) }

// MaxTTL is the maximum number of times a message can be repeated
func (f Flags) MaxTTL() int { return int(f & 0x03) }

func (f Flags) String() string {
	msg := "S"
	if f.Extended() {
		msg = "E"
	}

	return sprintf("%s%-5s %d:%d", msg, f.Type(), f.MaxTTL(), f.TTL())
}

// Message is a single insteon message
type Message struct {
	Src     Address
	Dst     Address
	Flags   Flags
	Command Command
	Payload []byte
}

func (m *Message) Ack() bool {
	return m.Flags&0xe0 == 0x20 || m.Flags&m.Flags&0xc0 == 0xa0
}

func (m *Message) Nak() bool {
	return m.Flags&0xe0 == 0xe0 || m.Flags&m.Flags&0xe0 == 0xa0
}

func (m *Message) Broadcast() bool {
	return m.Flags.Type().Broadcast()
}

func checksum(buf []byte) byte {
	sum := byte(0)
	for _, b := range buf {
		sum += b
	}
	return ^sum + 1
}

// MarshalBinary will convert the Message to a byte slice appropriate for
// sending out onto the insteon network
func (m *Message) MarshalBinary() (data []byte, err error) {
	data = make([]byte, StandardMsgLen)
	copy(data[0:3], m.Src[:])
	copy(data[3:6], m.Dst[:])
	data[6] = byte(m.Flags)
	data[7] = byte(m.Command >> 8)
	data[8] = byte(m.Command)
	if m.Flags.Extended() {
		data = append(data, make([]byte, 14)...)
		copy(data[9:23], m.Payload)
	}

	return data, err
}

// UnmarshalBinary will take a byte slice and unmarshal it into the Message
// fields
func (m *Message) UnmarshalBinary(data []byte) (err error) {
	// The CRC is not always present
	if len(data) < StandardMsgLen {
		return newBufError(ErrBufferTooShort, StandardMsgLen, len(data))
	}
	copy(m.Src[:], data[0:3])
	copy(m.Dst[:], data[3:6])
	m.Flags = Flags(data[6])
	m.Command = Command(data[7])<<8 | Command(data[8])

	if m.Flags.Extended() {
		if len(data) < ExtendedMsgLen {
			return newBufError(ErrBufferTooShort, ExtendedMsgLen, len(data))
		}
		m.Payload = make([]byte, 14)
		copy(m.Payload, data[9:])
	}
	return err
}

/*
func DumpMessage(m *Message) (str string) {
	if m.Broadcast() {
		if m.Flags.Type() == MsgTypeAllLinkBroadcast {
			str = sprintf("%s -> ff.ff.ff %v Group(%d)", m.Src, m.Flags, m.Dst[2])
		} else {
			devCat := DevCat{m.Dst[0], m.Dst[1]}
			firmware := FirmwareVersion(m.Dst[2])

			str = sprintf("%s -> ff.ff.ff %v DevCat %v Firmware %v", m.Src, m.Flags, devCat, firmware)
		}
	} else {
		str = sprintf("%s -> %s %v", m.Src, m.Dst, m.Flags)
	}

	// don't print the command in an ACK message because it doesn't
	// directly correspond to the command map that we have.  Return
	// commands can't really be looked up because the Command2 byte
	// might be different, or the ack might be a standard length
	// message when the request was extended length.  In any case,
	// much of the time, the command lookup on an ack message may
	// return a CommandByte that has an incorrect command name
	if !m.Ack() {
		str = sprintf("%s %v", str, m.Command)
	}

	if m.Flags.Extended() {
		payloadStr := make([]string, len(m.Payload))
		for i, value := range m.Payload {
			payloadStr[i] = fmt.Sprintf("%02x", value)
		}
		str = sprintf("%s [%v]", str, strings.Join(payloadStr, " "))
	}
	return str
}*/
