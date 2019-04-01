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
	"sync"
	"time"
)

// I2CsDevice can communicate with Version 2 (checksum) Insteon Engines
type I2CsDevice struct {
	sync.Mutex
	*I2Device
}

// NewI2CsDevice will initialize a new I2CsDevice object and make
// it ready for use
func NewI2CsDevice(address Address, connection Connection, timeout time.Duration) *I2CsDevice {
	return &I2CsDevice{I2Device: NewI2Device(address, connection, timeout)}
}

// EnterLinkingMode will put the device into linking mode. This is
// equivalent to holding down the set button until the device
// beeps and the indicator light starts flashing
func (i2cs *I2CsDevice) EnterLinkingMode(group Group) (err error) {
	return extractError(i2cs.SendCommand(CmdEnterLinkingModeExt.SubCommand(int(group)), make([]byte, 14)))
}

// String returns the string "I2CS Device (<address>)" where <address> is the destination
// address of the device
func (i2cs *I2CsDevice) String() string {
	return sprintf("I2CS Device (%s)", i2cs.Address())
}

// SendCommand will send the given command bytes to the device including
// a payload (for extended messages). If payload length is zero then a standard
// length message is used to deliver the commands. The command bytes from the
// response ack are returned as well as any error
func (i2cs *I2CsDevice) SendCommand(command Command, payload []byte) (response Command, err error) {
	flags := StandardDirectMessage

	if command[1] == CmdSetOperatingFlags[1] && len(payload) == 0 {
		payload = make([]byte, 14)
	}

	if len(payload) > 0 {
		flags = ExtendedDirectMessage
	}

	ack, err := i2cs.Send(&Message{
		Flags:   flags,
		Command: command,
		Payload: payload,
	})

	if err == nil {
		response = ack.Command
	}

	return response, err
}

func checksum(cmd Command, buf []byte) byte {
	sum := cmd[1] + cmd[2]
	for _, b := range buf {
		sum += b
	}
	return ^sum + 1
}

func (i2cs *I2CsDevice) Send(msg *Message) (ack *Message, err error) {
	i2cs.Lock()
	defer i2cs.Unlock()
	// set checksum
	if msg.Flags.Extended() {
		l := len(msg.Payload)
		msg.Payload[l-1] = checksum(msg.Command, msg.Payload)
	}
	return i2cs.I2Device.Send(msg)
}

func i2csErrLookup(msg *Message, err error) (*Message, error) {
	if err != nil {
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

func (i2cs *I2CsDevice) Receive() (*Message, error) {
	i2cs.Lock()
	defer i2cs.Unlock()
	return i2csErrLookup(i2cs.I2Device.Receive())
}
