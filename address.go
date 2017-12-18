package insteon

import "fmt"

// Address is a 3 byte insteon address
type Address [3]byte

// String returns a human readable version of the address
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
