//go:generate go run ./internal/ devcats
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
	"encoding/json"
	"errors"
	"fmt"
)

var (
	// ErrBufferTooShort indicates a buffer underrun when unmarshalling data
	ErrBufferTooShort = errors.New("Buffer is too short")

	// ErrReadTimeout indicates the timeout period expired while waiting for
	// a specific message
	ErrReadTimeout = errors.New("Read Timeout")

	// ErrAckTimeout indicates the timeout period expired while waiting for
	// ack message for a previously sent Direct message
	ErrAckTimeout = errors.New("Ack Timeout")

	// ErrWriteTimeout indicates the timeout period expired while waiting to
	// write a message
	ErrWriteTimeout = errors.New("Write Timeout")

	// ErrSendTimeout indicates the timeout period expired while trying to
	// send a message
	ErrSendTimeout = errors.New("Send Timeout")

	// ErrNotSupported indicates some feature (namely an updateable All-Link database) is
	// not supported by the underyling Insteon device
	ErrNotSupported = errors.New("Feature is not supported by the device")

	// ErrNotImplemented indicates that a device function has not yet been implemented
	ErrNotImplemented = errors.New("Command is not yet implemented")

	// ErrUnexpectedResponse is returned when a Nak is not understood
	ErrUnexpectedResponse = errors.New("Unexpected response from device")

	// ErrNotLinked indicates the device does not have an all-link entry in its
	// database
	ErrNotLinked = errors.New("Not in All-Link group")

	// ErrNotLinkable indicates a linking function was requested on a non-linkable device
	ErrNotLinkable = errors.New("Device is not remotely linkable")

	// ErrNoLoadDetected is an error returned by the device (this error condition is not documented)
	ErrNoLoadDetected = errors.New("No load detected")

	// ErrUnknownCommand is returned by the device (as a Nak) in response to an unknown command byte
	ErrUnknownCommand = errors.New("Unknown command")

	// ErrNak indicates a negative acknowledgement was received in response to a sent message
	ErrNak = errors.New("NAK received")

	// ErrUnknown is returned by a connection when a NAK occurred but the error code
	// is not known
	ErrUnknown = errors.New("Device returned unknown error")

	// ErrIllegalValue is returned by I2Cs devices (this error condition is not documented)
	ErrIllegalValue = errors.New("Illegal value in command")

	// ErrIncorrectChecksum is returned by I2Cs devices when an invalid checksum is detected
	ErrIncorrectChecksum = errors.New("I2CS invalid checksum")

	// ErrPreNak is returned by I2Cs devices (this error condition is not documented)
	ErrPreNak = errors.New("Database search took too long")

	// ErrAddrFormat is returned when unmarshalling an address from text and the
	// text is in an unsupported format
	ErrAddrFormat = errors.New("address format is xx.xx.xx (digits in hex)")

	// ErrInvalidMemAddress indicates a link record memory address is invalid
	ErrInvalidMemAddress = errors.New("Invalid memory address")

	// ErrVersion is returned when an engine version value is not known
	ErrVersion = errors.New("Unknown Insteon Engine Version")

	// ErrLinkIndexOutOfRange indicates that the index exceeds the length of the all-link database
	ErrLinkIndexOutOfRange = errors.New("Link index is beyond the bounds of the link database")

	// ErrReceiveComplete is used when calling the Receive() utility function.  If the callback is finished
	// receiving then it returns ErrReceiveComplete to indicate the Receive() function can return
	ErrReceiveComplete = errors.New("Completed receiving")

	// ErrReceiveContinue is used when calling the Receive() utility function.  If the callback
	// wants to continue receiving it will return this error.  This causes the Receive() function
	// to update the timeout and wait for a new message
	ErrReceiveContinue = errors.New("Continue receiving")

	// ErrInvalidThermostatMode indicates an unknown mode was supplied to the SetMode function
	ErrInvalidThermostatMode = errors.New("invalid mode")

	// ErrInvalidUnit indicates the given value for Unit is not either Fahrenheit or Celsius
	ErrInvalidUnit = errors.New("Invalid temperature unit")

	// ErrInvalidFanSpeed indicates the value provided for FanSpeed is either unsupported or
	// unknown
	ErrInvalidFanSpeed = errors.New("Invalid fan speed")
)

var sprintf = fmt.Sprintf

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
	return sprintf("0x%02x%02x%02x", p[0], p[1], p[2])
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

// String returns a string representation of the DevCat in the
// form of category.subcategory where those fields are the 2 digit
// hex representation of their corresponding values
func (dc DevCat) String() string {
	return sprintf("%02x.%02x", dc[0], dc[1])
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

// Domain represents an entire domain of similar devices (dimmers, switches, thermostats, etc)
type Domain byte

// Category indicates the specific kind of device within a domain.  For instance, a LampLing and
// a SwitchLinc Dimmer are both within the Dimmable device domain
type Category byte

// MarshalJSON will convert the DevCat to a valid JSON byte string
func (dc DevCat) MarshalJSON() ([]byte, error) {
	return json.Marshal(sprintf("%02x.%02x", dc[0], dc[1]))
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

// ProductData contains information about the device including its
// product key and device category
type ProductData struct {
	Key    ProductKey
	DevCat DevCat
}

// UnmarshalBinary takes the input byte buffer and unmarshals it into the
// ProductData object
func (pd *ProductData) UnmarshalBinary(buf []byte) error {
	if len(buf) < 14 {
		return newBufError(ErrBufferTooShort, 14, len(buf))
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
