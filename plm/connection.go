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
		if match.Cmd[0] == msg.Command.Cmd[0] {
			return true
		}
	}
	return false
}

type Connection struct {
	dst      insteon.Address
	plm      *PLM
	category insteon.Category

	txCh        chan *insteon.Message
	rxCh        chan *insteon.Message
	txReqCh     chan *txReq
	msgSubReqCh chan *msgSubReq
	closeCh     chan chan error
}

func NewConnection(plm *PLM, dst insteon.Address) *Connection {
	conn := &Connection{
		dst:  dst,
		plm:  plm,
		txCh: make(chan *insteon.Message),
		rxCh: make(chan *insteon.Message),

		txReqCh:     make(chan *txReq),
		msgSubReqCh: make(chan *msgSubReq),
		closeCh:     make(chan chan error),
	}

	go conn.readWriteLoop(plm, conn.dst)

	return conn
}

func (conn *Connection) writePacket(message *insteon.Message) error {
	if message.Flags.Type() == insteon.MsgTypeDirect {
		message.Dst = conn.dst
	}

	payload, err := message.MarshalBinary()
	if err == nil {
		// PLM expects that the payload begins with the
		// destinations address so we have to slice off
		// the src address
		payload = payload[3:]
		packet := &Packet{
			Command: CmdSendInsteonMsg,
			payload: payload,
		}
		_, err = conn.plm.Retry(packet, 3)
	}
	return err
}

func (conn *Connection) readWriteLoop(plm *PLM, dst insteon.Address) {
	var closeCh chan error
	txReqs := make([]*txReq, 0)
	rxChannels := make(map[<-chan *insteon.Message]*msgSubscription)
	rxCh := plm.Subscribe([]byte{0x50, dst[0], dst[1], dst[2]}, []byte{0x51, dst[0], dst[1], dst[2]})
	defer plm.Unsubscribe(rxCh)

loop:
	for {
		select {
		case txReq := <-conn.txReqCh:
			if len(txReqs) == 0 {
				conn.writePacket(txReq.message)
			}
			txReqs = append(txReqs, txReq)
		case msgSubReq := <-conn.msgSubReqCh:
			if msgSubReq.unsubscribe {
				if sub, found := rxChannels[msgSubReq.rxCh]; found {
					delete(rxChannels, msgSubReq.rxCh)
					close(sub.ch)
				}
			} else {
				ch := make(chan *insteon.Message, 10)
				msgSubReq.rxCh = ch
				rxChannels[msgSubReq.rxCh] = &msgSubscription{matches: msgSubReq.matches, ch: ch}
				msgSubReq.respCh <- true
				close(msgSubReq.respCh)
			}
		case pkt := <-rxCh:
			msg := &insteon.Message{}
			msg.DevCat = conn.category
			err := msg.UnmarshalBinary(pkt.payload)
			if err == nil {
				insteon.Log.Debugf("Connection received %v", msg)
				if msg.Flags.Type() == insteon.MsgTypeDirectAck || msg.Flags.Type() == insteon.MsgTypeDirectNak {
					if len(txReqs) > 0 {
						ch := txReqs[0].ackCh
						txReqs = txReqs[1:]
						insteon.Log.Debugf("Dispatching insteon ACK/NAK %v", msg)
						select {
						case ch <- msg:
						default:
							insteon.Log.Debugf("insteon ACK/NAK channel was not ready, discarding %v", msg)
						}
						close(ch)

						if len(txReqs) > 0 {
							conn.writePacket(txReqs[0].message)
						}
					} else {
						insteon.Log.Debugf("received ACK/NAK but no ack channel present, discarding %v", msg)
					}
				} else {
					for _, sub := range rxChannels {
						if sub.match(msg) {
							select {
							case sub.ch <- msg:
							default:
								insteon.Log.Infof("Insteon subscription exists, but buffer is full. discarding %v", msg)
							}
						}
					}
				}
			} else {
				insteon.Log.Infof("Failed to unmarshal message: %v", err)
			}
		case closeCh = <-conn.closeCh:
			break loop
		}
	}

	for _, req := range txReqs {
		close(req.ackCh)
	}

	for _, sub := range rxChannels {
		close(sub.ch)
	}

	closeCh <- nil
	close(closeCh)
}

func (conn *Connection) DevCat() (devCat insteon.Category, err error) {
	if [2]byte(conn.category) == [2]byte{0x00, 0x00} {
		conn.category, err = insteon.NewI1Device(conn.dst, conn).IDRequest()
	}

	return conn.category, err
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
		err = insteon.ErrAckTimeout
	}
	return
}
