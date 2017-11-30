package insteon

import (
	"fmt"
)

const (
	StandardMsgLen        = 9
	ExtendedMsgLen        = 23
	StandardDirectMessage = Flags(0x0f)
	ExtendedDirectMessage = Flags(0x1f)
)

type MessageType int

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

type Flags byte

func (f Flags) Type() MessageType { return MessageType((f & 0xf0) >> 5) }
func (f Flags) IsStandard() bool  { return f&0x10 == 0x00 }
func (f Flags) IsExtended() bool  { return f&0x10 == 0x10 }
func (f Flags) TTL() int          { return int((f & 0x0f) >> 2) }
func (f Flags) MaxTTL() int       { return int(f & 0x03) }
func (f Flags) String() string {
	msg := "S"
	if f.IsExtended() {
		msg = "E"
	}

	return fmt.Sprintf("%s%-5s %d:%d", msg, f.Type(), f.MaxTTL(), f.TTL())
}

const ()

type Message struct {
	Src     Address
	Dst     Address
	Flags   Flags
	Command *Command
	Payload Payload
}

func (m *Message) String() string {
	str := fmt.Sprintf("%s -> %s %s %s", m.Src, m.Dst, m.Flags, m.Command)
	if m.Flags.IsExtended() {
		str = fmt.Sprintf("%s %v", str, m.Payload)
	}
	return str
}

func (m *Message) MarshalBinary() (data []byte, err error) {
	data = make([]byte, StandardMsgLen)
	copy(data[0:3], m.Src[:])
	copy(data[3:6], m.Dst[:])
	data[6] = byte(m.Flags)
	copy(data[7:9], m.Command.cmd[:])
	if m.Flags.IsExtended() {
		var payload []byte
		payload, err = m.Payload.MarshalBinary()
		data = append(data, payload...)
	}
	return data, err
}

func (m *Message) UnmarshalBinary(data []byte) (err error) {
	// The CRC is not always present
	if len(data) < StandardMsgLen {
		return ErrBufferTooShort
	}
	copy(m.Src[:], data[0:3])
	copy(m.Dst[:], data[3:6])
	m.Flags = Flags(data[6])
	if m.Flags.IsStandard() {
		m.Command = Commands.FindStd(data[7:9])
	} else {
		m.Command = Commands.FindExt(data[7:9])
	}
	if m.Flags.IsExtended() {
		if len(data) < ExtendedMsgLen {
			return ErrExtendedBufferTooShort
		}
		payload := m.Command.generator()
		err = payload.UnmarshalBinary(data[9:23])
		m.Payload = payload
	}
	return err
}
