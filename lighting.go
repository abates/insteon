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

var (
	// LightingCategories match the two device categories known to be lighting
	// devices.  0x01 are dimmable devices and 0x02 are switched devices
	LightingCategories = []Category{Category(1), Category(2)}
)

func init() {
	Devices.Register(0x01, dimmableDeviceFactory)
	Devices.Register(0x02, switchedDeviceFactory)
}

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

	// String returns a string representation of the device
	String() string
}

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

// Dimmer is any device that satisfies the following interface
type Dimmer interface {
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

type i1SwitchedDevice struct {
	*I1Device
	Switch
}

func (i1 *i1SwitchedDevice) String() string { return i1.Switch.String() }

type i2SwitchedDevice struct {
	*I2Device
	Switch
}

func (i2 *i2SwitchedDevice) String() string { return i2.Switch.String() }

type i2CsSwitchedDevice struct {
	*I2CsDevice
	Switch
}

func (i2 *i2CsSwitchedDevice) String() string { return i2.Switch.String() }

type switchedDevice struct {
	Commandable
	firmwareVersion FirmwareVersion

	recvCh           <-chan *Message
	downstreamRecvCh chan<- *Message
}

func (sd *switchedDevice) process() {
	for message := range sd.recvCh {
		sd.downstreamRecvCh <- message
	}
}

func (sd *switchedDevice) On() error {
	_, err := sd.SendCommand(CmdLightOn, nil)
	return err
}

func (sd *switchedDevice) Off() error {
	_, err := sd.SendCommand(CmdLightOff, nil)
	return err
}

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
	address := ""
	if addr, ok := sd.Commandable.(Addressable); ok {
		address = fmt.Sprintf(" (%s)", addr.Address())
	}
	return fmt.Sprintf("Switch%s", address)
}

func (sd *switchedDevice) SetX10Address(button int, houseCode, unitCode byte) error {
	_, err := sd.SendCommand(CmdExtendedGetSet, []byte{byte(button), 0x04, houseCode, unitCode})
	return err
}

