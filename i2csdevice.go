package insteon

import "fmt"

type I2CsDevice struct {
	*I2Device
}

func NewI2CsDevice(i2device *I2Device) *I2CsDevice {
	return &I2CsDevice{i2device}
}

func (i2cs *I2CsDevice) EnterLinkingMode(group Group) (err error) {
	_, err = SendExtendedCommand(i2cs.I1Device.Connection, CmdEnterLinkingModeExtended.SubCommand(int(group)), NewBufPayload(14))
	return err
}

func (i2cs *I2CsDevice) EnterUnlinkingMode(group Group) error {
	_, err := SendStandardCommand(i2cs.I1Device.Connection, CmdEnterUnlinkingMode.SubCommand(int(group)))
	return err
}

func (i2cs *I2CsDevice) String() string {
	return fmt.Sprintf("I2CS Device (%s)", i2cs.Address())
}
