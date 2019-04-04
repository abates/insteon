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

	// ErrWriteTimeout indicates the timeout period expired while waiting to
	// write a message
	ErrWriteTimeout = errors.New("Write Timeout")

	// ErrSendTimeout indicates the timeout period expired while trying to
	// send a message
	ErrSendTimeout = errors.New("Send Timeout")

	// ErrNotImplemented indicates that a device function has not yet been implemented
	ErrNotImplemented = errors.New("Command is not yet implemented")

	// ErrUnexpectedResponse is returned when a Nak is not understood
	ErrUnexpectedResponse = errors.New("Unexpected response from device")

	// ErrNotLinked indicates the device does not have an all-link entry in its
	// database
	ErrNotLinked = errors.New("Not in All-Link group")

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

	// ErrEndOfLinks occurs when unmarshalling a link database record and that record
	// appears to be the last record in the database
	ErrEndOfLinks = errors.New("reached end of ALDB links")

	// ErrInvalidMemAddress indicates a link record memory address is invalid
	ErrInvalidMemAddress = errors.New("Invalid memory address")

	// ErrVersion is returned when an engine version value is not known
	ErrVersion = errors.New("Unknown Insteon Engine Version")
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

// Category returns the Category that the DevCat falls under
func (dc DevCat) Category() Category {
	return Category(dc[0])
}

// SubCategory returns the specific SubCategory for this DevCat
func (dc DevCat) SubCategory() SubCategory {
	return SubCategory(dc[1])
}

// String returns a string representation of the DevCat in the
// form of category.subcategory where those fields are the 2 digit
// hex representation of their corresponding values
func (dc DevCat) String() string {
	return sprintf("%02x.%02x", dc[0], dc[1])
}

// Category is type for the Category byte in the DevCat
type Category byte

// SubCategory is the type for the SubCategory byte in the DevCat
type SubCategory byte

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
