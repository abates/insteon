package insteon

import (
	"time"
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
		if match.cmd == msg.Command.cmd {
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
	Subscribe(match ...*Command) chan *Message
	Unsubscribe(chan *Message)
	Close() error
}

type I1Connection struct {
	destination Address
	timeout     time.Duration

	txCh    chan *Message
	rxCh    chan *Message
	txReqCh chan *txReq
	rxReqCh chan *rxReq
	closeCh chan chan error
}

func NewI1Connection(destination Address, txCh, rxCh chan *Message) Connection {
	conn := &I1Connection{
		destination: destination,
		timeout:     5 * time.Second,

		txCh:    txCh,
		rxCh:    rxCh,
		txReqCh: make(chan *txReq),
		rxReqCh: make(chan *rxReq),
		closeCh: make(chan chan error),
	}

	go conn.readWriteLoop()
	return conn
}

func (conn *I1Connection) readWriteLoop() {
	var closeCh chan error
	ackChannels := make(map[byte]chan *Message)
	rxChannels := make(map[chan *Message]*rxReq)

loop:
	for {
		select {
		case txReq := <-conn.txReqCh:
			ackChannels[txReq.message.Command.cmd[0]] = txReq.ackCh
			Log.Debugf("Adding ACK %+v to %+v", txReq.message.Command.cmd[0], ackChannels)
			if txReq.message.Flags.Type() == MsgTypeDirect {
				txReq.message.Dst = conn.destination
			}
			conn.txCh <- txReq.message
		case rxReq := <-conn.rxReqCh:
			if rxReq.unsubscribe {
				if req, found := rxChannels[rxReq.rxCh]; found {
					delete(rxChannels, req.rxCh)
					close(req.rxCh)
				}
			} else {
				rxChannels[rxReq.rxCh] = rxReq
			}
		case msg := <-conn.rxCh:
			if msg.Flags.Type() == MsgTypeDirectAck || msg.Flags.Type() == MsgTypeDirectNak {
				cmd := msg.Command.cmd[0]
				if ch, found := ackChannels[cmd]; found {
					Log.Debugf("Dispatching ACK %v", msg)
					select {
					case ch <- msg:
					default:
					}
					close(ch)
					delete(ackChannels, cmd)
				} else {
					Log.Debugf("No one was waiting for ACK: %v", cmd)
					Log.Debugf("Ack channels are %+v", ackChannels)
				}
			} else {
				for _, req := range rxChannels {
					if req.match(msg) {
						req.dispatch(msg)
					}
				}
			}
		case closeCh = <-conn.closeCh:
			break loop
		}
	}
	close(conn.txCh)
	closeCh <- nil
}

func (conn *I1Connection) Subscribe(matches ...*Command) chan *Message {
	rxCh := make(chan *Message, 1)
	conn.rxReqCh <- &rxReq{matches: matches, rxCh: rxCh}
	return rxCh
}

func (conn *I1Connection) Unsubscribe(rxCh chan *Message) {
	conn.rxReqCh <- &rxReq{rxCh: rxCh, unsubscribe: true}
}

func (conn *I1Connection) Close() error {
	ch := make(chan error)
	conn.closeCh <- ch
	return <-ch
}

func (conn *I1Connection) Write(message *Message) (ack *Message, err error) {
	ackCh := make(chan *Message)
	conn.txReqCh <- &txReq{ackCh, message}

	select {
	case ack = <-ackCh:
		if ack.Flags.Type() == MsgTypeDirectNak {
			switch ack.Command.cmd[1] {
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
	case <-time.After(conn.timeout):
		err = ErrAckTimeout
	}
	return
}

type I2CsConnection struct {
	Connection
}

func NewI2CsConnection(destination Address, tx, rx chan *Message) Connection {
	return &I2CsConnection{NewI1Connection(destination, tx, rx)}
}

func (i2csw *I2CsConnection) Write(message *Message) (*Message, error) {
	message.version = VerI2Cs
	ack, err := i2csw.Connection.Write(message)
	if ack != nil && ack.Flags.Type() == MsgTypeDirectNak {
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
	}
	return ack, err
}

func SendStandardCommandAndWait(conn Connection, command *Command, waitCmd *Command) (msg *Message, err error) {
	rxCh := conn.Subscribe(waitCmd)
	_, err = SendStandardCommand(conn, command)

	if err == nil {
		msg = <-rxCh
		conn.Unsubscribe(rxCh)
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
