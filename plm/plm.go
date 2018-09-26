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
	ErrAckTimeout     = errors.New("Timeout waiting for Ack from the PLM")
	ErrNak            = errors.New("PLM responded with a NAK.  Resend command")

	MaxRetries = 3
)

const (
	writeDelay = 500 * time.Millisecond
)

func hexDump(buf []byte) string {
	str := make([]string, len(buf))
	for i, b := range buf {
		str[i] = fmt.Sprintf("%02x", b)
	}
	return strings.Join(str, " ")
}

type txPacketReq struct {
	packet *Packet
	ackCh  chan *Packet
}

type PLM struct {
	in      *bufio.Reader
	out     io.Writer
	timeout time.Duration

	txPktCh  chan *txPacketReq
	rxPktCh  chan []byte
	closeCh  chan chan error
	listenCh chan *connection
}

func New(port io.ReadWriter, timeout time.Duration) *PLM {
	plm := &PLM{
		in:      bufio.NewReader(port),
		out:     port,
		timeout: timeout,

		txPktCh:  make(chan *txPacketReq, 1),
		rxPktCh:  make(chan []byte, 10),
		listenCh: make(chan *connection),
		closeCh:  make(chan chan error),
	}
	go plm.readPktLoop()
	go plm.readWriteLoop()
	return plm
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
		insteon.Log.Tracef("RX Packet %s", hexDump(packet))
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
		insteon.Log.Tracef("TX %s", hexDump(payload))
	}
	return err
}

func (plm *PLM) readWriteLoop() {
	connections := make([]*connection, 0)
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
				for _, connection := range connections {
					connection.notify(packet)
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
		case connection := <-plm.listenCh:
			connections = append(connections, connection)
		case closeCh = <-plm.closeCh:
			break loop
		}
	}

	for _, ch := range ackChannels {
		close(ch)
	}

	closeCh <- nil
}

// Retry will deliver a packet to the Insteon network. If delivery fails (due
// to a NAK from the PLM) then we will retry and decrement retries. This
// continues until the packet is sent (as acknowledged by the PLM) or retries
// reaches zero
func (plm *PLM) Retry(packet *Packet, retries int) (ack *Packet, err error) {
	for retries := 3; retries > 0; retries-- {
		ack, err = plm.send(packet)
		if ack.NAK() {
			insteon.Log.Debugf("PLM NAK received, resending packet")
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	return ack, err
}

// SendMessage will create a packet with the payload set
// to the contents of the message and then attempt to
// deliver the packet to the network.  Delivery will
// be retried according to the MaxRetries variable
func (plm *PLM) SendMessage(msg *insteon.Message) (err error) {
	buf, err := msg.MarshalBinary()
	if err == nil {
		// PLM expects that the payload begins with the
		// destinations address so we have to slice off
		// the src address
		buf = buf[3:]
		packet := &Packet{
			Command: CmdSendInsteonMsg,
			payload: buf,
		}
		_, err = plm.Retry(packet, MaxRetries)
	}
	return err
}

func (plm *PLM) Listen(command Command) Connection {
	connection := newConnection(plm, plm.timeout, command)
	plm.listenCh <- connection
	return connection
}

/*func (plm *PLM) ListenInsteon() insteon.Bridge {
	return newInsteonBridge(plm)
}*/

func (plm *PLM) send(packet *Packet) (ack *Packet, err error) {
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
	ack, err := plm.send(&Packet{
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

	ack, err := plm.send(&Packet{
		Command: CmdReset,
	})

	if err == nil && ack.NAK() {
		err = ErrNak
	}
	plm.timeout = timeout
	return err
}

func (plm *PLM) Config() (*Config, error) {
	ack, err := plm.send(&Packet{
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
	ack, err := plm.send(&Packet{
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

func (plm *PLM) Close() error {
	insteon.Log.Debugf("Closing PLM")
	errCh := make(chan error)
	plm.closeCh <- errCh
	err := <-errCh
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
