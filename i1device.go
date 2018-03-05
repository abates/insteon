package insteon

import (
	"strings"
	"time"
)

// I1Device provides remote communication to version 1 engines
type I1Device struct {
	Connection
	address         Address
	devCat          DevCat
	engineVersion   EngineVersion
	firmwareVersion FirmwareVersion
}

// NewI1Device will construct an I1Device for the given address and connection
func NewI1Device(address Address, connection Connection) *I1Device {
	return &I1Device{
		Connection:      connection,
		address:         address,
		devCat:          DevCat{0xff, 0xff},
		engineVersion:   EngineVersion(0xff),
		firmwareVersion: FirmwareVersion(0x00),
	}
}

// Address is the Insteon address of the device
func (i1 *I1Device) Address() Address {
	return i1.address
}

// AssignToAllLinkGroup will inform the device what group should be used during an All-Linking
// session
func (i1 *I1Device) AssignToAllLinkGroup(group Group) error {
	_, err := SendSubCommand(i1, CmdAssignToAllLinkGroup, int(group))
	return err
}

// DeleteFromAllLinkGroup will inform the device which group should be unlinked during an
// All-Link unlinking session
func (i1 *I1Device) DeleteFromAllLinkGroup(group Group) error {
	_, err := SendSubCommand(i1, CmdDeleteFromAllLinkGroup, int(group))
	return err
}

// ProductData will retrieve the device's product data
func (i1 *I1Device) ProductData() (*ProductData, error) {
	var data *ProductData
	msg, err := SendCommandAndWait(i1, CmdProductDataReq, CmdProductDataResp)
	if err == nil {
		data = &ProductData{}
		err = data.UnmarshalBinary(msg.Payload)
	}
	return data, err
}

// FXUsername will retrieve the device's FX username string
func (i1 *I1Device) FXUsername() (string, error) {
	username := ""
	msg, err := SendCommandAndWait(i1, CmdFxUsernameReq, CmdFxUsernameResp)
	if err == nil {
		username = string(msg.Payload)
	}
	return username, err
}

// TextString will retrieve the Device Text String
func (i1 *I1Device) TextString() (string, error) {
	text := ""
	msg, err := SendCommandAndWait(i1, CmdDeviceTextStringReq, CmdDeviceTextStringResp)
	if err == nil {
		text = string(msg.Payload)
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
	if i1.engineVersion == 0xff {
		var ack *Message
		ack, err = SendCommand(i1, CmdGetEngineVersion)
		if err == nil {
			i1.engineVersion = EngineVersion(ack.Command.Command2)
		}
	}
	return i1.engineVersion, err
}

// Ping will send a Ping command to the device
func (i1 *I1Device) Ping() error {
	_, err := SendCommand(i1, CmdPing)
	return err
}

func (i1 *I1Device) idRequest() {
	if i1.devCat == (DevCat{0xff, 0xff}) || i1.firmwareVersion == FirmwareVersion(0xff) {
		waitCommands := []CommandBytes{CmdSetButtonPressedController.Version(0), CmdSetButtonPressedResponder.Version(0)}
		rxCh := i1.Connection.Subscribe(waitCommands...)
		defer i1.Connection.Unsubscribe(rxCh)
		ack, err := sendCommand(i1.Connection, CmdIDReq.Version(0), StandardDirectMessage, nil)

		if err == nil {
			if ack.Ack() {
				select {
				case msg := <-rxCh:
					i1.devCat = DevCat{msg.Dst[0], msg.Dst[1]}
					i1.firmwareVersion = FirmwareVersion(msg.Dst[2])
				case <-time.After(Timeout):
				}
			}
		}
	}
	return
}

func (i1 *I1Device) DevCat() DevCat {
	i1.idRequest()
	return i1.devCat
}

func (i1 *I1Device) FirmwareVersion() FirmwareVersion {
	i1.idRequest()
	return i1.firmwareVersion
}

// SetTextString will set the device text string
func (i1 *I1Device) SetTextString(str string) error {
	textString := make([]byte, 14)
	copy(textString, []byte(str))
	_, err := SendExtendedCommand(i1, CmdSetDeviceTextString, textString)
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
	return sprintf("I1 Device (%s)", i1.Address())
}

// Close closes the underlying connection
func (i1 *I1Device) Close() (err error) {
	Log.Debugf("Closing I1Device connection")
	if closeable, ok := i1.Connection.(Closeable); ok {
		err = closeable.Close()
	}
	return err
}
