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

type Switch struct {
	Device
	timeout time.Duration
}

// NewSwitch is a factory function that will return the correctly
// configured switch based on the underlying device
func NewSwitch(info DeviceInfo, device Device, timeout time.Duration) *Switch {
	return &Switch{Device: device, timeout: timeout}
}

// Status sends a LightStatusRequest to determine the device's current
// level. For switched devices this is either 0 or 255, dimmable devices
// will be the current dim level between 0 and 255
func (sd *Switch) Status() (level int, err error) {
	ack, err := sd.Send(&Message{
		Flags:   StandardDirectMessage,
		Command: CmdLightStatusRequest,
	})
	if err == nil {
		level = int(ack.Command[2])
	}
	return level, err
}

func (sd *Switch) String() string {
	return fmt.Sprintf("Switch (%s)", sd.Address())
}

func (sd *Switch) Config() (config SwitchConfig, err error) {
	// SEE Dimmer.Config() notes for explanation of D1 and D2 (payload[0] and payload[1])
	err = sd.Device.SendCommand(CmdExtendedGetSet, []byte{0x00, 0x00})
	if err == nil {
		err = Receive(sd.Receive, sd.timeout, func(msg *Message) error {
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

func (sd *Switch) OperatingFlags() (flags LightFlags, err error) {
	commands := []Command{
		CmdGetOperatingFlags.SubCommand(0x00),
		CmdGetOperatingFlags.SubCommand(0x01),
		CmdGetOperatingFlags.SubCommand(0x02),
		CmdGetOperatingFlags.SubCommand(0x03),
		CmdGetOperatingFlags.SubCommand(0x05),
	}

	var ack *Message
	for i := 0; i < len(commands) && err == nil; i++ {
		ack, err = sd.Send(&Message{
			Flags:   StandardDirectMessage,
			Command: commands[i],
		})
		commands[i] = ack.Command
		flags[i] = commands[i][2]
	}
	return
}
