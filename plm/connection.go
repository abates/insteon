package plm

import (
	"github.com/abates/insteon"
)

type Connection struct {
	plm  *PLM
	txCh chan *insteon.Message
	rxCh chan *insteon.Message
}

func NewConnection(plm *PLM, dst insteon.Address) *Connection {
	conn := &Connection{
		txCh: make(chan *insteon.Message),
		rxCh: make(chan *insteon.Message),
	}

	go conn.readWriteLoop(plm, dst)
	return conn
}

func (conn *Connection) readWriteLoop(plm *PLM, dst insteon.Address) {
	rxCh := plm.Subscribe([]byte{0x50, dst[0], dst[1], dst[2]}, []byte{0x51, dst[0], dst[1], dst[2]})
loop:
	for {
		select {
		case pkt := <-rxCh:
			conn.rxCh <- pkt.Payload.(*insteon.Message)
		case msg, open := <-conn.txCh:
			if !open {
				break loop
			}
			packet := &Packet{
				Command: CmdSendInsteonMsg,
				Payload: msg,
			}
			plm.Send(packet)
		}
	}
	plm.Unsubscribe(rxCh)
	close(conn.rxCh)
}
