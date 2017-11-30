package plm

import (
	"fmt"

	"github.com/abates/insteon"
)

type Packet struct {
	retryCount int
	Command    Command
	Payload    insteon.Payload
	Ack        byte
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

	return fmt.Sprintf("%-24s %v", cmd, p.Payload)
}

func (p *Packet) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 2)
	buf[0] = 0x02
	buf[1] = byte(p.Command)
	payload, err := p.Payload.MarshalBinary()
	if _, ok := p.Payload.(*insteon.Message); ok {
		// slice off the source address
		payload = payload[3:]
	}
	buf = append(buf, payload...)
	return buf, err
}

func (p *Packet) UnmarshalBinary(buf []byte) (err error) {
	if buf[0] != 0x02 {
		return ErrNoSync
	}

	p.Command = Command(buf[1])
	switch {
	case p.Command == CmdStdMsgReceived || p.Command == CmdExtMsgReceived:
		msg := &insteon.Message{}
		err = msg.UnmarshalBinary(buf[2:])
		p.Payload = msg
	case p.Command == CmdSendInsteonMsg:
		msg := &insteon.Message{}
		data := make([]byte, len(buf[2:])+3)
		copy(data[3:], buf[2:len(buf)-1])
		err = msg.UnmarshalBinary(data)
		p.Payload = msg
		p.Ack = buf[len(buf)-1]
	default:
		payload := &insteon.BufPayload{}
		err = payload.UnmarshalBinary(buf[2 : len(buf)-1])
		p.Payload = payload
		p.Ack = buf[len(buf)-1]
	}
	return err
}
