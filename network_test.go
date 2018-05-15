package insteon

import (
	"reflect"
	"testing"
)

type testProductDB struct {
	updates    map[string]bool
	deviceInfo *DeviceInfo
}

func newTestProductDB() *testProductDB {
	return &testProductDB{updates: make(map[string]bool)}
}

func (tpd *testProductDB) WasUpdated(key string) bool {
	return tpd.updates[key]
}

func (tpd *testProductDB) UpdateDevCat(address Address, devCat DevCat) {
	tpd.updates["DevCat"] = true
}

func (tpd *testProductDB) UpdateEngineVersion(address Address, engineVersion EngineVersion) {
	tpd.updates["EngineVersion"] = true
}

func (tpd *testProductDB) UpdateFirmwareVersion(address Address, firmwareVersion FirmwareVersion) {
	tpd.updates["FirmwareVersion"] = true
}

func (tpd *testProductDB) Find(address Address) (deviceInfo DeviceInfo, found bool) {
	if tpd.deviceInfo == nil {
		return DeviceInfo{}, false
	}
	return *tpd.deviceInfo, true
}

type testBridge struct {
	sendError error
}

func (tb *testBridge) SendMessage(*Message) error {
	return tb.sendError
}

func TestProductDatabaseUpdateFind(t *testing.T) {
	address := Address{0, 1, 2}
	tests := []struct {
		update func(*productDatabase)
		test   func(DeviceInfo) bool
	}{
		{func(pdb *productDatabase) { pdb.UpdateFirmwareVersion(address, FirmwareVersion(42)) }, func(di DeviceInfo) bool { return di.FirmwareVersion == FirmwareVersion(42) }},
		{func(pdb *productDatabase) { pdb.UpdateEngineVersion(address, EngineVersion(42)) }, func(di DeviceInfo) bool { return di.EngineVersion == EngineVersion(42) }},
		{func(pdb *productDatabase) { pdb.UpdateDevCat(address, DevCat{42, 42}) }, func(di DeviceInfo) bool { return di.DevCat == DevCat{42, 42} }},
	}

	for i, test := range tests {
		pdb := NewProductDB().(*productDatabase)
		if _, found := pdb.Find(address); found {
			t.Errorf("tests[%d] expected not found but got found", i)
		} else {
			test.update(pdb)
			if deviceInfo, found := pdb.Find(address); found {
				if !test.test(deviceInfo) {
					t.Errorf("tests[%d] failed", i)
				}
			} else {
				t.Errorf("tests[%d] did not find device for address %s", i, address)
			}
		}
	}
}

type testNetwork struct {
	sentMessages []*Message
}

func (tn *testNetwork) Dial(Address) (Device, error) {
	return nil, nil
}

func (tn *testNetwork) Connect(Address) (Device, error) {
	return nil, nil
}

func (tn *testNetwork) Notify([]byte) error {
	return nil
}

func (tn *testNetwork) SendMessage(msg *Message) error {
	tn.sentMessages = append(tn.sentMessages, msg)
	return nil
}

func TestNetworkDecode(t *testing.T) {
	tests := []struct {
		input         []byte
		expectedError bool
	}{
		{[]byte{0x0, 0x00}, true},
		{[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x00, 0x00, 0x01}, false},
		{[]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x10, 0x00, 0x01, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, false},
	}

	for i, test := range tests {
		network := New(nil)
		err := network.Notify(test.input)
		if err == nil && test.expectedError {
			t.Errorf("tests[%d] expected error got nil", i)
		} else if err != nil && !test.expectedError {
			t.Errorf("tests[%d] expected no error got %v", i, err)
		}
	}
}

func TestNetworkProcess(t *testing.T) {
	tests := []struct {
		input                  *Message
		expectedUpdates        []string
		expectedEngineQueue    bool
		expectedDeviceDelivery bool
	}{
		{TestMessageSetButtonPressedController, []string{"FirmwareVersion", "DevCat"}, false, false},
		{TestMessageEngineVersionAck, []string{"EngineVersion"}, true, false},
		{TestMessagePingAck, nil, false, true},
	}

	for i, test := range tests {
		testDb := newTestProductDB()
		engineVersionCh := make(chan *Message, 1)
		device := &TestDevice{}
		network := &NetworkImpl{
			ProductDatabase: testDb,
			idRequestCh:     make(chan *Message, 1),
			engineVersionCh: engineVersionCh,
			devices:         make(map[Address]Device),
		}

		if test.expectedDeviceDelivery {
			network.devices[test.input.Src] = device
		}

		if len(engineVersionCh) > 0 || len(device.messages) > 0 {
			t.Errorf("tests[%d] Expected EngineVersionCh and Device to be empty", i)
		}

		network.process(test.input)

		for _, update := range test.expectedUpdates {
			if !testDb.WasUpdated(update) {
				t.Errorf("tests[%d] expected %v to be updated in the database", i, update)
			}
		}

		if test.expectedDeviceDelivery && len(device.messages) == 0 {
			t.Errorf("tests[%d] Expected device to receive message", i)
		}

		if test.expectedEngineQueue && len(engineVersionCh) == 0 {
			t.Errorf("tests[%d] Expected EngineVersionCh not to be empty", i)
		}
	}
}

