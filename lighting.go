package insteon

var (
	LightingCategories = []Category{Category(1), Category(2)}

	CmdLightOn             = Commands.RegisterStd("Light On", LightingCategories, MsgTypeDirect, 0x11, 0xff)
	CmdLightOnFast         = Commands.RegisterStd("Light On Fast", LightingCategories, MsgTypeDirect, 0x12, 0x00)
	CmdLightOff            = Commands.RegisterStd("Light Off", LightingCategories, MsgTypeDirect, 0x13, 0x00)
	CmdLightOffFast        = Commands.RegisterStd("Light Off Fast", LightingCategories, MsgTypeDirect, 0x14, 0x00)
	CmdLightBrighten       = Commands.RegisterStd("Light Brighten", LightingCategories, MsgTypeDirect, 0x15, 0x00)
	CmdLightDim            = Commands.RegisterStd("Light Dim", LightingCategories, MsgTypeDirect, 0x16, 0x00)
	CmdLightStartManual    = Commands.RegisterStd("Light Start Manual Change", LightingCategories, MsgTypeDirect, 0x17, 0x00)
	CmdLightStopManual     = Commands.RegisterStd("Light Stop Manual Change", LightingCategories, MsgTypeDirect, 0x18, 0x00)
	CmdLightStatusRequest  = Commands.RegisterStd("Light Status Req", LightingCategories, MsgTypeDirect, 0x19, 0x00)
	CmdLightInstantChange  = Commands.RegisterStd("Light Instant Change", LightingCategories, MsgTypeDirect, 0x21, 0x00)
	CmdLightManualOn       = Commands.RegisterStd("Tap Set Button Once", LightingCategories, MsgTypeDirect, 0x22, 0x01)
	CmdLightManualOff      = Commands.RegisterStd("Tap Set Button Once", LightingCategories, MsgTypeDirect, 0x23, 0x01)
	CmdTapSetButtonOnce    = Commands.RegisterStd("Tap Set Button Once", LightingCategories, MsgTypeDirect, 0x25, 0x01)
	CmdTapSetButtonTwice   = Commands.RegisterStd("Tap Set Button Twice", LightingCategories, MsgTypeDirect, 0x25, 0x02)
	CmdLightSetStatus      = Commands.RegisterStd("Update dimmer LEDs", LightingCategories, MsgTypeDirect, 0x27, 0x00)
	CmdLightOnAtRamp       = Commands.RegisterStd("Light On Ramp", LightingCategories, MsgTypeDirect, 0x2e, 0x00)
	CmdLightOffAtRamp      = Commands.RegisterStd("Light Off Ramp", LightingCategories, MsgTypeDirect, 0x2f, 0x00)
	CmdLightExtendedSetGet = Commands.RegisterExt("Extended Set/Get", LightingCategories, MsgTypeDirect, 0x2e, 0x00)
)

func init() {
	Commands.RegisterVersionStd(CmdLightOnAtRamp, 0x43, 0x34, 0x00)
	Commands.RegisterVersionStd(CmdLightOffAtRamp, 0x43, 0x35, 0x00)

	Devices.Register(0x01, dimmableLightingFactory)
	Devices.Register(0x02, switchedLightingFactory)
}

type SwitchConfig struct {
	HouseCode int
	UnitCode  int
}

func (sc *SwitchConfig) UnmarshalBinary(buf []byte) error {
	if len(buf) < 12 {
		return ErrBufferTooShort
	}
	sc.HouseCode = int(buf[4])
	sc.UnitCode = int(buf[5])
	return nil
}

func (sc *SwitchConfig) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 14)
	buf[4] = byte(sc.HouseCode)
	buf[5] = byte(sc.UnitCode)
	return buf, nil
}

type Switch interface {
	On() error
	ManualOn() error
	Off() error
	ManualOff() error
	Status() (level int, err error)
	OperatingFlags() (*LightFlags, error)
	SetOperatingFlags(*LightFlags) error
	SetX10Address(button int, houseCode, unitCode byte) error
	SwitchConfig() (SwitchConfig, error)
}

type DimmerConfig struct {
	HouseCode int
	UnitCode  int
	Ramp      int
	OnLevel   int
	SNR       int
}

func (dc *DimmerConfig) UnmarshalBinary(buf []byte) error {
	if len(buf) < 12 {
		return ErrBufferTooShort
	}
	dc.HouseCode = int(buf[4])
	dc.UnitCode = int(buf[5])
	dc.Ramp = int(buf[6])
	dc.OnLevel = int(buf[7])
	dc.SNR = int(buf[8])
	return nil
}

func (dc *DimmerConfig) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 14)
	buf[4] = byte(dc.HouseCode)
	buf[5] = byte(dc.UnitCode)
	buf[6] = byte(dc.Ramp)
	buf[7] = byte(dc.OnLevel)
	buf[8] = byte(dc.SNR)
	return buf, nil
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
	SetDefaultRamp(level int) error
	SetDefaultOnLevel(level int) error
	DimmerConfig() (DimmerConfig, error)
}

type SwitchedDevice struct {
	Device
}

func (sd *SwitchedDevice) On() error {
	_, err := SendCommand(sd, CmdLightOn)
	return err
}

func (sd *SwitchedDevice) Off() error {
	_, err := SendCommand(sd, CmdLightOff)
	return err
}

