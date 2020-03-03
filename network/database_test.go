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

package network

import (
	"testing"

	"github.com/abates/insteon"
)

/*func TestDeviceInfoComplete(t *testing.T) {
	tests := []struct {
		input    insteon.DeviceInfo
		expected bool
	}{
		{insteon.DeviceInfo{DevCat: insteon.DevCat{0x00, 0x00}, FirmwareVersion: insteon.FirmwareVersion(0)}, false},
		{insteon.DeviceInfo{DevCat: insteon.DevCat{0x00, 0x01}, FirmwareVersion: insteon.FirmwareVersion(0)}, false},
		{insteon.DeviceInfo{DevCat: insteon.DevCat{0x00, 0x01}, FirmwareVersion: insteon.FirmwareVersion(1)}, true},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v %+v", test.input.DevCat, test.input.FirmwareVersion), func(t *testing.T) {
			if test.input.Complete() != test.expected {
				t.Errorf("got %v, want %v ", test.input.Complete(), test.expected)
			}
		})
	}
}*/

/*type testProductDB struct {
	updates    sync.Map
	deviceInfo *insteon.DeviceInfo
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

func (tpd *testProductDB) UpdateDevCat(address insteon.Address, devCat insteon.DevCat) {
	tpd.updates.Store("DevCat", true)
}

func (tpd *testProductDB) UpdateEngineVersion(address insteon.Address, engineVersion insteon.EngineVersion) {
	tpd.updates.Store("EngineVersion", true)
}

func (tpd *testProductDB) UpdateFirmwareVersion(address insteon.Address, firmwareVersion insteon.FirmwareVersion) {
	tpd.updates.Store("FirmwareVersion", true)
}

func (tpd *testProductDB) Find(address insteon.Address) (deviceInfo insteon.DeviceInfo, found bool) {
	if tpd.deviceInfo == nil {
		return insteon.DeviceInfo{}, false
	}
	return *tpd.deviceInfo, true
}*/

func TestProductDatabaseUpdateFind(t *testing.T) {
	address := insteon.Address{0, 1, 2}
	tests := []struct {
		desc   string
		update func(*productDatabase)
		test   func(insteon.DeviceInfo) bool
	}{
		{"UpdateFirmwareVersion", func(pdb *productDatabase) {
			pdb.Update(address, func(info *insteon.DeviceInfo) { info.FirmwareVersion = insteon.FirmwareVersion(42) })
		}, func(di insteon.DeviceInfo) bool { return di.FirmwareVersion == insteon.FirmwareVersion(42) }},
		{"UpdateEngineVersion", func(pdb *productDatabase) {
			pdb.Update(address, func(info *insteon.DeviceInfo) { info.EngineVersion = insteon.EngineVersion(42) })
		}, func(di insteon.DeviceInfo) bool { return di.EngineVersion == insteon.EngineVersion(42) }},
		{"UpdateDevCat", func(pdb *productDatabase) {
			pdb.Update(address, func(info *insteon.DeviceInfo) { info.DevCat = insteon.DevCat{42, 42} })
		}, func(di insteon.DeviceInfo) bool { return di.DevCat == insteon.DevCat{42, 42} }},
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
