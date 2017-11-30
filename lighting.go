package insteon

var (
	CmdLightOn              = Commands.RegisterStd("Light On", 0x11, 0x00)
	CmdLightOnFast          = Commands.RegisterStd("Light On Fast", 0x12, 0x00)
	CmdLightOff             = Commands.RegisterStd("Light Off", 0x13, 0x00)
	CmdLightOffFast         = Commands.RegisterStd("Light Off Fast", 0x14, 0x00)
	CmdLightBrighten        = Commands.RegisterStd("Light Brighten", 0x15, 0x00)
	CmdLightDim             = Commands.RegisterStd("Light Dim", 0x16, 0x00)
	CmdLightStartManualDown = Commands.RegisterStd("Light Manual Down", 0x17, 0x00)
	CmdLightStartManualUp   = Commands.RegisterStd("Light Manual Up", 0x17, 0x01)
	CmdLightStopManual      = Commands.RegisterStd("Light Stop Manual", 0x18, 0x00)
	CmdLightStatusRequest   = Commands.RegisterStd("Light Status Req", 0x19, 0x00)
	CmdGetOperatingFlags    = Commands.RegisterStd("Get Operating Flags", 0x1f, 0x00)
	CmdSetOperatingFlags    = Commands.RegisterStd("Set Operating Flags", 0x20, 0x00)
	CmdLightInstantChange   = Commands.RegisterStd("Light Instant Change", 0x21, 0x00)
	CmdTapSetButtonOnce     = Commands.RegisterStd("Tap Set Button Once", 0x25, 0x01)
	CmdTapSetButtonTwice    = Commands.RegisterStd("Tap Set Button Twice", 0x25, 0x02)
	CmdLightOnRamp          = Commands.RegisterStd("Light On Ramp", 0x2e, 0x00)
	CmdLightOffRamp         = Commands.RegisterStd("Light Off Ramp", 0x2f, 0x00)
)

type DimmableDevice struct {
	*StandardDevice
}

func dimmableLightingFactory(conn Connection, address Address, pd *ProductData) Device {
	return &DimmableDevice{NewStandardDevice(conn, address)}
}

type SwitchedDevice struct {
	*StandardDevice
}

func switchedLightingFactory(conn Connection, address Address, pd *ProductData) Device {
	subCategory := pd.Category.SubCategory()
	if 0x06 <= subCategory && subCategory <= 0x15 {
		return &SwitchedDevice{NewStandardDevice(conn, address)}
	}

	//TODO log something here about an unknown dimmable device sub-category
	return defaultFactory(conn, address, pd)
}
