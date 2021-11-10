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

type packetWriter interface {
	ReadPacket() (*Packet, error)
	WritePacket(*Packet) (ack *Packet, err error)
}

type delayWriter struct {
	packetWriter
	lastTTL  uint8
	lastLen  int
	minDelay time.Duration
	lastRead time.Time
	lastPkt  *Packet
}

func (dw *delayWriter) ReadPacket() (*Packet, error) {
	pkt, err := dw.packetWriter.ReadPacket()
	if err == nil {
		dw.lastRead = time.Now()
		dw.lastLen = len(pkt.Payload)
		if pkt.Command == CmdStdMsgReceived || pkt.Command == CmdExtMsgReceived {
			dw.lastTTL = uint8(pkt.Payload[6] & 0x03)
		} else if pkt.Command == CmdStartAllLink {
			dw.lastTTL = 3
		}
	}
	return pkt, err
}

func (dw *delayWriter) writeDelay(pkt *Packet) (delay time.Duration) {
	if dw.lastPkt == nil {
		return 0
	}

	if dw.lastPkt.Command == CmdSendInsteonMsg || dw.lastPkt.Command == CmdStartAllLink {
		delay = insteon.PropagationDelay(dw.lastTTL, dw.lastLen)
		delay = time.Now().Sub(dw.lastRead.Add(delay))
		delay = time.Now().Sub(dw.lastRead.Add(delay))
	}

	if delay < dw.minDelay {
		delay = dw.minDelay
	}
	return delay
}

func (dw *delayWriter) Write(pkt *Packet) (ack *Packet, err error) {
	delay := dw.writeDelay(pkt)
	LogDebug.Printf("Write delay %v)", delay)
	time.Sleep(delay)

	dw.lastPkt = pkt
	return dw.packetWriter.WritePacket(pkt)
}

type retryWriter struct {
	packetWriter
	retries   int
	ignoreNak bool
}

func (rw *retryWriter) WritePacket(packet *Packet) (ack *Packet, err error) {
	retries := rw.retries
	for 0 <= retries {
		ack, err = rw.packetWriter.WritePacket(packet)
		if (err == ErrNak && rw.ignoreNak) || err == ErrReadTimeout {
			// TODO add exponential backoff
			LogDebug.Printf("Got %v retrying", err)
			time.Sleep(time.Second)
			retries--
		} else {
			break
		}
	}
	return
}

func retry(writer packetWriter, retries int, ignoreNak bool) packetWriter {
	return &retryWriter{writer, retries, ignoreNak}
}

type logReader struct {
	io.Reader
}

func (lr logReader) Read(buf []byte) (n int, err error) {
	n, err = lr.Reader.Read(buf)
	if n > 0 {
		LogDebug.Printf("RX %s", hexDump("%02x", buf[0:n], " "))
	}
	return
}

type logWriter struct {
	io.Writer
}

// Write writes len(p) bytes from p to the underlying data stream.
// and logs what was written. Write will return the number of bytes
// written and any associated error
func (lw logWriter) Write(buf []byte) (int, error) {
	LogDebug.Printf("TX %s", hexDump("%02x", buf, " "))
	n, err := lw.Writer.Write(buf)
	if err != nil {
		Log.Printf("Failed to write: %v", err)
	}
	return n, err
}

// packetReader reads PLM packets from a given io.Reader
type packetReader struct {
	reader *bufio.Reader
	buf    [maxPaclen]byte

	// ignoreAck instructs the packet reader to forgoe looking
	// for ack bytes in packets.  This is *only* useful when
	// running the packet reader on the transmit side of a
	// snooped connection
	ignoreAck bool
}

// newPacketReader will create and initialize a packetReader for the
// given io.Reader
func newPacketReader(reader io.Reader, ignoreAck bool) *packetReader {
	return &packetReader{reader: bufio.NewReader(logReader{reader}), ignoreAck: ignoreAck}
}

// sync will advance the reader until a start of text character is seen
func (pr *packetReader) sync() (n int, paclen int, err error) {
	for err == nil {
		var b byte
		b, err = pr.reader.ReadByte()
		if err != nil {
			break
		}

		// first byte of PLM packets is always 0x02
		if b != 0x02 {
			LogDebug.Printf("(syncronizing) Expected Start of Text (0x02) got 0x%02x", b)
			continue
		} else {
			b, err = pr.reader.ReadByte()
			var found bool
			if paclen, found = commandLens[Command(b)]; found {
				pr.buf[0] = 0x02
				pr.buf[1] = b
				// length will be one less for packets originating from
				// the host since no ack will be on the end.  Thi is
				// to support snooping
				if pr.ignoreAck && 0x60 <= b && b <= 0x64 {
					paclen--
				}
				n = 2
				break
			} else {
				err = pr.reader.UnreadByte()
			}
		}
	}
	return
}

func (pr *packetReader) read() (int, error) {
	// synchronize
	n, paclen, err := pr.sync()
	if err != nil {
		return n, err
	}

	nn, err := io.ReadAtLeast(pr.reader, pr.buf[n:n+paclen], paclen)
	n += nn
	if err == nil {
		// read some more if it's an extended message (this conditional is
		// necessary because the PLM echos everything back to us and the
		// send insteon command does not distinguish between standard and
		// extended messages like the received command does
		if Command(pr.buf[1]) == CmdSendInsteonMsg && insteon.Flags(pr.buf[5]).Extended() {
			nn, err = io.ReadFull(pr.reader, pr.buf[n:n+14])
			n += nn
		}
	}
	return n, err
}

func (pr *packetReader) ReadPacket() (packet *Packet, err error) {
	n, err := pr.read()
	if err == nil {
		packet = &Packet{}
		err = packet.UnmarshalBinary(pr.buf[0:n])
		if err == nil {
			LogDebug.Printf("RX Packet %v", packet)
		}
	}
	return packet, err
}
