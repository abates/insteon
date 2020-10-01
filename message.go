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
	"fmt"
	"strings"
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
	MsgTypeDirect            MessageType = 0    // D
	MsgTypeDirectAck                     = 0x20 // D (Ack)
	MsgTypeDirectNak                     = 0xA0 // D (Nak)
	MsgTypeAllLinkCleanup                = 0x40 // C
	MsgTypeAllLinkCleanupAck             = 0x60 // C (Ack)
	MsgTypeAllLinkCleanupNak             = 0xE0 // C (Nak)
	MsgTypeBroadcast                     = 0x80 // B
	MsgTypeAllLinkBroadcast              = 0xC0 // A
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
	return m&0x80 > 0 && m&0x20 == 0
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

// Flag allows building of MessageFlags from component parts.
func Flag(messageType MessageType, extended bool, hopsLeft, maxHops uint8) Flags {
	if hopsLeft > 3 || maxHops > 3 {
		return 0
	}
	var e uint8
	if extended {
		e = 1
	}

	return Flags(uint8(messageType) | e<<4 | hopsLeft<<2 | maxHops)
}

// Ack indicates if the message was an acknowledgement
func (f Flags) Ack() bool { return f&0x20 == 0x20 }

// Nak indicates if the message was a negative-acknowledgement
func (f Flags) Nak() bool { return f.Ack() && (f&0x80 == 0x80) }

// Type will return the MessageType of the flags
func (f Flags) Type() MessageType { return MessageType(f & 0xe0) }

// Standard will indicate if the insteon message is standard length
func (f Flags) Standard() bool { return f&0x10 == 0x00 }

// Extended will indicate if the insteon message is extended length
func (f Flags) Extended() bool { return f&0x10 == 0x10 }

// TTL is the remaining number of times an insteon message will be
// retransmitted. This is decremented each time a message is repeated
func (f Flags) TTL() uint8 { return uint8((f & 0x0f) >> 2) }

func (f *Flags) SetTTL(ttl uint8) {

}

// MaxTTL is the maximum number of times a message can be repeated
func (f Flags) MaxTTL() uint8 { return uint8(f & 0x03) }

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

// Ack indicates if the message is an acknowledgement of a previously sent
// message
func (m *Message) Ack() bool {
	return m.Flags.Ack() && !m.Flags.Nak()
}

// Nak indicates a negative acknowledgement.  This indicates the device
// is rejecting a previously sent command
func (m *Message) Nak() bool {
	return m.Flags.Nak()
}

// Broadcast indicates if the message is a broadcast message, as
// opposed to a direct message (sent directly to the local device)
func (m *Message) Broadcast() bool {
	return m.Flags.Type().Broadcast()
}

// MarshalBinary will convert the Message to a byte slice appropriate for
// sending out onto the insteon network
func (m *Message) MarshalBinary() (data []byte, err error) {
	data = make([]byte, StandardMsgLen)
	copy(data[0:3], m.Src[:])
	copy(data[3:6], m.Dst[:])
	data[6] = byte(m.Flags)
	data[7] = byte(m.Command.Command1())
	data[8] = byte(m.Command.Command2())
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
	// magic numbers, oh la la
	// data[6] is the flags field of the message which contains
	// the message flags as well as the max hops and hops left information
	// 0xe0 masks the 3 most significant bits, which are the actual message flags
	// 0xa0 - Direct NAK
	// 0xe0 - All-Link Cleanup NAK
	// The command lookup won't have the NAK bit set, so the first conditional
	// sets the command without the NAK bit (masking it with 0x70 instead of 0xf0)
	if m.Flags.Nak() {
		m.Command = Command(int(0x70&data[6])<<12 | int(data[7])<<8 | int(data[8]))
	} else if m.Flags.Type() == MsgTypeAllLinkCleanup {
		m.Command = Command(int(data[7])<<8 | int(data[8]))
	} else {
		m.Command = Command(int(0xf0&data[6])<<12 | int(data[7])<<8 | int(data[8]))
	}

	if m.Flags.Extended() {
		if len(data) < ExtendedMsgLen {
			return newBufError(ErrBufferTooShort, ExtendedMsgLen, len(data))
		}
		m.Payload = make([]byte, 14)
		copy(m.Payload, data[9:])
	}
	return err
}

func (m *Message) String() (str string) {
	if m.Flags.Type() == MsgTypeAllLinkBroadcast {
		str = sprintf("%s %s -> ff.ff.ff", m.Flags, m.Src)
	} else if m.Flags.Type() == MsgTypeBroadcast {
		devCat := DevCat{m.Dst[0], m.Dst[1]}
		firmware := FirmwareVersion(m.Dst[2])

		str = sprintf("%s %s -> ff.ff.ff DevCat %v Firmware %v", m.Flags, m.Src, devCat, firmware)
	} else if m.Flags.Type() == MsgTypeAllLinkCleanup {
		str = sprintf("%s %s -> %s Cleanup", m.Flags, m.Src, m.Dst)
	} else {
		str = sprintf("%s %s -> %s", m.Flags, m.Src, m.Dst)
	}

	// don't print the command in an ACK message because it doesn't
	// directly correspond to the command map that we have.  Return
	// commands can't really be looked up because the Command2 byte
	// might be different, or the ack might be a standard length
	// message when the request was extended length.  In any case,
	// much of the time, the command lookup on an ack message may
	// return a CommandByte that has an incorrect command name
	if m.Ack() {
		str = sprintf("%s %d.%d", str, m.Command.Command1(), m.Command.Command2())
	} else {
		str = sprintf("%s %v", str, m.Command)
	}

	if m.Flags.Type() == MsgTypeAllLinkBroadcast {
		str = sprintf("%s Group(%d)", str, m.Dst[2])
	}

	if m.Flags.Extended() {
		payloadStr := make([]string, len(m.Payload))
		for i, value := range m.Payload {
			payloadStr[i] = fmt.Sprintf("%02x", value)
		}
		str = sprintf("%s [%v]", str, strings.Join(payloadStr, " "))
	}
	return str
}
