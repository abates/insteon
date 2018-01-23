package insteon

import "testing"

func TestEngineVersionString(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "I1"},
		{1, "I2"},
		{2, "I2CS"},
		{3, "Unknown"},
	}

	for i, test := range tests {
		ver := EngineVersion(test.input)
		if ver.String() != test.expected {
			t.Errorf("tests[%d] expected %q got %q", i, test.expected, ver.String())
		}
	}
}

type TestDevice struct {
	productData    *ProductData
	productDataErr error
}

func (td *TestDevice) Address() Address                      { return Address{0x01, 0x02, 0x03} }
func (td *TestDevice) AssignToAllLinkGroup(Group) error      { return nil }
func (td *TestDevice) DeleteFromAllLinkGroup(Group) error    { return nil }
func (td *TestDevice) EnterLinkingMode(Group) error          { return nil }
func (td *TestDevice) EnterUnlinkingMode(Group) error        { return nil }
func (td *TestDevice) ExitLinkingMode() error                { return nil }
func (td *TestDevice) LinkDB() (LinkDB, error)               { return nil, nil }
func (td *TestDevice) ProductData() (*ProductData, error)    { return td.productData, td.productDataErr }
func (td *TestDevice) FXUsername() (string, error)           { return "", nil }
func (td *TestDevice) TextString() (string, error)           { return "", nil }
func (td *TestDevice) SetTextString(string) error            { return nil }
func (td *TestDevice) EngineVersion() (EngineVersion, error) { return 1, nil }
func (td *TestDevice) IDRequest() (Category, error)          { return td.productData.Category, td.productDataErr }
func (td *TestDevice) Ping() error                           { return nil }
func (td *TestDevice) Close() error                          { return nil }
func (td *TestDevice) Connection() Connection                { return nil }

func TestInitializeDevice(t *testing.T) {
	testPD := func(key, category byte) *ProductData {
		return &ProductData{ProductKey{key, key, key}, Category{category, category}}
	}

	testDevice := func(category byte, pdErr error) Device {
		return &TestDevice{productData: testPD(0x22, category), productDataErr: pdErr}
	}

	tests := []struct {
		category    byte
		initializer DeviceInitializer
		expected    *ProductData
		pdErr       error
	}{
		{0x22, func(Device) Device { return testDevice(0x33, nil) }, testPD(0x22, 0x33), nil},
		{0x22, func(Device) Device { return nil }, testPD(0x22, 0x33), ErrReadTimeout},
	}

	for i, test := range tests {
		dr := &DeviceRegistry{}
		dr.Register(test.category, test.initializer)
		device := testDevice(test.category, test.pdErr)
		lvl := Log.level
		Log.level = LevelNone
		initDevice, err := dr.Initialize(device)
		Log.level = lvl

		if test.pdErr == ErrReadTimeout {
			if err != nil {
				t.Errorf("tests[%d] expected Read Timeout error to have been cleared.  Got a %v", i, err)
			}

			if initDevice != device {
				t.Errorf("tests[%d] expected timeout call to force original device to be returned", i)
			}
		} else {
			if td, ok := initDevice.(*TestDevice); ok {
				if *td.productData != *test.expected {
					t.Errorf("tests[%d] expected %v got %v", i, test.expected, td)
				}
			} else {
				t.Errorf("tests[%d] expected TestDevice got %T", i, initDevice)
			}
		}
	}
}

func TestChecksum(t *testing.T) {
	tests := []struct {
		input    []byte
		expected byte
	}{
		{[]byte{0x2E, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xd1},
		{[]byte{0x2F, 0x00, 0x00, 0x00, 0x0F, 0xFF, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xC2},
		{[]byte{0x2F, 0x00, 0x01, 0x01, 0x0F, 0xFF, 0x00, 0xA2, 0x00, 0x19, 0x70, 0x1A, 0xFF, 0x1F, 0x01}, 0x5D},
		{[]byte{0x2F, 0x00, 0x00, 0x00, 0x0F, 0xF7, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xCA},
		{[]byte{0x2F, 0x00, 0x01, 0x01, 0x0F, 0xF7, 0x00, 0xE2, 0x01, 0x19, 0x70, 0x1A, 0xFF, 0x1F, 0x01}, 0x24},
		{[]byte{0x2F, 0x00, 0x00, 0x00, 0x0F, 0xEF, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xD2},
		{[]byte{0x2F, 0x00, 0x01, 0x01, 0x0F, 0xEF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xD1},
		{[]byte{0x2F, 0x00, 0x01, 0x02, 0x0F, 0xFF, 0x08, 0xE2, 0x01, 0x08, 0xB6, 0xEA, 0x00, 0x1B, 0x01}, 0x11},
		{[]byte{0x09, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xF6},
	}

	for i, test := range tests {
		got := checksum(test.input)
		if got != test.expected {
			t.Errorf("tests[%d] expected %02x got %02d", i, test.expected, got)
		}
	}
}
