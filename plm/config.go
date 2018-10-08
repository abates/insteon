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

import "fmt"

type Config byte

func (config *Config) setBit(pos uint) {
	*config |= (1 << pos)
}

func (config *Config) clearBit(pos uint) {
	*config &= ^(1 << pos)
}

func (config *Config) AutomaticLinking() bool { return (*config)&0x80 == 0x80 }
func (config *Config) setAutomaticLinking()   { config.setBit(7) }
func (config *Config) clearAutomaticLinking() { config.clearBit(7) }
func (config *Config) MonitorMode() bool      { return (*config)&0x40 == 0x40 }
func (config *Config) setMonitorMode()        { config.setBit(6) }
func (config *Config) clearMonitorMode()      { config.clearBit(6) }
func (config *Config) AutomaticLED() bool     { return (*config)&0x20 == 0x20 }
func (config *Config) setAutomaticLED()       { config.setBit(5) }
func (config *Config) clearAutomaticLED()     { config.clearBit(5) }
func (config *Config) DeadmanMode() bool      { return (*config)&0x10 == 0x10 }
func (config *Config) setDeadmanMode()        { config.setBit(4) }
func (config *Config) clearDeadmanMode()      { config.clearBit(4) }

func (config *Config) String() string {
	str := ""
	if config.AutomaticLinking() {
		str += "L"
	} else {
		str += "."
	}

	if config.MonitorMode() {
		str += "M"
	} else {
		str += "."
	}

	if config.AutomaticLED() {
		str += "A"
	} else {
		str += "."
	}

	if config.DeadmanMode() {
		str += "D"
	} else {
		str += "."
	}

	return str
}

func (config *Config) MarshalBinary() ([]byte, error) {
	return []byte{byte(*config) & 0xf0}, nil
}

func (config *Config) UnmarshalBinary(buf []byte) error {
	if len(buf) < 1 {
		return fmt.Errorf("config is 1 byte, got %d", len(buf))
	}
	*config = Config(buf[0])
	return nil
}
