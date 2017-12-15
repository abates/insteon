package plm

import (
	"time"

	"github.com/abates/insteon"
)

type plmBridge struct {
	plm *PLM
	ch  chan *Packet
}

func NewBridge(plm *PLM, dst insteon.Address) *plmBridge {
	ch := make(chan *Packet, 1)
	plm.Subscribe(ch, []byte{0x50, dst[0], dst[1], dst[2]}, []byte{0x51, dst[0], dst[1], dst[2]})
	return &plmBridge{
		plm: plm,
		ch:  ch,
	}
}

func (pb *plmBridge) Send(payload insteon.Payload) error {
	packet := &Packet{
		Command: CmdSendInsteonMsg,
		Payload: payload,
	}
	_, err := pb.plm.Send(packet)
	return err
}

func (pb *plmBridge) Receive() (payload insteon.Payload, err error) {
	select {
	case packet := <-pb.ch:
		payload = packet.Payload
	case <-time.After(pb.plm.timeout):
		err = insteon.ErrReadTimeout
	}
	return
}

func (pb *plmBridge) Close() error {
	pb.plm.Unsubscribe(pb.ch)
	return nil
}
