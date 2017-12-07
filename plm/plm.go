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

type connectionInfo struct {
	address insteon.Address
	ch      chan *Packet
}

type txPacketInfo struct {
	packet *Packet
	ackCh  chan *Packet
}

type PLM struct {
	in      *bufio.Reader
	out     io.Writer
	timeout time.Duration

	txPktCh      chan *txPacketInfo
	rxPktCh      chan *Packet
	plmCh        chan *Packet
	connectionCh chan connectionInfo

	linkDb *PLMLinkDB
}

func New(port io.ReadWriter, timeout time.Duration) *PLM {
	plm := &PLM{
		in:      bufio.NewReader(port),
		out:     port,
		timeout: timeout,

		txPktCh:      make(chan *txPacketInfo, 1),
		rxPktCh:      make(chan *Packet, 1),
		plmCh:        make(chan *Packet, 1),
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

func tracePkt(prefix string, packet *Packet) {
	insteon.Log.Tracef("%-05s %s", prefix, packet)
}

func traceMsg(prefix string, msg *insteon.Message) {
	insteon.Log.Tracef("%-05s %s", prefix, msg)
}

func (plm *PLM) read(buf []byte) error {
	_, err := io.ReadAtLeast(plm.in, buf, len(buf))
	return err
}

func (plm *PLM) readPacket() (packet *Packet, err error) {
	var buf []byte
	b, err := plm.in.ReadByte()
	if err == nil && b != 0x02 {
		return nil, fmt.Errorf("Expected first byte to be 0x02 got 0x%02x", b)
	}
	buf = append(buf, b)

	b, err = plm.in.ReadByte()

	if err == nil {
		buf = append(buf, b)
		// TODO commandLens should only be written during
		// initialization, but, technically speaking, this
		// access could cause a concurrent access violation
		if packetLen, ok := commandLens[b]; ok {
			buf = append(buf, make([]byte, packetLen)...)
			_, err = io.ReadAtLeast(plm.in, buf[2:], packetLen)
			if err == nil {
				traceBuf("RX", buf)
				// read some more if it's an extended message
				if buf[1] == 0x62 && insteon.Flags(buf[5]).Extended() {
					buf = append(buf, make([]byte, 14)...)
					_, err = io.ReadAtLeast(plm.in, buf[9:], 14)
				}
				packet = &Packet{}
				err = packet.UnmarshalBinary(buf)
			}
		} else {
			err = fmt.Errorf("PLM Received unknown command 0x%02x", b)
		}
	}
	return packet, err
}

func (plm *PLM) readPktLoop() {
	for {
		packet, err := plm.readPacket()
		if err == nil {
			tracePkt("RX", packet)
			plm.rxPktCh <- packet
		} else {
			insteon.Log.Infof("Error reading packet: %v", err)
		}
	}
}

func (plm *PLM) writePacket(packet *Packet) error {
	payload, err := packet.MarshalBinary()
	traceBuf("TX", payload)

	if err == nil {
		_, err = plm.out.Write(payload)
	}
	return err
}

func (plm *PLM) readWriteLoop() {
	connections := make(map[insteon.Address]chan *Packet)
	ackChannels := make(map[Command]chan *Packet)
	for {
		var packet *Packet
		insteon.Log.Debugf("readWriteLoop wait...")
		select {
		case send := <-plm.txPktCh:
			ackChannels[send.packet.Command] = send.ackCh
			tracePkt("TX", send.packet)
			err := plm.writePacket(send.packet)
			if err != nil {
				insteon.Log.Infof("Failed to write packet: %v", err)
			}
		case packet = <-plm.rxPktCh:
			switch {
			case packet.Command == 0x50 || packet.Command == 0x51:
				msg := packet.Payload.(*insteon.Message)
				insteon.Log.Debugf("Received INSTEON Message %v", msg)
				if conn, ok := connections[msg.Src]; ok {
					insteon.Log.Debugf("Dispatching message to device connection")
					conn <- packet
				}
			case 0x52 <= packet.Command && packet.Command <= 0x58:
				// 0x52 to 0x58 are modem commands and should be dispatched
				// to functions communicating with the modem itself, however
				// we don't want to hold things up
				select {
				case plm.plmCh <- packet:
				default:
					insteon.Log.Infof("Received modem response, but no one was listening for it")
				}
			default:
				// handle ack/nak
				if ackCh, ok := ackChannels[packet.Command]; ok {
					select {
					case ackCh <- packet:
						close(ackCh)
						ackChannels[packet.Command] = nil
					default:
					}
				}
			}
		case info := <-plm.connectionCh:
			connections[info.address] = info.ch
		}
	}
}

func (plm *PLM) Receive() (packet *Packet, err error) {
	select {
	case packet = <-plm.plmCh:
		tracePkt("PLM Receive", packet)
	case <-time.After(plm.timeout):
		err = insteon.ErrAckTimeout
	}
	return packet, err
}

func (plm *PLM) Send(packet *Packet) (ack *Packet, err error) {
	tracePkt("PLM Send", packet)
	ackCh := make(chan *Packet, 1)
	txPktInfo := &txPacketInfo{
		packet: packet,
		ackCh:  ackCh,
	}

	select {
	case plm.txPktCh <- txPktInfo:
		select {
		case ack = <-ackCh:
			insteon.Log.Debugf("PLM ACK Received")
		case <-time.After(plm.timeout):
			err = insteon.ErrAckTimeout
		}
	case <-time.After(plm.timeout):
		err = insteon.ErrWriteTimeout
	}
	return
}

func (plm *PLM) Info() (*IMInfo, error) {
	ack, err := plm.Send(&Packet{
		Command: CmdGetInfo,
	})
	return ack.Payload.(*IMInfo), err
}

func (plm *PLM) Reset() error {
	return ErrNotImplemented
}

func (plm *PLM) Config() (Config, error) {
	return Config(0x00), ErrNotImplemented
}

func (plm *PLM) SetConfig(Config) error {
	return ErrNotImplemented
}

func (plm *PLM) SetDeviceCategory(insteon.Category) error {
	return ErrNotImplemented
}

func (plm *PLM) RFSleep() error {
	return ErrNotImplemented
}

type plmBridge struct {
	plm *PLM
	rx  chan *Packet
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
	case packet := <-pb.rx:
		payload = packet.Payload
	case <-time.After(pb.plm.timeout):
		err = insteon.ErrReadTimeout
	}
	return
}

func (plm *PLM) Dial(dst insteon.Address) insteon.Bridge {
	rx := make(chan *Packet, 1)
	bridge := &plmBridge{
		plm: plm,
		rx:  rx,
	}
	plm.connectionCh <- connectionInfo{dst, rx}
	return bridge
}

func (plm *PLM) Connect(dst insteon.Address) (insteon.Device, error) {
	bridge := plm.Dial(dst)
	device := insteon.Device(insteon.NewI1Device(dst, bridge))
	version, err := device.EngineVersion()

	// ErrNotLinked here is only returned by i2cs devices
	if err == insteon.ErrNotLinked {
		err = nil
		device = insteon.NewI2CsDevice(dst, bridge)
	} else {
		switch version {
		case insteon.VerI2:
			device = insteon.NewI2Device(dst, bridge)
		case insteon.VerI2Cs:
			device = insteon.NewI2CsDevice(dst, bridge)
		}
	}
	return device, err
}

func (plm *PLM) ConnectAndInitialize(dst insteon.Address) (insteon.Device, error) {
	device, err := plm.Connect(dst)
	if err == nil {
		device, err = insteon.InitializeDevice(device)
	}
	return device, err
}

func (plm *PLM) LinkDB() (ldb insteon.LinkDB, err error) {
	if plm.linkDb == nil {
		plm.linkDb = &PLMLinkDB{plm: plm}
		err = plm.linkDb.Refresh()
	}
	return plm.linkDb, err
}

func (plm *PLM) AssignToAllLinkGroup(insteon.Group) error   { return ErrNotImplemented }
func (plm *PLM) DeleteFromAllLinkGroup(insteon.Group) error { return ErrNotImplemented }

type AllLinkReq struct {
	Flags byte
	Group insteon.Group
}

func (alr *AllLinkReq) MarshalBinary() ([]byte, error) {
	return []byte{alr.Flags, byte(alr.Group)}, nil
}

func (alr *AllLinkReq) UnmarshalBinary(buf []byte) error {
	if len(buf) < 2 {
		return fmt.Errorf("Needed 2 bytes to unmarshal all link request.  Got %d", len(buf))
	}
	alr.Flags = buf[0]
	alr.Group = insteon.Group(buf[1])
	return nil
}

func (alr *AllLinkReq) String() string {
	return fmt.Sprintf("%02x %d", alr.Flags, alr.Group)
}

func (plm *PLM) EnterLinkingMode(group insteon.Group) error {
	ack, err := plm.Send(&Packet{
		Command: CmdStartAllLink,
		Payload: &AllLinkReq{Flags: 0x01, Group: group},
	})

	if ack.NAK() {
		err = insteon.ErrNak
	}
	return err
}

func (plm *PLM) CancelLinkingMode() error {
	ack, err := plm.Send(&Packet{
		Command: CmdCancelAllLink,
	})

	if ack.NAK() {
		err = insteon.ErrNak
	}
	return err
}

func (plm *PLM) EnterUnlinkingMode(group insteon.Group) error {
	ack, err := plm.Send(&Packet{
		Command: CmdStartAllLink,
		Payload: &AllLinkReq{Flags: 0xff, Group: group},
	})

	if ack.NAK() {
		err = insteon.ErrNak
	}
	return err
}

func (plm *PLM) Address() insteon.Address {
	info, err := plm.Info()
	if err == nil {
		return info.Address
	}
	return insteon.Address([3]byte{})
}
