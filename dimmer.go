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

// Dimmer is any device that satisfies the following interface
type Dimmer interface {
	// Switch is the underlying Switch object that this dimmer is
	// composed of
	Switch

	// OnLevel turns the dimmer on to the specified level (0-255)
	OnLevel(level int) error

	// OnFast will turn the dimmer on to the specified level as fast as
	// possible (bypassing the default ramp rate)
	OnFast(level int) error

	// Brighten will increase the brightness one step
	Brighten() error

	// Dim will decrease the brightness one step
	Dim() error

	// StartBrighten behaves as if the dimmers' brighten button is being
	// continuously pressed
	StartBrighten() error

	// StartDim behaves as if the dimmer's dim button is being
	// continuously pressed
	StartDim() error

	// StopChange cancels the previous StartBrighten or StartDim commands
	StopChange() error

	// InstantChange will set the dimmer's on level instantaneously (with no
	// ramp rate whatsoever
	InstantChange(level int) error

	// SetStatus changes the dimmer's status LED
	SetStatus(level int) error

	// OnAtRamp will turn the dimmer on to the given level at the given ramp rate
	OnAtRamp(level, ramp int) error

	// OffAtRamp will turn the dimmer off using the specified ramp rate
	OffAtRamp(ramp int) error

	// SetDefaultRamp sets the ramp rate that is used for the default on-level
	// when the on or off button is tapped
	SetDefaultRamp(level int) error

	// SetDefaultOnLevel sets the lighting level that is recalled when the on
	// button is tapped
	SetDefaultOnLevel(level int) error

	// DimmerConfig queries the dimmer and returns the configuration
	DimmerConfig() (DimmerConfig, error)
}

// LinkableDimmer represents a Dimmer switch that supports remote
// linking (Insteon Engine version 2 or higher)
type LinkableDimmer interface {
	Dimmer
	Linkable
}

type dimmer struct {
	Switch
	timeout         time.Duration
	firmwareVersion FirmwareVersion
}

type linkableDimmer struct {
	LinkableSwitch
	*dimmer
}

// NewDimmer is a factory function that will return a dimmer switch configured
// appropriately for the given firmware version.  All dimmers are switches, so
// the first argument is a Switch object used to compose the new dimmer
func NewDimmer(sw Switch, timeout time.Duration, firmwareVersion FirmwareVersion) Dimmer {
	dd := &dimmer{
		Switch:          sw,
		timeout:         timeout,
		firmwareVersion: firmwareVersion,
	}
	if linkable, ok := sw.(LinkableSwitch); ok {
		return &linkableDimmer{LinkableSwitch: linkable, dimmer: dd}
	}
	return dd
}

func (dd *dimmer) OnLevel(level int) error {
	_, err := dd.SendCommand(CmdLightOn.SubCommand(level), nil)
	return err
}

func (dd *dimmer) OnFast(level int) error {
	_, err := dd.SendCommand(CmdLightOnFast.SubCommand(level), nil)
	return err
}

func (dd *dimmer) Brighten() error {
	_, err := dd.SendCommand(CmdLightBrighten, nil)
	return err
}

func (dd *dimmer) Dim() error {
	_, err := dd.SendCommand(CmdLightDim, nil)
	return err
}

func (dd *dimmer) StartBrighten() error {
	_, err := dd.SendCommand(CmdLightStartManual.SubCommand(1), nil)
	return err
}

func (dd *dimmer) StartDim() error {
	_, err := dd.SendCommand(CmdLightStartManual.SubCommand(0), nil)
	return err
}

func (dd *dimmer) StopChange() error {
	_, err := dd.SendCommand(CmdLightStopManual, nil)
	return err
}

func (dd *dimmer) InstantChange(level int) error {
	_, err := dd.SendCommand(CmdLightInstantChange.SubCommand(level), nil)
	return err
}

func (dd *dimmer) SetStatus(level int) error {
	_, err := dd.SendCommand(CmdLightSetStatus.SubCommand(level), nil)
	return err
}

func (dd *dimmer) OnAtRamp(level, ramp int) (err error) {
	levelRamp := byte(level) << 4
	levelRamp |= byte(ramp) & 0x0f
	if dd.firmwareVersion >= 0x43 {
		_, err = dd.SendCommand(CmdLightOnAtRampV67.SubCommand(int(levelRamp)), nil)
	} else {
		_, err = dd.SendCommand(CmdLightOnAtRamp.SubCommand(int(levelRamp)), nil)
	}
	return err
}

func (dd *dimmer) OffAtRamp(ramp int) (err error) {
	if dd.firmwareVersion >= 0x43 {
		_, err = dd.SendCommand(CmdLightOffAtRampV67.SubCommand(0x0f&ramp), nil)
	} else {
		_, err = dd.SendCommand(CmdLightOffAtRamp.SubCommand(0x0f&ramp), nil)
	}
	return err
}

func (dd *dimmer) String() string {
	return fmt.Sprintf("Dimmer (%s)", dd.Address())
}

func (dd *dimmer) SetDefaultRamp(rate int) error {
	// See notes on DimmerConfig() for information about D1 (payload[0])
	_, err := dd.SendCommand(CmdExtendedGetSet, []byte{0x01, 0x05, byte(rate)})
	return err
}

func (dd *dimmer) SetDefaultOnLevel(level int) error {
	// See notes on DimmerConfig() for information about D1 (payload[0])
	_, err := dd.SendCommand(CmdExtendedGetSet, []byte{0x01, 0x06, byte(level)})
	return err
}

func (dd *dimmer) DimmerConfig() (config DimmerConfig, err error) {
	// The documentation talks about D1 (payload[0]) being the button/group number, but my
	// SwitchLinc dimmers all return the same information regardless of
	// the value of D1.  I think D1 is maybe only relevant on KeyPadLinc dimmers.
	//
	// D2 is 0x00 for requests
	_, err = dd.Switch.SendCommand(CmdExtendedGetSet, []byte{0x01, 0x00})
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
