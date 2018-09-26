package insteon

import (
	"sync"
	"testing"
)

type testProductDB struct {
	updates    sync.Map
	deviceInfo *DeviceInfo
}

func newTestProductDB() *testProductDB {
	return &testProductDB{}
}

func (tpd *testProductDB) WasUpdated(key string) bool {
	if v, found := tpd.updates.Load(key); found {
		return v.(bool)
	}
	return false
}

func (tpd *testProductDB) UpdateDevCat(address Address, devCat DevCat) {
	tpd.updates.Store("DevCat", true)
}

func (tpd *testProductDB) UpdateEngineVersion(address Address, engineVersion EngineVersion) {
	tpd.updates.Store("EngineVersion", true)
}

func (tpd *testProductDB) UpdateFirmwareVersion(address Address, firmwareVersion FirmwareVersion) {
	tpd.updates.Store("FirmwareVersion", true)
}

func (tpd *testProductDB) Find(address Address) (deviceInfo DeviceInfo, found bool) {
	if tpd.deviceInfo == nil {
		return DeviceInfo{}, false
	}
	return *tpd.deviceInfo, true
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