// Status sends a LightStatusRequest to determine the device's current
// level. For switched devices this is either 0 or 255, dimmable devices
// will be the current dim level between 0 and 255
func (sd *SwitchedDevice) Status() (level int, err error) {
	ack, err := SendCommand(sd, CmdLightStatusRequest)
	if err == nil {
		level = int(ack.Command.Command1)
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
	ack, err := SendCommand(sd, CmdGetOperatingFlags)
	if err == nil {
		flags.UnmarshalBinary([]byte{ack.Command.Command1})
	}
	return flags, err
}

func (sd *SwitchedDevice) SetOperatingFlags(flags *LightFlags) error {
	buf, err := flags.MarshalBinary()
	if err == nil {
		_, err = SendSubCommand(sd, CmdSetOperatingFlags, int(buf[0]))
	}
	return err
}

func (sd *SwitchedDevice) ManualOn() error {
	_, err := SendCommand(sd, CmdLightManualOn)
	return err
}

func (sd *SwitchedDevice) ManualOff() error {
	_, err := SendCommand(sd, CmdLightManualOff)
	return err
}

func (sd *SwitchedDevice) String() string {
	return sprintf("Switch (%s)", sd.Address())
}

func (sd *SwitchedDevice) SetX10Address(button int, houseCode, unitCode byte) error {
	_, err := SendExtendedCommand(sd, CmdLightExtendedSetGet, []byte{byte(button), 0x04, houseCode, unitCode})
	return err
}

func (sd *SwitchedDevice) SwitchConfig() (config SwitchConfig, err error) {
	// SEE DimmerConfig() notes for explanation of D1 and D2 (payload[0] and payload[1])
	msg, err := SendExtendedCommandAndWait(sd, CmdLightExtendedSetGet, []byte{0x00, 0x00}, CmdLightExtendedSetGet)
	if err == nil {
		err = config.UnmarshalBinary(msg.Payload)
	}
	return config, err
}

type DimmableDevice struct {
	*SwitchedDevice
}

func (dd *DimmableDevice) OnLevel(level int) error {
	_, err := SendSubCommand(dd, CmdLightOn, level)
	return err
}

func (dd *DimmableDevice) OnFast(level int) error {
	_, err := SendCommand(dd, CmdLightOnFast)
	return err
}

func (dd *DimmableDevice) Brighten() error {
	_, err := SendCommand(dd, CmdLightBrighten)
	return err
}

func (dd *DimmableDevice) Dim() error {
	_, err := SendCommand(dd, CmdLightDim)
	return err
}

func (dd *DimmableDevice) StartBrighten() error {
	_, err := SendSubCommand(dd, CmdLightStartManual, 1)
	return err
}

func (dd *DimmableDevice) StartDim() error {
	_, err := SendSubCommand(dd, CmdLightStartManual, 0)
	return err
}

func (dd *DimmableDevice) StopChange() error {
	_, err := SendCommand(dd, CmdLightStopManual)
	return err
}

func (dd *DimmableDevice) InstantChange(level int) error {
	_, err := SendSubCommand(dd, CmdLightInstantChange, level)
	return err
}

func (dd *DimmableDevice) SetStatus(level int) error {
	_, err := SendSubCommand(dd, CmdLightSetStatus, level)
	return err
}

func (dd *DimmableDevice) OnAtRamp(level, ramp int) error {
	levelRamp := byte(level) << 4
	levelRamp |= byte(ramp) & 0x0f
	_, err := SendSubCommand(dd, CmdLightOnAtRamp, int(levelRamp))
	return err
}

func (dd *DimmableDevice) OffAtRamp(ramp int) error {
	_, err := SendSubCommand(dd, CmdLightOffAtRamp, 0x0f&ramp)
	return err
}

func (dd *DimmableDevice) String() string {
	return sprintf("Dimmable Light (%s)", dd.Address())
}

func (dd *DimmableDevice) SetDefaultRamp(rate int) error {
	// See notes on DimmerConfig() for information about D1 (payload[0])
	_, err := SendExtendedCommand(dd, CmdLightExtendedSetGet, []byte{0x01, 0x05, byte(rate)})
	return err
}

func (dd *DimmableDevice) SetDefaultOnLevel(level int) error {
	// See notes on DimmerConfig() for information about D1 (payload[0])
	_, err := SendExtendedCommand(dd, CmdLightExtendedSetGet, []byte{0x01, 0x06, byte(level)})
	return err
}

func (dd *DimmableDevice) DimmerConfig() (config DimmerConfig, err error) {
	// The documentation talks about D1 (payload[0]) being the button/group number, but my
	// SwitchLinc dimmers all return the same information regardless of
	// the value of D1.  I think D1 is maybe only relevant on KeyPadLinc dimmers.
	//
	// D2 is 0x00 for requests
	msg, err := SendExtendedCommandAndWait(dd, CmdLightExtendedSetGet, []byte{0x01, 0x00}, CmdLightExtendedSetGet)
	if err == nil {
		err = config.UnmarshalBinary(msg.Payload)
	}
	return config, err
}

func dimmableLightingFactory(device Device) Device {
	return &DimmableDevice{&SwitchedDevice{device}}
}

func switchedLightingFactory(device Device) Device {
	return &SwitchedDevice{device}
}
