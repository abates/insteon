package plm

import (
	"bytes"
	"testing"
	"time"
)

func TestOptions(t *testing.T) {
	want := 1234 * time.Millisecond

	without := New(&bytes.Buffer{})
	if without.writeDelay != 0 {
		t.Errorf("writeDelay is %v, expected %v", without.writeDelay, time.Duration(0))
	}

	with := New(&bytes.Buffer{}, WriteDelay(want))
	if with.writeDelay != want {
		t.Errorf("writeDelay is %v, want %v", without.writeDelay, want)
	}

	without = New(&bytes.Buffer{})
	if without.timeout != (time.Second * 3) {
		t.Errorf("timeout is %v, expected %v", without.timeout, time.Second*3)
	}

	with = New(&bytes.Buffer{}, Timeout(want))
	if with.timeout != want {
		t.Errorf("timeout is %v, expected %v", with.timeout, want)
	}
}
