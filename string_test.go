package insteon

import (
	"fmt"
	"testing"
)

func TestDeviceString(t *testing.T) {
	tests := []struct {
		desc   string
		device fmt.Stringer
		want   string
	}{
		{"I1Device", NewI1Device(&testConnection{addr: Address{1, 2, 3}}, 0), "I1 Device (01.02.03)"},
		{"I2Device", NewI2Device(&testConnection{addr: Address{1, 2, 3}}, 0), "I2 Device (01.02.03)"},
		{"I2CsDevice", NewI2CsDevice(&testConnection{addr: Address{1, 2, 3}}, 0), "I2CS Device (01.02.03)"},
		{"Switch", NewSwitch(&testConnection{addr: Address{1, 2, 3}}, 0).(*switchedDevice), "Switch (01.02.03)"},
		{"Dimmer", NewDimmer(NewSwitch(&testConnection{addr: Address{1, 2, 3}}, 0), 0, 0).(*dimmer), "Dimmer (01.02.03)"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := test.device.String()
			if test.want != got {
				t.Errorf("want %q got %q", test.want, got)
			}
		})
	}
}
