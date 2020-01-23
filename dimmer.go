// Copyright 2019 Andrew Bates
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
	"time"
)

// DimmerConfig includes the X10 configuration as well as default ramp
// and on levels
type DimmerConfig struct {
	// HouseCode is the device X10 house code
	HouseCode int

	// UnitCode is the device X10 unit code
	UnitCode int

	// Ramp is the default ramp rate
	Ramp int

	// OnLevel is the default on level
	OnLevel int

	// SNT is the Signal to Noise Threshold
	SNT int
}

// UnmarshalBinary will parse the byte buffer into the receiver
func (dc *DimmerConfig) UnmarshalBinary(buf []byte) error {
	if len(buf) < 14 {
		return ErrBufferTooShort
	}
	dc.HouseCode = int(buf[4])
	dc.UnitCode = int(buf[5])
	dc.Ramp = int(buf[6])
	dc.OnLevel = int(buf[7])
	dc.SNT = int(buf[8])
	return nil
}

// MarshalBinary will convert the DimmerConfig receiver to a byte string
func (dc *DimmerConfig) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 14)
	buf[4] = byte(dc.HouseCode)
	buf[5] = byte(dc.UnitCode)
	buf[6] = byte(dc.Ramp)
	buf[7] = byte(dc.OnLevel)
	buf[8] = byte(dc.SNT)
	return buf, nil
}

type Dimmer struct {
	*Switch
	timeout         time.Duration
	firmwareVersion FirmwareVersion
}

// NewDimmer is a factory function that will return a dimmer switch configured
// appropriately for the given firmware version.  All dimmers are switches, so
// the first argument is a Switch object used to compose the new dimmer
func NewDimmer(sw *Switch, timeout time.Duration, firmwareVersion FirmwareVersion) *Dimmer {
	dd := &Dimmer{
		Switch:          sw,
		timeout:         timeout,
		firmwareVersion: firmwareVersion,
	}
	return dd
}

func (dd *Dimmer) DimmerConfig() (config DimmerConfig, err error) {
	// The documentation talks about D1 (payload[0]) being the button/group number, but my
	// SwitchLinc dimmers all return the same information regardless of
	// the value of D1.  I think D1 is maybe only relevant on KeyPadLinc dimmers.
	//
	// D2 is 0x00 for requests
	err = dd.Switch.SendCommand(CmdExtendedGetSet, []byte{0x01, 0x00})
	if err == nil {
		err = Receive(dd.Switch, dd.timeout, func(msg *Message) error {
			if msg.Command == CmdExtendedGetSet {
				err = config.UnmarshalBinary(msg.Payload)
				if err == nil {
					err = ErrReceiveComplete
				}
			}
			return err
		})
	}
	return config, err
}

func (dd *Dimmer) SendCommand(cmd Command, payload []byte) error {
	if cmd[1] == CmdLightOnAtRamp[1] {
		if dd.firmwareVersion >= 0x43 {
			cmd = CmdLightOnAtRampV67.SubCommand(int(cmd[2]))
		}
	} else if cmd[1] == CmdLightOffAtRamp[1] {
		if dd.firmwareVersion >= 0x43 {
			cmd = CmdLightOffAtRampV67.SubCommand(int(cmd[2]))
		}
	}
	return dd.Switch.SendCommand(cmd, payload)
}

func (dd *Dimmer) String() string {
	return fmt.Sprintf("Dimmer (%s)", dd.Address())
}
