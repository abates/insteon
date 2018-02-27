package insteon

var (
	CmdLightOn            = Commands.RegisterStd("Light On", []byte{0x01, 0x02}, MsgTypeDirect, 0x11, 0xff)
	CmdLightOnFast        = Commands.RegisterStd("Light On Fast", []byte{0x01, 0x02}, MsgTypeDirect, 0x12, 0x00)
	CmdLightOff           = Commands.RegisterStd("Light Off", []byte{0x01, 0x02}, MsgTypeDirect, 0x13, 0x00)
	CmdLightOffFast       = Commands.RegisterStd("Light Off Fast", []byte{0x01, 0x02}, MsgTypeDirect, 0x14, 0x00)
	CmdLightBrighten      = Commands.RegisterStd("Light Brighten", []byte{0x01, 0x02}, MsgTypeDirect, 0x15, 0x00)
	CmdLightDim           = Commands.RegisterStd("Light Dim", []byte{0x01, 0x02}, MsgTypeDirect, 0x16, 0x00)
	CmdLightStartManual   = Commands.RegisterStd("Light Start Manual Change", []byte{0x01, 0x01}, MsgTypeDirect, 0x17, 0x00)
	CmdLightStopManual    = Commands.RegisterStd("Light Stop Manual Change", []byte{0x01, 0x02}, MsgTypeDirect, 0x18, 0x00)
	CmdLightStatusRequest = Commands.RegisterStd("Light Status Req", []byte{0x01, 0x02}, MsgTypeDirect, 0x19, 0x00)
	CmdLightInstantChange = Commands.RegisterStd("Light Instant Change", []byte{0x01, 0x02}, MsgTypeDirect, 0x21, 0x00)
	CmdLightManualOn      = Commands.RegisterStd("Tap Set Button Once", []byte{0x01, 0x02}, MsgTypeDirect, 0x22, 0x01)
	CmdLightManualOff     = Commands.RegisterStd("Tap Set Button Once", []byte{0x01, 0x02}, MsgTypeDirect, 0x23, 0x01)
	CmdTapSetButtonOnce   = Commands.RegisterStd("Tap Set Button Once", []byte{0x01, 0x02}, MsgTypeDirect, 0x25, 0x01)
	CmdTapSetButtonTwice  = Commands.RegisterStd("Tap Set Button Twice", []byte{0x01, 0x02}, MsgTypeDirect, 0x25, 0x02)
	CmdLightSetStatus     = Commands.RegisterStd("Update dimmer LEDs", []byte{0x01}, MsgTypeDirect, 0x27, 0x00)
	CmdLightOnAtRamp      = Commands.RegisterStd("Light On Ramp", []byte{0x01, 0x02}, MsgTypeDirect, 0x2e, 0x00)
	CmdLightOffAtRamp     = Commands.RegisterStd("Light Off Ramp", []byte{0x01, 0x02}, MsgTypeDirect, 0x2f, 0x00)
)

func init() {
	Devices.Register(0x01, dimmableLightingFactory)
	Devices.Register(0x02, switchedLightingFactory)
}

type Switch interface {
	On() error
	ManualOn() error
	Off() error
	ManualOff() error
	Status() (level int, err error)
	OperatingFlags() (*LightFlags, error)
	SetOperatingFlags(*LightFlags) error
}

type Dimmer interface {
	Switch
	OnLevel(level int) error
	OnFast(level int) error
	Brighten() error
	Dim() error
	StartBrighten() error
	StartDim() error
	StopChange() error
	InstantChange(level int) error
	SetStatus(level int) error
	OnAtRamp(level, ramp int) error
	OffAtRamp(ramp int) error
}

type SwitchedDevice struct {
	Device
}

func (sd *SwitchedDevice) DevCat() (Category, error) { return Category{0x02, 0x00}, nil }

func (sd *SwitchedDevice) On() error {
	_, err := SendStandardCommand(sd.Connection(), CmdLightOn)
	return err
}

func (sd *SwitchedDevice) Off() error {
	_, err := SendStandardCommand(sd.Connection(), CmdLightOff)
	return err
}

// Status sends a LightStatusRequest to determine the device's current
// level. For switched devices this is either 0 or 255, dimmable devices
// will be the current dim level between 0 and 255
func (sd *SwitchedDevice) Status() (level int, err error) {
	ack, err := SendStandardCommand(sd.Connection(), CmdLightStatusRequest)
	if err == nil {
		level = int(ack.Command.Cmd[1])
	}
	return level, err
}

type LightFlags struct {
	ProgramLock bool `json:program_lock`
	TransmitLED bool `json:transmit_led`
	ResumeDim   bool `json:resume_dim`
	LED         bool `json:led`
	LoadSense   bool `json:load_sense`
}

