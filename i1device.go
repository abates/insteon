package insteon

import (
	"fmt"
	"strings"
)

type I1Device struct {
	Connection
	address Address
}

func NewI1Device(address Address, connection Connection) *I1Device {
	return &I1Device{
		Connection: connection,
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
	msg, err := SendStandardCommandAndWait(i1.Connection, CmdProductDataReq, CmdProductDataResp)
	if err == nil {
		data = msg.Payload.(*ProductData)
	}
	return data, err
}

func (i1 *I1Device) FXUsername() (string, error) {
	username := ""
	msg, err := SendStandardCommandAndWait(i1.Connection, CmdFxUsernameReq, CmdFxUsernameResp)
	if err == nil {
		buf := msg.Payload.(*BufPayload)
		username = string(buf.Buf)
	}
	return username, err
}

func (i1 *I1Device) TextString() (string, error) {
	text := ""
	msg, err := SendStandardCommandAndWait(i1.Connection, CmdDeviceTextStringReq, CmdDeviceTextStringResp)
	if err == nil {
		buf := msg.Payload.(*BufPayload)
		text = string(buf.Buf)
	}
	return strings.TrimSpace(text), err
}

func (*I1Device) EnterLinkingMode(Group) error   { return ErrNotImplemented }
func (*I1Device) EnterUnlinkingMode(Group) error { return ErrNotImplemented }
func (*I1Device) ExitLinkingMode() error         { return ErrNotImplemented }

func (i1 *I1Device) EngineVersion() (EngineVersion, error) {
	ack, err := SendStandardCommand(i1.Connection, CmdGetEngineVersion)
	version := EngineVersion(0)
	if err == nil {
		version = EngineVersion(ack.Command.Cmd[1])
	}
	return version, err
}

func (i1 *I1Device) Ping() error {
	_, err := SendStandardCommand(i1.Connection, CmdPing)
	return err
}

func (i1 *I1Device) IDRequest() error {
	_, err := SendStandardCommand(i1.Connection, CmdIDReq)
	return err
}

func (i1 *I1Device) SetTextString(str string) error {
	textString := NewBufPayload(14)
	copy(textString.Buf, []byte(str))
	_, err := SendExtendedCommand(i1.Connection, CmdSetDeviceTextString, textString)
	return err
}

func (i1 *I1Device) SetAllLinkCommandAlias(match, replace *Command) error {
	// TODO implement
	return ErrNotImplemented
}

func (i1 *I1Device) SetAllLinkCommandAliasData(data []byte) error {
	// TODO implement
	return ErrNotImplemented
}

func (i1 *I1Device) BlockDataTransfer() ([]byte, error) {
	// TODO implement
	return nil, ErrNotImplemented
}

func (*I1Device) LinkDB() (LinkDB, error) { return nil, ErrNotImplemented }

func (i1 *I1Device) String() string {
	return fmt.Sprintf("I1 Device (%s)", i1.Address())
}

func (i1 *I1Device) Close() error {
	Log.Debugf("Closing I1Device connection")
	return i1.Connection.Close()
}
