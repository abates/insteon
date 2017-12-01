package insteon

import (
	"errors"
	"fmt"
)

var (
	ErrBufferTooShort         = errors.New("Buffer is too short to be an Insteon Message")
	ErrExtendedBufferTooShort = errors.New("Buffer is too short to be an Insteon Extended Message")
	ErrReadTimeout            = errors.New("Read Timeout")
	ErrWriteTimeout           = errors.New("Write Timeout")
	ErrAckTimeout             = errors.New("Timeout waiting for ACK")
	ErrNotImplemented         = errors.New("Command is not yet implemented")
	ErrUnexpectedResponse     = errors.New("Unexpected response from device")
	ErrNotLinked              = errors.New("Not in All-Link group")
	ErrNoLoadDetected         = errors.New("No load detected")
	ErrUnknownCommand         = errors.New("Unknown command")
)

type ProductKey [3]byte

func (p ProductKey) String() string {
	return fmt.Sprintf("0x%02x%02x%02x", p[0], p[1], p[2])
}

type Category [2]byte

func (c Category) Category() byte {
	return c[0]
}

func (c Category) SubCategory() byte {
	return c[1]
}

func (c Category) String() string {
	return fmt.Sprintf("%02x.%02x", c[0], c[1])
}

type ProductData struct {
	Key      ProductKey
	Category Category
}

func (pd *ProductData) String() string {
	return fmt.Sprintf("Category:%s Product Key:%s", pd.Category, pd.Key)
}

func (pd *ProductData) UnmarshalBinary(buf []byte) error {
	if len(buf) < 14 {
		return ErrExtendedBufferTooShort
	}

	copy(pd.Key[:], buf[1:4])
	copy(pd.Category[:], buf[4:6])
	return nil
}

func (pd *ProductData) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 7)
	copy(buf[1:4], pd.Key[:])
	copy(buf[4:6], pd.Category[:])
	buf[6] = 0xff
	return buf, nil
}
