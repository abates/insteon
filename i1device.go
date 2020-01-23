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

// i1Device provides remote communication to version 1 engines
type i1Device struct {
	Connection
	devCat          DevCat
	firmwareVersion FirmwareVersion
	timeout         time.Duration
}

// newI1Device will construct an I1Device for the given connection
func newI1Device(connection Connection, timeout time.Duration) *i1Device {
	i1 := &i1Device{
		Connection:      connection,
		devCat:          DevCat{0xff, 0xff},
		firmwareVersion: FirmwareVersion(0x00),
		timeout:         timeout,
	}

	return i1
}

// sendCommand will send the given command bytes to the device including
// a payload (for extended messages). If payload length is zero then a standard
// length message is used to deliver the commands. Any error encountered sending
// the command is returned (eg. ack timeout, etc)
func (i1 *i1Device) SendCommand(command Command, payload []byte) error {
	_, err := i1.sendCommand(command, payload)
	return err
}

// sendCommand will send the given command bytes to the device including
// a payload (for extended messages). If payload length is zero then a standard
// length message is used to deliver the commands. The command bytes from the
// response ack are returned as well as any error
func (i1 *i1Device) sendCommand(command Command, payload []byte) (response Command, err error) {
	flags := StandardDirectMessage
	if len(payload) > 0 {
		flags = ExtendedDirectMessage
		if len(payload) < 14 {
			tmp := make([]byte, 14)
			copy(tmp, payload)
			payload = tmp
		}
	}

	ack, err := i1.Connection.Send(&Message{
		Flags:   flags,
		Command: command,
		Payload: payload,
	})

	if err == nil {
		response = ack.Command
	}

	return response, err
}

func errLookup(msg *Message, err error) (*Message, error) {
	if err == nil && msg.Flags.Type() == MsgTypeDirectNak {
		switch msg.Command[2] & 0xff {
		case 0xfd:
			err = ErrUnknownCommand
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

// ProductData will retrieve the device's product data
func (i1 *i1Device) ProductData() (data *ProductData, err error) {
	err = i1.SendCommand(CmdProductDataReq, nil)
	if err == nil {
		err = Receive(i1.Connection.Receive, i1.timeout, func(msg *Message) error {
			if msg.Command[1] == CmdProductDataResp[1] {
				data = &ProductData{}
				err = data.UnmarshalBinary(msg.Payload)
				if err == nil {
					err = ErrReceiveComplete
				}
			}
			return err
		})
	}
	return data, err
}

// String returns the string "I1 Device (<address>)" where <address> is the destination
// address of the device
func (i1 *i1Device) String() string {
	return sprintf("I1 Device (%s)", i1.Address())
}

// Receive waits for the next message from the device.  Receive
// always returns, but may return with an error (such as ErrReadTimeout)
func (i1 *i1Device) Receive() (*Message, error) {
	return errLookup(i1.Connection.Receive())
}
