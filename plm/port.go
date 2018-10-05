package plm

import (
	"bufio"
	"io"
	"time"

	"github.com/abates/insteon"
)

type Port struct {
	in      *bufio.Reader
	out     io.Writer
	timeout time.Duration

	sendCh  chan []byte
	recvCh  chan []byte
	closeCh chan chan error
}

func NewPort(readWriter io.ReadWriter, timeout time.Duration) *Port {
	port := &Port{
		in:      bufio.NewReader(readWriter),
		out:     readWriter,
		timeout: timeout,

		sendCh:  make(chan []byte, 1),
		recvCh:  make(chan []byte, 1),
		closeCh: make(chan chan error),
	}
	go port.readLoop()
	go port.writeLoop()
	return port
}

func (port *Port) readLoop() {
	for {
		select {
		case closeCh := <-port.closeCh:
			closeCh <- nil
			return
		default:
			packet, err := port.readPacket()
			insteon.Log.Tracef("RX Packet %s", hexDump("%02x", packet, " "))
			if err == nil {
				port.recvCh <- packet
			} else {
				insteon.Log.Infof("Error reading packet: %v", err)
			}
		}
	}
}

func (port *Port) readPacket() (buf []byte, err error) {
	timeout := time.Now().Add(port.timeout)

	// synchronize
	for err == nil {
		var b byte
		b, err = port.in.ReadByte()
		if b != 0x02 {
			continue
		} else {
			b, err = port.in.ReadByte()
			if packetLen, found := commandLens[Command(b)]; found {
				buf = append(buf, []byte{0x02, b}...)
				buf = append(buf, make([]byte, packetLen)...)
				_, err = io.ReadAtLeast(port.in, buf[2:], packetLen)
				break
			} else {
				err = port.in.UnreadByte()
			}
		}
		// I don't remember why this is here...
		if time.Now().After(timeout) {
			err = insteon.ErrReadTimeout
			break
		}
	}

	if err == nil {
		// read some more if it's an extended message
		if buf[1] == 0x62 && insteon.Flags(buf[5]).Extended() {
			buf = append(buf, make([]byte, 14)...)
			_, err = io.ReadAtLeast(port.in, buf[9:], 14)
		}
	}
	return buf, err
}

func (port *Port) writeLoop() {
	for buf := range port.sendCh {
		_, err := port.out.Write(buf)
		if err == nil {
			insteon.Log.Tracef("TX %s", hexDump("%02x", buf, " "))
		} else {
			insteon.Log.Infof("Failed to write: %v", err)
		}
	}

	if closer, ok := port.out.(io.Closer); ok {
		err := closer.Close()
		if err != nil {
			insteon.Log.Infof("Failed to close io writer: %v", err)
		}
	}
	return
}

func (port *Port) Close() error {
	close(port.sendCh)
	closeCh := make(chan error)
	port.closeCh <- closeCh
	return <-closeCh
}
