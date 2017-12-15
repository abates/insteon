package plm

import (
	"bufio"
	"bytes"
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

type subscription struct {
	matches     [][]byte
	unsubscribe bool
	ch          chan *Packet
}

func (s *subscription) match(buf []byte) bool {
	for _, match := range s.matches {
		if bytes.Equal(match, buf[0:len(match)]) {
			return true
		}
	}
	return false
}

func (s *subscription) dispatch(pkt *Packet) {
	s.ch <- pkt
}

func (s *subscription) Close() error {
	close(s.ch)
	return nil
}

type txPacketInfo struct {
	packet *Packet
	ackCh  chan *Packet
}

type PLM struct {
	in      *bufio.Reader
	out     io.Writer
	timeout time.Duration

	txPktCh chan *txPacketInfo
	rxPktCh chan []byte
	subCh   chan *subscription
	closeCh chan chan error

	linkDb *LinkDB
}

func New(port io.ReadWriter, timeout time.Duration) *PLM {
	plm := &PLM{
		in:      bufio.NewReader(port),
		out:     port,
		timeout: timeout,

		txPktCh: make(chan *txPacketInfo, 1),
		rxPktCh: make(chan []byte, 1),
		subCh:   make(chan *subscription),
		closeCh: make(chan chan error),
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

func (plm *PLM) readPacket() (buf []byte, err error) {
	timeout := time.Now().Add(plm.timeout)

	// synchronize
	for err == nil {
		var b byte
		b, err = plm.in.ReadByte()
		if b != 0x02 {
			continue
		} else {
			b, err = plm.in.ReadByte()
			if packetLen, found := commandLens[b]; found {
				buf = append(buf, []byte{0x02, b}...)
				buf = append(buf, make([]byte, packetLen)...)
				_, err = io.ReadAtLeast(plm.in, buf[2:], packetLen)
				break
			} else {
				err = plm.in.UnreadByte()
			}
		}
		if time.Now().After(timeout) {
			err = insteon.ErrReadTimeout
			break
		}
	}

	if err == nil {
		// read some more if it's an extended message
		if buf[1] == 0x62 && insteon.Flags(buf[5]).Extended() {
			buf = append(buf, make([]byte, 14)...)
			_, err = io.ReadAtLeast(plm.in, buf[9:], 14)
		}
	}
	return buf, err
}

func (plm *PLM) readPktLoop() {
	for {
		packet, err := plm.readPacket()
		if err == nil {
			plm.rxPktCh <- packet
		} else {
			insteon.Log.Infof("Error reading packet: %v", err)
		}
	}
}

func (plm *PLM) writePacket(packet *Packet) error {
	payload, err := packet.MarshalBinary()

	if err == nil {
		_, err = plm.out.Write(payload)
	}
	if err == nil {
		tracePkt("TX", packet)
	}
	return err
}

func (plm *PLM) readWriteLoop() {
	subscriptions := make(map[chan *Packet]*subscription)
	ackChannels := make(map[Command]chan *Packet)
	var closeCh chan error

loop:
	for {
		var packet *Packet
		select {
		case send := <-plm.txPktCh:
			ackChannels[send.packet.Command] = send.ackCh
			err := plm.writePacket(send.packet)
			if err != nil {
				insteon.Log.Infof("Failed to write packet: %v", err)
			}
		case buf := <-plm.rxPktCh:
			packet = &Packet{}
			err := packet.UnmarshalBinary(buf)

			if err != nil {
				insteon.Log.Infof("Failed to unmarshal packet: %v", err)
				continue
			}

			tracePkt("RX", packet)
			if 0x50 <= packet.Command && packet.Command <= 0x58 {
				insteon.Log.Debugf("Looking for subscriptions to %v", packet)
				for _, subscription := range subscriptions {
					// make sure to slice off the leading 0x02 from the
					// buffer
					if subscription.match(buf[1:]) {
						subscription.dispatch(packet)
					}
				}
			} else {
				// handle ack/nak
				if ackCh, ok := ackChannels[packet.Command]; ok {
					select {
					case ackCh <- packet:
						close(ackCh)
						delete(ackChannels, packet.Command)
					default:
					}
				}
			}
		case sub := <-plm.subCh:
			if sub.unsubscribe {
				for ch, _ := range subscriptions {
					if ch == sub.ch {
						delete(subscriptions, ch)
						close(ch)
					}
				}
			} else {
				subscriptions[sub.ch] = sub
			}
		case closeCh = <-plm.closeCh:
			break loop
		}
	}

	for ch, _ := range subscriptions {
		close(ch)
	}

	for _, ch := range ackChannels {
		close(ch)
	}

	closeCh <- nil
}

func (plm *PLM) Send(packet *Packet) (ack *Packet, err error) {
	ackCh := make(chan *Packet, 1)
	txPktInfo := &txPacketInfo{
		packet: packet,
		ackCh:  ackCh,
	}

	select {
	case plm.txPktCh <- txPktInfo:
		select {
		case ack = <-ackCh:
			if ack.NAK() {
				insteon.Log.Debugf("PLM NAK Received!")
			} else {
				insteon.Log.Debugf("PLM ACK Received")
			}
		case <-time.After(plm.timeout):
			err = insteon.ErrAckTimeout
		}
	case <-time.After(plm.timeout):
		err = insteon.ErrWriteTimeout
	}
	return
}

func (plm *PLM) Info() (*Info, error) {
	ack, err := plm.Send(&Packet{
		Command: CmdGetInfo,
	})
	if err == nil {
		return ack.Payload.(*Info), nil
	}
	return nil, err
}

func (plm *PLM) Reset() error {
	return ErrNotImplemented
}

func (plm *PLM) Config() (*Config, error) {
	ack, err := plm.Send(&Packet{
		Command: CmdGetConfig,
	})
	if ack.NAK() {
		err = insteon.ErrNak
	} else if err == nil {
		return ack.Payload.(*Config), nil
	}
	return nil, err
}

func (plm *PLM) SetConfig(config *Config) error {
	ack, err := plm.Send(&Packet{
		Command: CmdSetConfig,
	})
	if ack.NAK() {
		err = insteon.ErrNak
	}
	return err
}

func (plm *PLM) SetDeviceCategory(insteon.Category) error {
	return ErrNotImplemented
}

func (plm *PLM) RFSleep() error {
	return ErrNotImplemented
}

func (plm *PLM) Subscribe(ch chan *Packet, matches ...[]byte) {
	plm.subCh <- &subscription{ch: ch, matches: matches}
}

func (plm *PLM) Unsubscribe(ch chan *Packet) {
	plm.subCh <- &subscription{ch: ch, unsubscribe: true}
}

func (plm *PLM) Dial(dst insteon.Address) (insteon.Device, error) {
	bridge := NewBridge(plm, dst)
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

func (plm *PLM) Connect(dst insteon.Address) (insteon.Device, error) {
	device, err := plm.Dial(dst)
	if err == nil {
		device, err = insteon.InitializeDevice(device)
	}
	return device, err
}

func (plm *PLM) LinkDB() (ldb insteon.LinkDB, err error) {
	if plm.linkDb == nil {
		plm.linkDb = NewLinkDB(plm)
		err = plm.linkDb.Refresh()
	}
	return plm.linkDb, err
}

func (plm *PLM) AssignToAllLinkGroup(insteon.Group) error   { return ErrNotImplemented }
func (plm *PLM) DeleteFromAllLinkGroup(insteon.Group) error { return ErrNotImplemented }

func (plm *PLM) Close() error {
	errCh := make(chan error)
	plm.closeCh <- errCh
	return <-errCh
}

type LinkingMode byte

type AllLinkReq struct {
	Mode  LinkingMode
	Group insteon.Group
}

func (alr *AllLinkReq) MarshalBinary() ([]byte, error) {
	return []byte{byte(alr.Mode), byte(alr.Group)}, nil
}

func (alr *AllLinkReq) UnmarshalBinary(buf []byte) error {
	if len(buf) < 2 {
		return fmt.Errorf("Needed 2 bytes to unmarshal all link request.  Got %d", len(buf))
	}
	alr.Mode = LinkingMode(buf[0])
	alr.Group = insteon.Group(buf[1])
	return nil
}

func (alr *AllLinkReq) String() string {
	return fmt.Sprintf("%02x %d", alr.Mode, alr.Group)
}

func (plm *PLM) AddManualLink(group insteon.Group) error {
	return plm.EnterLinkingMode(group)
}

func (plm *PLM) EnterLinkingMode(group insteon.Group) error {
	ack, err := plm.Send(&Packet{
		Command: CmdStartAllLink,
		Payload: &AllLinkReq{Mode: LinkingMode(0x03), Group: group},
	})

	if ack.NAK() {
		err = insteon.ErrNak
	}
	return err
}

func (plm *PLM) ExitLinkingMode() error {
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
		Payload: &AllLinkReq{Mode: LinkingMode(0xff), Group: group},
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

func (plm *PLM) String() string {
	return fmt.Sprintf("PLM (%s)", plm.Address())
}
