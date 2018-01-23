package plm

import (
	"fmt"
)

type Packet struct {
	payload []byte
	Command Command
	Ack     byte
}

func (p *Packet) ACK() bool {
	if p.Command >= 0x60 {
		return p.Ack == 0x06
	}
	return false
}

func (p *Packet) NAK() bool {
	if p.Command >= 0x60 {
		return p.Ack == 0x15
	}
	return false
}

func (p *Packet) String() string {
	cmd := p.Command.String()
	if p.Command >= 0x60 {
		if p.ACK() {
			cmd = fmt.Sprintf("%s ACK", cmd)
		} else if p.NAK() {
			cmd = fmt.Sprintf("%s NAK", cmd)
		}
	}

	// TODO dump payload
	return fmt.Sprintf("%-24s %s", cmd, hexDump(p.payload))
}

func (p *Packet) MarshalBinary() (buf []byte, err error) {
	buf = make([]byte, 2)
	buf[0] = 0x02
	buf[1] = byte(p.Command)
	if len(p.payload) > 0 {
		buf = append(buf, p.payload...)
	}
	return buf, err
}

func (p *Packet) UnmarshalBinary(buf []byte) (err error) {
	if buf[0] != 0x02 {
		return ErrNoSync
	}

	p.Command = Command(buf[1])
	buf = buf[2:]

	// responses to locally generated insteon messages need
	// some padding at the front since the source address
	// is removed
	if p.Command == CmdSendInsteonMsg {
		newBuf := make([]byte, len(buf)+3)
		copy(newBuf[3:], buf)
		buf = newBuf
	}

	if 0x60 <= p.Command && p.Command <= 0x7f {
		p.Ack = buf[len(buf)-1]
		buf = buf[0 : len(buf)-1]
	}

	p.payload = make([]byte, len(buf))
	copy(p.payload, buf)

	return err
}
