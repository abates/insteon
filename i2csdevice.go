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

type i2csbus struct {
	Bus
}

func (i2cs *i2csbus) Publish(msg *Message) (ack *Message, err error) {
	if msg.Command.Command1() == CmdEnterLinkingMode.Command1() {
		msg.Command = CmdEnterLinkingModeExt.SubCommand(int(msg.Command.Command2()))
		msg.Payload = make([]byte, 14)
	}

	// set checksum
	if len(msg.Payload) > 0 {
		if len(msg.Payload) < 14 {
			tmp := make([]byte, 14)
			copy(tmp, msg.Payload)
			msg.Payload = tmp
		}
		l := len(msg.Payload)
		msg.Payload[l-1] = checksum(msg.Command, msg.Payload)
	}
	return i2cs.Bus.Publish(msg)
}

// i2CsDevice can communicate with Version 2 (checksum) Insteon Engines
type i2CsDevice struct {
	*i2Device
}

// newI2CsDevice will initialize a new I2CsDevice object and make
// it ready for use
func newI2CsDevice(bus Bus, info DeviceInfo) *i2CsDevice {
	i2cs := &i2CsDevice{
		i2Device: newI2Device(&i2csbus{bus}, info),
	}

	return i2cs
}

// String returns the string "I2CS Device (<address>)" where <address> is the destination
// address of the device
func (i2cs *i2CsDevice) String() string {
	return sprintf("I2CS Device (%s)", i2cs.Info().Address)
}

func checksum(cmd Command, buf []byte) byte {
	sum := byte(cmd.Command1() + cmd.Command2())
	for _, b := range buf {
		sum += b
	}
	return ^sum + 1
}

func i2csErrLookup(msg *Message, err error) (*Message, error) {
	if err == ErrNak {
		switch msg.Command.Command2() & 0xff {
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