func (lf *LightFlags) MarshalBinary() ([]byte, error) {
	buf := []byte{0x00}
	if lf.ProgramLock {
		buf[0] |= 0x01
	}
	if lf.TransmitLED {
		buf[0] |= 0x02
	}
	if lf.ResumeDim {
		buf[0] |= 0x08
	}
	if lf.LED {
		buf[0] |= 0x10
	}
	if lf.LoadSense {
		buf[0] |= 0x20
	}
	return buf, nil
}

func (lf *LightFlags) UnmarshalBinary(buf []byte) error {
	if len(buf) < 1 {
		return ErrBufferTooShort
	}
	lf.ProgramLock = buf[0]&0x01 == 0x01
	lf.TransmitLED = buf[0]&0x02 == 0x02
	lf.ResumeDim = buf[0]&0x08 == 0x08
	lf.LED = buf[0]&0x10 == 0x10
	lf.LoadSense = buf[0]&0x20 == 0x20
	return nil
}

func (sd *SwitchedDevice) OperatingFlags() (*LightFlags, error) {
	flags := &LightFlags{}
	ack, err := SendStandardCommand(sd.Connection(), CmdGetOperatingFlags)
	if err == nil {
		flags.UnmarshalBinary([]byte{ack.Command.Cmd[1]})
	}
	return flags, err
}

func (sd *SwitchedDevice) SetOperatingFlags(flags *LightFlags) error {
	buf, err := flags.MarshalBinary()
	if err == nil {
		_, err = SendStandardCommand(sd.Connection(), CmdSetOperatingFlags.SubCommand(int(buf[0])))
	}
	return err
}

func (sd *SwitchedDevice) ManualOn() error {
	_, err := SendStandardCommand(sd.Connection(), CmdLightManualOn)
	return err
}

func (sd *SwitchedDevice) ManualOff() error {
	_, err := SendStandardCommand(sd.Connection(), CmdLightManualOff)
	return err
}

func (sd *SwitchedDevice) String() string {
	return sprintf("Switch (%s)", sd.Address())
}

type DimmableDevice struct {
	*SwitchedDevice
}

func (dd *DimmableDevice) DevCat() (Category, error) { return Category{0x01, 0x00}, nil }

func (dd *DimmableDevice) OnLevel(level int) error {
	_, err := SendStandardCommand(dd.Connection(), CmdLightOn.SubCommand(level))
	return err
}

func (dd *DimmableDevice) OnFast(level int) error {
	_, err := SendStandardCommand(dd.Connection(), CmdLightOnFast)
	return err
}

func (dd *DimmableDevice) Brighten() error {
	_, err := SendStandardCommand(dd.Connection(), CmdLightBrighten)
	return err
}

func (dd *DimmableDevice) Dim() error {
	_, err := SendStandardCommand(dd.Connection(), CmdLightDim)
	return err
}

func (dd *DimmableDevice) StartBrighten() error {
	_, err := SendStandardCommand(dd.Connection(), CmdLightStartManual.SubCommand(0x01))
	return err
}

func (dd *DimmableDevice) StartDim() error {
	_, err := SendStandardCommand(dd.Connection(), CmdLightStartManual.SubCommand(0x00))
	return err
}

func (dd *DimmableDevice) StopChange() error {
	_, err := SendStandardCommand(dd.Connection(), CmdLightStopManual)
	return err
}

func (dd *DimmableDevice) InstantChange(level int) error {
	_, err := SendStandardCommand(dd.Connection(), CmdLightInstantChange.SubCommand(level))
	return err
}

func (dd *DimmableDevice) SetStatus(level int) error {
	_, err := SendStandardCommand(dd.Connection(), CmdLightSetStatus.SubCommand(level))
	return err
}

func (dd *DimmableDevice) OnAtRamp(level, ramp int) error {
	levelRamp := byte(level) << 4
	levelRamp |= byte(ramp) & 0x0f
	_, err := SendStandardCommand(dd.Connection(), CmdLightOnAtRamp.SubCommand(int(levelRamp)))
	return err
}

func (dd *DimmableDevice) OffAtRamp(ramp int) error {
	_, err := SendStandardCommand(dd.Connection(), CmdLightOffAtRamp.SubCommand(0x0f&ramp))
	return err
}

func (dd *DimmableDevice) String() string {
	return sprintf("Dimmable Light (%s)", dd.Address())
}

func dimmableLightingFactory(device Device) Device {
	Log.Debugf("Returning dimmable device with underlying device %T", device)
	return &DimmableDevice{&SwitchedDevice{device}}
}

func switchedLightingFactory(device Device) Device {
	return &SwitchedDevice{device}
}
