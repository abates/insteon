package insteon

// I1Device provides remote communication to version 1 engines
type I1Device struct {
	address         Address
	network         Network
	devCat          DevCat
	firmwareVersion FirmwareVersion

	ackCh         chan *Message
	productDataCh chan *Message
}

// NewI1Device will construct an I1Device for the given address and connection
func NewI1Device(address Address, network Network) *I1Device {
	return &I1Device{
		address:         address,
		network:         network,
		devCat:          DevCat{0xff, 0xff},
		firmwareVersion: FirmwareVersion(0x00),

		ackCh:         make(chan *Message),
		productDataCh: make(chan *Message),
	}
}

func (i1 *I1Device) Notify(msg *Message) error {
	if msg.Ack() || msg.Nak() {
		writeToCh(i1.ackCh, msg)
	} else if msg.Flags.Extended() && msg.Command[0] == 0x03 {
		// Product Data Response
		writeToCh(i1.productDataCh, msg)
	}
	return nil
}

func (i1 *I1Device) SendCommand(command Command, payload []byte) (response Command, err error) {
	flags := StandardDirectMessage
	if len(payload) > 0 {
		flags = ExtendedDirectMessage
	}

	err = i1.network.SendMessage(&Message{
		Src:     i1.address,
		Flags:   flags,
		Command: command,
		Payload: payload,
	})

	if err == nil {
		var ack *Message
		ack, err = readFromCh(i1.ackCh)
		if err == nil {
			response = ack.Command
		}
	}

	return
}

// Address is the Insteon address of the device
func (i1 *I1Device) Address() Address {
	return i1.address
}

func extractError(v interface{}, err error) error {
	return err
}

// AssignToAllLinkGroup will inform the device what group should be used during an All-Linking
// session
func (i1 *I1Device) AssignToAllLinkGroup(group Group) error {
	return extractError(i1.SendCommand(CmdAssignToAllLinkGroup.SubCommand(int(group)), nil))
}

// DeleteFromAllLinkGroup will inform the device which group should be unlinked during an
// All-Link unlinking session
func (i1 *I1Device) DeleteFromAllLinkGroup(group Group) (err error) {
	return extractError(i1.SendCommand(CmdDeleteFromAllLinkGroup.SubCommand(int(group)), nil))
}

// ProductData will retrieve the device's product data
func (i1 *I1Device) ProductData() (data *ProductData, err error) {
	_, err = i1.SendCommand(CmdProductDataReq, nil)
	if err == nil {
		var msg *Message
		msg, err = readFromCh(i1.productDataCh)
		if err == nil {
			data = &ProductData{}
			err = data.UnmarshalBinary(msg.Payload)
		}
	}
	return data, err
}

// Ping will send a Ping command to the device
func (i1 *I1Device) Ping() (err error) {
	return extractError(i1.SendCommand(CmdPing, nil))
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

func (i1 *I1Device) String() string {
	return sprintf("I1 Device (%s)", i1.Address())
}
