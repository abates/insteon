package insteon

import (
	"time"
)

var (
	Timeout = 5 * time.Second
)

type txReq struct {
	ackCh   chan *Message
	message *Message
}

type rxReq struct {
	matches     []*Command
	unsubscribe bool
	rxCh        chan *Message
}

func (req *rxReq) match(msg *Message) bool {
	for _, match := range req.matches {
		if match.Cmd == msg.Command.Cmd {
			return true
		}
	}
	return false
}

func (req *rxReq) dispatch(msg *Message) {
	req.rxCh <- msg
}

type Connection interface {
	Write(*Message) (ack *Message, err error)
	Subscribe(match ...*Command) <-chan *Message
	Unsubscribe(<-chan *Message)
	Close() error
}

type I1Connection struct {
	Connection
}

func NewI1Connection(conn Connection) Connection {
	return &I1Connection{conn}
}

func (i1conn *I1Connection) Write(message *Message) (ack *Message, err error) {
	ack, err = i1conn.Connection.Write(message)
	if ack != nil && ack.Flags.Type() == MsgTypeDirectNak {
		switch ack.Command.Cmd[1] {
		case 0xfd:
			err = ErrUnknownCommand
		case 0xfe:
			err = ErrNoLoadDetected
		case 0xff:
			err = ErrNotLinked
		default:
			err = TraceError(ErrUnexpectedResponse)
		}
	}
	return
}

type I2CsConnection struct {
	Connection
}

func NewI2CsConnection(conn Connection) Connection {
	return &I2CsConnection{conn}
}

func (i2csw *I2CsConnection) Write(message *Message) (*Message, error) {
	message.version = VerI2Cs
	ack, err := i2csw.Connection.Write(message)
	if ack != nil && ack.Flags.Type() == MsgTypeDirectNak {
		switch ack.Command.Cmd[1] {
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
	}
	return ack, err
}

func SendStandardCommandAndWait(conn Connection, command *Command, waitCmd *Command) (msg *Message, err error) {
	rxCh := conn.Subscribe(waitCmd)
	_, err = SendStandardCommand(conn, command)

	if err == nil {
		Log.Debugf("Waiting for %v", waitCmd)
		select {
		case msg = <-rxCh:
		case <-time.After(Timeout):
			err = ErrAckTimeout
		}
		Log.Debugf("Got response %v", msg)
		conn.Unsubscribe(rxCh)
		Log.Debugf("Unsubscribed!")
	}
	return
}

func SendStandardCommand(conn Connection, command *Command) (*Message, error) {
	return conn.Write(&Message{
		Flags:   StandardDirectMessage,
		Command: command,
	})
}

func SendExtendedCommand(conn Connection, command *Command, payload Payload) (*Message, error) {
	return conn.Write(&Message{
		Flags:   ExtendedDirectMessage,
		Command: command,
		Payload: payload,
	})
}
