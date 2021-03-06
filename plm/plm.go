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
	"strings"
	"sync"
	"time"

	"github.com/abates/insteon"
	"github.com/abates/insteon/db"
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

func hexDump(format string, buf []byte, sep string) string {
	str := make([]string, len(buf))
	for i, b := range buf {
		str[i] = fmt.Sprintf(format, b)
	}
	return strings.Join(str, sep)
}

type PLM struct {
	insteon.Bus
	address   insteon.Address
	reader    PacketReader
	writer    io.Writer
	writeLock sync.Mutex

	linkdb
	timeout          time.Duration
	retries          int
	writeDelay       time.Duration
	lastRead         time.Time
	lastInsteonRead  time.Time
	lastInsteonFlags insteon.Flags

	connOptions []insteon.ConnectionOption
	plmCh       chan *Packet
	messages    chan *insteon.Message
	db          db.Database
}

// New creates a new PLM instance.
func New(reader io.Reader, writer io.Writer, timeout time.Duration, options ...Option) (plm *PLM, err error) {
	plm = &PLM{
		reader:          NewPacketReader(reader),
		writer:          LogWriter{writer, insteon.Log},
		timeout:         timeout,
		retries:         3,
		lastRead:        time.Now(),
		lastInsteonRead: time.Now(),

		plmCh:    make(chan *Packet, 1),
		messages: make(chan *insteon.Message, 10),
		db:       db.NewMemDB(),
	}

	for _, o := range options {
		err := o(plm)
		if err != nil {
			insteon.Log.Infof("error setting plm option: %v", err)
			return nil, err
		}
	}

	plm.linkdb.plm = plm
	plm.linkdb.retries = plm.retries
	plm.linkdb.timeout = plm.timeout

	if plm.Bus == nil {
		plm.Bus, err = insteon.NewBus(plm, plm.messages, plm.connOptions...)
		if err != nil {
			return nil, err
		}
	}

	if insteon.Log.Level >= insteon.LevelDebug {
		insteon.Log.Debugf("Staring PLM with config:")
		insteon.Log.Debugf("                Timeout: %s", plm.timeout)
		insteon.Log.Debugf("                Retries: %d", plm.retries)
		insteon.Log.Debugf("             WriteDelay: %s", plm.writeDelay)
	}
	go plm.readLoop()

	if insteon.Log.Level >= insteon.LevelDebug {
		address := plm.Address()
		insteon.Log.Debugf("         PLM Address is: %v", address)
	}

	return plm, nil
}

func (plm *PLM) readLoop() {
	for {
		packet, err := plm.reader.ReadPacket()
		plm.lastRead = time.Now()
		if err == nil {
			insteon.Log.Tracef("%v", packet)
			if packet.Command == CmdStdMsgReceived || packet.Command == CmdExtMsgReceived {
				msg := &insteon.Message{}
				err = msg.UnmarshalBinary(packet.Payload)
				if err == nil {
					plm.writeLock.Lock()
					plm.lastInsteonRead = time.Now()
					plm.lastInsteonFlags = msg.Flags
					plm.writeLock.Unlock()
					insteon.Log.Debugf("PLM Insteon RX %v", msg)
					plm.messages <- msg
				}
			} else {
				insteon.Log.Debugf("PLM CMD RX %v", packet)
				if packet.Command == CmdAllLinkComplete {
					continue
				}
				select {
				case plm.plmCh <- packet:
				default:
					insteon.Log.Infof("Nothing listening on PLM channel")
				}
			}
		} else {
			if err != io.EOF {
				insteon.Log.Infof("PLM Read Error: %v", err)
			}
			break
		}
	}
}

// WriteMessage to the network
func (plm *PLM) WriteMessage(msg *insteon.Message) error {
	buf, err := msg.MarshalBinary()
	if err == nil {
		insteon.Log.Debugf("PLM Insteon TX %v", msg)
		// slice off the source address since the PLM doesn't want it
		buf = buf[3:]
		_, err = RetryWriter(plm, plm.retries, true).WritePacket(&Packet{Command: CmdSendInsteonMsg, Payload: buf})
	}

	return err
}

//func writeDelay(pkt *Packet, last time.Time, maxDelay time.Duration) (delay time.Duration) {
func writeDelay(pkt *Packet, lastFlags insteon.Flags, minDelay time.Duration, lastRead time.Time) (delay time.Duration) {
	if pkt.Command == CmdSendInsteonMsg {
		delay = insteon.PropagationDelay(lastFlags.TTL(), lastFlags.Extended())
		delay = time.Now().Sub(lastRead.Add(delay))
		delay = time.Now().Sub(lastRead.Add(delay))
	}
	if delay < minDelay {
		delay = minDelay
	}
	return delay
}

// WritePacket and wait for the PLM to ack that the packet was
// sent.  This is a blocking function. Only callers that have acquired
// the mutex should call this function
func (plm *PLM) WritePacket(txPacket *Packet) (ack *Packet, err error) {
	buf, err := txPacket.MarshalBinary()
	if err == nil {
		plm.writeLock.Lock()
		time.Sleep(writeDelay(txPacket, plm.lastInsteonFlags, plm.writeDelay, plm.lastInsteonRead))

		insteon.Log.Tracef("Sending packet %v (write delay %v)", txPacket, writeDelay)
		if txPacket.Command != CmdSendInsteonMsg {
			insteon.Log.Debugf("PLM CMD TX %v", txPacket)
		}
		_, err = plm.writer.Write(buf)
		plm.writeLock.Unlock()

		if err == nil {
			insteon.Log.Tracef("PLM TX %v", txPacket)
			ack, err = plm.ReadPacket()

			if err == nil {
				// these things happen rarely, but we can (a least in the
				// case of ErrWrongAck) usually do something about it
				if !ack.ACK() {
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
	return
}

func (plm *PLM) ReadPacket() (pkt *Packet, err error) {
	select {
	case pkt = <-plm.plmCh:
	case <-time.After(plm.timeout):
		return nil, ErrReadTimeout
	}

	if pkt.NAK() {
		err = ErrNak
	}
	return pkt, err
}

func (plm *PLM) Open(dst insteon.Address) (insteon.Device, error) {
	return plm.db.Open(plm.Bus, dst)
}

func (plm *PLM) Info() (info *Info, err error) {
	ack, err := RetryWriter(plm, plm.retries, true).WritePacket(&Packet{Command: CmdGetInfo})
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

	_, err := RetryWriter(plm, plm.retries, true).WritePacket(&Packet{Command: CmdReset})
	plm.timeout = timeout
	return err
}

func (plm *PLM) Config() (config Config, err error) {
	ack, err := RetryWriter(plm, plm.retries, true).WritePacket(&Packet{Command: CmdGetConfig})
	if err == nil {
		err = config.UnmarshalBinary(ack.Payload)
	}
	return config, err
}

func (plm *PLM) SetConfig(config Config) error {
	payload, _ := config.MarshalBinary()
	_, err := RetryWriter(plm, plm.retries, true).WritePacket(&Packet{Command: CmdSetConfig, Payload: payload})
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
	address := insteon.Address{}
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

func (plm *PLM) Close() {
	//close(plm.insteonTxCh)
	if closer, ok := plm.writer.(io.Closer); ok {
		closer.Close()
	}
}

func (plm *PLM) LinkDatabase() (insteon.Linkable, error) {
	return plm, nil
}
