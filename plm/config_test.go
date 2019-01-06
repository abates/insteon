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
	"testing"
)

func TestSettingConfigFlags(t *testing.T) {
	t.Parallel()
	config := Config(0x00)

	tests := []struct {
		desc     string
		getter   func() bool
		setter   func()
		clearer  func()
		expected byte
	}{
		{"AutomaticLinking", config.AutomaticLinking, config.setAutomaticLinking, config.clearAutomaticLinking, 0x80},
		{"MonitorMode", config.MonitorMode, config.setMonitorMode, config.clearMonitorMode, 0x40},
		{"AutomaticLED", config.AutomaticLED, config.setAutomaticLED, config.clearAutomaticLED, 0x20},
		{"DeadmanMode", config.DeadmanMode, config.setDeadmanMode, config.clearDeadmanMode, 0x10},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			if test.getter() {
				t.Errorf("getter got %v, want false", test.getter())
			}

			test.setter()
			if !test.getter() {
				t.Errorf("getter got %v, want true", test.getter())
			}

			if byte(config) != test.expected {
				t.Errorf("config got 0x%02x, want 0x%02x", byte(config), test.expected)
			}

			test.clearer()
			if byte(config) != 0x00 {
				t.Errorf("config got 0x%02x, want 0x00", byte(config))
			}
		})
	}
}

func TestConfigString(t *testing.T) {
	t.Parallel()
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

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			t.Parallel()
			config := Config(test.input)
			if config.String() != test.expected {
				t.Errorf("got %q, expected %q", config.String(), test.expected)
			}
		})
	}
}

func TestConfigMarshalUnmarshal(t *testing.T) {
	t.Parallel()
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

	for _, test := range tests {
		t.Run(fmt.Sprintf("0x%02x", test.input), func(t *testing.T) {
			t.Parallel()
			config.UnmarshalBinary([]byte{test.input})

			if byte(config) != test.input {
				t.Errorf("got 0x%02x, want 0x%02x", config, test.input)
			}

			buf, _ := config.MarshalBinary()
			if buf[0] != test.input {
				t.Errorf("got 0x%02x, want 0x%02x", buf[0], test.input)
			}
		})
	}
}
