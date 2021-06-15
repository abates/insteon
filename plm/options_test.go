package plm

import (
	"bytes"
	"testing"
	"time"
)

func TestOptions(t *testing.T) {
	want := 1234 * time.Millisecond
	reader := bytes.NewReader(nil)
	writer := bytes.NewBuffer(nil)

	without, err := New(reader, writer, 5*time.Second)
	if err != nil {
		t.Errorf("unexpected error from plm.New(): %v", err)
	}

	if without.writeDelay == want {
		t.Errorf("writeDelay is %v, expected anything else", without.writeDelay)
	}

	with, err := New(reader, writer, 5*time.Second, WriteDelay(want))
	if err != nil {
		t.Errorf("unexpected error from plm.New(): %v", err)
	}

	if with.writeDelay != want {
		t.Errorf("writeDelay is %v, want %v", without.writeDelay, want)
	}

}
