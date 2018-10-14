// Copyright 2018 Andrew Bates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package insteon

import (
	"sync"
	"testing"
)

func TestDeviceInfoComplete(t *testing.T) {
	tests := []struct {
		input    DeviceInfo
		expected bool
	}{
		{DeviceInfo{DevCat: DevCat{0x00, 0x00}, FirmwareVersion: FirmwareVersion(0)}, false},
		{DeviceInfo{DevCat: DevCat{0x00, 0x01}, FirmwareVersion: FirmwareVersion(0)}, false},
		{DeviceInfo{DevCat: DevCat{0x00, 0x01}, FirmwareVersion: FirmwareVersion(1)}, true},
		{DeviceInfo{DevCat: DevCat{0x00, 0x01}, FirmwareVersion: FirmwareVersion(0)}, false},
	}

	for i, test := range tests {
		if test.input.Complete() != test.expected {
			t.Errorf("tests[%d] expected %v got %v", i, test.expected, test.input.Complete())
		}
	}
}

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
