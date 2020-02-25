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
	conn    Connection
	timeout time.Duration
}

// newI2Device will construct an device object that can communicate with version 2
// Insteon engines
func newI2Device(connection Connection, timeout time.Duration) *i2Device {
	i2 := &i2Device{i1Device: newI1Device(connection, timeout), timeout: timeout}
	i2.conn = connection
	i2.linkdb.device = i2
	i2.linkdb.timeout = timeout
	return i2
}

func (i2 *i2Device) linkingMode(cmd Command, payload []byte) (err error) {
	err = i2.i1Device.SendCommand(cmd, payload)
	if err == nil {
		Log.Tracef("Waiting %s for response (Set-Button Pressed Controller/Responder)", i2.timeout)
		<-time.After(PropagationDelay(3, len(payload) > 0))
	}
	return err
}

// String returns the string "I2 Device (<address>)" where <address> is the destination
// address of the device
func (i2 *i2Device) String() string {
	return sprintf("I2 Device (%s)", i2.Address())
}

func (i2 *i2Device) EnterLinkingMode(group Group) error {
	return i2.linkingMode(EnterLinkingMode(group))
}

func (i2 *i2Device) EnterUnlinkingMode(group Group) error {
	return i2.linkingMode(EnterUnlinkingMode(group))
}

func (i2 *i2Device) ExitLinkingMode() error {
	return i2.SendCommand(ExitLinkingMode())
}

func (i2 *i2Device) LinkDatabase() (Linkable, error) {
	return i2, nil
}
