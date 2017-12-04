package insteon

import (
	"encoding"
	"fmt"
	"strings"
)

type PayloadGenerator func() Payload
type Payload interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	String() string
}

func NewBufPayload(size int) *BufPayload {
	return &BufPayload{
		Buf: make([]byte, size),
	}
}

type BufPayload struct {
	Buf []byte
}

func (bp *BufPayload) MarshalBinary() ([]byte, error) {
	return bp.Buf, nil
}

func (bp *BufPayload) UnmarshalBinary(buf []byte) error {
	bp.Buf = make([]byte, len(buf))
	copy(bp.Buf, buf)
	return nil
}

func (bp *BufPayload) String() string {
	str := make([]string, len(bp.Buf))
	for i, b := range bp.Buf {
		str[i] = fmt.Sprintf("%02x", b)
	}
	return strings.Join(str, " ")
}
