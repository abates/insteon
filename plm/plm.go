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
	reader PacketReader
	writer io.Writer

	linkdb
	timeout    time.Duration
	retries    int
	writeDelay time.Duration
	nextWrite  time.Time
	lastRead   time.Time

	connOptions []insteon.ConnectionOption
	plmCh       chan *Packet
	messages    chan *insteon.Message
}

// The Option mechanism is based on the method described at https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type Option func(p *PLM) error

// New creates a new PLM instance.
func New(reader io.Reader, writer io.Writer, timeout time.Duration, options ...Option) (plm *PLM, err error) {
	plm = &PLM{
		reader:     NewPacketReader(reader),
		writer:     LogWriter{writer},
		timeout:    timeout,
		retries:    3,
		writeDelay: 0,

		plmCh:    make(chan *Packet, 1),
		messages: make(chan *insteon.Message),
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

	go plm.readLoop()
	return plm, nil
}

// WriteDelay can be passed as a parameter to New to change the delay used after writing a command before reading the response.
func WriteDelay(d time.Duration) Option {
	return func(p *PLM) error {
		p.writeDelay = d
		return nil
	}
}

func ConnectionOptions(options ...insteon.ConnectionOption) Option {
	return func(p *PLM) error {
		p.connOptions = options
		return nil
	}
}

func (plm *PLM) readLoop() {
	for {
		packet, err := plm.reader.ReadPacket()
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

	insteon.Log.Infof("PLM Read Done")
}

// WriteMessage to the network
func (plm *PLM) WriteMessage(msg *insteon.Message) error {
	buf, err := msg.MarshalBinary()
	if err == nil {
		// slice off the source address since the PLM doesn't want it
		buf = buf[3:]
		_, err = RetryWriter(plm, plm.retries, true).WritePacket(&Packet{Command: CmdSendInsteonMsg, Payload: buf})
	}

	return err
}

// WritePacket and wait for the PLM to ack that the packet was
// sent.  This is a blocking function. Only callers that have acquired
// the mutex should call this function
func (plm *PLM) WritePacket(txPacket *Packet) (ack *Packet, err error) {
	writeDelay := time.Duration(0)
	if txPacket.Command == CmdSendInsteonMsg {
		writeDelay = plm.writeDelay
		if writeDelay == 0 {
			// flags is the 4th byte in an insteon message and max ttl/hops is the
			// least significant 2 bits
			flags := insteon.Flags(txPacket.Payload[3])
			writeDelay = insteon.PropagationDelay(flags.TTL(), flags.Extended())
		}
	}

	buf, err := txPacket.MarshalBinary()
	if err == nil {
		if time.Now().Before(plm.nextWrite) {
			delay := plm.nextWrite.Sub(time.Now())
			insteon.Log.Tracef("Delaying write for %s", delay)
			time.Sleep(delay)
		}

		insteon.Log.Tracef("Sending packet %v (write delay %v)", txPacket, writeDelay)
		_, err = plm.writer.Write(buf)
		plm.nextWrite = time.Now().Add(writeDelay)

		if err == nil {
			insteon.Log.Debugf("PLM TX %v", txPacket)
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
	return insteon.Open(plm.Bus, dst)
}

func (plm *PLM) Monitor(cb func()) error {
	config, err := plm.Config()
	if err != nil {
		return err
	}

	config.SetMonitorMode()
	err = plm.SetConfig(config)
	if err != nil {
		return err
	}

	/*conn, err := plm.Demux.Dial(insteon.Wildcard)
	if err == nil {
		cb(conn)
		config.clearMonitorMode()
		err = plm.SetConfig(config)
	}*/
	return err
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
	info, err := plm.Info()
	if err == nil {
		return info.Address
	}
	return insteon.Address([3]byte{})
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
