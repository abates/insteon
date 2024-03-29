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
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/abates/insteon"
	"github.com/abates/insteon/devices"
)

var (
	ErrReadTimeout        = errors.New("Timeout reading from plm")
	ErrNoSync             = errors.New("No sync byte received")
	ErrNotImplemented     = errors.New("IM command not implemented")
	ErrAckTimeout         = errors.New("Timeout waiting for Ack from the PLM")
	ErrRetryCountExceeded = errors.New("Retry count exceeded sending command")
	ErrNoAck              = errors.New("Received non-ack packet after transmit")
	ErrWrongAck           = errors.New("Command in ACK does not match TX packet")
	ErrWrongPayload       = errors.New("Payload in ACK does not match TX packet")
	ErrNak                = errors.New("PLM responded with a NAK.  Resend command")
)

var (
	Log      = log.New(os.Stderr, "", log.LstdFlags)
	LogDebug = log.New(ioutil.Discard, "DEBUG ", log.LstdFlags)
)

func hexDump(format string, buf []byte, sep string) string {
	str := make([]string, len(buf))
	for i, b := range buf {
		str[i] = fmt.Sprintf(format, b)
	}
	return strings.Join(str, sep)
}

type PLM struct {
	writer  io.Writer
	reader  *packetReader
	address insteon.Address

	linkdb
	timeout    time.Duration
	retries    int
	writeDelay time.Duration

	msgBuf    chan *Packet
	packetBuf chan *Packet
}

// New creates a new PLM instance.
func New(rw io.ReadWriter, options ...Option) (plm *PLM) {
	plm = &PLM{
		writer: logWriter{rw},
		reader: newPacketReader(rw, false),

		timeout:   time.Second * 3,
		retries:   3,
		msgBuf:    make(chan *Packet, 10),
		packetBuf: make(chan *Packet, 10),
	}

	for _, o := range options {
		o(plm)
	}

	plm.linkdb.plm = plm
	plm.linkdb.retries = plm.retries
	plm.linkdb.timeout = plm.timeout
	go plm.readLoop()
	return plm
}

func (plm *PLM) readLoop() {
	for {
		pkt, err := plm.reader.ReadPacket()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				Log.Printf("Read error: %v", err)
			}
			close(plm.msgBuf)
			close(plm.packetBuf)
			return
		}

		if pkt.Command == CmdStdMsgReceived || pkt.Command == CmdExtMsgReceived {
			select {
			case plm.msgBuf <- pkt:
			default:
				Log.Printf("PLM Packet dropped, no one listening")
			}
		} else {
			select {
			case plm.packetBuf <- pkt:
			default:
				Log.Printf("PLM Packet dropped, no one listening")
			}
		}
	}
}

// Write insteon message
func (plm *PLM) Write(msg *insteon.Message) (ack *insteon.Message, err error) {
	buf, err := msg.MarshalBinary()
	if err == nil {
		LogDebug.Printf("TX Message %v", msg)
		// slice off the source address since the PLM doesn't want it
		buf = buf[3:]
		_, err = retry(plm, plm.retries, true).WritePacket(&Packet{Command: CmdSendInsteonMsg, Payload: buf})

		if err == nil {
			// get the ACK
			ack, err = plm.Read()
			for ; err == nil; ack, err = plm.Read() {
				if ack.Src == msg.Dst && (ack.Ack() || ack.Nak()) {
					if ack.Nak() {
						err = devices.ErrNak
					}
					break
				}
			}
		}
	}

	return ack, err
}

func (plm *PLM) WritePacket(pkt *Packet) (ack *Packet, err error) {
	buf, err := pkt.MarshalBinary()
	if err == nil {
		LogDebug.Printf("TX Packet %v", pkt)
		_, err = plm.writer.Write(buf)

		if err == nil {
			ack, err = plm.ReadPacket()
			if err == nil {
				// these things happen rarely, but we can (a least in the
				// case of ErrWrongAck) usually do something about it
				if !ack.ACK() && !ack.NAK() {
					err = ErrNoAck
				} else if ack.Command != pkt.Command {
					err = ErrWrongAck
				} else if ack.Command != CmdGetInfo && ack.Command != CmdGetConfig {
					payload := ack.Payload
					if ack.Command == CmdSendInsteonMsg {
						payload = payload[3:]
					}
					if !bytes.Equal(payload, pkt.Payload) {
						err = ErrWrongPayload
					}
				}

				if ack.NAK() {
					err = ErrNak
				}
			}
		}
	}

	return
}

func (plm *PLM) ReadPacket() (pkt *Packet, err error) {
	select {
	case pkt = <-plm.packetBuf:
	case <-time.After(plm.timeout):
		err = fmt.Errorf("PLM ACK %w", ErrReadTimeout)
	}
	return
}

func (plm *PLM) Read() (msg *insteon.Message, err error) {
	select {
	case pkt := <-plm.msgBuf:
		msg = &insteon.Message{}
		err = msg.UnmarshalBinary(pkt.Payload)
		if err == nil {
			LogDebug.Printf("RX Insteon Message %v", msg)
		}
	case <-time.After(plm.timeout):
		err = fmt.Errorf("Device ACK %w", insteon.ErrReadTimeout)
	}

	return msg, err
}

func (plm *PLM) Info() (info *Info, err error) {
	ack, err := retry(plm, plm.retries, true).WritePacket(&Packet{Command: CmdGetInfo})
	if err == nil {
		info = &Info{}
		err = info.UnmarshalBinary(ack.Payload)
	}
	return info, err
}

// Reset will factory reset and erase all data from the PLM. ⚠️ Use with care.
func (plm *PLM) Reset() error {
	timeout := plm.timeout
	plm.timeout = 20 * time.Second

	_, err := retry(plm, plm.retries, true).WritePacket(&Packet{Command: CmdReset})
	plm.timeout = timeout
	return err
}

func (plm *PLM) Config() (config Config, err error) {
	ack, err := retry(plm, plm.retries, true).WritePacket(&Packet{Command: CmdGetConfig})
	if err == nil {
		err = config.UnmarshalBinary(ack.Payload)
	}
	return config, err
}

func (plm *PLM) SetConfig(config Config) error {
	payload, _ := config.MarshalBinary()
	_, err := retry(plm, plm.retries, true).WritePacket(&Packet{Command: CmdSetConfig, Payload: payload})
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

func (plm *PLM) Address() insteon.Address {
	address := insteon.Address(0)
	if plm.address == address {
		info, err := plm.Info()
		if err == nil {
			plm.address = info.Address
		}
	}
	return plm.address
}

func (plm *PLM) String() string {
	return fmt.Sprintf("PLM (%s)", plm.Address())
}
