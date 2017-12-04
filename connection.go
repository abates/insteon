package insteon

func SendStandardCommandAndWait(conn Connection, command *Command) (msg *Message, err error) {
	_, err = SendStandardCommand(conn, command)
	if err == nil {
		msg, err = conn.Receive()
	}
	return
}

func SendStandardCommand(conn Connection, command *Command) (*Message, error) {
	return conn.Send(&Message{
		Flags:   StandardDirectMessage,
		Command: command,
	})
}

func SendExtendedCommand(conn Connection, command *Command, payload Payload) (*Message, error) {
	return conn.Send(&Message{
		Flags:   ExtendedDirectMessage,
		Command: command,
		Payload: payload,
	})
}

type Bridge interface {
	Send(Payload) error
	Receive() (Payload, error)
}

type Connection interface {
	Send(*Message) (*Message, error)
	Receive() (*Message, error)
}

func NewI1Connection(address Address, bridge Bridge) Connection {
	return &I1Connection{
		address: address,
		bridge:  bridge,
	}
}

type I1Connection struct {
	address Address
	bridge  Bridge
}

func (i1c *I1Connection) Send(msg *Message) (ack *Message, err error) {
	if msg.Flags.Type().Direct() {
		msg.Dst = i1c.address
	}

	err = i1c.bridge.Send(msg)

	if err == nil {
		ack, err = i1c.Receive()
		if err == nil {
			if ack.Flags.Type() == MsgTypeDirectAck {
				if ack.Command.cmd[0] == msg.Command.cmd[0] {
					Log.Debugf("INSTEON ACK received")
				} else {
					err = ErrUnexpectedResponse
				}
			} else if ack.Flags.Type() == MsgTypeDirectNak {
				if ack.Command.cmd[1] == 0xfd {
					err = ErrUnknownCommand
				} else if ack.Command.cmd[1] == 0xfe {
					err = ErrNoLoadDetected
				} else if ack.Command.cmd[1] == 0xff {
					err = ErrNotLinked
				}
			} else {
				err = ErrUnexpectedResponse
			}
		} else if err == ErrReadTimeout {
			// timed out waiting to read the Ack
			err = ErrAckTimeout
		}
	}

	return
}

func (i1c *I1Connection) Receive() (message *Message, err error) {
	payload, err := i1c.bridge.Receive()
	if msg, ok := payload.(*Message); ok {
		return msg, err
	}
	return nil, ErrUnexpectedResponse
}

func NewI2Connection(address Address, bridge Bridge) Connection {
	return &I2Connection{
		address: address,
		bridge:  bridge,
	}
}

type I2Connection struct {
	address Address
	bridge  Bridge
}

func (i2c *I2Connection) Send(msg *Message) (ack *Message, err error) {
	if msg.Flags.Type().Direct() {
		msg.Dst = i2c.address
	}

	// update checksum prior to sending
	if msg.Flags.Extended() {
		err = i2c.bridge.Send(&I2CsMessage{msg})
	} else {
		err = i2c.bridge.Send(msg)
	}

	if err == nil {
		ack, err = i2c.Receive()
		if err == nil {
			if ack.Flags.Type() == MsgTypeDirectAck {
				if ack.Command.cmd[0] == msg.Command.cmd[0] {
					Log.Debugf("INSTEON ACK received")
				} else {
					err = ErrUnexpectedResponse
				}
			} else if ack.Flags.Type() == MsgTypeDirectNak {
				switch ack.Command.cmd[1] {
				case 0xfb:
					err = ErrIllegalValue
				case 0xfc:
					err = ErrPreNak
				case 0xfd:
					err = ErrIncorrectChecksum
				case 0xfe:
					err = ErrNoLoadDetected
				case 0xff:
					err = ErrNotLinked
				default:
					err = ErrUnknown
				}
			} else {
				err = ErrUnexpectedResponse
			}
		} else if err == ErrReadTimeout {
			// timed out waiting to read the Ack
			err = ErrAckTimeout
		}
	}

	return
}

func (i2c *I2Connection) Receive() (message *Message, err error) {
	payload, err := i2c.bridge.Receive()
	if msg, ok := payload.(*Message); ok {
		return msg, err
	}
	return nil, ErrUnexpectedResponse
}
