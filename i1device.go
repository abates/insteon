package insteon

import (
	"strings"
)

type I1Device struct {
	Connection
	address Address
	ldb     *LinearLinkDB
}

func NewI1Device(address Address, bridge Bridge) *I1Device {
	return &I1Device{
		Connection: NewI1Connection(address, bridge),
		address:    address,
	}
}

func (i1 *I1Device) Address() Address {
	return i1.address
}

func (i1 *I1Device) AssignToAllLinkGroup(group Group) error {
	_, err := SendStandardCommand(i1.Connection, CmdAssignToAllLinkGroup.SubCommand(int(group)))
	return err
}

func (i1 *I1Device) DeleteFromAllLinkGroup(group Group) error {
	_, err := SendStandardCommand(i1.Connection, CmdDeleteFromAllLinkGroup.SubCommand(int(group)))
	return err
}

func (i1 *I1Device) ProductData() (*ProductData, error) {
	var data *ProductData
	msg, err := SendStandardCommandAndWait(i1.Connection, CmdProductDataReq)
	if err == nil {
		data = msg.Payload.(*ProductData)
	}
	return data, err
}

func (i1 *I1Device) FXUsername() (string, error) {
	username := ""
	msg, err := SendStandardCommandAndWait(i1.Connection, CmdFxUsernameReq)
	if err == nil {
		buf := msg.Payload.(*BufPayload)
		username = string(buf.Buf)
	}
	return username, err
}

func (i1 *I1Device) DeviceTextString() (string, error) {
	text := ""
	msg, err := SendStandardCommandAndWait(i1.Connection, CmdDeviceTextStringReq)
	if err == nil {
		buf := msg.Payload.(*BufPayload)
		text = string(buf.Buf)
	}
	return strings.TrimSpace(text), err
}

func (i1 *I1Device) EnterLinkingMode(group Group) error {
	_, err := SendStandardCommand(i1.Connection, CmdEnterLinkingMode.SubCommand(int(group)))
	return err
}

func (i1 *I1Device) EnterUnlinkingMode(group Group) error {
	_, err := SendStandardCommand(i1.Connection, CmdEnterUnlinkingMode.SubCommand(int(group)))
	return err
}

func (i1 *I1Device) EngineVersion() (EngineVersion, error) {
	ack, err := SendStandardCommand(i1.Connection, CmdGetEngineVersion)
	version := EngineVersion(0)
	if err == nil {
		version = EngineVersion(ack.Command.cmd[1])
	}
	return version, err
}

func (i1 *I1Device) Ping() error {
	// TODO implement
	return ErrNotImplemented
}

func (i1 *I1Device) LinkDB() (ldb LinkDB, err error) {
	if i1.ldb == nil {
		i1.ldb = &LinearLinkDB{conn: i1.Connection}
		err = i1.ldb.Refresh()
	}
	return i1.ldb, err
}
