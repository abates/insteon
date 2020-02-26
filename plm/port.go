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

	"github.com/abates/insteon"
)

// LogWriter is a pass-through io.Writer that just logs what
// is written
type LogWriter struct {
	io.Writer
}

// Write writes len(p) bytes from p to the underlying data stream.
// and logs what was written. Write will return the number of bytes
// written and any associated error
func (lw LogWriter) Write(buf []byte) (int, error) {
	insteon.Log.Tracef("TX %s", hexDump("%02x", buf, " "))
	n, err := lw.Writer.Write(buf)
	if err != nil {
		insteon.Log.Infof("Failed to write: %v", err)
	}
	return n, err
}

// PacketReader reads PLM packets from a given io.Reader
type PacketReader struct {
	reader *bufio.Reader
	buf    [maxPaclen]byte
}

// NewPacketReader will create and initialize a PacketReader for the
// given io.Reader
func NewPacketReader(reader io.Reader) PacketReader {
	return PacketReader{reader: bufio.NewReader(reader)}
}

// sync will advance the reader until a start of text character is seen
func (pr *PacketReader) sync() (n int, paclen int, err error) {
	for err == nil {
		var b byte
		b, err = pr.reader.ReadByte()
		if err != nil {
			break
		}

		// first byte of PLM packets is always 0x02
		if b != 0x02 {
			insteon.Log.Tracef("(syncronizing) Expected Start of Text (0x02) got 0x%02x", b)
			continue
		} else {
			b, err = pr.reader.ReadByte()
			var found bool
			if paclen, found = commandLens[Command(b)]; found {
				pr.buf[0] = 0x02
				pr.buf[1] = b
				n = 2
				insteon.Log.Tracef("Successfully synchronized with input stream")
				break
			} else {
				err = pr.reader.UnreadByte()
			}
		}
	}
	return
}

func (pr *PacketReader) read() (int, error) {
	// synchronize
	n, paclen, err := pr.sync()
	if err != nil {
		return n, err
	}

	insteon.Log.Tracef("Attempting to read %d more bytes", paclen)
	nn, err := io.ReadAtLeast(pr.reader, pr.buf[2:2+paclen], paclen)
	n += nn
	insteon.Log.Tracef("Completed read (err %v): %d %s", err, n, hexDump("%02x", pr.buf[0:n], " "))

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
			insteon.Log.Tracef("RX %s", hexDump("%02x", pr.buf[0:n], " "))
		}
	}

	return n, err
}

func (pr *PacketReader) ReadPacket() (packet *Packet, err error) {
	n, err := pr.read()
	if err == nil {
		packet = &Packet{}
		err = packet.UnmarshalBinary(pr.buf[0:n])
	}
	return packet, err
}
