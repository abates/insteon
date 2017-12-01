package insteon

import (
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

func (dc *DeviceConnection) send(msg *Message) error {
	err := dc.bridge.Send(dc.timeout, msg)

	if err == nil {
		var rxMsg *Message
		rxMsg, err = dc.Receive()
		if err == nil {
			if rxMsg.Flags.Type() == MsgTypeDirectAck {
				if rxMsg.Command.cmd == msg.Command.cmd {
					Log.Debugf("INSTEON ACK received")
				} else {
					err = ErrUnexpectedResponse
				}
			} else if rxMsg.Flags.Type() == MsgTypeDirectNak {
				if rxMsg.Command.cmd[1] == 0xfd {
					err = ErrUnknownCommand
				} else if rxMsg.Command.cmd[1] == 0xfe {
					err = ErrNoLoadDetected
				} else if rxMsg.Command.cmd[1] == 0xff {
					err = ErrNotLinked
				}
			} else {
				err = ErrUnexpectedResponse
			}
		}
	}

	return err
}

func (dc *DeviceConnection) SendStandardCommandAndWait(command *Command) (msg *Message, err error) {
	err = dc.SendStandardCommand(command)
	if err == nil {
		msg, err = dc.Receive()
	}
	return
}

func (dc *DeviceConnection) SendStandardCommand(command *Command) error {
	return dc.send(&Message{
		Dst:     dc.address,
		Flags:   StandardDirectMessage,
		Command: command,
	})
}

func (dc *DeviceConnection) SendExtendedCommand(command *Command, payload Payload) error {
	return dc.send(&Message{
		Dst:     dc.address,
		Flags:   ExtendedDirectMessage,
		Command: command,
		Payload: payload,
	})
}

func (dc *DeviceConnection) Receive() (message *Message, err error) {
	return dc.bridge.Receive(dc.timeout)
}
