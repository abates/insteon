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
)

// DeviceInfo is a record of information about known
// devices on the network
type DeviceInfo struct {
	Address         Address
	DevCat          DevCat
	FirmwareVersion FirmwareVersion
	EngineVersion   EngineVersion
}

// Complete indicates whether or not a record appears to be complete.  A complete
// record will have a non-zero DevCat and a non-zero FirmwareVersion
func (info *DeviceInfo) Complete() bool {
	return info.DevCat != DevCat{0x00, 0x00} && info.FirmwareVersion != FirmwareVersion(0x00)
}

// ProductDatabase is a registry of all the devices that have been
// seend on the local Insteon network.  The database includes the
// device category and firmware number. ProductDatabase implementations
// must be thread safe as the methods can be called from multiple
// go routines
type ProductDatabase interface {
	UpdateDevCat(address Address, devCat DevCat)
	UpdateEngineVersion(address Address, engineVersion EngineVersion)
	UpdateFirmwareVersion(address Address, firmwareVersion FirmwareVersion)
	Find(address Address) (deviceInfo DeviceInfo, found bool)
}

type productDatabase struct {
	devices map[Address]*DeviceInfo
	mutex   sync.Mutex
}

// NewProductDB will initialize a product database for
// use in the network object
func NewProductDB() ProductDatabase {
	return &productDatabase{
		devices: make(map[Address]*DeviceInfo),
	}
}

func (pdb *productDatabase) Find(address Address) (deviceInfo DeviceInfo, found bool) {
	pdb.mutex.Lock()
	di, found := pdb.devices[address]
	pdb.mutex.Unlock()
	if found {
		deviceInfo = *di
	}
	return deviceInfo, found
}

func (pdb *productDatabase) update(address Address, callback func(*DeviceInfo)) {
	pdb.mutex.Lock()
	deviceInfo, found := pdb.devices[address]
	if !found {
		deviceInfo = &DeviceInfo{
			Address: address,
		}
		pdb.devices[address] = deviceInfo
	}
	callback(deviceInfo)
	pdb.mutex.Unlock()
}

func (pdb *productDatabase) UpdateFirmwareVersion(address Address, firmwareVersion FirmwareVersion) {
	pdb.update(address, func(deviceInfo *DeviceInfo) { deviceInfo.FirmwareVersion = firmwareVersion })
}

func (pdb *productDatabase) UpdateEngineVersion(address Address, engineVersion EngineVersion) {
	pdb.update(address, func(deviceInfo *DeviceInfo) { deviceInfo.EngineVersion = engineVersion })
}

func (pdb *productDatabase) UpdateDevCat(address Address, devCat DevCat) {
	pdb.update(address, func(deviceInfo *DeviceInfo) { deviceInfo.DevCat = devCat })
}
