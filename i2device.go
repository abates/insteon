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
}

// newI2Device will construct an device object that can communicate with version 2
// Insteon engines
func newI2Device(bus Bus, info DeviceInfo) *i2Device {
	i2 := &i2Device{i1Device: newI1Device(bus, info)}
	i2.linkdb.device = i2
	i2.linkdb.config = bus.Config()
	return i2
}

func (i2 *i2Device) linkingMode(cmd Command, payload []byte) (err error) {
	_, err = i2.SendCommand(cmd, payload)
	if err == nil {
		// allow linking mode to activate
		time.Sleep(PropagationDelay(i2.i1Device.bus.Config().TTL, len(payload) > 0))
	}
	return err
}

// String returns the string "I2 Device (<address>)" where <address> is the destination
// address of the device
func (i2 *i2Device) String() string {
	return sprintf("I2 Device (%s)", i2.Info().Address)
}

func (i2 *i2Device) EnterLinkingMode(group Group) error {
	return i2.linkingMode(CmdEnterLinkingMode.SubCommand(int(group)), nil)
}

func (i2 *i2Device) EnterUnlinkingMode(group Group) error {
	return i2.linkingMode(CmdEnterUnlinkingMode.SubCommand(int(group)), nil)
}

func (i2 *i2Device) ExitLinkingMode() error {
	_, err := i2.SendCommand(CmdExitLinkingMode, nil)
	return err
}

func (i2 *i2Device) LinkDatabase() (Linkable, error) {
	return i2, nil
}
