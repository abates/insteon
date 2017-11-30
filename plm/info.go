package plm

import "github.com/abates/insteon"

type Version byte

type IMInfo struct {
	Address  insteon.Address
	Category insteon.Category
	Firmware Version
}

func (imi *IMInfo) MarshalBinary() ([]byte, error) {
	data := make([]byte, 6)

	copy(data[0:3], imi.Address[:])
	copy(data[3:5], imi.Category[:])
	data[5] = byte(imi.Firmware)
	return data, nil
}

func (imi *IMInfo) UnmarshalBinary(data []byte) error {
	copy(imi.Address[:], data[0:3])
	copy(imi.Category[:], data[3:5])
	imi.Firmware = Version(data[5])
	return nil
}
