package plm

import (
	"fmt"

	"github.com/abates/insteon"
)

type Version byte

func (v Version) String() string { return fmt.Sprintf("%d", byte(v)) }

type Info struct {
	Address  insteon.Address
	Category insteon.Category
	Firmware Version
}

func (info *Info) String() string {
	return fmt.Sprintf("%s category %s version %s", info.Address, info.Category, info.Firmware)
}

func (info *Info) MarshalBinary() ([]byte, error) {
	data := make([]byte, 6)

	copy(data[0:3], info.Address[:])
	copy(data[3:5], info.Category[:])
	data[5] = byte(info.Firmware)
	return data, nil
}

func (info *Info) UnmarshalBinary(data []byte) error {
	copy(info.Address[:], data[0:3])
	copy(info.Category[:], data[3:5])
	info.Firmware = Version(data[5])
	return nil
}
