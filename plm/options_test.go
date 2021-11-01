package plm

import (
	"bytes"
	"testing"
	"time"
)

func TestOptions(t *testing.T) {
	want := 1234 * time.Millisecond

	without, err := New(&bytes.Buffer{})
	if err != nil {
		t.Errorf("unexpected error from plm.New(): %v", err)
	}

	if without.writeDelay != 0 {
		t.Errorf("writeDelay is %v, expected %v", without.writeDelay, time.Duration(0))
	}

	with, err := New(&bytes.Buffer{}, WriteDelay(want))
	if err != nil {
		t.Errorf("unexpected error from plm.New(): %v", err)
	}

	if with.writeDelay != want {
		t.Errorf("writeDelay is %v, want %v", without.writeDelay, want)
	}

}
