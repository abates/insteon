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
	linkdb
	timeout    time.Duration
	retries    int
	writeDelay time.Duration
	nextWrite  time.Time
	port       *Port
	demux      insteon.Demux

	plmCh chan *Packet
}

// The Option mechanism is based on the method described at https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type Option func(p *PLM) error

// New creates a new PLM instance.
func New(port *Port, timeout time.Duration, options ...Option) (*PLM, error) {
	plm := &PLM{
		timeout:    timeout,
		retries:    3,
		writeDelay: 500 * time.Millisecond,
		port:       port,

		plmCh: make(chan *Packet),
	}
	plm.demux.Sender = plm
	plm.linkdb.plm = plm
	plm.linkdb.timeout = timeout

	for _, o := range options {
		err := o(plm)
		if err != nil {
			insteon.Log.Infof("error setting plm option: %v", err)
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

func (plm *PLM) readLoop() {
	for {
		buf, err := plm.port.Read()
		if err == nil {
			packet := &Packet{}
			err := packet.UnmarshalBinary(buf)

			if err == nil {
				insteon.Log.Tracef("%v", packet)
				if packet.Command == 0x50 || packet.Command == 0x51 {
					msg := &insteon.Message{}
					err := msg.UnmarshalBinary(packet.Payload)
					if err == nil {
						plm.demux.Dispatch(msg)
					} else {
						insteon.Log.Infof("Failed to unmarshal Insteon Message: %v", err)
					}
				} else {
					plm.plmCh <- packet
				}
			} else {
				insteon.Log.Infof("Failed to unmarshal packet: %v", err)
			}
		} else {
			if err != io.EOF {
				insteon.Log.Infof("Failed to read from PLM port: %v", err)
			}
			break
		}
	}
}

func (plm *PLM) Send(msg *insteon.Message) error {
	buf, err := msg.MarshalBinary()
	if err == nil {
		// slice off the source address since the PLM doesn't want it
		buf = buf[3:]
		retries := 0
		err = ErrRetryCountExceeded
		for retries < plm.retries {
			_, err = plm.send(&Packet{Command: CmdSendInsteonMsg, Payload: buf})
			if err == ErrNak || err == ErrReadTimeout {
				// TODO add exponential backoff
				retries++
			} else {
				break
			}
		}
	}

	return err
}

// send a packet and wait for the PLM to ack that the packet was
// sent.  This is a blocking function. Only callers that have acquired
// the mutex should call this function
func (plm *PLM) send(txPacket *Packet) (ack *Packet, err error) {
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
			<-time.After(delay)
		}

		insteon.Log.Tracef("Sending packet %v (write delay %v)", txPacket, writeDelay)
		plm.port.Write(buf)
		plm.nextWrite = time.Now().Add(writeDelay)

		ack, err = plm.ReadPacket()
	}
	return
}

func (plm *PLM) ReadPacket() (pkt *Packet, err error) {
	select {
	case pkt = <-plm.plmCh:
		if pkt.NAK() {
			err = ErrNak
		}
	case <-time.After(plm.timeout):
		err = ErrReadTimeout
	}

	return
}

func (plm *PLM) Connect(addr insteon.Address, options ...insteon.ConnectionOption) (insteon.Connection, error) {
	return plm.demux.New(addr, options...)
}

func (plm *PLM) Open(addr insteon.Address, options ...insteon.ConnectionOption) (insteon.Device, error) {
	conn, err := plm.Connect(addr, options...)
	if err != nil {
		return nil, err
	}

	return insteon.Open(conn, plm.timeout)
}

func (plm *PLM) Monitor(ch chan *insteon.Message) error {
	conn, err := plm.demux.New(insteon.Wildcard)
	if err == nil {
		conn.AddHandler(ch, insteon.Command{0x00, 0xff, 0xff})
	}
	return err
}

func (plm *PLM) Info() (info *Info, err error) {
	ack, err := plm.send(&Packet{Command: CmdGetInfo})
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

	_, err := plm.send(&Packet{Command: CmdReset})
	plm.timeout = timeout
	return err
}

func (plm *PLM) Config() (config Config, err error) {
	ack, err := plm.send(&Packet{Command: CmdGetConfig})
	if err == nil {
		err = config.UnmarshalBinary(ack.Payload)
	}
	return config, err
}

func (plm *PLM) SetConfig(config Config) error {
	payload, _ := config.MarshalBinary()
	_, err := plm.send(&Packet{Command: CmdSetConfig, Payload: payload})
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
	plm.port.Close()
}
