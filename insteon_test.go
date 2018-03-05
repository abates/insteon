package insteon

import (
	"bytes"
	"reflect"
	"testing"
)

func testStr(stringer func() string, expectedFmts []string, expectedArgs []interface{}) (msg string, failed bool) {
	var gotFmts []string
	var gotArgs []interface{}

	oldSprintf := sprintf
	sprintf = func(format string, a ...interface{}) string {
		gotFmts = append(gotFmts, format)
		gotArgs = append(gotArgs, a...)
		return ""
	}
	stringer()
	if !reflect.DeepEqual(expectedFmts, gotFmts) {
		msg = oldSprintf("expected %q got %q", expectedFmts, gotFmts)
		failed = true
	} else if !reflect.DeepEqual(expectedArgs, gotArgs) {
		if len(expectedArgs) != len(gotArgs) {
			msg = oldSprintf("expected %d arguments but got %d", len(expectedArgs), len(gotArgs))
			failed = true
		} else {
			for i, expectedArg := range expectedArgs {
				if expectedArg != gotArgs[i] {
					msg = oldSprintf("expected '%v' got '%v'", expectedArg, gotArgs[i])
					failed = true
				}
			}
		}
	}
	sprintf = oldSprintf
	return
}

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

func TestProductKey(t *testing.T) {
	tests := []struct {
		input          [3]byte
		expectedString string
	}{
		{[3]byte{0x01, 0x02, 0x03}, "0x010203"},
	}

	for i, test := range tests {
		key := ProductKey(test.input)
		if key.String() != test.expectedString {
			t.Errorf("tests[%d] expectdd %q got %q", i, test.expectedString, key.String())
		}
	}
}

func TestDevCat(t *testing.T) {
	tests := []struct {
		input               [2]byte
		expectedCategory    Category
		expectedSubCategory SubCategory
		expectedString      string
	}{
		{[2]byte{0x01, 0x02}, Category(0x01), SubCategory(0x02), "01.02"},
	}

	for i, test := range tests {
		devCat := DevCat(test.input)
		if devCat.Category() != test.expectedCategory {
			t.Errorf("tests[%d] expected 0x%02x got 0x%02x", i, test.expectedCategory, devCat.Category())
		}

		if devCat.SubCategory() != test.expectedSubCategory {
			t.Errorf("tests[%d] expected 0x%02x got 0x%02x", i, test.expectedSubCategory, devCat.SubCategory())
		}

		if devCat.String() != test.expectedString {
			t.Errorf("tests[%d] expected %q got %q", i, test.expectedString, devCat.String())
		}
	}
}

func TestDevCatMarshaling(t *testing.T) {
	tests := []struct {
		input          string
		expectedDevCat DevCat
		expectedJSON   string
		expectedError  string
	}{
		{"\"01.02\"", DevCat{1, 2}, "\"01.02\"", ""},
		{"\"01\"", DevCat{0, 0}, "", "Expected Scanf to parse 2 digits, got 1"},
	}

	for i, test := range tests {
		var devCat DevCat
		err := devCat.UnmarshalJSON([]byte(test.input))
		if err == nil {
			if devCat != test.expectedDevCat {
				t.Errorf("tests[%d] expected %q got %q", i, test.expectedDevCat, devCat)
			} else {
				data, _ := devCat.MarshalJSON()
				if string(data) != test.expectedJSON {
					t.Errorf("tests[%d] expected %q got %q", i, test.expectedJSON, string(data))
				}
			}
		} else if err.Error() != test.expectedError {
			t.Errorf("tests[%d] expected error %v got %v", i, test.expectedError, err)
		}
	}
}

func TestProductDataString(t *testing.T) {
	tests := []struct {
		key            [3]byte
		devCat         [2]byte
		expectedString string
	}{
		{[3]byte{0x01, 0x02, 0x03}, [2]byte{0x04, 0x05}, "DevCat:04.05 Product Key:0x010203"},
	}

	for i, test := range tests {
		pd := &ProductData{ProductKey(test.key), DevCat(test.devCat)}
		if pd.String() != test.expectedString {
			t.Errorf("tests[%d] expected %q got %q", i, test.expectedString, pd.String())
		}
	}
}

func TestProductDataMarshaling(t *testing.T) {
	tests := []struct {
		input          []byte
		expectedDevCat [2]byte
		expectedKey    [3]byte
		expectedError  error
	}{
		{[]byte{0, 1, 2, 3, 4, 5, 255, 0, 0, 0, 0, 0, 0, 0}, [2]byte{4, 5}, [3]byte{1, 2, 3}, nil},
		{[]byte{}, [2]byte{0, 0}, [3]byte{0, 0, 0}, ErrBufferTooShort},
	}

	for i, test := range tests {
		pd := &ProductData{}
		err := pd.UnmarshalBinary(test.input)
		if !IsError(test.expectedError, err) {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedError, err)
		}

		if err == nil {
			if pd.Key != ProductKey(test.expectedKey) {
				t.Errorf("tests[%d] expected %x got %x", i, test.expectedKey, pd.Key)
			}

			if pd.DevCat != DevCat(test.expectedDevCat) {
				t.Errorf("tests[%d] expected %x got %x", i, test.expectedDevCat, pd.DevCat)
			}

			buf, _ := pd.MarshalBinary()
			if !bytes.Equal(buf, test.input[0:7]) {
				t.Errorf("tests[%d] expected %x got %x", i, test.input[0:7], buf)
			}
		}
	}
}
