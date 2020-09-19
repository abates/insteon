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
	"html/template"
	"strings"
)

var i1DumpTpl = template.Must(template.New("name").Parse(`
        Device: {{.String}}
      Category: {{ .Info.DevCat }}
      Firmware: {{ .Info.FirmwareVersion }}
Engine Version: {{ .Info.EngineVersion }}
`[1:]))

// i1Device provides remote communication to version 1 engines
type i1Device struct {
	dial Dialer
	info DeviceInfo
}

// newI1Device will construct an I1Device for the given connection
func newI1Device(dial Dialer, info DeviceInfo) *i1Device {
	i1 := &i1Device{
		dial: dial,
		info: info,
	}

	return i1
}

func (i1 *i1Device) Info() DeviceInfo {
	return i1.info
}

func (i1 *i1Device) Dial(cmds ...Command) (Connection, error) {
	return i1.dial.Dial(i1.info.Address, cmds...)
}

// SendCommand will send the given command bytes to the device including
// a payload (for extended messages). If payload length is zero then a standard
// length message is used to deliver the commands. Any error encountered sending
// the command is returned (eg. ack timeout, etc)
func (i1 *i1Device) SendCommand(command Command, payload []byte) (Command, error) {
	return i1.sendCommand(command, payload, errLookup)
}

// sendCommand will send the given command bytes to the device including
// a payload (for extended messages). If payload length is zero then a standard
// length message is used to deliver the commands. The command bytes from the
// response ack are returned as well as any error
func (i1 *i1Device) sendCommand(command Command, payload []byte, errLookup func(*Message, error) (*Message, error)) (response Command, err error) {
	conn, err := i1.Dial(command)
	if err == nil {
		defer conn.Close()
		var ack *Message
		ack, err = errLookup(conn.Send(&Message{
			Command: command,
			Payload: payload,
		}))

		if err == nil {
			response = ack.Command
		}
	}

	return response, err
}

func errLookup(msg *Message, err error) (*Message, error) {
	if err == ErrNak {
		switch msg.Command.Command2() & 0xff {
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
	conn, err := i1.Dial(CmdProductDataReq, CmdProductDataResp)
	if err == nil {
		defer conn.Close()
		_, err = conn.Send(&Message{Command: CmdProductDataReq})
		if err == nil {
			var msg *Message
			msg, err = conn.Receive()
			if err == nil {
				data = &ProductData{}
				err = data.UnmarshalBinary(msg.Payload)
			}
		}
	}
	return data, err
}

func (i1 *i1Device) Address() Address {
	return i1.info.Address
}

// String returns the string "I1 Device (<address>)" where <address> is the destination
// address of the device
func (i1 *i1Device) String() string {
	return sprintf("I1 Device (%s)", i1.info.Address)
}

func (i1 *i1Device) Dump() string {
	builder := &strings.Builder{}
	i1DumpTpl.Execute(builder, i1)
	return builder.String()
}

func (i1 *i1Device) LinkDatabase() (Linkable, error) {
	return nil, ErrNotSupported
}
