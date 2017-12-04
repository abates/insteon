package insteon

import (
	"strings"
)

type DeviceInitializer func(Connection, Address, *ProductData) Device

var Devices DeviceRegistry

type DeviceRegistry struct {
	// devices key is the first byte of the
	// Category.  Documentation simply calls this
	// the category and the second byte the sub
	// category, but we've combined both bytes
	// into a single type
	devices map[byte]DeviceInitializer
}

func (dr *DeviceRegistry) Register(category byte, initializer DeviceInitializer) {
	if dr.devices == nil {
		dr.devices = make(map[byte]DeviceInitializer)
	}
	dr.devices[category] = initializer
}

func (dr *DeviceRegistry) Find(category Category) DeviceInitializer {
	if initializer, found := dr.devices[category[0]]; found {
		return initializer
	}
	// always return a factory
	return StandardDeviceInitializer
}

func StandardDeviceInitializer(conn Connection, address Address, pd *ProductData) Device {
	return NewStandardDevice(conn, address)
}

type Device interface {
	Linkable
	ProductData() (*ProductData, error)
	FXUsername() (string, error)
	DeviceTextString() (string, error)
	InsteonEngineVersion() (int, error)
	Ping() error
}

func DeviceFactory(conn Connection, address Address) (Device, error) {
	device := Device(NewStandardDevice(conn, address))
	// query the device
	pd, err := device.ProductData()

	// construct device for device type
	if err == nil {
		device = Devices.Find(pd.Category)(conn, address, pd)
	} else if err == ErrReadTimeout {
		Log.Infof("Timed out waiting for product data response. Returning standard device")
		err = nil
	}

	return device, err
}

type StandardDevice struct {
	Connection
	address Address
	ldb     *LinearLinkDB
}

func NewStandardDevice(conn Connection, address Address) *StandardDevice {
	return &StandardDevice{
		Connection: conn,
		address:    address,
	}
}

func (sd *StandardDevice) Address() Address {
	return sd.address
}

func (sd *StandardDevice) AssignToAllLinkGroup(group Group) error {
	return sd.SendStandardCommand(CmdAssignToAllLinkGroup.SubCommand(int(group)))
}

func (sd *StandardDevice) DeleteFromAllLinkGroup(group Group) error {
	return sd.SendStandardCommand(CmdDeleteFromAllLinkGroup.SubCommand(int(group)))
}

func (sd *StandardDevice) ProductData() (*ProductData, error) {
	var data *ProductData
	msg, err := sd.SendStandardCommandAndWait(CmdProductDataReq)
	if err == nil {
		data = msg.Payload.(*ProductData)
	}
	return data, err
}

func (sd *StandardDevice) FXUsername() (string, error) {
	username := ""
	msg, err := sd.SendStandardCommandAndWait(CmdFxUsernameReq)
	if err == nil {
		buf := msg.Payload.(*BufPayload)
		username = string(buf.Buf)
	}
	return username, err
}

func (sd *StandardDevice) DeviceTextString() (string, error) {
	text := ""
	msg, err := sd.SendStandardCommandAndWait(CmdDeviceTextStringReq)
	if err == nil {
		buf := msg.Payload.(*BufPayload)
		text = string(buf.Buf)
	}
	return strings.TrimSpace(text), err
}

func (sd *StandardDevice) EnterLinkingMode(group Group) error {
	return sd.SendStandardCommand(CmdEnterLinkingMode.SubCommand(int(group)))
}

func (sd *StandardDevice) EnterUnlinkingMode(group Group) error {
	return sd.SendStandardCommand(CmdEnterUnlinkingMode.SubCommand(int(group)))
}

func (sd *StandardDevice) InsteonEngineVersion() (int, error) {
	// TODO implement
	return 0, ErrNotImplemented
}

func (sd *StandardDevice) Ping() error {
	// TODO implement
	return ErrNotImplemented
}

func (sd *StandardDevice) LinkDB() (ldb LinkDB, err error) {
	if sd.ldb == nil {
		sd.ldb = &LinearLinkDB{conn: sd.Connection}
		err = sd.ldb.Refresh()
	}
	return sd.ldb, err
}
