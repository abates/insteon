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
		{"I1 EngineVersion", VerI1, "I1"},
		{"I2 EngineVersion", VerI2, "I2"},
		{"I2Cs EngineVersion", VerI2Cs, "I2Cs"},
		{"Unknown EngineVersion", EngineVersion(3), "unknown"},
		{"I1Device", newDevice(nil, DeviceInfo{Address: Address{1, 2, 3}}), "I1 Device (01.02.03)"},
		{"Switch", NewSwitch(&device{}, DeviceInfo{Address: Address{1, 2, 3}}), "Switch (01.02.03)"},
		{"Dimmer", NewDimmer(&device{}, DeviceInfo{Address: Address{1, 2, 3}}), "Dimmer (01.02.03)"},
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
		{"All-Link Broadcast Message", &Message{Address{1, 2, 3}, Address{4, 5, 14}, StandardAllLinkBroadcast, CmdAllLinkRecall, nil}, "SA     2:2 01.02.03 -> ff.ff.ff All-link recall Group(14)"},
		{"All-Link Cleanup Message", &Message{Address{1, 2, 3}, Address{4, 5, 14}, Flag(MsgTypeAllLinkCleanup, false, 2, 2), CmdAllLinkRecall, nil}, "SC     2:2 01.02.03 -> 04.05.0e Cleanup All-link recall"},
		{"Extended Direct", &Message{Address{1, 2, 3}, Address{4, 5, 6}, ExtendedDirectMessage, CmdEnterLinkingModeExt, make([]byte, 14)}, "ED     2:2 01.02.03 -> 04.05.06 Enter Linking Mode (i2cs) [00 00 00 00 00 00 00 00 00 00 00 00 00 00]"},
		{"Standard Ack", &Message{Address{1, 2, 3}, Address{4, 5, 6}, StandardDirectAck, CmdEnterLinkingMode, nil}, "SD Ack 2:2 01.02.03 -> 04.05.06 9.0"},
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
