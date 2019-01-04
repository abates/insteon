// Copyright 2018 Andrew Bates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	readCh  chan []byte
	recvCh  chan []byte
	closeCh chan chan error
}

func NewPort(readWriter io.ReadWriter, timeout time.Duration) *Port {
	port := &Port{
		in:      bufio.NewReader(readWriter),
		out:     readWriter,
		timeout: timeout,

		sendCh:  make(chan []byte, 1),
		readCh:  make(chan []byte, 1),
		recvCh:  make(chan []byte, 1),
		closeCh: make(chan chan error),
	}
	go port.readLoop()
	go port.process()
	return port
}

func (port *Port) readLoop() {
	for {
		packet, err := port.readPacket()
		if err == nil {
			insteon.Log.Tracef("RX %s", hexDump("%02x", packet, " "))
			port.readCh <- packet
		} else {
			if err == io.EOF {
				close(port.readCh)
				break
			}
			insteon.Log.Infof("Error reading packet: %v", err)
		}
	}
	insteon.Log.Debugf("Port exiting read loop")
}

func (port *Port) process() {
	defer func() {
		close(port.recvCh)
		if closer, ok := port.out.(io.Closer); ok {
			err := closer.Close()
			if err != nil {
				insteon.Log.Infof("Failed to close io writer: %v", err)
			}
		}
		insteon.Log.Debugf("Port exiting process loop")
	}()

	for {
		select {
		case packet, open := <-port.readCh:
			if !open {
				return
			}
			port.recvCh <- packet
		case buf, open := <-port.sendCh:
			if !open {
				return
			}
			port.send(buf)
		case closeCh := <-port.closeCh:
			closeCh <- nil
			return
		}
	}
}

func (port *Port) send(buf []byte) {
	insteon.Log.Tracef("TX %s", hexDump("%02x", buf, " "))
	_, err := port.out.Write(buf)
	if err != nil {
		insteon.Log.Infof("Failed to write: %v", err)
	}
}

func (port *Port) readPacket() (buf []byte, err error) {
	timeout := time.Now().Add(port.timeout)

	// synchronize
	for err == nil {
		var b byte
		b, err = port.in.ReadByte()
		if err != nil {
			break
		}

		// first byte of PLM packets is always 0x02
		if b != 0x02 {
			insteon.Log.Tracef("Expected STX (0x02) got 0x%02x", b)
			continue
		} else {
			b, err = port.in.ReadByte()
			if packetLen, found := commandLens[Command(b)]; found {
				buf = append(buf, []byte{0x02, b}...)
				buf = append(buf, make([]byte, packetLen)...)
				insteon.Log.Tracef("Attempting to read %d more bytes", packetLen)
				_, err = io.ReadAtLeast(port.in, buf[2:], packetLen)
				insteon.Log.Tracef("Completed read: %s", hexDump("%02x", buf, " "))
				break
			} else {
				err = port.in.UnreadByte()
			}
		}

		// prevent infinite loop while trying to synchronize
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

func (port *Port) Close() error {
	close(port.sendCh)
	closeCh := make(chan error)
	port.closeCh <- closeCh
	return <-closeCh
}
