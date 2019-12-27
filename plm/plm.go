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

type Resetable interface {
	Reset() error
}

type Configurable interface {
	Config() (config *Config, err error)
	SetConfig(config *Config) error
}

type Categorical interface {
	SetDeviceCategory(insteon.Category) error
}

type Sleepable interface {
	RFSleep() error
}

type Monitorable interface {
	Monitor() (<-chan *insteon.Message, error)
}

type PLM struct {
	sync.Mutex
	linkdb
	portMutex  sync.Mutex
	timeout    time.Duration
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
		writeDelay: 500 * time.Millisecond,
		port:       port,

		plmCh: make(chan *Packet),
	}
	plm.demux = insteon.NewDemux(plm)
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
		_, err = plm.send(&Packet{Command: CmdSendInsteonMsg, Payload: buf})
	}
	return err
}

// send a packet and wait for the PLM to ack that the packet was
// sent.  This is a blocking function. Only callers that have acquired
// the mutex should call this function
func (plm *PLM) send(txPacket *Packet) (ack *Packet, err error) {
	plm.portMutex.Lock()
	defer plm.portMutex.Unlock()

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

func (plm *PLM) receive(timeout time.Duration) (pkt *Packet, err error) {
	plm.portMutex.Lock()
	defer plm.portMutex.Unlock()
	select {
	case pkt = <-plm.plmCh:
	case <-time.After(timeout):
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

/*func (plm *plm) Monitor() (<-chan *insteon.Message, error) {
	var err error
	addr := insteon.Address{0, 0, 0}
	if conn, found := plm.connections[addr]; found {
		return conn.rxCh, nil
	}
	conn := &connection{plm: plm, rxCh: make(chan *insteon.Message, 0)}
	plm.connections[addr] = conn
	return conn.rxCh, err
}*/

// retry will deliver a packet to the Insteon network. If delivery fails (due
// to a NAK from the PLM) then we will retry and decrement retries. This
// continues until the packet is sent (as acknowledged by the PLM) or retries
// reaches zero
func retry(plm *PLM, packet *Packet, retries int) (ack *Packet, err error) {
	for err == ErrNak || retries > 0 {
		ack, err = plm.send(packet)
		if err == nil {
			break
		}
		retries--
	}

	if err == ErrNak {
		insteon.Log.Debugf("Retry count exceeded")
		err = ErrRetryCountExceeded
	}
	return ack, err
}

func (plm *PLM) Info() (info *Info, err error) {
	ack, err := plm.send(&Packet{Command: CmdGetInfo})
	if err == nil {
		info = &Info{}
		err = info.UnmarshalBinary(ack.Payload)
	}
	return info, err
}

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
