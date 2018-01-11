package insteon

import (
	"fmt"
	"strings"
)

// I1Device provides remote communication to version 1 engines
type I1Device struct {
	Connection
	address Address
}

// NewI1Device will construct an I1Device for the given address and connection
func NewI1Device(address Address, connection Connection) *I1Device {
	return &I1Device{
		Connection: connection,
		address:    address,
	}
}

// Address is the Insteon address of the device
func (i1 *I1Device) Address() Address {
	return i1.address
}

// AssignToAllLinkGroup will inform the device what group should be used during an All-Linking
// session
func (i1 *I1Device) AssignToAllLinkGroup(group Group) error {
	_, err := SendStandardCommand(i1.Connection, CmdAssignToAllLinkGroup.SubCommand(int(group)))
	return err
}

// DeleteFromAllLinkGroup will inform the device which group should be unlinked during an
// All-Link unlinking session
func (i1 *I1Device) DeleteFromAllLinkGroup(group Group) error {
	_, err := SendStandardCommand(i1.Connection, CmdDeleteFromAllLinkGroup.SubCommand(int(group)))
	return err
}

// ProductData will retrieve the device's product data
func (i1 *I1Device) ProductData() (*ProductData, error) {
	var data *ProductData
	msg, err := SendStandardCommandAndWait(i1.Connection, CmdProductDataReq, CmdProductDataResp)
	if err == nil {
		data = msg.Payload.(*ProductData)
	}
	return data, err
}

// FXUsername will retrieve the device's FX username string
func (i1 *I1Device) FXUsername() (string, error) {
	username := ""
	msg, err := SendStandardCommandAndWait(i1.Connection, CmdFxUsernameReq, CmdFxUsernameResp)
	if err == nil {
		buf := msg.Payload.(*BufPayload)
		username = string(buf.Buf)
	}
	return username, err
}

// TextString will retrieve the Device Text String
func (i1 *I1Device) TextString() (string, error) {
	text := ""
	msg, err := SendStandardCommandAndWait(i1.Connection, CmdDeviceTextStringReq, CmdDeviceTextStringResp)
	if err == nil {
		buf := msg.Payload.(*BufPayload)
		text = string(buf.Buf)
	}
	return strings.TrimSpace(text), err
}

// EnterLinkingMode is unsupported on I1Devices
func (*I1Device) EnterLinkingMode(Group) error { return ErrNotImplemented }

// EnterUnlinkingMode is unsupported on I1Devices
func (*I1Device) EnterUnlinkingMode(Group) error { return ErrNotImplemented }

// ExitLinkingMode is unsupported on I1Devices
func (*I1Device) ExitLinkingMode() error { return ErrNotImplemented }

// EngineVersion will retrieve the device's EngineVersion
func (i1 *I1Device) EngineVersion() (EngineVersion, error) {
	ack, err := SendStandardCommand(i1.Connection, CmdGetEngineVersion)
	version := EngineVersion(0)
	if err == nil {
		version = EngineVersion(ack.Command.Cmd[1])
	}
	return version, err
}

// Ping will send a Ping command to the device
func (i1 *I1Device) Ping() error {
	_, err := SendStandardCommand(i1.Connection, CmdPing)
	return err
}

// IDRequest will send an ID request to the device
func (i1 *I1Device) IDRequest() error {
	_, err := SendStandardCommand(i1.Connection, CmdIDReq)
	return err
}

// SetTextString will set the device text string
func (i1 *I1Device) SetTextString(str string) error {
	textString := NewBufPayload(14)
	copy(textString.Buf, []byte(str))
	_, err := SendExtendedCommand(i1.Connection, CmdSetDeviceTextString, textString)
	return err
}

// SetAllLinkCommandAlias will set the device's standard command to be used
// when the given alias command is sent
func (i1 *I1Device) SetAllLinkCommandAlias(match, replace *Command) error {
	// TODO implement
	return ErrNotImplemented
}

// SetAllLinkCommandAliasData will set any extended data required by the alias
// command
func (i1 *I1Device) SetAllLinkCommandAliasData(data []byte) error {
	// TODO implement
	return ErrNotImplemented
}

// BlockDataTransfer will retrieve a block of memory from the device
func (i1 *I1Device) BlockDataTransfer(start, end MemAddress, length int) ([]byte, error) {
	// TODO implement
	return nil, ErrNotImplemented
}

// LinkDB is unsupported on I1Devices
func (*I1Device) LinkDB() (LinkDB, error) { return nil, ErrNotImplemented }

// String will return a string containing the device address
func (i1 *I1Device) String() string {
	return fmt.Sprintf("I1 Device (%s)", i1.Address())
}

// Close closes the underlying connection
func (i1 *I1Device) Close() error {
	Log.Debugf("Closing I1Device connection")
	return i1.Connection.Close()
}
