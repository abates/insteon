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

// DeviceConstructor will return an initialized device for the given input arguments.  The DeviceInfo
// allows a factory to select different
type DeviceConstructor func(info DeviceInfo, device Device, timeout time.Duration) (Device, error)

// Devices is a global DeviceRegistry. This device registry should only be used
// if you are adding a new device category to the system
var Devices DeviceRegistry

// DeviceRegistry is a mechanism to keep track of specific constructors for different
// device categories
type DeviceRegistry struct {
	// Devices are mapped by their domain
	devices map[Domain]DeviceConstructor
}

// Register will assign the given constructor to the supplied category
func (dr *DeviceRegistry) Register(domain Domain, constructor DeviceConstructor) {
	if dr.devices == nil {
		dr.devices = make(map[Domain]DeviceConstructor)
	}
	dr.devices[domain] = constructor
}

// Delete will remove a device constructor from the registry
func (dr *DeviceRegistry) Delete(domain Domain) {
	delete(dr.devices, domain)
}

// Find looks for a constructor corresponding to the given category
func (dr *DeviceRegistry) Find(domain Domain) (DeviceConstructor, bool) {
	constructor, found := dr.devices[domain]
	return constructor, found
}

// New will look in the registry for a device constructor matching the
// given device category (supplied by the DeviceInfo argument).  If found,
// the constructor is called and the specific device type is returned.  If
// not found, then a base device (I1Device, I2Device, I2CsDevice) is returned
//
// Errors are only returned if the device category is found in the registry and
// that type's constructor returns an error
func (dr *DeviceRegistry) New(info DeviceInfo, conn Connection, timeout time.Duration) (Device, error) {
	device, err := New(info.EngineVersion, conn, timeout)
	if err == nil {
		if constructor, found := dr.Find(info.DevCat.Domain()); found {
			device, err = constructor(info, device, timeout)
		}
	}
	return device, err
}

// Addressable is any receiver that can be queried for its address
type Addressable interface {
	// Address will return the 3 byte destination address of the device.
	// All device implemtaions must be able to return their address
	Address() Address
}

// Device is the most basic capability that any device must implement. Devices
// can be sent commands and can receive messages
type Device interface {
	Connection

	// SendCommand will send the given command bytes to the device including
	// a payload (for extended messages). If payload length is zero then a standard
	// length message is used to deliver the commands.
	SendCommand(cmd Command, payload []byte) error
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

// AllLinkable is any device that has an all-link database that
// can be programmed remotely
type AllLinkable interface {
	// AssignToAllLinkGroup should be called after the set button
	// has been pressed on a responder. If the set button was pressed
	// then this method will assign the responder to the given
	// All-Link Group
	AssignToAllLinkGroup(Group) error

	// DeleteFromAllLinkGroup removes an All-Link record from a responding
	// device during an Unlinking session
	DeleteFromAllLinkGroup(Group) error
}

// LinkableDevice represents a Device that contains an All-Link database
// that can be accessed over the network.  Devices with Insteon Engine
// version 2 and higher are Linkable
type LinkableDevice interface {
	Device
	Linkable
}

// AddressableLinkable represent Linkable devices (may be devices such
// as a PLM) that also implement an Address() method to retrieve the
// device's Insteon address
type AddressableLinkable interface {
	Addressable
	Linkable
}

// Linkable is any device that can be put into
// linking mode and the link database can be managed remotely
type Linkable interface {
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

	// UpdateLinks will write the given links to the device's all-link
	// database.  Links will be written to available records
	// (link records marked with an Available flag).  If no more
	// available records are found, then the links will be appended
	// to the all-link database.  If a communication failure occurs then
	// the appropriate error is returned (ErrReadTimeout, ErrAckTimeout, etc.)
	// If an existing link is found that has different flags then the existing
	// record is updated to reflect the new flags
	UpdateLinks(...*LinkRecord) error

	// WriteLinks will overwrite the entire device all-link database
	// with the list of links provided.  If a communication failure occurs
	// then the appropriate error is returned (ErrReadTimeout, ErrAckTimeout,
	// etc).
	WriteLinks(...*LinkRecord) error
}

// DeviceInfo is a record of information about known
// devices on the network
type DeviceInfo struct {
	Address         Address
	DevCat          DevCat
	FirmwareVersion FirmwareVersion
	EngineVersion   EngineVersion
}

// Open will create a new device that is ready to be used. Open tries to contact
// the device to determine the device category and firmware version.  If successful,
// then a specific device type (dimmer, switch, thermostat, etc) is returned.  If
// the device responds with a NAK/NotLinked error, then a basic I2CsDevice is
// returned.  Only I2CsDevices will respond with a "Not Linked" NAK when being
// queried for the EngineVersion.
//
// If no spefici device type is found in the registry, then the base device (I1Device,
// I2Device or I2CsDevice) is returned.  If, in opening the device, a "Not Linked" NAK
// is encountered, then the I2CsDevice is returned with an ErrNotLinked error.  This
// allows the application to initiate linking, if desired
func Open(conn Connection, timeout time.Duration) (device Device, err error) {
	version, err := conn.EngineVersion()
	if err == nil {
		info := DeviceInfo{
			Address:       conn.Address(),
			EngineVersion: version,
		}
		info.FirmwareVersion, info.DevCat, err = conn.IDRequest()
		if err == nil {
			device, err = Devices.New(info, conn, timeout)
		}
	} else if err == ErrNotLinked {
		device, _ = New(VerI2Cs, conn, timeout)
	}
	return device, err
}

// New will return either an I1Device, an I2Device or an I2CsDevice based on the
// supplied EngineVersion
func New(version EngineVersion, conn Connection, timeout time.Duration) (device Device, err error) {
	switch version {
	case VerI1:
		device = newI1Device(conn, timeout)
	case VerI2:
		device = newI2Device(conn, timeout)
	case VerI2Cs:
		device = newI2CsDevice(conn, timeout)
	default:
		err = ErrVersion
	}
	return
}
