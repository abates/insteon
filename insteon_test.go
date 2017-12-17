package insteon

import "testing"

func TestAddress(t *testing.T) {
	tests := []struct {
		input [3]byte
		str   string
	}{
		{[3]byte{0x47, 0x2d, 0x10}, "47.2d.10"},
	}

	for i, test := range tests {
		address := Address(test.input)

		if address.String() != test.str {
			t.Errorf("tests[%d] expected %q got %q", i, test.str, address.String())
		}
	}
}
