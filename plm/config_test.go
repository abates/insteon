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

import "testing"

func TestSettingConfigFlags(t *testing.T) {
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

func TestConfigString(t *testing.T) {
	tests := []struct {
		input    byte
		expected string
	}{
		{0x80, "L..."},
		{0x40, ".M.."},
		{0x20, "..A."},
		{0x10, "...D"},
		{0xf0, "LMAD"},
	}

	for i, test := range tests {
		config := Config(test.input)
		if config.String() != test.expected {
			t.Errorf("tests[%d] expected %q got %q", i, test.expected, config.String())
		}
	}
}

func TestConfigMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		input byte
	}{
		{0x00},
		{0x80},
		{0x40},
		{0x20},
		{0x10},
		{0xf0},
	}

	var config Config
	err := config.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	for i, test := range tests {
		config.UnmarshalBinary([]byte{test.input})

		if byte(config) != test.input {
			t.Errorf("tests[%d] expected 0x%02x got 0x%02x", i, test.input, config)
		}

		buf, _ := config.MarshalBinary()
		if buf[0] != test.input {
			t.Errorf("tests[%d] expected 0x%02x got 0x%02x", i, test.input, buf[0])
		}
	}
}
