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

// i2CsDevice can communicate with Version 2 (checksum) Insteon Engines
type i2CsDevice struct {
	*i2Device
	connection Connection
}

// newI2CsDevice will initialize a new I2CsDevice object and make
// it ready for use
func newI2CsDevice(connection Connection, timeout time.Duration) *i2CsDevice {
	i2cs := &i2CsDevice{connection: connection}
	// pass i2cs in here so that the downstream devices (I2Device and its I1Device) will
	// get checksums set for extended messages
	i2cs.i2Device = newI2Device(i2cs, timeout)
	i2cs.linkdb.device = i2cs
	i2cs.linkdb.timeout = timeout

	return i2cs
}

// Send will send the message to the device and wait for the
// device to respond with an Ack/Nak.  Send will always return
// but may return with a read timeout or other communication error
// In the case of the I2CsDevice, if an extended message is being
// sent, then the checksum of the message is computed and set as
// the last byte of the payload
func (i2cs *i2CsDevice) Send(message *Message) (*Message, error) {
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
	return i2cs.connection.Send(message)
}

// Address returns the unique Insteon address of the device
func (i2cs *i2CsDevice) Address() Address {
	return i2cs.connection.Address()
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

func i2csErrLookup(msg *Message, err error) (*Message, error) {
	if err == nil && msg.Flags.Type() == MsgTypeDirectNak {
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

// IDRequest sends an IDRequest command to the device and waits for
// the corresponding Set Button Pressed Controller/Responder message.
// The response is parsed and the Firmware version and DevCat are
// returned.  A ReadTimeout may occur if the device doesn't respond
// with the appropriate broadcast message, or if the local system
// doesn't receive it
func (i2cs *i2CsDevice) IDRequest() (FirmwareVersion, DevCat, error) {
	return i2cs.connection.IDRequest()
}

// Receive waits for the next message from the device.  Receive
// always returns, but may return with an error (such as ErrReadTimeout)
func (i2cs *i2CsDevice) Receive() (*Message, error) {
	return i2csErrLookup(i2cs.connection.Receive())
}
