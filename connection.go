package insteon

import (
	"time"
)

var (
	// Timeout is a time.Duration that indicates how long
	// various operations should wait on a device to respond
	// defaults to 5 seconds
	Timeout = 5 * time.Second
)

// Connection is the interface that must be implemented by
// any device bridging the local program and the Insteon
// network. Any receiver implementing this interface must
// only communicate with a single device
type Connection interface {
	// Write sends a Message to a specific device on the network
	Write(*Message) (ack *Message, err error)

	// Subscribe provides a way to receive messages where the command
	// fields match one of the specified commands. Channels returned
	// by Subscribe must be closed using the Unsubscribe method
	Subscribe(match ...*Command) <-chan *Message

	// Unsubscribe removes a channel from its subscription.  Calling unsubscribe
	// will close the channel on the Connection end
	Unsubscribe(<-chan *Message)

	// Close will close the connection and any associated channels.  To
	// prevent reading from closed channels, Unsubscribe should be called
	// on any subscription channels prior to calling Close on the Connection
	Close() error
}

// I1Connection is used as a base connection for all devices.
// It represents the capabilities provided by devices with
// Version 1 engines.  Most functions of the I1Connection are
// used in the I2CSConnection
type I1Connection struct {
	Connection
}

// NewI1Connection creates a connection for a device having a version 1
// engine.
func NewI1Connection(conn Connection) Connection {
	return &I1Connection{conn}
}

// Write a message to the device and return the ACK. If a NAK
// is received, return one of ErrUnknownCommand, ErrNoLoadDetected or
// ErrNotLinked. These errors correspond to the associated Insteon
// error codes
//
// If the underlying connection returns an error (such as a timeout) then
// this error is returned back to the caller
func (i1conn *I1Connection) Write(message *Message) (ack *Message, err error) {
	ack, err = i1conn.Connection.Write(message)

	if ack != nil && ack.Flags.Type() == MsgTypeDirectNak {
		switch ack.Command.Cmd[1] {
		case 0xfd:
			err = ErrUnknownCommand
		case 0xfe:
			err = ErrNoLoadDetected
		case 0xff:
			err = ErrNotLinked
		default:
			err = TraceError(ErrUnexpectedResponse)
		}
	}
	return
}

// I2CsConnection is used for any Insteon Version 2CS (checksum)
// devices. This connection will ensure that the checksum is
// calculated for all outgoing insteon extended messages. It
// also uses slightly different Commands as required by I2CS engines
type I2CsConnection struct {
	Connection
}

// NewI2CsConnection creates a connection appropriate for communicating
// with I2CS devices. All extended messages through an I2CsConnection will
// have their checksum field computed and updated
func NewI2CsConnection(conn Connection) Connection {
	return &I2CsConnection{conn}
}

// Write a message to the device and return the ACK. If a NAK
// is received, return one of ErrIllegalValue, ErrPreNak, ErrIncorrectChecksum
// ErrNoLoadDetected or ErrNotLinked. These errors correspond to the
// associated Insteon error codes. The primary difference between this
// Write and an I1Connection write is that this Write will set the outgoing
// message version to 2 which triggers a checksum computation for the message
//
// If the underlying connection returns an error (such as a timeout) then
// this error is returned back to the caller
func (i2cs *I2CsConnection) Write(message *Message) (*Message, error) {
	message.version = VerI2Cs
	ack, err := i2cs.Connection.Write(message)
	if ack != nil && ack.Flags.Type() == MsgTypeDirectNak {
		switch ack.Command.Cmd[1] {
		case 0xfb:
			err = ErrIllegalValue
		case 0xfc:
			err = ErrPreNak
		case 0xfd:
			err = ErrIncorrectChecksum
		case 0xfe:
			err = ErrNoLoadDetected
		case 0xff:
			err = ErrNotLinked
		default:
			err = TraceError(ErrUnexpectedResponse)
		}
	}
	return ack, err
}

// SendStandardCommandAndWait will prepare a standard length direct message
// with the specified command and wait for a response matching waitCmd. If
// no response is received prior to a Timeout then SendStandardCommandAndWait will return
// ErrReadTimeout
func SendStandardCommandAndWait(conn Connection, command *Command, waitCmds ...*Command) (msg *Message, err error) {
	Log.Debugf("Subscribing to traffic for command %v", waitCmds)
	rxCh := conn.Subscribe(waitCmds...)
	defer conn.Unsubscribe(rxCh)
	ack, err := SendStandardCommand(conn, command)

	if err == nil {
		if ack.Nak() {
			err = ErrNak
		} else {
			Log.Debugf("Waiting for %v response", waitCmds)
			select {
			case msg = <-rxCh:
				Log.Debugf("Received message %v", msg)
			case <-time.After(Timeout):
				err = ErrReadTimeout
			}
		}
	}
	return
}

// SendStandardCommand will prepare a standard length direct message and
// send it to the device.  Upstream errors are returned
func SendStandardCommand(conn Connection, command *Command) (*Message, error) {
	Log.Debugf("Sending standard command %v", command)
	return conn.Write(&Message{
		Flags:   StandardDirectMessage,
		Command: command,
	})
}

// SendExtendedCommand will prepare an extended length direct message with
// the specified commands and payload and deliver it to the device.  Upstream
// errors are returned
func SendExtendedCommand(conn Connection, command *Command, payload Payload) (*Message, error) {
	Log.Debugf("Sending extended command %v %v", command, payload)
	return conn.Write(&Message{
		Flags:   ExtendedDirectMessage,
		Command: command,
		Payload: payload,
	})
}
