package insteon

type I2Device struct {
	*I1Device
	linkCh chan *Message
}

func NewI2Device(address Address, network Network) *I2Device {
	return &I2Device{
		I1Device: NewI1Device(address, network),
		linkCh:   make(chan *Message),
	}
}

func (i2 *I2Device) Notify(msg *Message) error {
	if msg.Flags.Extended() && msg.Command[0] == 0x2f {
		writeToCh(i2.linkCh, msg)
	}
	return i2.I1Device.Notify(msg)
}

// AddLink will either add the link to the All-Link database
// or it will replace an existing link-record that has been marked
// as deleted
func (i2 *I2Device) AddLink(newLink *LinkRecord) error {
	return ErrNotImplemented
}

// RemoveLinks will either remove the link records from the device
// All-Link database, or it will simply mark them as deleted
func (i2 *I2Device) RemoveLinks(oldLinks ...*LinkRecord) error {
	return ErrNotImplemented
}

// Links will retrieve the link-database from the device and
// return a list of LinkRecords
func (i2 *I2Device) Links() (links []*LinkRecord, err error) {
	Log.Debugf("Retrieving Device link database")
	lastAddress := MemAddress(0)
	buf, _ := (&LinkRequest{Type: ReadLink, NumRecords: 0}).MarshalBinary()
	_, err = i2.SendCommand(CmdReadWriteALDB, buf)

	var msg *Message
	for err == nil {
		msg, err = readFromCh(i2.linkCh)
		if err == nil {
			lr := &LinkRequest{}
			err = lr.UnmarshalBinary(msg.Payload)
			if err == nil && lr.MemAddress != lastAddress {
				lastAddress = lr.MemAddress
				links = append(links, lr.Link)
			}
		}
	}

	if err == ErrEndOfLinks {
		err = nil
	}
	return links, err
}

func (i2 *I2Device) EnterLinkingMode(group Group) error {
	return extractError(i2.SendCommand(CmdEnterLinkingMode.SubCommand(int(group)), nil))
}

func (i2 *I2Device) EnterUnlinkingMode(group Group) error {
	return extractError(i2.SendCommand(CmdEnterUnlinkingMode.SubCommand(int(group)), nil))
}

func (i2 *I2Device) ExitLinkingMode() error {
	return extractError(i2.SendCommand(CmdExitLinkingMode, nil))
}

func (i2 *I2Device) WriteLink(link *LinkRecord) (err error) {
	if link.memAddress == MemAddress(0x0000) {
		err = ErrInvalidMemAddress
	} else {
		buf, _ := (&LinkRequest{MemAddress: link.memAddress, Type: WriteLink, Link: link}).MarshalBinary()
		_, err = i2.SendCommand(CmdReadWriteALDB, buf)
	}
	return err
}

func (i2 *I2Device) String() string {
	return sprintf("I2 Device (%s)", i2.Address())
}
