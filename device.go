package insteon

import (
	"strings"
	"time"
)

type FactoryFunc func(Connection, Address, *ProductData) Device

var (
	deviceFactories = make(map[byte]FactoryFunc)
)

func defaultFactory(conn Connection, address Address, pd *ProductData) Device {
	return NewStandardDevice(conn, address)
}

func init() {
	registerFactory(0x00, defaultFactory)
	registerFactory(0x01, dimmableLightingFactory)
	registerFactory(0x02, switchedLightingFactory)
}

func registerFactory(category byte, factory FactoryFunc) {
	deviceFactories[category] = factory
}

func getFactory(pd *ProductData) FactoryFunc {
	if deviceFactory, ok := deviceFactories[pd.Category[0]]; ok {
		return deviceFactory
	}

	return defaultFactory
}

var (
	CmdAssignToAllLinkGroup    = Commands.RegisterStd("All Link Assign", 0x01, 0x00)
	CmdDeleteFromAllLinkGroup  = Commands.RegisterStd("All Link Delete", 0x02, 0x00)
	CmdProductDataReq          = Commands.RegisterStd("Product Data Req", 0x03, 0x00)
	CmdProductDataResp         = Commands.RegisterExt("Product Data Resp", 0x03, 0x00, func() Payload { return &ProductData{} })
	CmdFxUsernameReq           = Commands.RegisterStd("FX Username Req", 0x03, 0x01)
	CmdFxUsernameResp          = Commands.RegisterExt("FX Username Resp", 0x03, 0x01, nil)
	CmdDeviceTextStringReq     = Commands.RegisterStd("Text String Req", 0x03, 0x02)
	CmdDeviceTextStringResp    = Commands.RegisterExt("Text String Resp", 0x03, 0x02, nil)
	CmdEnterLinkingMode        = Commands.RegisterStd("Enter Link Mode", 0x09, 0x00)
	CmdEnterUnlinkingMode      = Commands.RegisterStd("Enter Unlink Mode", 0x0a, 0x00)
	CmdGetInsteonEngineVersion = Commands.RegisterStd("Get INSTEON Ver", 0x0d, 0x00)
	CmdPing                    = Commands.RegisterStd("Ping", 0x0f, 0x00)
	CmdIDReq                   = Commands.RegisterStd("ID Req", 0x10, 0x00)
	CmdReadWriteALDB           = Commands.RegisterExt("Read/Write ALDB", 0x2f, 0x00, func() Payload { return &LinkRequest{} })

	DeviceTimeout = time.Second * 10
)

type Device interface {
	Connection
	AssignToAllLinkGroup(Group) error
	DeleteFromAllLinkGroup(Group) error
	ProductData() (*ProductData, error)
	FXUsername() (string, error)
	DeviceTextString() (string, error)
	EnterLinkingMode(Group) error
	EnterUnlinkingMode(Group) error
	InsteonEngineVersion() (int, error)
	Ping() error
	ReadLink(MemAddress) (*Link, error)
	WriteLink(MemAddress, *Link) error
	Links() []*Link
}

func DeviceFactory(conn Connection, address Address) (Device, error) {
	var device Device
	sd := NewStandardDevice(conn, address)
	// query the device
	pd, err := sd.ProductData()

	// construct device for device type
	if err == nil {
		device = getFactory(pd)(conn, address, pd)
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

func (sd *StandardDevice) ReadLink(memAddress MemAddress) (*Link, error) {
	// TODO implement
	return nil, ErrNotImplemented
}

func (sd *StandardDevice) WriteLink(memAddress MemAddress, link *Link) error {
	request := &LinkRequest{Type: WriteLink, Link: link}
	return sd.SendExtendedCommand(CmdReadWriteALDB, request)
}

func (sd *StandardDevice) RefreshLinkDB() error {
	if sd.ldb == nil {
		sd.ldb = &LinearLinkDB{device: sd}
	}
	return sd.ldb.Refresh()
}

func (sd *StandardDevice) Links() []*Link {
	if sd.ldb == nil {
		sd.RefreshLinkDB()
	}
	return sd.ldb.Links()
}
