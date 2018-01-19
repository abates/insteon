package insteon

import (
	"bytes"
	"testing"
)

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

func TestCategory(t *testing.T) {
	tests := []struct {
		input               [2]byte
		expectedCategory    byte
		expectedSubCategory byte
		expectedString      string
	}{
		{[2]byte{0x01, 0x02}, 0x01, 0x02, "01.02"},
	}

	for i, test := range tests {
		category := Category(test.input)
		if category.Category() != test.expectedCategory {
			t.Errorf("tests[%d] expected 0x%02x got 0x%02x", i, test.expectedCategory, category.Category())
		}

		if category.SubCategory() != test.expectedSubCategory {
			t.Errorf("tests[%d] expected 0x%02x got 0x%02x", i, test.expectedSubCategory, category.SubCategory())
		}

		if category.String() != test.expectedString {
			t.Errorf("tests[%d] expected %q got %q", i, test.expectedString, category.String())
		}
	}
}

func TestProductDataString(t *testing.T) {
	tests := []struct {
		key            [3]byte
		category       [2]byte
		expectedString string
	}{
		{[3]byte{0x01, 0x02, 0x03}, [2]byte{0x04, 0x05}, "Category:04.05 Product Key:0x010203"},
	}

	for i, test := range tests {
		pd := &ProductData{ProductKey(test.key), Category(test.category)}
		if pd.String() != test.expectedString {
			t.Errorf("tests[%d] expected %q got %q", i, test.expectedString, pd.String())
		}
	}
}

func TestProductDataMarshaling(t *testing.T) {
	tests := []struct {
		input            []byte
		expectedCategory [2]byte
		expectedKey      [3]byte
		expectedError    error
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

			if pd.Category != Category(test.expectedCategory) {
				t.Errorf("tests[%d] expected %x got %x", i, test.expectedCategory, pd.Category)
			}

			buf, _ := pd.MarshalBinary()
			if !bytes.Equal(buf, test.input[0:7]) {
				t.Errorf("tests[%d] expected %x got %x", i, test.input[0:7], buf)
			}
		}
	}
}
