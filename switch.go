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

// SwitchConfig contains the HouseCode and UnitCode for a switch's
// X10 configuration
type SwitchConfig struct {
	// HouseCode is the X10 house code of the switch or dimmer
	HouseCode int

	// UnitCode is the X10 unit code of the switch or dimmer
	UnitCode int
}

// UnmarshalBinary takes the given byte buffer and unmarshals it into
// the receiver
func (sc *SwitchConfig) UnmarshalBinary(buf []byte) error {
	if len(buf) < 14 {
		return ErrBufferTooShort
	}
	sc.HouseCode = int(buf[4])
	sc.UnitCode = int(buf[5])
	return nil
}

// MarshalBinary will convert the receiver into a serialized byte buffer
func (sc *SwitchConfig) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 14)
	buf[4] = byte(sc.HouseCode)
	buf[5] = byte(sc.UnitCode)
	return buf, nil
}

// Switch is any implementation that satisfies the following switch functions
type Switch interface {
	Device

	// On changes the device state to on
	On() error

	// Off changes the device state to off
	Off() error

	// Status returns the device's current status. The returned level is 0 for off and
	// 255 for fully on. A switch alternates only between 0 and 255 while a dimmer can
	// be any value in between
	Status() (level int, err error)

	// OperatingFlags queries the device and returns the LightFlags
	OperatingFlags() (LightFlags, error)

	// SetLED enable or disables the device status LED. When false, the
	// status LED should be extuingished and when the flag is true the device
	// status LED should be lit
	SetLED(flag bool) error

	// SetLoadSense enables or disables the device's load sense
	SetLoadSense(flag bool) error

	// SetProgramLock will set the program lock flag on the device. Program lock prevents
	// local changes at the device
	SetProgramLock(flag bool) error

	// SetResumeDim sets the resume dim flag on the device.  If set, the device will
	// return to the previous ramp level when turned on
	SetResumeDim(flag bool) error

	// SetTxLED will enable the device status LED to flash on
	// insteon traffic
	SetTxLED(flag bool) error

	// SetX10Address sets the X10 house and unit codes
	SetX10Address(button int, houseCode, unitCode byte) error

	// SwitchConfig queries the device and returns the returned configuration
	SwitchConfig() (SwitchConfig, error)
}

// LinkableSwitch represents a switch that contains an Insteon version 2
// (or higher) engine.  A LinkableSwitch contains a remotely manageable
// All-Link database
type LinkableSwitch interface {
	Switch
	Linkable
}

// LightFlags are the operating flags for a switch or dimmer
type LightFlags [5]byte

// ProgramLock indicates if the Program Lock flag is set
func (lf LightFlags) ProgramLock() bool { return lf[0]&01 == 0x01 }

// TxLED indicates whether the status LED will flash when Insteon traffic is received
func (lf LightFlags) TxLED() bool { return lf[0]&0x02 == 0x02 }

// ResumeDim indicates if the switch will return to the previous on level or
// will return to the default on level
func (lf LightFlags) ResumeDim() bool { return lf[0]&0x04 == 0x04 }

// LED indicates if the status LED is enabled
func (lf LightFlags) LED() bool { return lf[0]&0x08 == 0x08 }

// LoadSense indicates if the device should activate when a load is
// added
func (lf LightFlags) LoadSense() bool { return lf[0]&0x10 == 0x10 }

// DBDelta indicates the number of changes that have been written to the all-link
// database
func (lf LightFlags) DBDelta() int { return int(lf[1]) }

// SNR indicates the current signal-to-noise ratio
func (lf LightFlags) SNR() int { return int(lf[2]) }

// SNRFailCount ...
// @TODO research what this value means
func (lf LightFlags) SNRFailCount() int { return int(lf[3]) }

// X10Enabled indicates if the device will respond to X10 commands
func (lf LightFlags) X10Enabled() bool { return lf[4]&0x01 != 0x01 }

// ErrorBlink enables the device to blink the status LED when errors occur
// TODO: Confirm this description is correct
func (lf LightFlags) ErrorBlink() bool { return lf[4]&0x02 == 0x02 }

