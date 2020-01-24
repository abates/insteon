// Copyright 2019 Andrew Bates
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
	"time"
)

var (
	// LightingCategories match the two device categories known to be lighting
	// devices.  0x01 are dimmable devices and 0x02 are switched devices
	LightingDomains = []Domain{Domain(1), Domain(2)}
)

func init() {
	Devices.Register(0x01, dimmableDeviceFactory)
	Devices.Register(0x02, switchedDeviceFactory)
}

func switchedDeviceFactory(info DeviceInfo, device Device, timeout time.Duration) (Device, error) {
	return NewSwitch(info, device, timeout), nil
}

func dimmableDeviceFactory(info DeviceInfo, device Device, timeout time.Duration) (Device, error) {
	return NewDimmer(info, device, timeout), nil
}