func (sd *switchedDevice) SwitchConfig() (config SwitchConfig, err error) {
	// SEE DimmerConfig() notes for explanation of D1 and D2 (payload[0] and payload[1])
	recvCh, err := sd.SendCommandAndListen(CmdExtendedGetSet, []byte{0x00, 0x00})
	for response := range recvCh {
		if response.Message.Command == CmdExtendedGetSet {
			err = config.UnmarshalBinary(response.Message.Payload)
			response.DoneCh <- response
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

type i1DimmableDevice struct {
	Dimmer
	*i1SwitchedDevice
}

func (i1 *i1DimmableDevice) String() string { return i1.Dimmer.String() }

type i2DimmableDevice struct {
	Dimmer
	*i2SwitchedDevice
}

func (i2 *i2DimmableDevice) String() string { return i2.Dimmer.String() }

type i2CsDimmableDevice struct {
	Dimmer
	*i2CsSwitchedDevice
}

func (i2cs *i2CsDimmableDevice) String() string { return i2cs.Dimmer.String() }

type dimmableDevice struct {
	Switch
	Commandable
	firmwareVersion FirmwareVersion

	recvCh           <-chan *Message
	downstreamRecvCh chan<- *Message
}

func (dd *dimmableDevice) process() {
	for message := range dd.recvCh {
		dd.downstreamRecvCh <- message
	}
}

func (dd *dimmableDevice) OnLevel(level int) error {
	_, err := dd.SendCommand(CmdLightOn.SubCommand(level), nil)
	return err
}

func (dd *dimmableDevice) OnFast(level int) error {
	_, err := dd.SendCommand(CmdLightOnFast.SubCommand(level), nil)
	return err
}

func (dd *dimmableDevice) Brighten() error {
	_, err := dd.SendCommand(CmdLightBrighten, nil)
	return err
}

func (dd *dimmableDevice) Dim() error {
	_, err := dd.SendCommand(CmdLightDim, nil)
	return err
}

func (dd *dimmableDevice) StartBrighten() error {
	_, err := dd.SendCommand(CmdLightStartManual.SubCommand(1), nil)
	return err
}

func (dd *dimmableDevice) StartDim() error {
	_, err := dd.SendCommand(CmdLightStartManual.SubCommand(0), nil)
	return err
}

func (dd *dimmableDevice) StopChange() error {
	_, err := dd.SendCommand(CmdLightStopManual, nil)
	return err
}

func (dd *dimmableDevice) InstantChange(level int) error {
	_, err := dd.SendCommand(CmdLightInstantChange.SubCommand(level), nil)
	return err
}

func (dd *dimmableDevice) SetStatus(level int) error {
	_, err := dd.SendCommand(CmdLightSetStatus.SubCommand(level), nil)
	return err
}

func (dd *dimmableDevice) OnAtRamp(level, ramp int) (err error) {
	levelRamp := byte(level) << 4
	levelRamp |= byte(ramp) & 0x0f
	if dd.firmwareVersion >= 0x43 {
		_, err = dd.SendCommand(CmdLightOnAtRampV67.SubCommand(int(levelRamp)), nil)
	} else {
		_, err = dd.SendCommand(CmdLightOnAtRamp.SubCommand(int(levelRamp)), nil)
	}
	return err
}

func (dd *dimmableDevice) OffAtRamp(ramp int) (err error) {
	if dd.firmwareVersion >= 0x43 {
		_, err = dd.SendCommand(CmdLightOffAtRampV67.SubCommand(0x0f&ramp), nil)
	} else {
		_, err = dd.SendCommand(CmdLightOffAtRamp.SubCommand(0x0f&ramp), nil)
	}
	return err
}

func (dd *dimmableDevice) String() string {
	address := ""
	if addr, ok := dd.Commandable.(Addressable); ok {
		address = fmt.Sprintf(" (%s)", addr.Address())
	}
	return fmt.Sprintf("Dimmable Light%s", address)
}

func (dd *dimmableDevice) SetDefaultRamp(rate int) error {
	// See notes on DimmerConfig() for information about D1 (payload[0])
	_, err := dd.SendCommand(CmdExtendedGetSet, []byte{0x01, 0x05, byte(rate)})
	return err
}

func (dd *dimmableDevice) SetDefaultOnLevel(level int) error {
	// See notes on DimmerConfig() for information about D1 (payload[0])
	_, err := dd.SendCommand(CmdExtendedGetSet, []byte{0x01, 0x06, byte(level)})
	return err
}

func (dd *dimmableDevice) DimmerConfig() (config DimmerConfig, err error) {
	// The documentation talks about D1 (payload[0]) being the button/group number, but my
	// SwitchLinc dimmers all return the same information regardless of
	// the value of D1.  I think D1 is maybe only relevant on KeyPadLinc dimmers.
	//
	// D2 is 0x00 for requests
	recvCh, err := dd.SendCommandAndListen(CmdExtendedGetSet, []byte{0x01, 0x00})
	for response := range recvCh {
		if response.Message.Command == CmdExtendedGetSet {
			err = config.UnmarshalBinary(response.Message.Payload)
			response.DoneCh <- response
		}
	}
	return config, err
}

func switchedDeviceFactory(info DeviceInfo, address Address, sendCh chan<- *MessageRequest, recvCh <-chan *Message, timeout time.Duration) (device Device, err error) {
	downstreamRecvCh := make(chan *Message, 1)
	sd := &switchedDevice{
		firmwareVersion: info.FirmwareVersion,

		recvCh:           recvCh,
		downstreamRecvCh: downstreamRecvCh,
	}

	switch info.EngineVersion {
	case VerI1:
		device = &i1SwitchedDevice{
			I1Device: NewI1Device(address, sendCh, downstreamRecvCh, timeout),
			Switch:   sd,
		}
	case VerI2:
		device = &i2SwitchedDevice{
			I2Device: NewI2Device(address, sendCh, downstreamRecvCh, timeout),
			Switch:   sd,
		}
	case VerI2Cs:
		device = &i2CsSwitchedDevice{
			I2CsDevice: NewI2CsDevice(address, sendCh, downstreamRecvCh, timeout),
			Switch:     sd,
		}
	}

	sd.Commandable = device
	go sd.process()
	return
}

func dimmableDeviceFactory(info DeviceInfo, address Address, sendCh chan<- *MessageRequest, recvCh <-chan *Message, timeout time.Duration) (device Device, err error) {
	downstreamRecvCh := make(chan *Message, 1)
	sd, err := switchedDeviceFactory(info, address, sendCh, downstreamRecvCh, timeout)
	dd := &dimmableDevice{
		Commandable:     sd,
		firmwareVersion: info.FirmwareVersion,

		recvCh:           recvCh,
		downstreamRecvCh: downstreamRecvCh,
	}

	switch sw := sd.(type) {
	case *i1SwitchedDevice:
		dd.Switch = sw.Switch
		device = &i1DimmableDevice{
			i1SwitchedDevice: sw,
			Dimmer:           dd,
		}
	case *i2SwitchedDevice:
		dd.Switch = sw.Switch
		device = &i2DimmableDevice{
			i2SwitchedDevice: sw,
			Dimmer:           dd,
		}
	case *i2CsSwitchedDevice:
		dd.Switch = sw.Switch
		device = &i2CsDimmableDevice{
			i2CsSwitchedDevice: sw,
			Dimmer:             dd,
		}
	}
	go dd.process()
	return device, err
}
