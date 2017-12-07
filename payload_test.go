package insteon

import "testing"

func TestNewBufPayload(t *testing.T) {
	p := NewBufPayload(3141)
	if len(p.Buf) != 3141 {
		t.Errorf("test expected %d got %d", 3141, len(p.Buf))
	}
}
