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

package commands

import (
	"fmt"
	"strconv"
)

// Commandable indicates that the implementation exists to send commands
type Commandable interface {
	// SendCommand will send the given command bytes to the device including
	// a payload (for extended messages). If payload length is zero then a standard
	// length message is used to deliver the commands.
	SendCommand(cmd Command, payload []byte) (err error)

	Send(cmd Command, payload []byte) (ack Command, err error)
}

// Command is a 3 byte sequence that indicates command flags (direct, all-link or broadcast, standard/extended)
// and two byte commands
type Command int

const (
	StandardDirect = Command(0x000000)
	ExtendedDirect = Command(0x010000)
	Broadcast      = Command(0x080000)
	AllLink        = Command(0x0c0000)
)

// SubCommand will return a new command where the subcommand byte is updated
// to reflect command2 from the arguments
func (cmd Command) SubCommand(command2 int) Command {
	return (cmd & 0xffff00) | (0xff & Command(command2))
}

func (cmd *Command) Set(value string) error {
	i, err := strconv.Atoi(value)
	if err == nil {
		*cmd = (*cmd & 0xffff00) | (0xff & Command(i))
	}
	return err
}

func (cmd Command) Command0() int {
	return int(cmd >> 16 & 0xff)
}

func (cmd Command) Command1() int {
	return int((cmd >> 8) & 0xff)
}

func (cmd Command) Command2() int {
	return int(cmd & 0xff)
}

func (cmd Command) Matches(other Command) bool {
	return cmd.Command1() == other.Command1()
}

func (cmd Command) String() string {
	if str, found := cmdStrings[cmd]; found {
		return str
	} else if str, found := cmdStrings[cmd&0xffff00]; found {
		return fmt.Sprintf("%s(%d)", str, cmd.Command2())
	}
	return fmt.Sprintf("Command(0x%02x, 0x%02x, 0x%02x)", cmd.Command0(), cmd.Command1(), cmd.Command2())
}

func From(flagsByte byte, cmd1Byte byte, cmd2Byte byte) Command {
	// the lower 4 bits of insteon message flags are the
	// message hop bits that we don't need for command lookup
	// the 5th bit is the extended message bit
	var flags Command
	if flagsByte&0x10 == 0x10 {
		flags = ExtendedDirect
	} else if flagsByte > 0 {
		flagsByte = flagsByte & 0xe0
		switch {
		case flagsByte == 0x00 || flagsByte == 0x20:
			flags = StandardDirect
		case flagsByte == 0x80:
			flags = Broadcast
		case flagsByte == 0x40 || flagsByte == 0x60 || flagsByte == 0xc0 || flagsByte == 0xe0:
			flags = AllLink
		}
	}
	cmd1 := Command(cmd1Byte) << 8
	cmd2 := Command(cmd2Byte)

	return flags | cmd1 | cmd2
}
