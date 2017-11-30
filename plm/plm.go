package plm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/abates/insteon"
)

var (
	ErrNoSync         = errors.New("No sync byte received")
	ErrNotImplemented = errors.New("IM command not implemented")
)

type Config byte

type PLM interface {
	Info() (*IMInfo, error)
	Reset() error
	Config() (Config, error)
	SetConfig(Config) error
	SetDeviceCategory(insteon.Category) error
	RFSleep() error
	Connect(insteon.Address) (insteon.Device, error)
	//Links() []*insteon.Link
}

type connectionInfo struct {
	address insteon.Address
	ch      chan *Packet
}

type txPacketInfo struct {
	pkt   *Packet
	ackCh chan *Packet
}

type plm struct {
	in  *bufio.Reader
	out io.Writer

	txPktCh      chan *txPacketInfo
	rxPktCh      chan *Packet
	connectionCh chan connectionInfo
}

func New(port io.ReadWriter) PLM {
	plm := &plm{
		in:  bufio.NewReader(port),
		out: port,

		txPktCh:      make(chan *txPacketInfo, 1),
		rxPktCh:      make(chan *Packet, 1),
		connectionCh: make(chan connectionInfo, 1),
	}
	go plm.readPktLoop()
	go plm.readWriteLoop()
	return plm
}

func traceBuf(prefix string, buf []byte) {
	bb := make([]string, len(buf))
	for i, b := range buf {
		bb[i] = fmt.Sprintf("%02x", b)
	}
	insteon.Log.Tracef("%-05s BUFFER %s", prefix, strings.Join(bb, " "))
}

func tracePkt(prefix string, pkt *Packet) {
	insteon.Log.Tracef("%-05s %s", prefix, pkt)
}

func traceMsg(prefix string, msg *insteon.Message) {
	insteon.Log.Tracef("%-05s %s", prefix, msg)
}

func (p *plm) read(buf []byte) error {
	_, err := io.ReadAtLeast(p.in, buf, len(buf))
	return err
}

func (p *plm) readPacket() (pkt *Packet, err error) {
	var buf []byte
	b, err := p.in.ReadByte()
	if err == nil && b != 0x02 {
		return nil, fmt.Errorf("Expected first byte to be 0x02 got 0x%02x", b)
	}
	buf = append(buf, b)

	b, err = p.in.ReadByte()

	if err == nil {
		buf = append(buf, b)
		// TODO commandLens should only be written during
		// initialization, but, technically speaking, this
		// access could cause a concurrent access violation
		if pktLen, ok := commandLens[b]; ok {
			buf = append(buf, make([]byte, pktLen)...)
			_, err = io.ReadAtLeast(p.in, buf[2:], pktLen)
			if err == nil {
				traceBuf("RX", buf)
				// read some more if it's an extended message
				if buf[1] == 0x62 && insteon.Flags(buf[5]).IsExtended() {
					buf = append(buf, make([]byte, 14)...)
					_, err = io.ReadAtLeast(p.in, buf[9:], 14)
				}
				pkt = &Packet{}
				err = pkt.UnmarshalBinary(buf)
			}
		} else {
			err = fmt.Errorf("PLM Received unknown command 0x%02x", b)
		}
	}
	return pkt, err
}

func (p *plm) readPktLoop() {
	for {
		pkt, err := p.readPacket()
		if err == nil {
			tracePkt("RX", pkt)
			p.rxPktCh <- pkt
		} else {
			insteon.Log.Infof("Error reading packet: %v", err)
		}
	}
}

func (p *plm) writePacket(pkt *Packet) error {
	payload, err := pkt.MarshalBinary()
	traceBuf("TX", payload)

	if err == nil {
		_, err = p.out.Write(payload)
	}
	return err
}

func (p *plm) readWriteLoop() {
	connections := make(map[insteon.Address]chan *Packet)
	ackChannels := make(map[Command]chan *Packet)
	for {
		var pkt *Packet
		insteon.Log.Debugf("readWriteLoop wait...")
		select {
		case send := <-p.txPktCh:
			ackChannels[send.pkt.Command] = send.ackCh
			err := p.writePacket(send.pkt)
			if err == nil {
				tracePkt("TX", send.pkt)
			}
		case pkt = <-p.rxPktCh:
			switch {
			case pkt.Command == 0x50 || pkt.Command == 0x51:
				msg := pkt.Payload.(*insteon.Message)
				insteon.Log.Debugf("Received INSTEON Message %v", msg)
				if conn, ok := connections[msg.Src]; ok {
					insteon.Log.Debugf("Dispatching message to device connection")
					conn <- pkt
				}
			case 0x52 <= pkt.Command && pkt.Command <= 0x58:
				// handle event
			default:
				// handle ack/nak
				if ackCh, ok := ackChannels[pkt.Command]; ok {
					select {
					case ackCh <- pkt:
						close(ackCh)
						ackChannels[pkt.Command] = nil
					default:
					}
				}
			}
		case info := <-p.connectionCh:
			connections[info.address] = info.ch
		}
	}
}

func (p *plm) Info() (*IMInfo, error) {
	return nil, ErrNotImplemented
}

func (p *plm) Reset() error {
	return ErrNotImplemented
}

func (p *plm) Config() (Config, error) {
	return Config(0x00), ErrNotImplemented
}

func (p *plm) SetConfig(Config) error {
	return ErrNotImplemented
}

func (p *plm) SetDeviceCategory(insteon.Category) error {
	return ErrNotImplemented
}

func (p *plm) RFSleep() error {
	return ErrNotImplemented
}

type plmBridge struct {
	rx chan *Packet
	tx chan *txPacketInfo
}

func (pb *plmBridge) Send(timeout time.Duration, msg *insteon.Message) (err error) {
	pkt := &Packet{
		retryCount: 3,
		Command:    CmdSendInsteonMsg,
		Payload:    msg,
	}
	tracePkt("BR TX", pkt)
	ackCh := make(chan *Packet, 1)
	txPktInfo := &txPacketInfo{
		pkt:   pkt,
		ackCh: ackCh,
	}

	select {
	case pb.tx <- txPktInfo:
		select {
		case <-ackCh:
			insteon.Log.Debugf("PLM ACK Received")
		case <-time.After(timeout):
			err = insteon.ErrAckTimeout
		}
	case <-time.After(timeout):
		err = insteon.ErrWriteTimeout
	}
	return err
}

func (pb *plmBridge) Receive(timeout time.Duration) (msg *insteon.Message, err error) {
	select {
	case pkt := <-pb.rx:
		msg = pkt.Payload.(*insteon.Message)
	case <-time.After(timeout):
		err = insteon.ErrReadTimeout
	}
	return
}

func (p *plm) Connect(dst insteon.Address) (insteon.Device, error) {
	rx := make(chan *Packet, 1)
	bridge := &plmBridge{
		tx: p.txPktCh,
		rx: rx,
	}
	connection := insteon.NewDeviceConnection(insteon.DeviceTimeout, dst, bridge)
	p.connectionCh <- connectionInfo{dst, rx}
	return insteon.DeviceFactory(connection, dst)
}
