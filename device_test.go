package insteon

type TestDevice struct {
	messages []*Message
}

func (td *TestDevice) SendCommand(cmd Command, payload []byte) (response Command, err error) {
	return
}

func (td *TestDevice) Notify(msg *Message) error {
	td.messages = append(td.messages, msg)
	return nil
}

func (te *TestDevice) Address() Address { return Address{} }
