package insteon

import (
	"fmt"
	"time"
)

// Insteon Engine Versions
const (
	VerI1 EngineVersion = iota
	VerI2
	VerI2Cs
)

// EngineVersion indicates the Insteon engine version that the
// device is running
type EngineVersion int

// String converts the version number to one of I1, I2 or I2CS corresponding to
// insteon engine version 1, version 2 and version 2 with checksum (CS)
func (ver EngineVersion) String() string {
	switch ver {
	case VerI1:
		return "I1"
	case VerI2:
		return "I2"
	case VerI2Cs:
		return "I2CS"
	}
	return fmt.Sprintf("Unknown(%v)", int(ver))
}

// The DeviceInitializer is a function that will return a fully initialized device
// using the input device as a template. DeviceInitializers are used to convert
// standard devices to category specific devices. For instance, when a PLM connects
// to a device, it uses an I1Device object to attempt to determine the device's
// category (by way of its product data).  If the product data is received, then
// the appropriate initializer is called for that device category in order to get
// a more specific device object (like a light)
type DeviceInitializer func(Device) Device

// Devices is a global DeviceRegistry. This device registry should only be used
// if you are adding a new device category to the system
var Devices DeviceRegistry

// DeviceRegistry is a mechanism to keep track of specific initializers for different
// device categories
type DeviceRegistry struct {
	// devices key is the first byte of the
	// Category.  Documentation simply calls this
	// the category and the second byte the sub
	// category, but we've combined both bytes
	// into a single type
	devices map[Category]DeviceInitializer
}

// Register will assign the given initializer to the supplied category
func (dr *DeviceRegistry) Register(category Category, initializer DeviceInitializer) {
	if dr.devices == nil {
		dr.devices = make(map[Category]DeviceInitializer)
	}
	dr.devices[category] = initializer
}

// Find looks for an initializer corresponding to the given category
func (dr *DeviceRegistry) Find(category Category) DeviceInitializer {
	return dr.devices[category]
}

// Initialize attempts to query the given device for its product data and then
// call the appropriate initializer for that device category. If the device fails
// to respond to the Product Data Request (generating an ErrReadTimeout) then the
// original device is returned
func (dr *DeviceRegistry) Initialize(device Device) (Device, error) {
	// query the device
	devCat, err := device.IDRequest()

	// construct device for device type
	if err == nil {
		initializer := dr.Find(devCat.Category())
		if initializer != nil {
			device = initializer(device)
		}
	} else if err == ErrReadTimeout {
		Log.Infof("Timed out waiting for device ID. Returning standard device")
		err = nil
	}

	return device, err
}

type Device interface {
	VersionedConnection
	Linkable
	ProductData() (*ProductData, error)
	FXUsername() (string, error)
	TextString() (string, error)
	SetTextString(string) error
	EngineVersion() (EngineVersion, error)
	Ping() error
	IDRequest() (DevCat, error)
}

func SendSubCommand(connection VersionedConnection, command *Command, subCommand int) (ack *Message, err error) {
	commandBytes := command.Version(connection.FirmwareVersion())
	return sendCommand(connection, commandBytes.SubCommand(subCommand), StandardDirectMessage, nil)
}

func SendCommand(connection VersionedConnection, command *Command) (ack *Message, err error) {
	return sendCommand(connection, command.Version(connection.FirmwareVersion()), StandardDirectMessage, nil)
}

func SendCommandAndWait(connection VersionedConnection, command *Command, waitCommands ...*Command) (msg *Message, err error) {
	Log.Debugf("Subscribing to traffic for command %v", waitCommands)
	waitBytes := make([]CommandBytes, len(waitCommands))
	for i, waitCommand := range waitCommands {
		waitBytes[i] = waitCommand.Version(connection.FirmwareVersion())
	}

	rxCh := connection.Subscribe(waitBytes...)
	defer connection.Unsubscribe(rxCh)
	ack, err := sendCommand(connection, command.Version(connection.FirmwareVersion()), StandardDirectMessage, nil)

	if err == nil {
		if ack.Nak() {
			err = ErrNak
		} else {
			Log.Debugf("Waiting for %v response", waitCommands)
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

func SendExtendedSubCommand(connection VersionedConnection, command *Command, subCommand int, payload []byte) (ack *Message, err error) {
	if len(payload) < 14 {
		payload = append(payload, make([]byte, 14-len(payload))...)
	}
	commandBytes := command.Version(connection.FirmwareVersion())
	return sendCommand(connection, commandBytes.SubCommand(subCommand), ExtendedDirectMessage, payload)
}

func SendExtendedCommand(connection VersionedConnection, command *Command, payload []byte) (ack *Message, err error) {
	if len(payload) < 14 {
		payload = append(payload, make([]byte, 14-len(payload))...)
	}
	return sendCommand(connection, command.Version(connection.FirmwareVersion()), ExtendedDirectMessage, payload)
}

func sendCommand(connection Connection, command CommandBytes, flags Flags, payload []byte) (ack *Message, err error) {
	return connection.Write(&Message{
		Flags:   flags,
		Command: command,
		Payload: payload,
	})
}
