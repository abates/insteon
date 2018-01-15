package insteon

import "fmt"

// Address is a 3 byte insteon address
type Address [3]byte

// String will format the Address object into a form
// common to Insteon devices: 00.00.00 where each byte
// is represented in hexadecimal form (e.g. 01.b4.a5) the
// string will always be 8 characters long, bytes are zero
// padded
func (a Address) String() string { return fmt.Sprintf("%02x.%02x.%02x", a[0], a[1], a[2]) }

// ParseAddress converts a human readable string into an
// Insteon address. If the address cannot be parsed then
// ParseAddress returns an ErrAddressFormat error
func ParseAddress(str string) (Address, error) {
	var b1, b2, b3 byte
	_, err := fmt.Sscanf(str, "%2x.%2x.%2x", &b1, &b2, &b3)
	if err == nil {
		return Address([3]byte{b1, b2, b3}), nil
	}
	return Address([3]byte{}), ErrAddrFormat
}
