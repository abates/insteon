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
	"fmt"
	"html/template"
	"strings"

	"github.com/abates/insteon/commands"
)

var dDumpTpl = template.Must(template.New("name").Parse(`
        Device: {{.String}}
      Category: {{ .Info.DevCat }}
      Firmware: {{ .Info.FirmwareVersion }}
Engine Version: {{ .Info.EngineVersion }}
`[1:]))

// BasicDevice provides remote communication to version 1 engines
type BasicDevice struct {
	MessageWriter
	DeviceInfo
	*linkdb
}

func NewDevice(mw MessageWriter, info DeviceInfo) *BasicDevice {
	d := &BasicDevice{
		MessageWriter: mw,
		linkdb:        &linkdb{},
		DeviceInfo:    info,
	}
	d.linkdb.MessageWriter = d
	return d
}

func (d *BasicDevice) Write(msg *Message) (ack *Message, err error) {
	msg.Dst = d.DeviceInfo.Address
	msg.Flags = StandardDirectMessage
	msg.SetMaxTTL(3)
	msg.SetTTL(3)
	if len(msg.Payload) > 0 {
		msg.Flags = ExtendedDirectMessage
	}

	if d.DeviceInfo.EngineVersion == VerI2Cs {
		return d.writeWithChecksum(msg)
	}
	return d.MessageWriter.Write(msg)
}

func (d *BasicDevice) writeWithChecksum(msg *Message) (ack *Message, err error) {
	// set checksum
	if len(msg.Payload) > 0 {
		if len(msg.Payload) < 14 {
			tmp := make([]byte, 14)
			copy(tmp, msg.Payload)
			msg.Payload = tmp
		}
		setChecksum(msg.Command, msg.Payload)
	}
	return d.MessageWriter.Write(msg)
}

func (d *BasicDevice) Info() DeviceInfo {
	return d.DeviceInfo
}

func (d *BasicDevice) SendCommand(command commands.Command, payload []byte) error {
	_, err := d.Send(command, payload)
	return err
}

// Send will send the given command bytes to the device including
// a payload (for extended messages). If payload length is zero then a standard
// length message is used to deliver the commands. Any error encountered sending
// the command is returned (eg. ack timeout, etc)
func (d *BasicDevice) Send(command commands.Command, payload []byte) (commands.Command, error) {
	return d.sendCommand(command, payload)
}

// sendCommand will send the given command bytes to the device including
// a payload (for extended messages). If payload length is zero then a standard
// length message is used to deliver the commands. The command bytes from the
// response ack are returned as well as any error
func (d *BasicDevice) sendCommand(command commands.Command, payload []byte) (response commands.Command, err error) {
	ack, err := d.Write(&Message{
		Command: command,
		Payload: payload,
	})

	if err == nil {
		response = ack.Command
	}

	return response, err
}

func (d *BasicDevice) errLookup(msg *Message, err error) (*Message, error) {
	if err == ErrNak {
		if d.DeviceInfo.EngineVersion == VerI2Cs {
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
		} else {
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
	}
	return msg, err
}

// ProductData will retrieve the device's product data
func (d *BasicDevice) ProductData() (data *ProductData, err error) {
	msg, err := d.Write(&Message{Command: commands.ProductDataReq})
	if err == nil {
		msg, err = Read(d, CmdMatcher(commands.ProductDataResp))
		if err == nil {
			data = &ProductData{}
			err = data.UnmarshalBinary(msg.Payload)
		}
	}
	return data, err
}

func (d *BasicDevice) ExtendedGet(data []byte) (buf []byte, err error) {
	msg, err := d.Write(&Message{Command: commands.ExtendedGetSet, Payload: data})
	if err == nil {
		msg, err = Read(d, CmdMatcher(commands.ExtendedGetSet))
		if err == nil {
			buf = make([]byte, len(msg.Payload))
			copy(buf, msg.Payload)
		}
	}
	return buf, err
}

func (d *BasicDevice) Address() Address {
	return d.DeviceInfo.Address
}

// String returns the string "<engine version> Device (<address>)" where <address> is the destination
// address of the device
func (d *BasicDevice) String() string {
	return fmt.Sprintf("%s Device (%s)", d.DeviceInfo.EngineVersion, d.DeviceInfo.Address)
}

func (d *BasicDevice) Dump() string {
	builder := &strings.Builder{}
	dDumpTpl.Execute(builder, d)
	return builder.String()
}

func (d *BasicDevice) linkingMode(cmd commands.Command, payload []byte) (err error) {
	return d.SendCommand(cmd, payload)
}

func (d *BasicDevice) EnterLinkingMode(group Group) error {
	payload := []byte{}
	cmd := commands.EnterLinkingMode.SubCommand(int(group))
	if d.DeviceInfo.EngineVersion == VerI2Cs {
		cmd = commands.EnterLinkingModeExt.SubCommand(int(group))
		payload = make([]byte, 14)
	}

	return d.linkingMode(cmd, payload)
}

func (d *BasicDevice) EnterUnlinkingMode(group Group) error {
	payload := []byte{}
	if d.DeviceInfo.EngineVersion == VerI2Cs {
		payload = make([]byte, 14)
	}
	return d.linkingMode(commands.EnterUnlinkingMode.SubCommand(int(group)), payload)
}

func (d *BasicDevice) ExitLinkingMode() error {
	return d.SendCommand(commands.ExitLinkingMode, nil)
}
