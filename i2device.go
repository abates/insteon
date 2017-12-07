package insteon

import "fmt"

type I2Device struct {
	*I1Device
}

func NewI2Device(address Address, bridge Bridge) *I2Device {
	return &I2Device{NewI1Device(address, bridge)}
}

func (i2 *I2Device) LinkDB() (ldb LinkDB, err error) {
	if i2.ldb == nil {
		i2.ldb = NewI2LinkDB(i2.Connection)
		err = i2.ldb.Refresh()
	}
	return i2.ldb, err
}

func (i2 *I2Device) EnterLinkingMode(group Group) error {
	_, err := SendStandardCommand(i2.Connection, CmdEnterLinkingMode.SubCommand(int(group)))
	return err
}

func (i2 *I2Device) EnterUnlinkingMode(group Group) error {
	_, err := SendStandardCommand(i2.Connection, CmdEnterUnlinkingMode.SubCommand(int(group)))
	return err
}

func (i2 *I2Device) ExitLinkingMode() error {
	_, err := SendStandardCommand(i2.I1Device.Connection, CmdExitLinkingMode)
	return err
}

func (i2 *I2Device) String() string {
	return fmt.Sprintf("I2 Device (%s)", i2.Address())
}
