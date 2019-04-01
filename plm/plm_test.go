package plm

import (
	"bufio"
	"bytes"
	"testing"
	"time"
)

func TestPlmOption(t *testing.T) {
	want := 1234 * time.Millisecond
	buf := bytes.NewBuffer(nil)

	without := New(&Port{in: bufio.NewReader(buf), out: buf}, 5*time.Second)
	if without.writeDelay == want {
		t.Errorf("writeDelay is %v, expected anything else", without.writeDelay)
	}

	with := New(&Port{in: bufio.NewReader(buf), out: buf}, 5*time.Second, WriteDelay(want))
	if with.writeDelay != want {
		t.Errorf("writeDelay is %v, want %v", without.writeDelay, want)
	}

}
