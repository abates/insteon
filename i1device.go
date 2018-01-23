package insteon

import (
	"fmt"
	"strings"
)

// I1Device provides remote communication to version 1 engines
type I1Device struct {
	conn     Connection
	address  Address
	category Category
	version  EngineVersion
}

// NewI1Device will construct an I1Device for the given address and connection
func NewI1Device(address Address, connection Connection) *I1Device {
	return &I1Device{
		conn:     connection,
		address:  address,
		category: Category{0xff, 0xff},
		version:  EngineVersion(0xff),
	}
}

// Address is the Insteon address of the device
func (i1 *I1Device) Address() Address {
	return i1.address
}

// AssignToAllLinkGroup will inform the device what group should be used during an All-Linking
// session
func (i1 *I1Device) AssignToAllLinkGroup(group Group) error {
	_, err := SendStandardCommand(i1.conn, CmdAssignToAllLinkGroup.SubCommand(int(group)))
	return err
}

// DeleteFromAllLinkGroup will inform the device which group should be unlinked during an
// All-Link unlinking session
func (i1 *I1Device) DeleteFromAllLinkGroup(group Group) error {
	_, err := SendStandardCommand(i1.conn, CmdDeleteFromAllLinkGroup.SubCommand(int(group)))
	return err
}

// ProductData will retrieve the device's product data
func (i1 *I1Device) ProductData() (*ProductData, error) {
	var data *ProductData
	msg, err := SendStandardCommandAndWait(i1.conn, CmdProductDataReq, CmdProductDataResp)
	if err == nil {
		data = msg.Payload.(*ProductData)
	}
	return data, err
}

// FXUsername will retrieve the device's FX username string
func (i1 *I1Device) FXUsername() (string, error) {
	username := ""
	msg, err := SendStandardCommandAndWait(i1.conn, CmdFxUsernameReq, CmdFxUsernameResp)
	if err == nil {
		buf := msg.Payload.(*BufPayload)
		username = string(buf.Buf)
	}
	return username, err
}

// TextString will retrieve the Device Text String
func (i1 *I1Device) TextString() (string, error) {
	text := ""
	msg, err := SendStandardCommandAndWait(i1.conn, CmdDeviceTextStringReq, CmdDeviceTextStringResp)
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
func (i1 *I1Device) EngineVersion() (version EngineVersion, err error) {
	if i1.version == 0xff {
		var ack *Message
		ack, err = SendStandardCommand(i1.conn, CmdGetEngineVersion)
		if err == nil {
			i1.version = EngineVersion(ack.Command.Cmd[1])
		}
	}
	return i1.version, err
}

// Ping will send a Ping command to the device
func (i1 *I1Device) Ping() error {
	_, err := SendStandardCommand(i1.conn, CmdPing)
	return err
}

// IDRequest will send an ID request to the device and return
// the device category
func (i1 *I1Device) IDRequest() (category Category, err error) {
	if i1.category == Category([2]byte{0xff, 0xff}) {
		var msg *Message
		msg, err = SendStandardCommandAndWait(i1.conn, CmdIDReq, CmdSetButtonPressedController, CmdSetButtonPressedResponder)
		if msg != nil {
			i1.category = Category([2]byte{msg.Dst[0], msg.Dst[1]})
		}
	}
	return i1.category, err
}

// SetTextString will set the device text string
func (i1 *I1Device) SetTextString(str string) error {
	textString := NewBufPayload(14)
	copy(textString.Buf, []byte(str))
	_, err := SendExtendedCommand(i1.conn, CmdSetDeviceTextString, textString)
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
	return i1.conn.Close()
}

func (i1 *I1Device) Connection() Connection {
	return i1.conn
}
