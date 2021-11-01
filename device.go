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

// Insteon Engine Versions
const (
	VerI1 EngineVersion = iota
	VerI2
	VerI2Cs
)

// EngineVersion indicates the Insteon engine version that the
// device is running
type EngineVersion int

func (ev EngineVersion) String() string {
	switch ev {
	case VerI1:
		return "I1"
	case VerI2:
		return "I2"
	case VerI2Cs:
		return "I2Cs"
	}
	return "unknown"
}

// Addressable is any receiver that can be queried for its address
type Addressable interface {
	// Address will return the 3 byte destination address of the device.
	// All device implemtaions must be able to return their address
	Address() Address
}

type Device interface {
	MessageWriter
	Addressable
	Commandable
	// Info will return the device's information
	Info() DeviceInfo
}

type ExtendedGetSet interface {
	ExtendedGet([]byte) ([]byte, error)
}

// Commandable indicates that the implementation exists to send commands
type Commandable interface {
	// SendCommand will send the given command bytes to the device including
	// a payload (for extended messages). If payload length is zero then a standard
	// length message is used to deliver the commands.
	SendCommand(cmd Command, payload []byte) (err error)

	Send(cmd Command, payload []byte) (ack Command, err error)
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

// Linkable is any device that can be put into
// linking mode and the link database can be managed remotely
type Linkable interface {
	Addressable

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

	// LinkDatabase will return a LinkDatabase if the underlying device
	// supports it.  If the underlying device (namely I1 devices) does
	// not support the All-Link database then an ErrNotSupported will
	// be returned
	//LinkDatabase() (Linkable, error)

	// Links will return a list of LinkRecords that are present in
	// the All-Link database
	Links() ([]LinkRecord, error)

	// UpdateLinks will write the given links to the device's all-link
	// database.  Links will be written to available records
	// (link records marked with an Available flag).  If no more
	// available records are found, then the links will be appended
	// to the all-link database.  If a communication failure occurs then
	// the appropriate error is returned (ErrReadTimeout, ErrAckTimeout, etc.)
	// If an existing link is found that has different flags then the existing
	// record is updated to reflect the new flags
	UpdateLinks(...LinkRecord) error

	// WriteLinks will overwrite the entire device all-link database
	// with the list of links provided.  If a communication failure occurs
	// then the appropriate error is returned (ErrReadTimeout, ErrAckTimeout,
	// etc).
	WriteLinks(...LinkRecord) error
}

// DeviceInfo is a record of information about known
// devices on the network
type DeviceInfo struct {
	Address         Address         `json:"address"`
	DevCat          DevCat          `json:"devCat"`
	FirmwareVersion FirmwareVersion `json:"firmwareVersion"`
	EngineVersion   EngineVersion   `json:"engineVersion"`
}

// Open will try to establish communication with the remote device.
// If the device responds, Open will request its engine version as
// well as device info in order to return the correct device type
// (Dimmer, switch, thermostat, etc).  Open requires a MessageWriter,
// such as a PLM to use to communicate with the Insteon network
func Open(mw MessageWriter, dst Address) (device Device, info DeviceInfo, err error) {
	info.Address = dst
	info.EngineVersion, err = GetEngineVersion(mw, dst)
	if err == nil {
		info.FirmwareVersion, info.DevCat, err = IDRequest(mw, dst)
	}

	if err == nil {
		device = New(mw, info)
	}
	return
}

// New will use the supplied DeviceInfo to create a device instance for the
// given connection.  For instance, if the DevCat is 0x01 with an I2CS
// EngineVersion then a Dimmer with an underlying i2CsDevice will be returned
func New(mw MessageWriter, info DeviceInfo) (device Device) {
	d := newDevice(mw, info)
	switch info.DevCat.Domain() {
	case DimmerDomain:
		device = NewDimmer(d, info)
	case SwitchDomain:
		if info.DevCat.Category() == 0x08 {
			device = NewOutlet(d, info)
		} else {
			device = NewSwitch(d, info)
		}
	case ThermostatDomain:
		device = NewThermostat(d, info)
	}
	return device
}
