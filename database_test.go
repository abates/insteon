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
	"fmt"
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
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v %+v", test.input.DevCat, test.input.FirmwareVersion), func(t *testing.T) {
			if test.input.Complete() != test.expected {
				t.Errorf("got %v, want %v ", test.input.Complete(), test.expected)
			}
		})
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
		desc   string
		update func(*productDatabase)
		test   func(DeviceInfo) bool
	}{
		{"UpdateFirmwareVersion", func(pdb *productDatabase) { pdb.UpdateFirmwareVersion(address, FirmwareVersion(42)) }, func(di DeviceInfo) bool { return di.FirmwareVersion == FirmwareVersion(42) }},
		{"UpdateEngineVersion", func(pdb *productDatabase) { pdb.UpdateEngineVersion(address, EngineVersion(42)) }, func(di DeviceInfo) bool { return di.EngineVersion == EngineVersion(42) }},
		{"UpdateDevCat", func(pdb *productDatabase) { pdb.UpdateDevCat(address, DevCat{42, 42}) }, func(di DeviceInfo) bool { return di.DevCat == DevCat{42, 42} }},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			pdb := NewProductDB().(*productDatabase)
			if _, found := pdb.Find(address); found {
				t.Error("got found, want not found ")
			} else {
				test.update(pdb)
				if deviceInfo, found := pdb.Find(address); found {
					if !test.test(deviceInfo) {
						t.Error("test func failed")
					}
				} else {
					t.Errorf("did not find device for address %s", address)
				}
			}
		})
	}
}
