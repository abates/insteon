package insteon

import (
	"encoding"
	"fmt"
	"strings"
)

// PayloadGenerator will return an instance of any type that
// implements the Payload interface. PayloadGenerator is used
// when unmarshaling insteon messages
type PayloadGenerator func() Payload

// A Payload is any object that can be marshaled to a byte slice or
// unmarshaled from a byte slice and also includes a String method
// for human readable representation
type Payload interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	String() string
}

// NewBufPayload returns a properly BufPayload instance
func NewBufPayload(size int) *BufPayload {
	return &BufPayload{
		Buf: make([]byte, size),
	}
}

// BufPayload is a generic payload type that will simply copy
// the bytes being unmarshaled into a local byte buffer
type BufPayload struct {
	Buf []byte
}

// MarshalBinary will return the underlying byte buffer
func (bp *BufPayload) MarshalBinary() ([]byte, error) {
	return bp.Buf, nil
}

// UnmarshalBinary will size the underlying buffer appropriately for the
// passed in byte slice. It will then copy the bytes locally. The buf argument
// can safely be changed/reused by the caller since a copy is held by
// BufPayload
func (bp *BufPayload) UnmarshalBinary(buf []byte) error {
	bp.Buf = make([]byte, len(buf))
	copy(bp.Buf, buf)
	return nil
}

// String will return a string representation of the underlying byte buffer
// in the format of "00 01 02 ..." where each item is the hex representation
// of the byte value
func (bp *BufPayload) String() string {
	str := make([]string, len(bp.Buf))
	for i, b := range bp.Buf {
		str[i] = fmt.Sprintf("%02x", b)
	}
	return strings.Join(str, " ")
}
