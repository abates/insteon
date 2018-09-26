package insteon

import (
	"bytes"
	"testing"
)

func init() {
	// turn off logging for tests
	Log.Level(LevelNone)
}

func TestFirmwareVersionString(t *testing.T) {
	ver := FirmwareVersion(0x42)
	if ver.String() != "0x42" {
		t.Errorf("expected %q got %q", "0x42", ver.String())
	}
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
		if !IsError(err, test.expectedError) {
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

/*func TestWriteToCh(t *testing.T) {
	tests := []struct {
		ch          chan *Message
		expectedErr error
	}{
		{make(chan *Message, 1), nil},
		{make(chan *Message), ErrWriteTimeout},
	}

	timeout := Timeout
	Timeout = time.Second
	for i, test := range tests {
		err := writeToCh(test.ch, &Message{})
		if err != test.expectedErr {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedErr, err)
		}
	}
	Timeout = timeout
}

func TestReadFromCh(t *testing.T) {
	fullCh := make(chan *Message, 1)
	fullCh <- &Message{}

	tests := []struct {
		ch          chan *Message
		expectedErr error
	}{
		{fullCh, nil},
		{make(chan *Message), ErrReadTimeout},
	}

	timeout := Timeout
	Timeout = time.Second
	for i, test := range tests {
		_, err := readFromCh(test.ch)
		if err != test.expectedErr {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedErr, err)
		}

	}
	Timeout = timeout
}*/
