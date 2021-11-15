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

package plm

import (
	"fmt"

	"github.com/abates/insteon"
)

type Version byte

func (v Version) String() string { return fmt.Sprintf("%d", byte(v)) }

type Info struct {
	Address  insteon.Address
	DevCat   insteon.DevCat
	Firmware Version
}

func (info *Info) String() string {
	return fmt.Sprintf("%s category %s version %s", info.Address, info.DevCat, info.Firmware)
}

func (info *Info) MarshalBinary() ([]byte, error) {
	data := make([]byte, 6)

	copy(data[0:3], info.Address.Bytes())
	copy(data[3:5], info.DevCat[:])
	data[5] = byte(info.Firmware)
	return data, nil
}

func (info *Info) UnmarshalBinary(data []byte) error {
	info.Address.Put(data[0:3])
	copy(info.DevCat[:], data[3:5])
	info.Firmware = Version(data[5])
	return nil
}
