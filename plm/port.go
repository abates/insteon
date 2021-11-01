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
	"bytes"
	"io"
	"log"
	"time"

	"github.com/abates/insteon"
)

type PacketReader interface {
	ReadPacket() (*Packet, error)
}

type PacketWriter interface {
	PacketReader
	WritePacket(*Packet) (ack *Packet, err error)
}

type packetWriter struct {
	PacketReader
	writer io.Writer
}

func (w *packetWriter) WritePacket(txPacket *Packet) (ack *Packet, err error) {
	buf, err := txPacket.MarshalBinary()
	if err == nil {
		insteon.LogTrace.Printf("Sending packet %v", txPacket)
		if txPacket.Command != CmdSendInsteonMsg {
			insteon.LogDebug.Printf("CMD TX %v", txPacket)
		}
		_, err = w.writer.Write(buf)

		if err == nil {
			insteon.LogTrace.Printf("TX %v", txPacket)
			ack, err = w.ReadPacket()

			if err == nil {
				// these things happen rarely, but we can (a least in the
				// case of ErrWrongAck) usually do something about it
				if !ack.ACK() && !ack.NAK() {
					err = ErrNoAck
				} else if ack.Command != txPacket.Command {
					err = ErrWrongAck
				} else if ack.Command != CmdGetInfo && ack.Command != CmdGetConfig {
					payload := ack.Payload
					if ack.Command == CmdSendInsteonMsg {
						payload = payload[3:]
					}
					if !bytes.Equal(payload, txPacket.Payload) {
						err = ErrWrongPayload
					}
				}
			}
		}
	}
	if ack.NAK() {
		err = ErrNak
	}

	return
}

type delayWriter struct {
	PacketWriter
	lastTTL  uint8
	lastLen  int
	minDelay time.Duration
	lastRead time.Time
	lastPkt  *Packet
}

func (dw *delayWriter) ReadPacket() (*Packet, error) {
	pkt, err := dw.PacketWriter.ReadPacket()
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

	if pkt.Command == CmdSendInsteonMsg || dw.lastPkt.Command == CmdStartAllLink {
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
	insteon.LogTrace.Printf("Write delay %v)", delay)
	time.Sleep(delay)

	if pkt.Command != CmdSendInsteonMsg {
		insteon.LogDebug.Printf("PLM CMD TX %v", pkt)
	}
	dw.lastPkt = pkt
	return dw.PacketWriter.WritePacket(pkt)
}

type retryWriter struct {
	PacketWriter
	retries   int
	ignoreNak bool
}

func (rw *retryWriter) WritePacket(packet *Packet) (ack *Packet, err error) {
	retries := rw.retries
	for 0 <= retries {
		ack, err = rw.PacketWriter.WritePacket(packet)
		if (err == ErrNak && rw.ignoreNak) || err == ErrReadTimeout {
			// TODO add exponential backoff
			insteon.LogDebug.Printf("Got %v retrying", err)
			time.Sleep(time.Second)
			retries--
		} else {
			break
		}
	}
	return
}

func RetryWriter(writer PacketWriter, retries int, ignoreNak bool) PacketWriter {
	return &retryWriter{writer, retries, ignoreNak}
}

type logWriter struct {
	io.Writer
	Log      *log.Logger
	LogDebug *log.Logger
	LogTrace *log.Logger
}

// Write writes len(p) bytes from p to the underlying data stream.
// and logs what was written. Write will return the number of bytes
// written and any associated error
func (lw logWriter) Write(buf []byte) (int, error) {
	lw.LogTrace.Printf("TX %s", hexDump("%02x", buf, " "))
	n, err := lw.Writer.Write(buf)
	if err != nil {
		lw.Log.Printf("Failed to write: %v", err)
	}
	return n, err
}

// PacketReader reads PLM packets from a given io.Reader
type packetReader struct {
	reader *bufio.Reader
	buf    [maxPaclen]byte
}

// NewPacketReader will create and initialize a PacketReader for the
// given io.Reader
func NewPacketReader(reader io.Reader) PacketReader {
	return &packetReader{reader: bufio.NewReader(reader)}
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
			insteon.LogTrace.Printf("(syncronizing) Expected Start of Text (0x02) got 0x%02x", b)
			continue
		} else {
			b, err = pr.reader.ReadByte()
			var found bool
			if paclen, found = commandLens[Command(b)]; found {
				pr.buf[0] = 0x02
				pr.buf[1] = b
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

	nn, err := io.ReadAtLeast(pr.reader, pr.buf[2:2+paclen], paclen)
	n += nn

	if err == nil {
		// read some more if it's an extended message (this conditional is
		// necessary because the PLM echos everything back to us and the
		// send insteon command does not distinguish between standard and
		// extended messages like the received command does
		if Command(pr.buf[1]) == CmdSendInsteonMsg && insteon.Flags(pr.buf[5]).Extended() {
			nn, err = io.ReadAtLeast(pr.reader, pr.buf[n:], 14)
			n += nn
		}

		if err == nil {
			insteon.LogTrace.Printf("RX %s", hexDump("%02x", pr.buf[0:n], " "))
		}
	}

	return n, err
}

func (pr *packetReader) ReadPacket() (packet *Packet, err error) {
	n, err := pr.read()
	if err == nil {
		packet = &Packet{}
		err = packet.UnmarshalBinary(pr.buf[0:n])
	}
	return packet, err
}

func defaultPacketWriter(rw io.ReadWriter, minDelay time.Duration) PacketWriter {
	return &delayWriter{
		minDelay: minDelay,
		lastRead: time.Now(),
		PacketWriter: &packetWriter{
			PacketReader: NewPacketReader(rw),
			writer:       logWriter{Writer: rw, Log: insteon.Log, LogDebug: insteon.LogDebug, LogTrace: insteon.LogTrace},
		},
	}
}
