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

package insteon

import (
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

// DeviceConstructor will return an initialized device for the given input arguments
type DeviceConstructor func(info DeviceInfo, address Address, sendCh chan<- *MessageRequest, recvCh <-chan *Message, timeout time.Duration) (Device, error)

// Devices is a global DeviceRegistry. This device registry should only be used
// if you are adding a new device category to the system
var Devices DeviceRegistry

// DeviceRegistry is a mechanism to keep track of specific constructors for different
// device categories
type DeviceRegistry struct {
	// devices key is the first byte of the
	// Category.  Documentation simply calls this
	// the category and the second byte the sub
	// category, but we've combined both bytes
	// into a single type
	devices map[Category]DeviceConstructor
}

// Register will assign the given constructor to the supplied category
func (dr *DeviceRegistry) Register(category Category, constructor DeviceConstructor) {
	if dr.devices == nil {
		dr.devices = make(map[Category]DeviceConstructor)
	}
	dr.devices[category] = constructor
}

// Delete will remove a device constructor from the registry
func (dr *DeviceRegistry) Delete(category Category) {
	delete(dr.devices, category)
}

// Find looks for a constructor corresponding to the given category
func (dr *DeviceRegistry) Find(category Category) (DeviceConstructor, bool) {
	constructor, found := dr.devices[category]
	return constructor, found
}

// CommandRequest is used to request that a given command and payload are sent to a device
type CommandRequest struct {
	// Command to send to the device
	Command Command

	// Payload to include, if set
	Payload []byte

	// RecvCh to receive subsuquent messages
	RecvCh chan<- *CommandResponse

	// DoneCh that will be written to by the connection once the request is complete
	DoneCh chan<- *CommandRequest

	// Ack contains the response Ack/Nak from the device
	Ack *Message

	// Err includes any error that occurred while trying to send the request
	Err error

	timeout time.Time
}

// CommandResponse is used for sending messages back to a caller in conjunction with a CommandRequest
type CommandResponse struct {
	// Message that was received
	Message *Message

	// DoneCh to indicate whether more messages should be received or not.
	DoneCh chan<- *CommandResponse

	request *CommandRequest
}

// Addressable is any receiver that can be queried for its address
type Addressable interface {
	// Address will return the 3 byte destination address of the device.
	// All device implemtaions must be able to return their address
	Address() Address
}

// Commandable is the most basic capability that any device must implement.  Commandable
// devices can be sent commands and can receive messages
type Commandable interface {
	// SendCommand will send the given command bytes to the device including
	// a payload (for extended messages). If payload length is zero then a standard
	// length message is used to deliver the commands. The command bytes from the
	// response ack are returned as well as any error
	SendCommand(cmd Command, payload []byte) (response Command, err error)

	// SendCommandAndListen performs the same function as SendCommand.  However, instead of returning
	// the Ack/Nak command, it returns a channel that can be read to get messages received after
	// the command was sent.  This is useful for things like retrieving the link database where the
	// response information is not in the Ack but in one or more ALDB responses.  Once all information
	// has been received the command response DoneCh should be sent a "false" value to indicate no
	// more messages are expected.
	SendCommandAndListen(cmd Command, payload []byte) (recvCh <-chan *CommandResponse, err error)
}

// Device is any implementation that returns the device address and can send commands to the
// destination addresss
type Device interface {
	Addressable
	Commandable
}

// PingableDevice is any device that implements the Ping method
type PingableDevice interface {
	// Ping sends a ping request to the device and waits for a single ACK
	Ping() error
}

// NameableDevice is any device that have a settable text string
type NameableDevice interface {
	// TextString returns the information assigned to the device
	TextString() (string, error)

	// SetTextString assigns the information to the device
	SetTextString(string) error
}

// FXDevice indicates the device is capable of user-defined FX commands
type FXDevice interface {
	FXUsername() (string, error)
}

// AllLinkableDevice is any device that has an all-link database that
// can be programmed remotely
type AllLinkableDevice interface {
	// AssignToAllLinkGroup should be called after the set button
	// has been pressed on a responder. If the set button was pressed
	// then this method will assign the responder to the given
	// All-Link Group
	AssignToAllLinkGroup(Group) error

	// DeleteFromAllLinkGroup removes an All-Link record from a responding
	// device during an Unlinking session
	DeleteFromAllLinkGroup(Group) error
}

// LinkableDevice is any device that can be put into
// linking mode and the link database can be managed remotely
type LinkableDevice interface {
	// Address is the remote/destination address of the device
	Address() Address

	// AppendLink will add a new link record to the end of the All-Link database
	AppendLink(link *LinkRecord) error

	// EnterLinkingMode is the programmatic equivalent of holding down
	// the set button for two seconds. If the device is the first
	// to enter linking mode, then it is the controller. The next
	// device to enter linking mode is the responder.  LinkingMode
	// is usually indicated by a flashing GREEN LED on the device
	EnterLinkingMode(Group) error

	// EnterUnlinkingMode puts a controller device into unlinking mode
	// when the set button is then pushed (EnterLinkingMode) on a linked
	// device the corresponding links in both the controller and responder
	// are deleted.  EnterUnlinkingMode is the programmatic equivalent
	// to pressing the set button until the device beeps, releasing, then
	// pressing the set button again until the device beeps again. UnlinkingMode
	// is usually indicated by a flashing RED LED on the device
	EnterUnlinkingMode(Group) error

	// ExitLinkingMode takes a controller out of linking/unlinking mode.
	ExitLinkingMode() error

	// Links will return a list of LinkRecords that are present in
	// the All-Link database
	Links() ([]*LinkRecord, error)

	// AddLink will either add the link to the All-Link database
	// or it will replace an existing link-record that has been marked
	// as deleted
	AddLink(newLink *LinkRecord) error

	// RemoveLinks will either remove the link records from the device
	// All-Link database, or it will simply mark them as deleted
	RemoveLinks(oldLinks ...*LinkRecord) error

	// WriteLink will write the link record to the device's link database
	WriteLink(*LinkRecord) error
}
