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

	without, err := New(bufio.NewReader(buf), buf, 5*time.Second)
	if err != nil {
		t.Errorf("unexpected error from plm.New(): %v", err)
	}
	if without.writeDelay == want {
		t.Errorf("writeDelay is %v, expected anything else", without.writeDelay)
	}

	with, err := New(bufio.NewReader(buf), buf, 5*time.Second, WriteDelay(want))
	if err != nil {
		t.Errorf("unexpected error from plm.New(): %v", err)
	}

	if with.writeDelay != want {
		t.Errorf("writeDelay is %v, want %v", without.writeDelay, want)
	}

}
