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

// i2Device can communicate with Version 2 Insteon Engines
type i2Device struct {
	*i1Device
	linkdb
	timeout time.Duration
}

// newI2Device will construct an device object that can communicate with version 2
// Insteon engines
func newI2Device(connection Connection, timeout time.Duration) *i2Device {
	i2 := &i2Device{i1Device: newI1Device(connection, timeout), timeout: timeout}
	i2.linkdb.device = i2
	i2.linkdb.timeout = timeout
	return i2
}

func (i2 *i2Device) linkingMode(cmd Command, payload ...byte) error {
	i2.Lock()
	defer i2.Unlock()
	setButton := i2.AddListener(MsgTypeBroadcast, CmdSetButtonPressedController, CmdSetButtonPressedResponder)
	defer i2.RemoveListener(setButton)
	_, err := i2.SendCommand(cmd, payload)
	if err == nil {
		_, err = readFromCh(setButton, i2.timeout)
	}
	return err
}

// EnterLinkingMode is the programmatic equivalent of holding down
// the set button for two seconds. If the device is the first
// to enter linking mode, then it is the controller. The next
// device to enter linking mode is the responder.  LinkingMode
// is usually indicated by a flashing GREEN LED on the device
func (i2 *i2Device) EnterLinkingMode(group Group) error {
	return i2.linkingMode(CmdEnterLinkingMode.SubCommand(int(group)))
}

// EnterUnlinkingMode puts a controller device into unlinking mode
// when the set button is then pushed (EnterLinkingMode) on a linked
// device the corresponding links in both the controller and responder
// are deleted.  EnterUnlinkingMode is the programmatic equivalent
// to pressing the set button until the device beeps, releasing, then
// pressing the set button again until the device beeps again. UnlinkingMode
// is usually indicated by a flashing RED LED on the device
func (i2 *i2Device) EnterUnlinkingMode(group Group) error {
	return i2.linkingMode(CmdEnterUnlinkingMode.SubCommand(int(group)))
}

// ExitLinkingMode takes a controller out of linking/unlinking mode.
func (i2 *i2Device) ExitLinkingMode() error {
	return extractError(i2.SendCommand(CmdExitLinkingMode, nil))
}

// String returns the string "I2 Device (<address>)" where <address> is the destination
// address of the device
func (i2 *i2Device) String() string {
	return sprintf("I2 Device (%s)", i2.Address())
}
