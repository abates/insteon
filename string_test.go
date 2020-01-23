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
		{"I1Device", newI1Device(&testConnection{addr: Address{1, 2, 3}}, 0), "I1 Device (01.02.03)"},
		{"I2Device", newI2Device(&testConnection{addr: Address{1, 2, 3}}, 0), "I2 Device (01.02.03)"},
		{"I2CsDevice", newI2CsDevice(&testConnection{addr: Address{1, 2, 3}}, 0), "I2CS Device (01.02.03)"},
		{"Switch", NewSwitch(&testConnection{addr: Address{1, 2, 3}}, 0), "Switch (01.02.03)"},
		{"Dimmer", NewDimmer(NewSwitch(&testConnection{addr: Address{1, 2, 3}}, 0), 0, 0), "Dimmer (01.02.03)"},
		{"Link Record", &LinkRecord{Flags: 0xd0, Group: Group(1), Address: Address{1, 2, 3}, Data: [3]byte{4, 5, 6}}, "UC 1 01.02.03 0x04 0x05 0x06"},
		{"Link Request Nil Link", &linkRequest{Type: readLink, MemAddress: BaseLinkDBAddress, NumRecords: 2, Link: nil}, "Link Read 0f.ff 2"},
		{"Link Request", &linkRequest{Type: readLink, MemAddress: BaseLinkDBAddress, NumRecords: 2, Link: &LinkRecord{Flags: 0xd0, Group: Group(1), Address: Address{1, 2, 3}, Data: [3]byte{4, 5, 6}}}, "Link Read 0f.ff 2 UC 1 01.02.03 0x04 0x05 0x06"},
		{"LevelNone", LevelNone, "NONE"},
		{"LevelInfo", LevelInfo, "INFO"},
		{"LevelDebug", LevelDebug, "DEBUG"},
		{"LevelTrace", LevelTrace, "TRACE"},
		{"LevelUnkown", LogLevel(-1), ""},
		{"StandardDirectFlag", StandardDirectMessage, "SD     2:2"},
		{"ExtendedDirectFlag", ExtendedDirectMessage, "ED     2:2"},
		{"AvailableController", AvailableController, "AC"},
		{"UnavailableController", UnavailableController, "UC"},
		{"AvailableResponder", AvailableResponder, "AR"},
		{"UnavailableResponder", UnavailableResponder, "UR"},
		{"Firmware Version", FirmwareVersion(42), "42"},
		{"Broadcast Message", &Message{Address{1, 2, 3}, Address{4, 5, 6}, StandardBroadcast, CmdSetButtonPressedController, nil}, "SB     2:2 01.02.03 -> ff.ff.ff DevCat 04.05 Firmware 6 Set-button Pressed (controller)"},
		{"All-Link Broadcast Message", &Message{Address{1, 2, 3}, Address{4, 5, 14}, StandardAllLinkBroadcast, CmdAllLinkRecall, nil}, "SA     2:2 01.02.03 -> ff.ff.ff Group(14) All-link recall"},
		{"Extended Direct", &Message{Address{1, 2, 3}, Address{4, 5, 6}, ExtendedDirectMessage, CmdEnterLinkingModeExt, make([]byte, 14)}, "ED     2:2 01.02.03 -> 04.05.06 Enter Linking Mode (i2cs) [00 00 00 00 00 00 00 00 00 00 00 00 00 00]"},
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
