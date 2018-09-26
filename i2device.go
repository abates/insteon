package insteon

type I2Device struct {
	*I1Device
}

func NewI2Device(address Address, sendCh chan<- *MessageRequest, recvCh <-chan *Message) *I2Device {
	return &I2Device{NewI1Device(address, sendCh, recvCh)}
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
	recvCh, err := i2.SendCommandAndListen(CmdReadWriteALDB, buf)

	for response := range recvCh {
		if response.Message.Flags.Extended() && response.Message.Command[0] == CmdReadWriteALDB[0] {
			lr := &LinkRequest{}
			err = lr.UnmarshalBinary(response.Message.Payload)
			if err == nil && lr.MemAddress != lastAddress {
				lastAddress = lr.MemAddress
				links = append(links, lr.Link)
				response.DoneCh <- false
			} else if err == ErrEndOfLinks {
				response.DoneCh <- true
				err = nil
			} else {
				response.DoneCh <- true
			}
		} else {
			response.DoneCh <- false
		}
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
