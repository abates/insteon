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

import "time"

// I2CsDevice can communicate with Version 2 (checksum) Insteon Engines
type I2CsDevice struct {
	*I2Device
}

// NewI2CsDevice will initialize a new I2CsDevice object and make
// it ready for use
func NewI2CsDevice(info DeviceInfo, address Address, sendCh chan<- *MessageRequest, recvCh <-chan *Message, timeout time.Duration) (Device, error) {
	i2Device, err := NewI2Device(info, address, sendCh, recvCh, timeout)
	return &I2CsDevice{i2Device.(*I2Device)}, err
}

// EnterLinkingMode will put the device into linking mode. This is
// equivalent to holding down the set button until the device
// beeps and the indicator light starts flashing
func (i2cs *I2CsDevice) EnterLinkingMode(group Group) (err error) {
	return extractError(i2cs.SendCommand(CmdEnterLinkingModeExt.SubCommand(int(group)), make([]byte, 14)))
}

// String returns the string "I2CS Device (<address>)" where <address> is the destination
// address of the device
func (i2cs *I2CsDevice) String() string {
	return sprintf("I2CS Device (%s)", i2cs.Address())
}
