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

package devices

import (
	"fmt"

	"github.com/abates/insteon"
	"github.com/abates/insteon/commands"
)

// ProductData contains information about the device including its
// product key and device category
type ProductData struct {
	Key    insteon.ProductKey
	DevCat insteon.DevCat
}

// UnmarshalBinary takes the input byte buffer and unmarshals it into the
// ProductData object
func (pd *ProductData) UnmarshalBinary(buf []byte) error {
	if len(buf) < 14 {
		return fmt.Errorf("%w wanted 14 bytes got %d", insteon.ErrBufferTooShort, len(buf))
	}

	copy(pd.Key[:], buf[1:4])
	copy(pd.DevCat[:], buf[4:6])
	return nil
}

// MarshalBinary will convert the ProductData to a binary byte string
// for sending on the network
func (pd *ProductData) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 7)
	copy(buf[1:4], pd.Key[:])
	copy(buf[4:6], pd.DevCat[:])
	buf[6] = 0xff
	return buf, nil
}

type Device interface {
	commands.Commandable

	// Address will return the 3 byte destination address of the device.
	// All device implemtaions must be able to return their address
	Address() insteon.Address

	// Info will return the device's information
	Info() DeviceInfo
}

type ExtendedGetSet interface {
	ExtendedGet([]byte) ([]byte, error)
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
	AssignToAllLinkGroup(insteon.Group) error

	// DeleteFromAllLinkGroup removes an All-Link record from a responding
	// device during an Unlinking session
	DeleteFromAllLinkGroup(insteon.Group) error
}

// Linkable is any device that can be put into
// linking mode and the link database can be managed remotely
type Linkable interface {
	// Address will return the 3 byte destination address of the device.
	// All device implemtaions must be able to return their address
	Address() insteon.Address

	// EnterLinkingMode is the programmatic equivalent of holding down
	// the set button for two seconds. If the device is the first
	// to enter linking mode, then it is the controller. The next
	// device to enter linking mode is the responder.  LinkingMode
	// is usually indicated by a flashing GREEN LED on the device
	EnterLinkingMode(insteon.Group) error

	// EnterUnlinkingMode puts a controller device into unlinking mode
	// when the set button is then pushed (EnterLinkingMode) on a linked
	// device the corresponding links in both the controller and responder
	// are deleted.  EnterUnlinkingMode is the programmatic equivalent
	// to pressing the set button until the device beeps, releasing, then
	// pressing the set button again until the device beeps again. UnlinkingMode
	// is usually indicated by a flashing RED LED on the device
	EnterUnlinkingMode(insteon.Group) error

	// ExitLinkingMode takes a controller out of linking/unlinking mode.
	ExitLinkingMode() error

	// Links will return a list of LinkRecords that are present in
	// the All-Link database
	Links() ([]insteon.LinkRecord, error)

	// UpdateLinks will write the given links to the device's all-link
	// database.  Links will be written to available records
	// (link records marked with an Available flag).  If no more
	// available records are found, then the links will be appended
	// to the all-link database.  If a communication failure occurs then
	// the appropriate error is returned (ErrReadTimeout, ErrAckTimeout, etc.)
	// If an existing link is found that has different flags then the existing
	// record is updated to reflect the new flags
	UpdateLinks(...insteon.LinkRecord) error

	// WriteLinks will overwrite the entire device all-link database
	// with the list of links provided.  If a communication failure occurs
	// then the appropriate error is returned (ErrReadTimeout, ErrAckTimeout,
	// etc).
	WriteLinks(...insteon.LinkRecord) error
}

type WriteLink interface {
	WriteLink(index int, record insteon.LinkRecord) error
}

// DeviceInfo is a record of information about known
// devices on the network
type DeviceInfo struct {
	Address         insteon.Address         `json:"address"`
	DevCat          insteon.DevCat          `json:"devCat"`
	FirmwareVersion insteon.FirmwareVersion `json:"firmwareVersion"`
	EngineVersion   insteon.EngineVersion   `json:"engineVersion"`
}

// Open will try to establish communication with the remote device.
// If the device responds, Open will request its engine version as
// well as device info in order to return the correct device type
// (Dimmer, switch, thermostat, etc).  Open requires a MessageWriter,
// such as a PLM to use to communicate with the Insteon network
func Open(mw MessageWriter, dst insteon.Address, filters ...Filter) (device *BasicDevice, info DeviceInfo, err error) {
	for _, filter := range filters {
		mw = filter.Filter(mw)
	}

	info.Address = dst
	info.EngineVersion, err = GetEngineVersion(mw, dst)
	if err == nil {
		info.FirmwareVersion, info.DevCat, err = IDRequest(mw, dst)
	}

	if err == nil || err == ErrNotLinked {
		device = New(mw, info)
	}
	return
}

// Lookup will convert the *BasicDevice to a more specific device (*Dimmer,
// *Switch, etc)
func Lookup(bd *BasicDevice) (device Device) {
	switch bd.DevCat.Domain() {
	case insteon.DimmerDomain:
		device = NewDimmer(bd)
	case insteon.SwitchDomain:
		if bd.DevCat.Category() == 0x08 {
			device = NewOutlet(bd)
		} else {
			device = NewSwitch(bd)
		}
	case insteon.ThermostatDomain:
		device = NewThermostat(bd)
	default:
		device = bd
	}
	return device
}
