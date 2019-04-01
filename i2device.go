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

// I2Device can communicate with Version 2 Insteon Engines
type I2Device struct {
	sync.Mutex
	*I1Device
}

// NewI2Device will construct an device object that can communicate with version 2
// Insteon engines
func NewI2Device(address Address, connection Connection, timeout time.Duration) *I2Device {
	return &I2Device{I1Device: NewI1Device(address, connection, timeout)}
}

// AddLink will either add the link to the All-Link database
// or it will replace an existing link-record that has been marked
// as deleted
func (i2 *I2Device) AddLink(newLink *LinkRecord) error {
	return ErrNotImplemented
}

// RemoveLinks will either remove the link records from the device
// All-Link database, or it will simply mark them as deleted
func (i2 *I2Device) RemoveLinks(oldLinks ...*LinkRecord) error {
	return ErrNotImplemented
}

// Links will retrieve the link-database from the device and
// return a list of LinkRecords
func (i2 *I2Device) Links() (links []*LinkRecord, err error) {
	i2.Lock()
	defer i2.Unlock()

	Log.Debugf("Retrieving Device link database")
	lastAddress := MemAddress(0)
	buf, _ := (&LinkRequest{Type: ReadLink, NumRecords: 0}).MarshalBinary()
	_, err = i2.I1Device.SendCommand(CmdReadWriteALDB, buf)

	for err == nil {
		var msg *Message
		msg, err = i2.I1Device.Receive()
		if msg.Flags.Extended() && msg.Command[1] == CmdReadWriteALDB[1] {
			lr := &LinkRequest{}
			err = lr.UnmarshalBinary(msg.Payload)
			// make sure there was no error unmarshalling, also make sure
			// that it's a new memory address.  Since insteon messages
			// are retransmitted, it is possible that the same ALDB response
			// will be received more than once
			if err == nil && lr.MemAddress != lastAddress {
				lastAddress = lr.MemAddress
				links = append(links, lr.Link)
			}
		}
	}

	if err == ErrEndOfLinks {
		err = nil
	}
	return links, err
}

// EnterLinkingMode is the programmatic equivalent of holding down
// the set button for two seconds. If the device is the first
// to enter linking mode, then it is the controller. The next
// device to enter linking mode is the responder.  LinkingMode
// is usually indicated by a flashing GREEN LED on the device
func (i2 *I2Device) EnterLinkingMode(group Group) error {
	return extractError(i2.SendCommand(CmdEnterLinkingMode.SubCommand(int(group)), nil))
}

// EnterUnlinkingMode puts a controller device into unlinking mode
// when the set button is then pushed (EnterLinkingMode) on a linked
// device the corresponding links in both the controller and responder
// are deleted.  EnterUnlinkingMode is the programmatic equivalent
// to pressing the set button until the device beeps, releasing, then
// pressing the set button again until the device beeps again. UnlinkingMode
// is usually indicated by a flashing RED LED on the device
func (i2 *I2Device) EnterUnlinkingMode(group Group) error {
	return extractError(i2.SendCommand(CmdEnterUnlinkingMode.SubCommand(int(group)), nil))
}

// ExitLinkingMode takes a controller out of linking/unlinking mode.
func (i2 *I2Device) ExitLinkingMode() error {
	return extractError(i2.SendCommand(CmdExitLinkingMode, nil))
}

// WriteLink will write the link record to the device's link database
func (i2 *I2Device) WriteLink(link *LinkRecord) (err error) {
	if link.memAddress == MemAddress(0x0000) {
		err = ErrInvalidMemAddress
	} else {
		buf, _ := (&LinkRequest{MemAddress: link.memAddress, Type: WriteLink, Link: link}).MarshalBinary()
		_, err = i2.SendCommand(CmdReadWriteALDB, buf)
	}
	return err
}

// AppendLink will add a new link record to the end of the All-Link database
func (i2 *I2Device) AppendLink(link *LinkRecord) (err error) {
	// determine address of last link record
	links, err := i2.Links()
	if err == nil {
		link.memAddress = BaseLinkDBAddress
		if len(links) > 0 {
			link.memAddress = links[len(links)-1].memAddress - LinkRecordSize
		}
		err = i2.WriteLink(link)
	}
	return err
}

// String returns the string "I2 Device (<address>)" where <address> is the destination
// address of the device
func (i2 *I2Device) String() string {
	return sprintf("I2 Device (%s)", i2.Address())
}

func (i2 *I2Device) SendCommand(command Command, payload []byte) (response Command, err error) {
	i2.Lock()
	defer i2.Unlock()
	return i2.I1Device.SendCommand(command, payload)
}

func (i2 *I2Device) Send(msg *Message) (ack *Message, err error) {
	i2.Lock()
	defer i2.Unlock()
	return i2.I1Device.Send(msg)
}

func (i2 *I2Device) Receive() (*Message, error) {
	i2.Lock()
	defer i2.Unlock()
	return i2.I1Device.Receive()
}
