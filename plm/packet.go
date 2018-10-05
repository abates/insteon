package plm

import (
	"fmt"
)

type Packet struct {
	Command Command
	Payload []byte
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

func (p *Packet) Format(f fmt.State, c rune) {
	if c == 'x' || c == 'X' {
		format := fmt.Sprintf("%%02%c%%s%%02%c", c, c)
		fmt.Fprintf(f, format, byte(p.Command), hexDump(fmt.Sprintf("%%02%c", c), p.Payload, ""), p.Ack)
	} else if c == 's' || c == 'v' || c == 'q' {
		if c == 's' {
			fmt.Fprintf(f, p.String())
		} else if c == 'v' {
			fmt.Fprintf(f, "%s %s", p.Command, hexDump("%02x", p.Payload, " "))
			if p.ACK() {
				fmt.Fprintf(f, " ACK")
			} else if p.NAK() {
				fmt.Fprintf(f, " NAK")
			}
		} else {
			fmt.Fprintf(f, "%q", p.String())
		}
	} else {
		fmt.Fprintf(f, "%%!%c(packet=%s)", c, p.String())
	}
}

func (p *Packet) String() string {
	cmd := fmt.Sprintf("%v", p.Command)
	if p.Command >= 0x60 {
		if p.ACK() {
			cmd = fmt.Sprintf("%s ACK", cmd)
		} else if p.NAK() {
			cmd = fmt.Sprintf("%s NAK", cmd)
		}
	}

	return cmd
}

func (p *Packet) MarshalBinary() (buf []byte, err error) {
	buf = make([]byte, 2)
	buf[0] = 0x02
	buf[1] = byte(p.Command)
	if len(p.Payload) > 0 {
		buf = append(buf, p.Payload...)
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

	p.Payload = make([]byte, len(buf))
	copy(p.Payload, buf)

	return err
}
