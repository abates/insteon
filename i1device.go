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
)

var dDumpTpl = template.Must(template.New("name").Parse(`
        Device: {{.String}}
      Category: {{ .Info.DevCat }}
      Firmware: {{ .Info.FirmwareVersion }}
Engine Version: {{ .Info.EngineVersion }}
`[1:]))

// Device provides remote communication to version 1 engines
type device struct {
	MessageWriter
	*linkdb
	info DeviceInfo
}

func newDevice(mw MessageWriter, info DeviceInfo) *device {
	d := &device{
		MessageWriter: mw,
		linkdb:        &linkdb{},
		info:          info,
	}
	d.linkdb.MessageWriter = d
	return d
}

func (d *device) Write(msg *Message) (ack *Message, err error) {
	msg.Dst = d.info.Address
	msg.Flags = StandardDirectMessage
	msg.SetMaxTTL(3)
	msg.SetTTL(3)
	if len(msg.Payload) > 0 {
		msg.Flags = ExtendedDirectMessage
	}

	if d.info.EngineVersion == VerI2Cs {
		return d.writeWithChecksum(msg)
	}
	return d.MessageWriter.Write(msg)
}

func (d *device) writeWithChecksum(msg *Message) (ack *Message, err error) {
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

func (d *device) Info() DeviceInfo {
	return d.info
}

func (d *device) SendCommand(command Command, payload []byte) error {
	_, err := d.Send(command, payload)
	return err
}

// Send will send the given command bytes to the device including
// a payload (for extended messages). If payload length is zero then a standard
// length message is used to deliver the commands. Any error encountered sending
// the command is returned (eg. ack timeout, etc)
func (d *device) Send(command Command, payload []byte) (Command, error) {
	return d.sendCommand(command, payload)
}

// sendCommand will send the given command bytes to the device including
// a payload (for extended messages). If payload length is zero then a standard
// length message is used to deliver the commands. The command bytes from the
// response ack are returned as well as any error
func (d *device) sendCommand(command Command, payload []byte) (response Command, err error) {
	ack, err := d.Write(&Message{
		Command: command,
		Payload: payload,
	})

	if err == nil {
		response = ack.Command
	}

	return response, err
}

func (d *device) errLookup(msg *Message, err error) (*Message, error) {
	if err == ErrNak {
		if d.info.EngineVersion == VerI2Cs {
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
func (d *device) ProductData() (data *ProductData, err error) {
	msg, err := d.Write(&Message{Command: CmdProductDataReq})
	if err == nil {
		msg, err = Read(d, CmdMatcher(CmdProductDataResp))
		if err == nil {
			data = &ProductData{}
			err = data.UnmarshalBinary(msg.Payload)
		}
	}
	return data, err
}

func (d *device) ExtendedGet(data []byte) (buf []byte, err error) {
	msg, err := d.Write(&Message{Command: CmdExtendedGetSet, Payload: data})
	if err == nil {
		msg, err = Read(d, CmdMatcher(CmdExtendedGetSet))
		if err == nil {
			buf = make([]byte, len(msg.Payload))
			copy(buf, msg.Payload)
		}
	}
	return buf, err
}

func (d *device) Address() Address {
	return d.info.Address
}

// String returns the string "<engine version> Device (<address>)" where <address> is the destination
// address of the device
func (d *device) String() string {
	return fmt.Sprintf("%s Device (%s)", d.info.EngineVersion, d.info.Address)
}

func (d *device) Dump() string {
	builder := &strings.Builder{}
	dDumpTpl.Execute(builder, d)
	return builder.String()
}

func (d *device) linkingMode(cmd Command, payload []byte) (err error) {
	return d.SendCommand(cmd, payload)
}

func (d *device) EnterLinkingMode(group Group) error {
	payload := []byte{}
	cmd := CmdEnterLinkingMode.SubCommand(int(group))
	if d.info.EngineVersion == VerI2Cs {
		cmd = CmdEnterLinkingModeExt.SubCommand(int(group))
		payload = make([]byte, 14)
	}

	return d.linkingMode(cmd, payload)
}

func (d *device) EnterUnlinkingMode(group Group) error {
	payload := []byte{}
	if d.info.EngineVersion == VerI2Cs {
		payload = make([]byte, 14)
	}
	return d.linkingMode(CmdEnterUnlinkingMode.SubCommand(int(group)), payload)
}

func (d *device) ExitLinkingMode() error {
	return d.SendCommand(CmdExitLinkingMode, nil)
}
