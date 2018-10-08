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

import "time"

// I2Device can communicate with Version 2 Insteon Engines
type I2Device struct {
	*I1Device
}

// NewI2Device will construct an device object that can communicate with version 2
// Insteon engines
func NewI2Device(info DeviceInfo, address Address, sendCh chan<- *MessageRequest, recvCh <-chan *Message, timeout time.Duration) (Device, error) {
	i1Device, err := NewI1Device(info, address, sendCh, recvCh, timeout)
	return &I2Device{i1Device.(*I1Device)}, err
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
	Log.Debugf("Retrieving Device link database")
	lastAddress := MemAddress(0)
	buf, _ := (&LinkRequest{Type: ReadLink, NumRecords: 0}).MarshalBinary()
	recvCh, err := i2.SendCommandAndListen(CmdReadWriteALDB, buf)

	for response := range recvCh {
		if response.Message.Flags.Extended() && response.Message.Command[1] == CmdReadWriteALDB[1] {
			lr := &LinkRequest{}
			err = lr.UnmarshalBinary(response.Message.Payload)
			if err == nil && lr.MemAddress != lastAddress {
				lastAddress = lr.MemAddress
				links = append(links, lr.Link)
				response.DoneCh <- false
			} else if err == ErrEndOfLinks {
				response.DoneCh <- true
				err = nil
			} else {
				response.DoneCh <- true
			}
		} else {
			response.DoneCh <- false
		}
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

// String returns the string "I2 Device (<address>)" where <address> is the destination
// address of the device
func (i2 *I2Device) String() string {
	return sprintf("I2 Device (%s)", i2.Address())
}
