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

package plm

import (
	"fmt"
)

const maxPaclen = 25

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
			fmt.Fprintf(f, "%v", p.Command)
			if len(p.Payload) > 0 {
				fmt.Fprintf(f, " %s", hexDump("%02x", p.Payload, " "))
			}
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
		p.Payload = make([]byte, 3)
	}

	// If the length of the buffer is 1 less than we're expecting, assume
	// no ack byte is present.  This change was made to support snooping
	if 0x60 <= p.Command && p.Command <= 0x7f {
		if p.Command == CmdSendInsteonMsg {
			if len(buf) == 21 || len(buf) == 7 {
				p.Ack = buf[len(buf)-1]
				buf = buf[0 : len(buf)-1]
			}
		} else if len(buf) == commandLens[p.Command] {
			p.Ack = buf[len(buf)-1]
			buf = buf[0 : len(buf)-1]
		}
	}

	p.Payload = append(p.Payload, buf...)
	return err
}
