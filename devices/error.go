package devices

import (
	"errors"
)

var (
	// ErrReadTimeout indicates the timeout period expired while waiting for
	// a specific message
	ErrReadTimeout = errors.New("Read Timeout")

	// ErrNotLinked indicates the device does not have an all-link entry in its
	// database
	ErrNotLinked = errors.New("Not in All-Link group")

	// ErrUnexpectedResponse is returned when a Nak is not understood
	ErrUnexpectedResponse = errors.New("Unexpected response from device")

	// ErrNoLoadDetected is an error returned by the device (this error condition is not documented)
	ErrNoLoadDetected = errors.New("No load detected")

	// ErrIllegalValue is returned by I2Cs devices (this error condition is not documented)
	ErrIllegalValue = errors.New("Illegal value in command")

	// ErrIncorrectChecksum is returned by I2Cs devices when an invalid checksum is detected
	ErrIncorrectChecksum = errors.New("I2CS invalid checksum")

	// ErrLinkIndexOutOfRange indicates that the index exceeds the length of the all-link database
	ErrLinkIndexOutOfRange = errors.New("Link index is beyond the bounds of the link database")

	// ErrInvalidThermostatMode indicates an unknown mode was supplied to the SetMode function
	ErrInvalidThermostatMode = errors.New("invalid mode")

	// ErrInvalidUnit indicates the given value for Unit is not either Fahrenheit or Celsius
	ErrInvalidUnit = errors.New("Invalid temperature unit")

	// ErrInvalidFanSpeed indicates the value provided for FanSpeed is either unsupported or
	// unknown
	ErrInvalidFanSpeed = errors.New("Invalid fan speed")

	// ErrInvalidResponse indicates the device responded in a way that the system
	// doesn't understand
	ErrInvalidResponse = errors.New("Invalid response received")

	// ErrNak indicates a negative acknowledgement was received in response to a sent message
	ErrNak = errors.New("NAK received")
)
