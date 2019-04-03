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
		{"Link Record", &LinkRecord{Flags: 0xd0, Group: Group(1), Address: Address{1, 2, 3}, Data: [3]byte{4, 5, 6}}, "UC 1 01.02.03 0x04 0x05 0x06"},
		{"Link Request Nil Link", &linkRequest{Type: readLink, MemAddress: BaseLinkDBAddress, NumRecords: 2, Link: nil}, "Link Read 0f.ff 2"},
		{"Link Request", &linkRequest{Type: readLink, MemAddress: BaseLinkDBAddress, NumRecords: 2, Link: &LinkRecord{Flags: 0xd0, Group: Group(1), Address: Address{1, 2, 3}, Data: [3]byte{4, 5, 6}}}, "Link Read 0f.ff 2 UC 1 01.02.03 0x04 0x05 0x06"},
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
