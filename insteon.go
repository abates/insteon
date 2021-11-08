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

//go:generate go run ./internal/ commands
//go:generate go run ./internal/ devcats
package insteon

import (
	"encoding/json"
	"fmt"
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

// FirmwareVersion indicates the software/firmware revision number of a device
type FirmwareVersion int

// String will return the hexadecimal string of the firmware version
func (fv FirmwareVersion) String() string {
	return fmt.Sprintf("%d", int(fv))
}

// ProductKey is a 3 byte code assigned by Insteon
type ProductKey [3]byte

// String returns the hexadecimal string for the product key
func (p ProductKey) String() string {
	return fmt.Sprintf("0x%02x%02x%02x", p[0], p[1], p[2])
}

// DevCat is a 2 byte value including a device category and
// sub-category.  Devices are grouped by categories (thermostat,
// light, etc) and then each category has specific types of devices
// such as on/off switches and dimmer switches
type DevCat [2]byte

// Domain returns what device domain a particular device belongs to
func (dc DevCat) Domain() Domain {
	return Domain(dc[0])
}

// Category returns the device's category.  For instance a DevCat with the
// Domain "Dimmable" may return the Category for LampLinc or SwitchLinc Dimmer
func (dc DevCat) Category() Category {
	return Category(dc[1])
}

// In determines if the DevCat domain is found in the list
func (dc DevCat) In(domains ...Domain) bool {
	for _, domain := range domains {
		if Domain(dc[0]) == domain {
			return true
		}
	}
	return false
}

// String returns a string representation of the DevCat in the
// form of category.subcategory where those fields are the 2 digit
// hex representation of their corresponding values
func (dc DevCat) String() string {
	return fmt.Sprintf("%02x.%02x", dc[0], dc[1])
}

// Domain represents an entire domain of similar devices (dimmers, switches, thermostats, etc)
type Domain byte

// Category indicates the specific kind of device within a domain.  For instance, a LampLing and
// a SwitchLinc Dimmer are both within the Dimmable device domain
type Category byte

// MarshalJSON will convert the DevCat to a valid JSON byte string
func (dc DevCat) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%02x.%02x", dc[0], dc[1]))
}

// UnmarshalJSON will unmarshal the input json byte string into the
// DevCat receiver
func (dc *DevCat) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err = json.Unmarshal(data, &s); err == nil {
		var n int
		n, err = fmt.Sscanf(s, "%02x.%02x", &dc[0], &dc[1])
		if n < 2 {
			err = fmt.Errorf("Expected Scanf to parse 2 digits, got %d", n)
		}
	}
	return err
}
