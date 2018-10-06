package insteon

var (
	LightingCategories = []Category{Category(1), Category(2)}
)

const (
	CmdLightOn             Command = 0x11ff
	CmdLightOnFast         Command = 0x1200
	CmdLightOff            Command = 0x1300
	CmdLightOffFast        Command = 0x1400
	CmdLightBrighten       Command = 0x1500
	CmdLightDim            Command = 0x1600
	CmdLightStartManual    Command = 0x1700
	CmdLightStopManual     Command = 0x1800
	CmdLightStatusRequest  Command = 0x1900
	CmdLightInstantChange  Command = 0x2100
	CmdLightManualOn       Command = 0x2201
	CmdLightManualOff      Command = 0x2301
	CmdTapSetButtonOnce    Command = 0x2501
	CmdTapSetButtonTwice   Command = 0x2502
	CmdLightSetStatus      Command = 0x2700
	CmdLightOnAtRamp       Command = 0x2e00
	CmdLightOnAtRampV67    Command = 0x3400
	CmdLightOffAtRamp      Command = 0x2f00
	CmdLightOffAtRampV67   Command = 0x3500
	CmdLightExtendedSetGet Command = 0x2e00
)

func init() {
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
	firmwareVersion FirmwareVersion
}

func (sd *SwitchedDevice) On() error {
	_, err := sd.SendCommand(CmdLightOn, nil)
	return err
}

func (sd *SwitchedDevice) Off() error {
	_, err := sd.SendCommand(CmdLightOff, nil)
	return err
}

// Status sends a LightStatusRequest to determine the device's current
// level. For switched devices this is either 0 or 255, dimmable devices
// will be the current dim level between 0 and 255
func (sd *SwitchedDevice) Status() (level int, err error) {
	response, err := sd.SendCommand(CmdLightStatusRequest, nil)
	if err == nil {
		level = int(response >> 8)
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
	response, err := sd.SendCommand(CmdGetOperatingFlags, nil)
	if err == nil {
		flags.UnmarshalBinary([]byte{byte(response >> 8)})
	}
	return flags, err
}

func (sd *SwitchedDevice) SetOperatingFlags(flags *LightFlags) error {
	buf, err := flags.MarshalBinary()
	if err == nil {
		_, err = sd.SendCommand(CmdSetOperatingFlags.SubCommand(int(buf[0])), nil)
	}
	return err
}

func (sd *SwitchedDevice) ManualOn() error {
	_, err := sd.SendCommand(CmdLightManualOn, nil)
	return err
}

func (sd *SwitchedDevice) ManualOff() error {
	_, err := sd.SendCommand(CmdLightManualOff, nil)
	return err
}

func (sd *SwitchedDevice) String() string {
	return sprintf("Switch (%s)", sd.Address())
}

func (sd *SwitchedDevice) SetX10Address(button int, houseCode, unitCode byte) error {
	_, err := sd.SendCommand(CmdLightExtendedSetGet, []byte{byte(button), 0x04, houseCode, unitCode})
	return err
}

func (sd *SwitchedDevice) SwitchConfig() (config SwitchConfig, err error) {
	// SEE DimmerConfig() notes for explanation of D1 and D2 (payload[0] and payload[1])
	recvCh, err := sd.SendCommandAndListen(CmdLightExtendedSetGet, []byte{0x00, 0x00})
	for response := range recvCh {
		if response.Message.Command == CmdLightExtendedSetGet {
			err = config.UnmarshalBinary(response.Message.Payload)
			response.DoneCh <- true
		}
		response.DoneCh <- false
	}
	return config, err
}

type DimmableDevice struct {
	*SwitchedDevice
	firmwareVersion FirmwareVersion
}

func (dd *DimmableDevice) OnLevel(level int) error {
	_, err := dd.SendCommand(CmdLightOn.SubCommand(level), nil)
	return err
}

func (dd *DimmableDevice) OnFast(level int) error {
	_, err := dd.SendCommand(CmdLightOnFast, nil)
	return err
}

func (dd *DimmableDevice) Brighten() error {
	_, err := dd.SendCommand(CmdLightBrighten, nil)
	return err
}

func (dd *DimmableDevice) Dim() error {
	_, err := dd.SendCommand(CmdLightDim, nil)
	return err
}

func (dd *DimmableDevice) StartBrighten() error {
	_, err := dd.SendCommand(CmdLightStartManual.SubCommand(1), nil)
	return err
}

func (dd *DimmableDevice) StartDim() error {
	_, err := dd.SendCommand(CmdLightStartManual.SubCommand(0), nil)
	return err
}

func (dd *DimmableDevice) StopChange() error {
	_, err := dd.SendCommand(CmdLightStopManual, nil)
	return err
}

func (dd *DimmableDevice) InstantChange(level int) error {
	_, err := dd.SendCommand(CmdLightInstantChange.SubCommand(level), nil)
	return err
}

func (dd *DimmableDevice) SetStatus(level int) error {
	_, err := dd.SendCommand(CmdLightSetStatus.SubCommand(level), nil)
	return err
}

func (dd *DimmableDevice) OnAtRamp(level, ramp int) (err error) {
	levelRamp := byte(level) << 4
	levelRamp |= byte(ramp) & 0x0f
	if dd.firmwareVersion >= 0x43 {
		_, err = dd.SendCommand(CmdLightOnAtRampV67.SubCommand(int(levelRamp)), nil)
	} else {
		_, err = dd.SendCommand(CmdLightOnAtRamp.SubCommand(int(levelRamp)), nil)
	}
	return err
}

func (dd *DimmableDevice) OffAtRamp(ramp int) (err error) {
	if dd.firmwareVersion >= 0x43 {
		_, err = dd.SendCommand(CmdLightOffAtRampV67.SubCommand(0x0f&ramp), nil)
	} else {
		_, err = dd.SendCommand(CmdLightOffAtRamp.SubCommand(0x0f&ramp), nil)
	}
	return err
}

func (dd *DimmableDevice) String() string {
	return sprintf("Dimmable Light (%s)", dd.Address())
}

func (dd *DimmableDevice) SetDefaultRamp(rate int) error {
	// See notes on DimmerConfig() for information about D1 (payload[0])
	_, err := dd.SendCommand(CmdLightExtendedSetGet, []byte{0x01, 0x05, byte(rate)})
	return err
}

func (dd *DimmableDevice) SetDefaultOnLevel(level int) error {
	// See notes on DimmerConfig() for information about D1 (payload[0])
	_, err := dd.SendCommand(CmdLightExtendedSetGet, []byte{0x01, 0x06, byte(level)})
	return err
}

func (dd *DimmableDevice) DimmerConfig() (config DimmerConfig, err error) {
	// The documentation talks about D1 (payload[0]) being the button/group number, but my
	// SwitchLinc dimmers all return the same information regardless of
	// the value of D1.  I think D1 is maybe only relevant on KeyPadLinc dimmers.
	//
	// D2 is 0x00 for requests
	recvCh, err := dd.SendCommandAndListen(CmdLightExtendedSetGet, []byte{0x01, 0x00})
	for response := range recvCh {
		if response.Message.Command == CmdLightExtendedSetGet {
			err = config.UnmarshalBinary(response.Message.Payload)
			response.DoneCh <- true
		}
		response.DoneCh <- false
	}
	return config, err
}

func dimmableLightingFactory(device Device, info DeviceInfo) Device {
	return &DimmableDevice{&SwitchedDevice{device, info.FirmwareVersion}, info.FirmwareVersion}
}

func switchedLightingFactory(device Device, info DeviceInfo) Device {
	return &SwitchedDevice{device, info.FirmwareVersion}
}
