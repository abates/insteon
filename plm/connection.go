package plm

import (
	"time"

	"github.com/abates/insteon"
)

type txReq struct {
	ackCh   chan *insteon.Message
	message *insteon.Message
}

type msgSubReq struct {
	matches     []*insteon.Command
	unsubscribe bool
	rxCh        <-chan *insteon.Message
	respCh      chan bool
}

type msgSubscription struct {
	matches []*insteon.Command
	ch      chan<- *insteon.Message
}

func (sub *msgSubscription) match(msg *insteon.Message) bool {
	for _, match := range sub.matches {
		if match.Cmd == msg.Command.Cmd {
			return true
		}
	}
	return false
}

type Connection struct {
	txCh        chan *insteon.Message
	rxCh        chan *insteon.Message
	txReqCh     chan *txReq
	msgSubReqCh chan *msgSubReq
	closeCh     chan chan error
}

func NewConnection(plm *PLM, dst insteon.Address) *Connection {
	conn := &Connection{
		txCh: make(chan *insteon.Message),
		rxCh: make(chan *insteon.Message),

		txReqCh:     make(chan *txReq),
		msgSubReqCh: make(chan *msgSubReq),
		closeCh:     make(chan chan error),
	}

	go conn.readWriteLoop(plm, dst)
	return conn
}

func (conn *Connection) readWriteLoop(plm *PLM, dst insteon.Address) {
	var closeCh chan error
	ackChannels := make(map[byte]chan *insteon.Message)
	rxChannels := make(map[<-chan *insteon.Message]*msgSubscription)
	rxCh := plm.Subscribe([]byte{0x50, dst[0], dst[1], dst[2]}, []byte{0x51, dst[0], dst[1], dst[2]})

loop:
	for {
		select {
		case txReq := <-conn.txReqCh:
			ackChannels[txReq.message.Command.Cmd[0]] = txReq.ackCh
			if txReq.message.Flags.Type() == insteon.MsgTypeDirect {
				txReq.message.Dst = dst
			}
			packet := &Packet{
				Command: CmdSendInsteonMsg,
				Payload: txReq.message,
			}
			plm.Send(packet)
		case msgSubReq := <-conn.msgSubReqCh:
			if msgSubReq.unsubscribe {
				if sub, found := rxChannels[msgSubReq.rxCh]; found {
					delete(rxChannels, msgSubReq.rxCh)
					close(sub.ch)
				}
			} else {
				ch := make(chan *insteon.Message, 1)
				msgSubReq.rxCh = ch
				rxChannels[msgSubReq.rxCh] = &msgSubscription{matches: msgSubReq.matches, ch: ch}
				msgSubReq.respCh <- true
				close(msgSubReq.respCh)
			}
		case pkt := <-rxCh:
			msg := pkt.Payload.(*insteon.Message)
			if msg.Flags.Type() == insteon.MsgTypeDirectAck || msg.Flags.Type() == insteon.MsgTypeDirectNak {
				cmd := msg.Command.Cmd[0]
				if ch, found := ackChannels[cmd]; found {
					insteon.Log.Debugf("Dispatching insteon ACK/NAK %v", msg)
					select {
					case ch <- msg:
					default:
					}
					close(ch)
					delete(ackChannels, cmd)
				}
			} else {
				for _, sub := range rxChannels {
					if sub.match(msg) {
						select {
						case sub.ch <- msg:
						default:
						}
					}
				}
			}
		case closeCh = <-conn.closeCh:
			break loop
		}
	}

	for _, ch := range ackChannels {
		close(ch)
	}

	for _, sub := range rxChannels {
		close(sub.ch)
	}

	plm.Unsubscribe(rxCh)
	closeCh <- nil
	close(closeCh)
}

func (conn *Connection) Subscribe(matches ...*insteon.Command) <-chan *insteon.Message {
	respCh := make(chan bool)
	req := &msgSubReq{matches: matches, respCh: respCh}
	conn.msgSubReqCh <- req
	<-respCh
	return req.rxCh
}

func (conn *Connection) Unsubscribe(rxCh <-chan *insteon.Message) {
	conn.msgSubReqCh <- &msgSubReq{rxCh: rxCh, unsubscribe: true}
}

func (conn *Connection) Close() error {
	insteon.Log.Debugf("Closing PLM Connection")
	ch := make(chan error)
	conn.closeCh <- ch
	return <-ch
}

func (conn *Connection) Write(message *insteon.Message) (ack *insteon.Message, err error) {
	ackCh := make(chan *insteon.Message)
	conn.txReqCh <- &txReq{ackCh, message}

	select {
	case ack = <-ackCh:
	case <-time.After(insteon.Timeout):
		err = ErrAckTimeout
	}
	return
}
