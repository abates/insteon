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
	ErrAckTimeout     = errors.New("Timeout waiting for Ack from the PLM")
	ErrNak            = errors.New("PLM responded with a NAK.  Resend command")
)

const (
	writeDelay = 500 * time.Millisecond
)

type pktSubReq struct {
	matches     [][]byte
	unsubscribe bool
	rxCh        <-chan *Packet
	respCh      chan<- bool
}

type pktSubscription struct {
	matches [][]byte
	ch      chan<- *Packet
}

func (sub *pktSubscription) match(buf []byte) bool {
	for _, match := range sub.matches {
		if bytes.Equal(match, buf[0:len(match)]) {
			return true
		}
	}
	return false
}

type txPacketReq struct {
	packet *Packet
	ackCh  chan *Packet
}

type PLM struct {
	in          *bufio.Reader
	out         io.Writer
	timeout     time.Duration
	txPktCh     chan *txPacketReq
	rxPktCh     chan []byte
	pktSubReqCh chan *pktSubReq
	closeCh     chan chan error

	linkDb   *LinkDB
	devCatDB map[insteon.Address]insteon.Category
}

func New(port io.ReadWriter, timeout time.Duration) *PLM {
	plm := &PLM{
		in:      bufio.NewReader(port),
		out:     port,
		timeout: timeout,

		txPktCh:     make(chan *txPacketReq, 1),
		rxPktCh:     make(chan []byte, 10),
		pktSubReqCh: make(chan *pktSubReq, 1),
		closeCh:     make(chan chan error),

		devCatDB: make(map[insteon.Address]insteon.Category),
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
	insteon.Log.Tracef("%s %s", prefix, strings.Join(bb, " "))
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
			insteon.Log.Debugf("delivered packet to read/write loop")
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
		insteon.Log.Tracef("TX %x", payload)
	}
	return err
}

func (plm *PLM) readWriteLoop() {
	pktSubscriptions := make(map[<-chan *Packet]*pktSubscription)
	ackChannels := make(map[Command]chan *Packet)
	var closeCh chan error

loop:
	for {
		select {
		case send := <-plm.txPktCh:
			ackChannels[send.packet.Command] = send.ackCh
			err := plm.writePacket(send.packet)
			if err != nil {
				insteon.Log.Infof("Failed to write packet: %v", err)
			}
		case buf := <-plm.rxPktCh:
			packet := &Packet{}
			err := packet.UnmarshalBinary(buf)

			if err != nil {
				insteon.Log.Infof("Failed to unmarshal packet: %v", err)
				continue
			}

			insteon.Log.Tracef("RX %v", packet)
			if 0x50 <= packet.Command && packet.Command <= 0x58 {
				for _, pktSubscription := range pktSubscriptions {
					// make sure to slice off the leading 0x02 from the
					// buffer
					if pktSubscription.match(buf[1:]) {
						select {
						case pktSubscription.ch <- packet:
						default:
							insteon.Log.Infof("PLM Subscription exists, but buffer is full. discarding %v", packet)
						}
					}
				}
			} else {
				// handle ack/nak
				insteon.Log.Debugf("Dispatching PLM ACK/NAK %v", packet)
				if ackCh, ok := ackChannels[packet.Command]; ok {
					select {
					case ackCh <- packet:
						close(ackCh)
						delete(ackChannels, packet.Command)
					default:
						insteon.Log.Debugf("PLM ACK/NAK channel was not ready, discarding %v", packet)
					}
				}
			}
		case pktSubReq := <-plm.pktSubReqCh:
			if pktSubReq.unsubscribe {
				if sub, found := pktSubscriptions[pktSubReq.rxCh]; found {
					delete(pktSubscriptions, pktSubReq.rxCh)
					close(sub.ch)
				}
			} else {
				ch := make(chan *Packet, 10)
				pktSubReq.rxCh = ch
				pktSubscriptions[pktSubReq.rxCh] = &pktSubscription{ch: ch, matches: pktSubReq.matches}
				pktSubReq.respCh <- true
				close(pktSubReq.respCh)
			}
		case closeCh = <-plm.closeCh:
			break loop
		}
	}

	for _, pktSubscription := range pktSubscriptions {
		close(pktSubscription.ch)
	}

	for _, ch := range ackChannels {
		close(ch)
	}

	closeCh <- nil
}

