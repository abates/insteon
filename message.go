package insteon

import (
	"fmt"
)

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
	StandardDirectMessage = Flags(0x0a)
	StandardDirectAck     = Flags(0x2a)
	StandardDirectNak     = Flags(0xaa)
	ExtendedDirectMessage = Flags(0x1a)
	ExtendedDirectAck     = Flags(0x3a)
	ExtendedDirectNak     = Flags(0xba)
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

	return fmt.Sprintf("%s%-5s %d:%d", msg, f.Type(), f.MaxTTL(), f.TTL())
}

// Message is a single insteon message
type Message struct {
	DevCat  Category
	version EngineVersion
	Src     Address
	Dst     Address
	Flags   Flags
	Command *Command
	Payload Payload
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

func (m *Message) String() (str string) {
	if m.Broadcast() {
		str = fmt.Sprintf("%s -> ff.ff.ff %s %s", m.Src, m.Flags, m.Command)
	} else {
		str = fmt.Sprintf("%s -> %s %s %s", m.Src, m.Dst, m.Flags, m.Command)
		if m.Flags.Extended() {
			str = fmt.Sprintf("%s %v", str, m.Payload)
		}
	}
	return str
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
	copy(data[7:9], m.Command.Cmd[:])
	if m.Flags.Extended() {
		var payload []byte
		payload, err = m.Payload.MarshalBinary()
		data = append(data, payload...)
	}

	if m.Flags.Extended() && m.version == VerI2Cs {
		data[len(data)-1] = checksum(data[7:])
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
	if m.Flags.Standard() {
		m.Command = Commands.FindStd(m.DevCat.Category(), m.Flags.Type(), data[7:9])
	} else {
		m.Command = Commands.FindExt(m.DevCat.Category(), m.Flags.Type(), data[7:9])
	}

	if m.Flags.Extended() {
		if len(data) < ExtendedMsgLen {
			return newBufError(ErrBufferTooShort, ExtendedMsgLen, len(data))
		}
		payload := m.Command.generator()
		err = payload.UnmarshalBinary(data[9:23])
		m.Payload = payload
	}
	return err
}
