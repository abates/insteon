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
	"sync"
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

	MaxRetries = 3
)

func hexDump(format string, buf []byte, sep string) string {
	str := make([]string, len(buf))
	for i, b := range buf {
		str[i] = fmt.Sprintf(format, b)
	}
	return strings.Join(str, sep)
}

type PLM struct {
	sync.Mutex
	timeout    time.Duration
	writeDelay time.Duration
	nextWrite  time.Time
	port       *Port

	insteonRxCh chan *insteon.Message
	insteonTxCh chan *insteon.Message
	plmCh       chan *Packet
}

// The Option mechanism is based on the method described at https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type Option func(p *PLM) error

// New creates a new PLM instance.
func New(port *Port, timeout time.Duration, options ...Option) *PLM {
	plm := &PLM{
		timeout:     timeout,
		writeDelay:  500 * time.Millisecond,
		port:        port,
		insteonTxCh: make(chan *insteon.Message),
		insteonRxCh: make(chan *insteon.Message),
		plmCh:       make(chan *Packet),
	}

	for _, o := range options {
		err := o(plm)
		if err != nil {
			insteon.Log.Infof("error setting option %v: %v", err, err)
			return nil
			// TODO: change New() to return an error if there's an error
		}
	}

	go plm.readLoop()
	go plm.writeLoop()
	return plm
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
						plm.insteonRxCh <- msg
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

func (plm *PLM) writeLoop() {
	for msg := range plm.insteonTxCh {
		buf, err := msg.MarshalBinary()
		if err == nil {
			// slice off the source address since the PLM doesn't want it
			buf = buf[3:]
			_, err = plm.send(&Packet{Command: 0x62, Payload: buf})
			if err != nil {
				insteon.Log.Infof("Failed to send packet: %v", err)
			}
		} else {
			insteon.Log.Infof("Failed to marshal insteon message: %v", err)
		}
	}
}

// transmit a packet and wait for the PLM to ack that the packet was
// sent.  This is a blocking function. Only callers that have acquired
// the mutex shoud call this function
func (plm *PLM) tx(txPacket *Packet) (ack *Packet, err error) {
	buf, err := txPacket.MarshalBinary()
	if err == nil {
		if time.Now().Before(plm.nextWrite) {
			<-time.After(plm.nextWrite.Sub(time.Now()))
		}

		plm.port.Write(buf)
		if 0x61 <= txPacket.Command && txPacket.Command <= 0x63 {
			plm.nextWrite = time.Now().Add(plm.writeDelay)
		} else {
			plm.nextWrite = time.Now()
		}

		// loop until either timeout or the appropriate ack is received
		timeout := time.Now().Add(plm.timeout)
		for err == nil {
			select {
			case rxPacket := <-plm.plmCh:
				if txPacket.Command == rxPacket.Command {
					ack = rxPacket
					if rxPacket.NAK() {
						err = ErrNak
					}
					return
				}
			case <-time.After(plm.timeout):
				err = ErrAckTimeout
			}

			if timeout.Before(time.Now()) {
				err = ErrAckTimeout
			}
		}
	}
	return
}

// send a packet and wait for the PLM to ack that the packet was
// sent.  This is a blocking function
func (plm *PLM) send(txPacket *Packet) (ack *Packet, err error) {
	plm.Lock()
	defer plm.Unlock()
	return plm.tx(txPacket)
}

func (plm *PLM) Connect(addr insteon.Address, options ...insteon.ConnectionOption) insteon.Connection {
	return insteon.NewConnection(plm.insteonTxCh, plm.insteonRxCh, addr, options...)
}

func (plm *PLM) Open(addr insteon.Address, options ...insteon.ConnectionOption) (insteon.Device, error) {
	conn := plm.Connect(addr, options...)
	return insteon.Open(conn, plm.timeout)
}

// Retry will deliver a packet to the Insteon network. If delivery fails (due
// to a NAK from the PLM) then we will retry and decrement retries. This
// continues until the packet is sent (as acknowledged by the PLM) or retries
// reaches zero
func (plm *PLM) Retry(packet *Packet, retries int) (ack *Packet, err error) {
	plm.Lock()
	defer plm.Unlock()
	for err == ErrNak && retries > 0 {
		ack, err = plm.tx(packet)
		retries--
	}

	if err == ErrNak {
		insteon.Log.Debugf("Retry count exceeded")
		err = ErrRetryCountExceeded
	}
	return ack, err
}

func (plm *PLM) Info() (*Info, error) {
	ack, err := plm.send(&Packet{Command: CmdGetInfo})
	if err == nil {
		info := &Info{}
		err := info.UnmarshalBinary(ack.Payload)
		return info, err
	}
	return nil, err
}

func (plm *PLM) Reset() error {
	timeout := plm.timeout
	plm.timeout = 20 * time.Second

	ack, err := plm.send(&Packet{Command: CmdReset})

	if err == nil && ack.NAK() {
		err = ErrNak
	}
	plm.timeout = timeout
	return err
}

func (plm *PLM) Config() (*Config, error) {
	ack, err := plm.send(&Packet{Command: CmdGetConfig})
	if err == nil && ack.NAK() {
		err = ErrNak
	} else if err == nil {
		var config Config
		err := config.UnmarshalBinary(ack.Payload)
		return &config, err
	}
	return nil, err
}

func (plm *PLM) SetConfig(config *Config) error {
	payload, _ := config.MarshalBinary()
	ack, err := plm.send(&Packet{Command: CmdSetConfig, Payload: payload})
	if err == nil && ack.NAK() {
		err = ErrNak
	}
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
	close(plm.insteonTxCh)
	plm.port.Close()
}
