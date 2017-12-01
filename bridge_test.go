package insteon

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestBufPayload(t *testing.T) {
	testBuf := []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21}
	bp := &BufPayload{}

	bp.UnmarshalBinary(testBuf)
	if !reflect.DeepEqual(testBuf, bp.Buf) {
		t.Errorf("Expected %x got %x", testBuf, bp.Buf)
	}

	buf, _ := bp.MarshalBinary()

	if !reflect.DeepEqual(buf, testBuf) {
		t.Errorf("Expected %x got %x", testBuf, bp.Buf)
	}

	var values []string
	for _, b := range testBuf {
		values = append(values, fmt.Sprintf("%x", b))
	}
	str := strings.Join(values, " ")
	if str != bp.String() {
		t.Errorf("Expected %s got %s", str, bp.String())
	}
}
