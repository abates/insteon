package insteon

import "fmt"

type I2Device struct {
	*I1Device
	ldb *DeviceLinkDB
}

func NewI2Device(i1device *I1Device) *I2Device {
	return &I2Device{I1Device: i1device}
}

func (i2 *I2Device) LinkDB() (ldb LinkDB, err error) {
	if i2.ldb == nil {
		i2.ldb = NewDeviceLinkDB(i2)
	}
	return i2.ldb, err
}

func (i2 *I2Device) EnterLinkingMode(group Group) (err error) {
	_, err = SendStandardCommand(i2, CmdEnterLinkingMode.SubCommand(int(group)))
	return err
}

func (i2 *I2Device) EnterUnlinkingMode(group Group) error {
	_, err := SendStandardCommand(i2, CmdEnterUnlinkingMode.SubCommand(int(group)))
	return err
}

func (i2 *I2Device) ExitLinkingMode() error {
	_, err := SendStandardCommand(i2, CmdExitLinkingMode)
	return err
}

func (i2 *I2Device) String() string {
	return fmt.Sprintf("I2 Device (%s)", i2.Address())
}

func (i2 *I2Device) Close() error {
	Log.Debugf("Closing I2Device")
	return i2.I1Device.Close()
}