func TestNetworkSendMessage(t *testing.T) {
	tests := []struct {
		input      *Message
		deviceInfo DeviceInfo
	}{
		{TestMessagePing, DeviceInfo{EngineVersion: VerI1}},
		{TestMessagePing, DeviceInfo{EngineVersion: VerI2}},
		{TestMessagePing, DeviceInfo{EngineVersion: VerI2Cs}},
	}

	for i, test := range tests {
		testDb := newTestProductDB()
		testDb.deviceInfo = &test.deviceInfo
		bridge := &testBridge{}

		network := &NetworkImpl{
			ProductDatabase: testDb,
			bridge:          bridge,
		}

		network.SendMessage(test.input)

		if test.input.version != test.deviceInfo.EngineVersion {
			t.Errorf("tests[%d] expected %v got %v", i, test.deviceInfo.EngineVersion, test.input.version)
		}
	}
}

func TestNetworkEngineVersion(t *testing.T) {
	testDb := newTestProductDB()
	engineVersionCh := make(chan *Message, 1)
	msg := *TestMessageEngineVersionAck
	msg.Command[1] = 0x02

	engineVersionCh <- &msg

	bridge := &testBridge{}

	network := &NetworkImpl{
		ProductDatabase: testDb,
		bridge:          bridge,
		engineVersionCh: engineVersionCh,
	}

	version, err := network.EngineVersion(Address{1, 2, 3})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if version != EngineVersion(msg.Command[1]) {
		t.Errorf("Expected %v got %v", EngineVersion(msg.Command[1]), version)
	}
}

func TestNetworkIDRequest(t *testing.T) {
	testDb := newTestProductDB()
	idRequestCh := make(chan *Message, 1)
	msg := *TestMessageSetButtonPressedController
	msg.Dst = Address{2, 3, 4}

	idRequestCh <- &msg

	bridge := &testBridge{}

	network := &NetworkImpl{
		ProductDatabase: testDb,
		bridge:          bridge,
		idRequestCh:     idRequestCh,
	}

	firmwareVersion, devCat, err := network.IDRequest(Address{1, 2, 3})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if firmwareVersion != FirmwareVersion(4) {
		t.Errorf("Expected %v got %v", FirmwareVersion(4), firmwareVersion)
	}

	expected := DevCat{2, 3}
	if devCat != expected {
		t.Errorf("Expected %v got %v", expected, devCat)
	}
}

func TestNetworkDial(t *testing.T) {
	tests := []struct {
		deviceInfo    *DeviceInfo
		engineVersion byte
		bridgeError   error
		expected      interface{}
	}{
		{&DeviceInfo{EngineVersion: VerI1}, 0, nil, &I1Device{}},
		{&DeviceInfo{EngineVersion: VerI2}, 0, nil, &I2Device{}},
		{&DeviceInfo{EngineVersion: VerI2Cs}, 0, nil, &I2CsDevice{}},
		{nil, 0, nil, &I1Device{}},
		{nil, 1, nil, &I2Device{}},
		{nil, 2, nil, &I2CsDevice{}},
		{nil, 0, ErrNotLinked, &I2CsDevice{}},
	}

	for i, test := range tests {
		testDb := newTestProductDB()

		bridge := &testBridge{sendError: test.bridgeError}

		engineVersionCh := make(chan *Message, 1)
		if test.deviceInfo == nil {
			msg := *TestMessageEngineVersionAck
			msg.Command[1] = test.engineVersion
			engineVersionCh <- &msg
		} else {
			testDb.deviceInfo = test.deviceInfo
		}

		network := &NetworkImpl{
			ProductDatabase: testDb,
			bridge:          bridge,
			engineVersionCh: engineVersionCh,
			devices:         make(map[Address]Device),
		}

		device, _ := network.Dial(Address{1, 2, 3})

		if reflect.TypeOf(device) != reflect.TypeOf(test.expected) {
			t.Fatalf("tests[%d] expected type %T got type %T", i, test.expected, device)
		}
	}
}

func TestNetworkConnect(t *testing.T) {
	tests := []struct {
		deviceInfo    *DeviceInfo
		engineVersion EngineVersion
		dst           Address
		expected      Device
	}{
		{&DeviceInfo{DevCat: DevCat{42, 2}}, VerI1, Address{}, &I2Device{}},
		{nil, VerI1, Address{42, 2, 3}, &I2Device{}},
	}

	for i, test := range tests {
		var category Category
		testDb := newTestProductDB()
		bridge := &testBridge{}
		engineVersionCh := make(chan *Message, 1)
		idRequestCh := make(chan *Message, 1)

		if test.deviceInfo == nil {
			msg := *TestMessageEngineVersionAck
			msg.Command[1] = byte(test.engineVersion)
			engineVersionCh <- &msg

			msg = *TestMessageSetButtonPressedController
			msg.Dst = test.dst
			idRequestCh <- &msg
			category = Category(test.dst[0])
		} else {
			testDb.deviceInfo = test.deviceInfo
			category = test.deviceInfo.DevCat.Category()
		}
		Devices.Register(category, func(Device, DeviceInfo) Device { return test.expected })

		network := &NetworkImpl{
			ProductDatabase: testDb,
			bridge:          bridge,
			engineVersionCh: engineVersionCh,
			idRequestCh:     idRequestCh,
			devices:         make(map[Address]Device),
		}

		device, _ := network.Connect(Address{1, 2, 3})

		if device != test.expected {
			t.Fatalf("tests[%d] expected %v got %v", i, test.expected, device)
		}
		Devices.Delete(category)
	}
}
