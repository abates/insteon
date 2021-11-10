package plm

import (
	"errors"
	"io"

	"github.com/abates/insteon"
	"github.com/abates/insteon/devices"
)

type snoop struct {
	msgBuf <-chan *Packet
}

func (s *snoop) Read() (*insteon.Message, error) {
	pkt := <-s.msgBuf
	msg := &insteon.Message{}
	err := msg.UnmarshalBinary(pkt.Payload)
	return msg, err
}

func (s *snoop) Write(*insteon.Message) (ack *insteon.Message, err error) {
	// We can't write to a snooped PLM
	return nil, ErrNotImplemented
}

func Snoop(rx, tx io.Reader) devices.MessageWriter {
	msgBuf := make(chan *Packet, 10)
	s := &snoop{
		msgBuf: msgBuf,
	}
	go s.readLoop(newPacketReader(tx, false), msgBuf)
	go s.readLoop(newPacketReader(rx, true), msgBuf)
	return s
}

func (s *snoop) readLoop(reader *packetReader, msgBuf chan<- *Packet) {
	for {
		pkt, err := reader.ReadPacket()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				Log.Printf("Read error: %v", err)
			}
			return
		}

		if pkt.Command == CmdStdMsgReceived || pkt.Command == CmdExtMsgReceived || pkt.Command == CmdSendInsteonMsg {
			msgBuf <- pkt
		}
	}
}