// CleanupReport enables sending All-link cleanup reports
// TODO: Confirm this description is correct
func (lf LightFlags) CleanupReport() bool { return lf[4]&0x04 == 0x04 }

type switchedDevice struct {
	Device
	timeout time.Duration
}

type linkableSwitch struct {
	LinkableDevice
	*switchedDevice
	timeout time.Duration
}

// NewSwitch is a factory function that will return the correctly
// configured switch based on the underlying device
func NewSwitch(device Device, timeout time.Duration) Switch {
	sw := &switchedDevice{Device: device, timeout: timeout}
	if linkable, ok := device.(LinkableDevice); ok {
		return &linkableSwitch{LinkableDevice: linkable, switchedDevice: sw}
	}
	return sw
}

func (sd *switchedDevice) On() error  { return extractError(sd.SendCommand(CmdLightOn, nil)) }
func (sd *switchedDevice) Off() error { return extractError(sd.SendCommand(CmdLightOff, nil)) }

// Status sends a LightStatusRequest to determine the device's current
// level. For switched devices this is either 0 or 255, dimmable devices
// will be the current dim level between 0 and 255
func (sd *switchedDevice) Status() (level int, err error) {
	response, err := sd.SendCommand(CmdLightStatusRequest, nil)
	if err == nil {
		level = int(response[2])
	}
	return level, err
}

func (sd *switchedDevice) String() string {
	return fmt.Sprintf("Switch (%s)", sd.Address())
}

func (sd *switchedDevice) SetX10Address(button int, houseCode, unitCode byte) error {
	return extractError(sd.SendCommand(CmdExtendedGetSet, []byte{byte(button), 0x04, houseCode, unitCode}))
}

func (sd *switchedDevice) SwitchConfig() (config SwitchConfig, err error) {
	// SEE DimmerConfig() notes for explanation of D1 and D2 (payload[0] and payload[1])
	_, err = sd.Device.SendCommand(CmdExtendedGetSet, []byte{0x00, 0x00})
	timeout := time.Now().Add(sd.timeout)
	for err == nil {
		var msg *Message
		msg, err = sd.Device.Receive()
		if err == nil {
			if msg.Command == CmdExtendedGetSet {
				err = config.UnmarshalBinary(msg.Payload)
				break
			} else if timeout.Before(time.Now()) {
				err = ErrReadTimeout
			}
		}
	}
	return config, err
}

func (sd *switchedDevice) setOperatingFlags(flags byte, conditional bool) error {
	if conditional {
		return extractError(sd.SendCommand(CmdSetOperatingFlags.SubCommand(int(flags)), nil))
	}
	return extractError(sd.SendCommand(CmdSetOperatingFlags.SubCommand(int(flags)+1), nil))
}

func (sd *switchedDevice) SetProgramLock(flag bool) error { return sd.setOperatingFlags(0, flag) }
func (sd *switchedDevice) SetTxLED(flag bool) error       { return sd.setOperatingFlags(2, flag) }
func (sd *switchedDevice) SetResumeDim(flag bool) error   { return sd.setOperatingFlags(4, flag) }
func (sd *switchedDevice) SetLoadSense(flag bool) error   { return sd.setOperatingFlags(6, !flag) }
func (sd *switchedDevice) SetLED(flag bool) error         { return sd.setOperatingFlags(8, !flag) }

func (sd *switchedDevice) OperatingFlags() (flags LightFlags, err error) {
	commands := []Command{
		CmdGetOperatingFlags.SubCommand(0x00),
		CmdGetOperatingFlags.SubCommand(0x01),
		CmdGetOperatingFlags.SubCommand(0x02),
		CmdGetOperatingFlags.SubCommand(0x03),
		CmdGetOperatingFlags.SubCommand(0x05),
	}

	for i := 0; i < len(commands) && err == nil; i++ {
		commands[i], err = sd.SendCommand(commands[i], nil)
		flags[i] = commands[i][2]
	}
	return
}
