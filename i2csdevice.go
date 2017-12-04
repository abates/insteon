package insteon

type I2CsDevice struct {
	*I1Device
}

func NewI2CsDevice(address Address, bridge Bridge) *I2CsDevice {
	return &I2CsDevice{
		I1Device: &I1Device{
			Connection: NewI2Connection(address, bridge),
			address:    address,
		},
	}
}

func (i2cs *I2CsDevice) EnterLinkingMode(group Group) error {
	_, err := SendExtendedCommand(i2cs.I1Device.Connection, CmdEnterLinkingModeExtended.SubCommand(int(group)), NewBufPayload(14))
	return err
}

func (i2cs *I2CsDevice) EnterUnlinkingMode(group Group) error {
	_, err := SendStandardCommand(i2cs.I1Device.Connection, CmdEnterUnlinkingMode.SubCommand(int(group)))
	return err
}
