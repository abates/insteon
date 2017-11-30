package insteon

import "testing"

func TestParseAddress(t *testing.T) {
	tests := []struct {
		input   string
		correct bool
	}{
		{"01.02.03", true},
		{"abcd", false},
		{"01b.02.03", false},
	}

	for i, test := range tests {
		address, err := ParseAddress(test.input)
		if err == nil && !test.correct {
			t.Errorf("tests[%d] expected failure", i)
		} else if err != nil && test.correct {
			t.Errorf("tests[%d] failed: %v", i, err)
		}

		if test.correct && test.input != address.String() {
			t.Errorf("tests[%d] expected %q got %q", i, test.input, address.String())
		}
	}
}
