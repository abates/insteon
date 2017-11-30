package insteon

import "fmt"

type Address [3]byte

func (a Address) String() string { return fmt.Sprintf("%02x.%02x.%02x", a[0], a[1], a[2]) }

func ParseAddress(str string) (Address, error) {
	var b1, b2, b3 byte
	_, err := fmt.Sscanf(str, "%2x.%2x.%2x", &b1, &b2, &b3)
	if err == nil {
		return Address([3]byte{b1, b2, b3}), nil
	}
	return Address([3]byte{}), fmt.Errorf("address format is xx.xx.xx (digits in hex)")
}
