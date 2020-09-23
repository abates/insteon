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
	"time"
)

var i1DumpTpl = template.Must(template.New("name").Parse(`
        Device: {{.String}}
      Category: {{ .Info.DevCat }}
      Firmware: {{ .Info.FirmwareVersion }}
Engine Version: {{ .Info.EngineVersion }}
`[1:]))

// i1Device provides remote communication to version 1 engines
type i1Device struct {
	bus  Bus
	info DeviceInfo
}

// newI1Device will construct an I1Device for the given connection
func newI1Device(bus Bus, info DeviceInfo) *i1Device {
	i1 := &i1Device{
		bus:  bus,
		info: info,
	}

	return i1
}

func (i1 *i1Device) Info() DeviceInfo {
	return i1.info
}

func (i1 *i1Device) Publish(msg *Message) (*Message, error) {
	msg.Dst = i1.info.Address
	msg.Flags = StandardDirectMessage
	if len(msg.Payload) > 0 {
		msg.Flags = ExtendedDirectMessage
	}

	return i1.bus.Publish(msg)
}

func (i1 *i1Device) Subscribe(matcher Matcher) <-chan *Message {
	ch := i1.bus.Subscribe(i1.info.Address, matcher)
	return ch
}

func (i1 *i1Device) Unsubscribe(ch <-chan *Message) {
	i1.bus.Unsubscribe(i1.info.Address, ch)
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
	ack, err := i1.Publish(&Message{
		Command: command,
		Payload: payload,
	})

	if err == nil {
		response = ack.Command
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
	rx := i1.Subscribe(And(Not(AckMatcher()), Or(CmdMatcher(CmdProductDataReq), CmdMatcher(CmdProductDataResp))))
	defer i1.Unsubscribe(rx)

	_, err = i1.Publish(&Message{Command: CmdProductDataReq})
	if err == nil {
		select {
		case msg := <-rx:
			data = &ProductData{}
			err = data.UnmarshalBinary(msg.Payload)
		case <-time.After(2 * i1.bus.Config().Timeout(false)):
			err = ErrReadTimeout
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
