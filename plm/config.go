package plm

import "fmt"

type Config byte

func (config *Config) setBit(pos uint) {
	*config |= (1 << pos)
}

func (config *Config) clearBit(pos uint) {
	*config &= ^(1 << pos)
}

func (config Config) AutomaticLinking() bool  { return config&0x80 == 0x80 }
func (config *Config) setAutomaticLinking()   { config.setBit(7) }
func (config *Config) clearAutomaticLinking() { config.clearBit(7) }
func (config Config) MonitorMode() bool       { return config&0x40 == 0x40 }
func (config *Config) setMonitorMode()        { config.setBit(6) }
func (config *Config) clearMonitorMode()      { config.clearBit(6) }
func (config Config) AutomaticLED() bool      { return config&0x20 == 0x20 }
func (config *Config) setAutomaticLED()       { config.setBit(5) }
func (config *Config) clearAutomaticLED()     { config.clearBit(5) }
func (config Config) DeadmanMode() bool       { return config&0x10 == 0x10 }
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
	return []byte{byte(*config)}, nil
}

func (config *Config) UnmarshalBinary(buf []byte) error {
	if len(buf) < 1 {
		return fmt.Errorf("config is 1 byte, got %d", len(buf))
	}
	*config = Config(buf[0])
	return nil
}
