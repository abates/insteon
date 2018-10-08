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

	MaxRetries = 3
)

func hexDump(format string, buf []byte, sep string) string {
	str := make([]string, len(buf))
	for i, b := range buf {
		str[i] = fmt.Sprintf(format, b)
	}
	return strings.Join(str, sep)
}

type CommandRequest struct {
	Command Command
	Payload []byte
	Err     error
	DoneCh  chan<- *CommandRequest
}

type PacketRequest struct {
	Packet  *Packet
	Retry   int
	Ack     *Packet
	Err     error
	DoneCh  chan<- *PacketRequest
	timeout time.Time
}

type PLM struct {
	timeout     time.Duration
	connections []chan<- *Packet
	queue       []*PacketRequest

	sendCh         chan *PacketRequest
	upstreamSendCh chan<- []byte
	upstreamRecvCh <-chan []byte
	connectCh      chan chan<- *Packet
	disconnectCh   chan chan<- *Packet

	Network *insteon.Network
}

func New(port *Port, timeout time.Duration) *PLM {
	plm := &PLM{
		timeout: timeout,

		sendCh:         make(chan *PacketRequest, 1),
		upstreamSendCh: port.sendCh,
		upstreamRecvCh: port.recvCh,
		connectCh:      make(chan chan<- *Packet),
		disconnectCh:   make(chan chan<- *Packet),
	}

	go plm.process()

	sendCh, recvCh := plm.Connect(CmdSendInsteonMsg, CmdStdMsgReceived, CmdExtMsgReceived)
	plm.Network = insteon.New(sendCh, recvCh, timeout)
	return plm
}

func (plm *PLM) process() {
	for {
		select {
		case buf, open := <-plm.upstreamRecvCh:
			if !open {
				plm.close()
				return
			}
			plm.receive(buf)
		case request, open := <-plm.sendCh:
			if !open {
				plm.close()
				return
			}
			plm.queue = append(plm.queue, request)
			if len(plm.queue) == 1 {
				plm.send()
			}
		case connection := <-plm.connectCh:
			plm.connections = append(plm.connections, connection)
		case connection := <-plm.disconnectCh:
			plm.disconnect(connection)
		case <-time.After(plm.timeout):
			if len(plm.queue) > 0 && plm.queue[0].timeout.Before(time.Now()) {
				request := plm.queue[0]
				request.Err = ErrReadTimeout
				request.DoneCh <- request
				plm.queue = plm.queue[1:]
			}
		}
	}
}

func (plm *PLM) send() {
	if len(plm.queue) > 0 {
		request := plm.queue[0]
		if buf, err := request.Packet.MarshalBinary(); err == nil {
			insteon.Log.Debugf("Sending packet to port")
			request.timeout = time.Now().Add(plm.timeout)
			plm.upstreamSendCh <- buf
		} else {
			insteon.Log.Infof("Failed to marshal packet: %v", err)
			request.Err = err
			request.DoneCh <- request
			plm.queue = plm.queue[1:]
			plm.send()
		}
	}
}

func (plm *PLM) disconnect(connection chan<- *Packet) {
	for i, conn := range plm.connections {
		if conn == connection {
			close(conn)
			plm.connections = append(plm.connections[0:i], plm.connections[i+1:]...)
			break
		}
	}
}

func (plm *PLM) receive(buf []byte) {
	packet := &Packet{}
	err := packet.UnmarshalBinary(buf)

	if err == nil {
		insteon.Log.Tracef("RX %v", packet)
		if 0x50 <= packet.Command && packet.Command <= 0x58 {
			for _, connection := range plm.connections {
				connection <- packet
			}
		} else {
			plm.receiveAck(packet)
		}
	} else {
		insteon.Log.Infof("Failed to unmarshal packet: %v", err)
	}
}

func (plm *PLM) receiveAck(packet *Packet) {
	if len(plm.queue) > 0 {
		request := plm.queue[0]
		if packet.Command == plm.queue[0].Packet.Command {
			request.Ack = packet
			if packet.NAK() {
				request.Err = ErrNak
			}
			request.DoneCh <- plm.queue[0]
			plm.queue = plm.queue[1:]
			plm.send()
		}
	}
}

// Retry will deliver a packet to the Insteon network. If delivery fails (due
// to a NAK from the PLM) then we will retry and decrement retries. This
// continues until the packet is sent (as acknowledged by the PLM) or retries
// reaches zero
func (plm *PLM) Retry(packet *Packet, retries int) (ack *Packet, err error) {
	doneCh := make(chan *PacketRequest, 1)
	request := &PacketRequest{
		Packet: packet,
		Retry:  retries,
		DoneCh: doneCh,
	}

	plm.sendCh <- request
	<-doneCh
	if request.Err == ErrNak && retries > 0 {
		for request.Err == ErrNak && retries > 0 {
			insteon.Log.Debugf("Received NAK sending %q. Retrying", packet)
			retries--
			plm.sendCh <- request
			<-doneCh
		}

		if request.Err == ErrNak {
			insteon.Log.Debugf("Retry count exceeded")
			request.Err = ErrRetryCountExceeded
		}
	}
	return request.Ack, request.Err
}

func (plm *PLM) Connect(sendCmd Command, recvCmds ...Command) (chan<- *insteon.PacketRequest, <-chan []byte) {
	conn := plm.connect(sendCmd, recvCmds...)
	return conn.sendCh, conn.recvCh
}

func (plm *PLM) connect(sendCmd Command, recvCmds ...Command) *connection {
	sendCh := make(chan *CommandRequest, 1)
	recvCh := make(chan *Packet, 1)
	plm.connectCh <- recvCh

	go func() {
		for request := range sendCh {
			_, request.Err = plm.Retry(&Packet{Command: request.Command, Payload: request.Payload}, 0)
			request.DoneCh <- request
		}
		plm.disconnectCh <- recvCh
	}()

	conn := newConnection(sendCh, recvCh, sendCmd, recvCmds...)
	return conn
}

func (plm *PLM) Info() (*Info, error) {
	ack, err := plm.Retry(&Packet{Command: CmdGetInfo}, 0)
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

	ack, err := plm.Retry(&Packet{Command: CmdReset}, 0)

	if err == nil && ack.NAK() {
		err = ErrNak
	}
	plm.timeout = timeout
	return err
}

func (plm *PLM) Config() (*Config, error) {
	ack, err := plm.Retry(&Packet{Command: CmdGetConfig}, 0)
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
	ack, err := plm.Retry(&Packet{Command: CmdSetConfig, Payload: payload}, 0)
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

func (plm *PLM) close() {
	for _, connection := range plm.connections {
		close(connection)
	}
	close(plm.upstreamSendCh)
}

func (plm *PLM) Close() error {
	close(plm.sendCh)
	return nil
}
