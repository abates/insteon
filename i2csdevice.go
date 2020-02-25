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
	"time"
)

type i2CsConnection struct {
	Connection
}

func (i2cs i2CsConnection) Send(message *Message) (*Message, error) {
	if message.Command[1] == CmdEnterLinkingMode[1] {
		message.Command = CmdEnterLinkingModeExt.SubCommand(int(message.Command[2]))
		message.Payload = make([]byte, 14)
		message.Flags = ExtendedDirectMessage
	}

	// set checksum
	if message.Flags.Extended() {
		l := len(message.Payload)
		message.Payload[l-1] = checksum(message.Command, message.Payload)
	}
	return i2cs.Connection.Send(message)
}

// i2CsDevice can communicate with Version 2 (checksum) Insteon Engines
type i2CsDevice struct {
	*i2Device
}

// newI2CsDevice will initialize a new I2CsDevice object and make
// it ready for use
func newI2CsDevice(connection Connection, timeout time.Duration) *i2CsDevice {
	i2cs := &i2CsDevice{
		i2Device: newI2Device(i2CsConnection{connection}, timeout),
	}
	i2cs.linkdb.device = i2cs
	i2cs.linkdb.timeout = timeout

	return i2cs
}

// SendCommand will send the given command bytes to the device including
// a payload (for extended messages). If payload length is zero then a standard
// length message is used to deliver the commands. Any error encountered sending
// the command is returned (eg. ack timeout, etc)
func (i2cs *i2CsDevice) SendCommand(command Command, payload []byte) error {
	_, err := i2cs.sendCommand(command, payload, i2csErrLookup)
	return err
}

// String returns the string "I2CS Device (<address>)" where <address> is the destination
// address of the device
func (i2cs *i2CsDevice) String() string {
	return sprintf("I2CS Device (%s)", i2cs.Address())
}

func checksum(cmd Command, buf []byte) byte {
	sum := cmd[1] + cmd[2]
	for _, b := range buf {
		sum += b
	}
	return ^sum + 1
}

func i2csErrLookup(msg *Message) (*Message, error) {
	var err error
	if msg.Flags.Type() == MsgTypeDirectNak {
		switch msg.Command[2] & 0xff {
		case 0xfb:
			err = ErrIllegalValue
		case 0xfc:
			err = ErrPreNak
		case 0xfd:
			err = ErrIncorrectChecksum
		case 0xfe:
			err = ErrNoLoadDetected
		case 0xff:
			err = ErrNotLinked
		default:
			err = newTraceError(ErrUnexpectedResponse)
		}
	}
	return msg, err
}
