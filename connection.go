package insteon

import (
	"fmt"
	"time"
)

type Bridge interface {
	Send(timeout time.Duration, message *Message) error
	Receive(timeout time.Duration) (*Message, error)
}

type Connection interface {
	SendStandardCommandAndWait(*Command) (*Message, error)
	SendStandardCommand(*Command) error
	SendExtendedCommand(*Command, Payload) error
	Receive() (*Message, error)
}

func NewDeviceConnection(timeout time.Duration, address Address, bridge Bridge) Connection {
	return &DeviceConnection{
		timeout: timeout,
		address: address,
		bridge:  bridge,
	}
}

type DeviceConnection struct {
	timeout time.Duration
	address Address
	bridge  Bridge
}

func (dc *DeviceConnection) SendStandardCommandAndWait(command *Command) (msg *Message, err error) {
	err = dc.SendStandardCommand(command)
	if err == nil {
		msg, err = dc.Receive()
	}
	return
}

func (dc *DeviceConnection) SendStandardCommand(command *Command) error {
	err := dc.bridge.Send(dc.timeout, &Message{
		Dst:     dc.address,
		Flags:   StandardDirectMessage,
		Command: command,
	})

	if err == nil {
		msg, err := dc.Receive()
		if err == nil {
			if msg.Command == command {
				if msg.Flags.Type() == MsgTypeDirectAck {
					Log.Debugf("INSTEON ACK received")
				} else if msg.Flags.Type() == MsgTypeDirectNak {
					err = ErrDeviceNak
				} else {
					err = fmt.Errorf("Unexpected response from device: %v", msg)
				}
			} else {
				err = fmt.Errorf("Unexpected messge: %v", msg)
			}
		}
	}

	return err
}

func (dc *DeviceConnection) SendExtendedCommand(command *Command, payload Payload) error {
	return dc.bridge.Send(dc.timeout, &Message{
		Dst:     dc.address,
		Flags:   ExtendedDirectMessage,
		Command: command,
		Payload: payload,
	})
}

func (dc *DeviceConnection) Receive() (message *Message, err error) {
	return dc.bridge.Receive(dc.timeout)
}