func (plm *PLM) Retry(packet *Packet, retries int) (ack *Packet, err error) {
	for retries := 3; retries > 0; retries-- {
		ack, err = plm.Send(packet)
		if ack.NAK() {
			insteon.Log.Debugf("PLM NAK received, resending packet")
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	return ack, err
}

func (plm *PLM) Send(packet *Packet) (ack *Packet, err error) {
	ackCh := make(chan *Packet, 1)
	txPktInfo := &txPacketReq{
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
			insteon.Log.Debugf("PLM ACK Timeout")
			err = ErrAckTimeout
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
		info := &Info{}
		err := info.UnmarshalBinary(ack.payload)
		return info, err
	}
	return nil, err
}

func (plm *PLM) Reset() error {
	timeout := plm.timeout
	plm.timeout = 20 * time.Second

	ack, err := plm.Send(&Packet{
		Command: CmdReset,
	})

	if err == nil && ack.NAK() {
		err = ErrNak
	}
	plm.timeout = timeout
	return err
}

func (plm *PLM) Monitor(callback func(buf []byte, msg *insteon.Message)) {
	ch := plm.Subscribe([]byte{0x50}, []byte{0x51})
	defer plm.Unsubscribe(ch)

	plm.StartMonitor()
	defer plm.StopMonitor()

	for pkt := range ch {
		msg := &insteon.Message{}
		msg.DevCat = plm.devCatDB[insteon.Address{pkt.payload[0], pkt.payload[1], pkt.payload[2]}]
		err := msg.UnmarshalBinary(pkt.payload)
		if err == nil {
			// slice off the packet header
			callback(pkt.payload, msg)
		}
	}
}

func (plm *PLM) StartMonitor() error {
	config, err := plm.Config()
	if err == nil {
		config.setMonitorMode()
		err = plm.SetConfig(config)
	}
	return err
}

func (plm *PLM) StopMonitor() error {
	config, err := plm.Config()
	if err == nil {
		config.clearMonitorMode()
		err = plm.SetConfig(config)
	}
	return err
}

func (plm *PLM) Config() (*Config, error) {
	ack, err := plm.Send(&Packet{
		Command: CmdGetConfig,
	})
	if err == nil && ack.NAK() {
		err = ErrNak
	} else if err == nil {
		var config Config
		err := config.UnmarshalBinary(ack.payload)
		return &config, err
	}
	return nil, err
}

func (plm *PLM) SetConfig(config *Config) error {
	payload, _ := config.MarshalBinary()
	ack, err := plm.Send(&Packet{
		Command: CmdSetConfig,
		payload: payload,
	})
	if err == nil && ack.NAK() {
		err = ErrNak
	}
	return err
}

func (plm *PLM) SetDeviceCategory(insteon.Category) error {
	// TODO
	return ErrNotImplemented
}

func (plm *PLM) RFSleep() error {
	// TODO
	return ErrNotImplemented
}

func (plm *PLM) Subscribe(matches ...[]byte) <-chan *Packet {
	respCh := make(chan bool)
	req := &pktSubReq{respCh: respCh, matches: matches}
	plm.pktSubReqCh <- req
	<-respCh
	return req.rxCh
}

func (plm *PLM) Unsubscribe(ch <-chan *Packet) {
	plm.pktSubReqCh <- &pktSubReq{rxCh: ch, unsubscribe: true}
}

func (plm *PLM) Dial(dst insteon.Address) (insteon.Device, error) {
	connection := NewConnection(plm, dst)
	i1Device := insteon.NewI1Device(dst, insteon.NewI1Connection(connection))
	device := insteon.Device(i1Device)

	version, err := device.EngineVersion()

	if err == nil {
		category, err := device.IDRequest()
		if err == nil {
			plm.devCatDB[dst] = category
		}
	}

	// ErrNotLinked here is only returned by i2cs devices
	if err == insteon.ErrNotLinked {
		insteon.Log.Debugf("Got ErrNotLinked, creating I2CS device")
		err = nil
		device = insteon.NewI2CsDevice(dst, connection)
	} else {
		switch version {
		case insteon.VerI2:
			insteon.Log.Debugf("Version 2 device detected")
			device = insteon.NewI2Device(dst, connection)
		case insteon.VerI2Cs:
			insteon.Log.Debugf("Version 2 CS device detected")
			device = insteon.NewI2CsDevice(dst, connection)
		}
	}
	return device, err
}

func (plm *PLM) Connect(dst insteon.Address) (insteon.Device, error) {
	device, err := plm.Dial(dst)
	if err == nil {
		device, err = insteon.Devices.Initialize(device)
	}
	return device, err
}

func (plm *PLM) LinkDB() (ldb insteon.LinkDB, err error) {
	if plm.linkDb == nil {
		plm.linkDb = NewLinkDB(plm)
	}
	return plm.linkDb, err
}

func (plm *PLM) AssignToAllLinkGroup(insteon.Group) error   { return ErrNotImplemented }
func (plm *PLM) DeleteFromAllLinkGroup(insteon.Group) error { return ErrNotImplemented }

func (plm *PLM) Close() error {
	insteon.Log.Debugf("Closing PLM")
	errCh := make(chan error)
	plm.closeCh <- errCh
	err := <-errCh
	return err
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
	lr := &AllLinkReq{Mode: LinkingMode(0x03), Group: group}
	payload, _ := lr.MarshalBinary()
	ack, err := plm.Retry(&Packet{
		Command: CmdStartAllLink,
		payload: payload,
	}, 3)

	if err == nil && ack.NAK() {
		err = ErrNak
	}
	return err
}

func (plm *PLM) ExitLinkingMode() error {
	ack, err := plm.Retry(&Packet{
		Command: CmdCancelAllLink,
	}, 3)

	if err == nil && ack.NAK() {
		err = ErrNak
	}
	return err
}

func (plm *PLM) EnterUnlinkingMode(group insteon.Group) error {
	lr := &AllLinkReq{Mode: LinkingMode(0xff), Group: group}
	payload, _ := lr.MarshalBinary()
	ack, err := plm.Retry(&Packet{
		Command: CmdStartAllLink,
		payload: payload,
	}, 3)

	if err == nil && ack.NAK() {
		err = ErrNak
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
