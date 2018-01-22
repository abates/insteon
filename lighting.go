package insteon

import "fmt"

var (
	CmdLightOn              = Commands.RegisterStd("Light On", []byte{0x01, 0x02}, MsgTypeDirect, 0x11, 0x00)
	CmdLightOnFast          = Commands.RegisterStd("Light On Fast", []byte{0x01, 0x02}, MsgTypeDirect, 0x12, 0x00)
	CmdLightOff             = Commands.RegisterStd("Light Off", []byte{0x01, 0x02}, MsgTypeDirect, 0x13, 0x00)
	CmdLightOffFast         = Commands.RegisterStd("Light Off Fast", []byte{0x01, 0x02}, MsgTypeDirect, 0x14, 0x00)
	CmdLightBrighten        = Commands.RegisterStd("Light Brighten", []byte{0x01, 0x02}, MsgTypeDirect, 0x15, 0x00)
	CmdLightDim             = Commands.RegisterStd("Light Dim", []byte{0x01, 0x02}, MsgTypeDirect, 0x16, 0x00)
	CmdLightStartManualDown = Commands.RegisterStd("Light Manual Down", []byte{0x01, 0x02}, MsgTypeDirect, 0x17, 0x00)
	CmdLightStartManualUp   = Commands.RegisterStd("Light Manual Up", []byte{0x01, 0x02}, MsgTypeDirect, 0x17, 0x01)
	CmdLightStopManual      = Commands.RegisterStd("Light Stop Manual", []byte{0x01, 0x02}, MsgTypeDirect, 0x18, 0x00)
	CmdLightStatusRequest   = Commands.RegisterStd("Light Status Req", []byte{0x01, 0x02}, MsgTypeDirect, 0x19, 0x00)
	CmdGetOperatingFlags    = Commands.RegisterStd("Get Operating Flags", []byte{0x01, 0x02}, MsgTypeDirect, 0x1f, 0x00)
	CmdSetOperatingFlags    = Commands.RegisterStd("Set Operating Flags", []byte{0x01, 0x02}, MsgTypeDirect, 0x20, 0x00)
	CmdLightInstantChange   = Commands.RegisterStd("Light Instant Change", []byte{0x01, 0x02}, MsgTypeDirect, 0x21, 0x00)
	CmdTapSetButtonOnce     = Commands.RegisterStd("Tap Set Button Once", []byte{0x01, 0x02}, MsgTypeDirect, 0x25, 0x01)
	CmdTapSetButtonTwice    = Commands.RegisterStd("Tap Set Button Twice", []byte{0x01, 0x02}, MsgTypeDirect, 0x25, 0x02)
	CmdLightOnRamp          = Commands.RegisterStd("Light On Ramp", []byte{0x01, 0x02}, MsgTypeDirect, 0x2e, 0x00)
	CmdLightOffRamp         = Commands.RegisterStd("Light Off Ramp", []byte{0x01, 0x02}, MsgTypeDirect, 0x2f, 0x00)
)

func init() {
	Devices.Register(0x01, dimmableLightingFactory)
	Devices.Register(0x02, switchedLightingFactory)
}

type SwitchedDevice struct {
	Device
}

func (sd *SwitchedDevice) On() error {
	return ErrNotImplemented
}

func (sd *SwitchedDevice) Off() error {
	return ErrNotImplemented
}

// Status sends a LightStatusRequest to determine the device's current
// level. For switched devices this is either 0 or 255, dimmable devices
// will be the current dim level between 0 and 255
func (sd *SwitchedDevice) Status() (int, error) {
	return 0, ErrNotImplemented
}

type LightFlags byte

func (sd *SwitchedDevice) OperatingFlags() (LightFlags, error) {
	return LightFlags(0x00), ErrNotImplemented
}

func (sd *SwitchedDevice) SetOperatingFlags(LightFlags) error {
	return ErrNotImplemented
}

func (sd *SwitchedDevice) ManualOn() error {
	return ErrNotImplemented
}

func (sd *SwitchedDevice) ManualOff() error {
	return ErrNotImplemented
}

func (sd *SwitchedDevice) FlashLED() error {
	return ErrNotImplemented
}

func (sd *SwitchedDevice) String() string {
	return fmt.Sprintf("Switch (%s)", sd.Address())
}

type DimmableDevice struct {
	*SwitchedDevice
}

func (dd *DimmableDevice) On(level int) error {
	return ErrNotImplemented
}

func (dd *DimmableDevice) OnFast(level int) error {
	return ErrNotImplemented
}

func (dd *DimmableDevice) Brighten() error {
	return ErrNotImplemented
}

func (dd *DimmableDevice) Dim() error {
	return ErrNotImplemented
}

func (dd *DimmableDevice) StartBrighten() error {
	return ErrNotImplemented
}

func (dd *DimmableDevice) StartDim() error {
	return ErrNotImplemented
}

func (dd *DimmableDevice) StopChange() error {
	return ErrNotImplemented
}

func (dd *DimmableDevice) InstantChange(level int) error {
	return ErrNotImplemented
}

func (dd *DimmableDevice) SetStatus(level int) error {
	return ErrNotImplemented
}

func (dd *DimmableDevice) OnAtRate(level, rate int) error {
	return ErrNotImplemented
}

func (dd *DimmableDevice) OffAtRate(rate int) error {
	return ErrNotImplemented
}

func (dd *DimmableDevice) String() string {
	return fmt.Sprintf("Dimmable Light (%s)", dd.Address())
}

func dimmableLightingFactory(device Device) Device {
	Log.Debugf("Returning dimmable device with underlying device %T", device)
	return &DimmableDevice{&SwitchedDevice{device}}
}

func switchedLightingFactory(device Device) Device {
	return &SwitchedDevice{device}
}
