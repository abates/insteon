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
	"sync"

	"github.com/abates/insteon"
)

// ProductDatabase is a registry of all the devices that have been
// seend on the local Insteon network.  The database includes the
// device category and firmware number. ProductDatabase implementations
// must be thread safe as the methods can be called from multiple
// go routines
type ProductDatabase interface {
	//UpdateDevCat(address insteon.Address, devCat insteon.DevCat)
	//UpdateEngineVersion(address insteon.Address, engineVersion insteon.EngineVersion)
	//UpdateFirmwareVersion(address insteon.Address, firmwareVersion insteon.FirmwareVersion)
	Update(address insteon.Address, cb func(deviceInfo *insteon.DeviceInfo))
	Find(address insteon.Address) (deviceInfo insteon.DeviceInfo, found bool)
}

type productDatabase struct {
	devices map[insteon.Address]*insteon.DeviceInfo
	mutex   sync.Mutex
}

// NewProductDB will initialize a product database for
// use in the network object
func NewProductDB() ProductDatabase {
	return &productDatabase{
		devices: make(map[insteon.Address]*insteon.DeviceInfo),
	}
}

func (pdb *productDatabase) Find(address insteon.Address) (deviceInfo insteon.DeviceInfo, found bool) {
	pdb.mutex.Lock()
	di, found := pdb.devices[address]
	pdb.mutex.Unlock()
	if found {
		deviceInfo = *di
	}
	return deviceInfo, found
}

func (pdb *productDatabase) Update(address insteon.Address, callback func(*insteon.DeviceInfo)) {
	pdb.mutex.Lock()
	deviceInfo, found := pdb.devices[address]
	if !found {
		deviceInfo = &insteon.DeviceInfo{
			Address: address,
		}
		pdb.devices[address] = deviceInfo
	}
	callback(deviceInfo)
	pdb.mutex.Unlock()
}

/*func (pdb *productDatabase) UpdateFirmwareVersion(address insteon.Address, firmwareVersion insteon.FirmwareVersion) {
	pdb.update(address, func(deviceInfo *insteon.DeviceInfo) { deviceInfo.FirmwareVersion = firmwareVersion })
}

func (pdb *productDatabase) UpdateEngineVersion(address insteon.Address, engineVersion insteon.EngineVersion) {
	pdb.update(address, func(deviceInfo *insteon.DeviceInfo) { deviceInfo.EngineVersion = engineVersion })
}

func (pdb *productDatabase) UpdateDevCat(address insteon.Address, devCat insteon.DevCat) {
	pdb.update(address, func(deviceInfo *insteon.DeviceInfo) { deviceInfo.DevCat = devCat })
}*/
