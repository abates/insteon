package plm

import (
	"bytes"
	"testing"

	"github.com/abates/insteon"
)

func TestInfoMarshalUnmarshalBinary(t *testing.T) {
	tests := []struct {
		desc  string
		input []byte
		want  *Info
	}{
		{"test 1", []byte{1, 2, 3, 4, 5, 6}, &Info{insteon.Address(0x010203), insteon.DevCat{4, 5}, 6}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := &Info{}
			err := got.UnmarshalBinary(test.input)
			if err == nil {
				if *test.want != *got {
					t.Errorf("Wanted info %+v got %+v", test.want, got)
				}

				buf, _ := got.MarshalBinary()
				if !bytes.Equal(test.input, buf) {
					t.Errorf("Wanted bytes %x got %x", test.input, buf)
				}
			} else {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
