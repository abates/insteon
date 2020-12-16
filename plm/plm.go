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
	"errors"
	"fmt"
	"io"
	"strings"
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
	address insteon.Address
	reader  PacketReader
	writer  io.Writer

	linkdb
	timeout    time.Duration
	retries    int
	writeDelay time.Duration
	lastWrite  time.Time
	lastRead   time.Time

	connOptions []insteon.ConnectionOption
	plmCh       chan *Packet
	messages    chan *insteon.Message
	db          db.Database
}

// New creates a new PLM instance.
func New(reader io.Reader, writer io.Writer, timeout time.Duration, options ...Option) (plm *PLM, err error) {
	plm = &PLM{
		reader:     NewPacketReader(reader),
		writer:     LogWriter{writer, insteon.Log},
		timeout:    timeout,
		retries:    3,
		lastRead:   time.Now(),
		lastWrite:  time.Now(),
		writeDelay: 0,

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

	insteon.Log.Debugf("Staring PLM with config:")
	insteon.Log.Debugf("                Timeout: %s", plm.timeout)
	insteon.Log.Debugf("                Retries: %d", plm.writeDelay)
	insteon.Log.Debugf("             WriteDelay: %s", plm.writeDelay)
	go plm.readLoop()

	address := plm.Address()
	insteon.Log.Debugf("         PLM Address is: %v", address)

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

func writeDelay(pkt *Packet, last time.Time, maxDelay time.Duration) (delay time.Duration) {
	if pkt.Command == CmdSendInsteonMsg {
		if maxDelay > 0 {
			// flags is the 4th byte in an insteon message and max ttl/hops is the
			// least significant 2 bits
			//flags := insteon.Flags(pkt.Payload[3])
			//delay = insteon.PropagationDelay(flags.TTL(), flags.Extended())
			//} else {
			delay = maxDelay
		}
		delay = time.Now().Sub(last.Add(delay))
	}
	return delay
}

// WritePacket and wait for the PLM to ack that the packet was
// sent.  This is a blocking function. Only callers that have acquired
// the mutex should call this function
func (plm *PLM) WritePacket(txPacket *Packet) (ack *Packet, err error) {
	buf, err := txPacket.MarshalBinary()
	if err == nil {
		time.Sleep(writeDelay(txPacket, plm.lastRead, plm.writeDelay))

		insteon.Log.Tracef("Sending packet %v (write delay %v)", txPacket, writeDelay)
		if txPacket.Command != CmdSendInsteonMsg {
			insteon.Log.Debugf("PLM CMD TX %v", txPacket)
		}
		_, err = plm.writer.Write(buf)
		plm.lastWrite = time.Now()

		if err == nil {
			insteon.Log.Tracef("PLM TX %v", txPacket)
			ack, err = plm.ReadPacket()
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
