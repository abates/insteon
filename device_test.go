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
		{3, "Unknown(3)"},
	}

	for i, test := range tests {
		ver := EngineVersion(test.input)
		if ver.String() != test.expected {
			t.Errorf("tests[%d] expected %q got %q", i, test.expected, ver.String())
		}
	}
}

type TestDevice struct {
	testConnection
	productData     *ProductData
	productDataErr  error
	firmwareVersion FirmwareVersion
}

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
func (td *TestDevice) Ping() error                           { return nil }
func (td *TestDevice) Close() error                          { return nil }
func (td *TestDevice) Connection() Connection                { return nil }
func (td *TestDevice) DevCat() DevCat                        { return DevCat{} }

func TestInitializeDevice(t *testing.T) {
	testPD := func(key byte, category Category) *ProductData {
		return &ProductData{ProductKey{key, key, key}, DevCat{byte(category), byte(category)}}
	}

	testDevice := func(category Category, pdErr error) Device {
		return &TestDevice{productData: testPD(0x22, category), productDataErr: pdErr}
	}

	tests := []struct {
		category    Category
		initializer DeviceInitializer
		expected    *ProductData
		pdErr       error
	}{
		{0x33, func(Device) Device { return testDevice(0x33, nil) }, testPD(0x22, 0x33), nil},
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
					t.Errorf("tests[%d] expected %v got %v", i, test.expected, td.productData)
				}
			} else {
				t.Errorf("tests[%d] expected TestDevice got %T", i, initDevice)
			}
		}
	}
}
