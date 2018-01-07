package plm

import "testing"

func TestSettingFlags(t *testing.T) {
	config := Config(0x00)

	tests := []struct {
		getter   func() bool
		setter   func()
		clearer  func()
		expected byte
	}{
		{config.AutomaticLinking, config.setAutomaticLinking, config.clearAutomaticLinking, 0x80},
		{config.MonitorMode, config.setMonitorMode, config.clearMonitorMode, 0x40},
		{config.AutomaticLED, config.setAutomaticLED, config.clearAutomaticLED, 0x20},
		{config.DeadmanMode, config.setDeadmanMode, config.clearDeadmanMode, 0x10},
	}

	for i, test := range tests {
		if test.getter() {
			t.Errorf("tests[%d] expected false got %v", i, test.getter())
		}

		test.setter()
		if !test.getter() {
			t.Errorf("tests[%d] expected true got %v", i, test.getter())
		}

		if byte(config) != test.expected {
			t.Errorf("tests[%d] expected 0x%02x got 0x%02x", i, test.expected, byte(config))
		}

		test.clearer()
		if byte(config) != 0x00 {
			t.Errorf("tests[%d] expected 0x00 got 0x%02x", i, byte(config))
		}
	}
}
