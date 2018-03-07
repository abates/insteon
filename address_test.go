package insteon

import "testing"

func TestAddressUnmarshalText(t *testing.T) {
	tests := []struct {
		input   string
		correct bool
	}{
		{"01.02.03", true},
		{"abcd", false},
		{"01b.02.03", false},
	}

	for i, test := range tests {
		address := Address{}
		err := address.UnmarshalText([]byte(test.input))
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

func TestAddressMarshaling(t *testing.T) {
	tests := []struct {
		input           string
		expectedAddress Address
		expectedJSON    string
		expectedError   string
	}{
		{"\"01.02.03\"", Address{1, 2, 3}, "\"01.02.03\"", ""},
		{"\"01.02\"", Address{0, 0, 0}, "", "Expected Scanf to parse 3 digits, got 2"},
	}

	for i, test := range tests {
		var address Address
		err := address.UnmarshalJSON([]byte(test.input))
		if err == nil {
			if address != test.expectedAddress {
				t.Errorf("tests[%d] expected %q got %q", i, test.expectedAddress, address)
			} else {
				data, _ := address.MarshalJSON()
				if string(data) != test.expectedJSON {
					t.Errorf("tests[%d] expected %q got %q", i, test.expectedJSON, string(data))
				}
			}
		} else if err.Error() != test.expectedError {
			t.Errorf("tests[%d] expected error %v got %v", i, test.expectedError, err)
		}
	}
}
