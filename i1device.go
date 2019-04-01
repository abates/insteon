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

// I1Device provides remote communication to version 1 engines
type I1Device struct {
	sync.Mutex
	address         Address
	devCat          DevCat
	firmwareVersion FirmwareVersion
	timeout         time.Duration
	connection      Connection
}

// NewI1Device will construct an I1Device for the given address
func NewI1Device(address Address, connection Connection, timeout time.Duration) *I1Device {
	i1 := &I1Device{
		address:         address,
		devCat:          DevCat{0xff, 0xff},
		firmwareVersion: FirmwareVersion(0x00),
		timeout:         timeout,
		connection:      connection,
	}

	return i1
}

func (i1 *I1Device) sendCommand(command Command, payload []byte) (response Command, err error) {
	flags := StandardDirectMessage
	if len(payload) > 0 {
		flags = ExtendedDirectMessage
	}

	ack, err := i1.connection.Send(&Message{
		Flags:   flags,
		Command: command,
		Payload: payload,
	})

	if err == nil {
		response = ack.Command
	}

	return response, err
}

func (i1 *I1Device) Send(msg *Message) (ack *Message, err error) {
	i1.Lock()
	defer i1.Unlock()
	return i1.connection.Send(msg)
}

func errLookup(msg *Message, err error) (*Message, error) {
	if err != nil {
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

func (i1 *I1Device) Receive() (*Message, error) {
	i1.Lock()
	defer i1.Unlock()
	return i1.connection.Receive()
}

// SendCommand will send the given command bytes to the device including
// a payload (for extended messages). If payload length is zero then a standard
// length message is used to deliver the commands. The command bytes from the
// response ack are returned as well as any error
func (i1 *I1Device) SendCommand(command Command, payload []byte) (response Command, err error) {
	i1.Lock()
	defer i1.Unlock()
	return i1.sendCommand(command, payload)
}

// Address is the Insteon address of the device
func (i1 *I1Device) Address() Address {
	return i1.address
}

func extractError(v interface{}, err error) error {
	return err
}

// AssignToAllLinkGroup will inform the device what group should be used during an All-Linking
// session
func (i1 *I1Device) AssignToAllLinkGroup(group Group) error {
	return extractError(i1.SendCommand(CmdAssignToAllLinkGroup.SubCommand(int(group)), nil))
}

// DeleteFromAllLinkGroup will inform the device which group should be unlinked during an
// All-Link unlinking session
func (i1 *I1Device) DeleteFromAllLinkGroup(group Group) (err error) {
	return extractError(i1.SendCommand(CmdDeleteFromAllLinkGroup.SubCommand(int(group)), nil))
}

// ProductData will retrieve the device's product data
func (i1 *I1Device) ProductData() (data *ProductData, err error) {
	i1.Lock()
	defer i1.Unlock()

	_, err = i1.sendCommand(CmdProductDataReq, nil)
	timeout := time.Now().Add(i1.timeout)
	for timeout.After(time.Now()) {
		msg, err := i1.connection.Receive()
		if msg.Command[1] == CmdProductDataResp[1] {
			data = &ProductData{}
			err = data.UnmarshalBinary(msg.Payload)
			return data, err
		}
	}
	err = ErrReadTimeout
	return data, err
}

// Ping will send a Ping command to the device
func (i1 *I1Device) Ping() (err error) {
	return extractError(i1.SendCommand(CmdPing, nil))
}

// SetAllLinkCommandAlias will set the device's standard command to be used
// when the given alias command is sent
func (i1 *I1Device) SetAllLinkCommandAlias(match, replace Command) error {
	// TODO implement
	return ErrNotImplemented
}

// SetAllLinkCommandAliasData will set any extended data required by the alias
// command
func (i1 *I1Device) SetAllLinkCommandAliasData(data []byte) error {
	// TODO implement
	return ErrNotImplemented
}

// BlockDataTransfer will retrieve a block of memory from the device
func (i1 *I1Device) BlockDataTransfer(start, end MemAddress, length int) ([]byte, error) {
	// TODO implement
	return nil, ErrNotImplemented
}

// String returns the string "I1 Device (<address>)" where <address> is the destination
// address of the device
func (i1 *I1Device) String() string {
	return sprintf("I1 Device (%s)", i1.Address())
}
