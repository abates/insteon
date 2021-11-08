package insteon

import (
	"flag"
	"fmt"
	"testing"

	"github.com/abates/insteon/commands"
)

func TestGetSet(t *testing.T) {
	tests := []struct {
		name    string
		getter  flag.Getter
		input   string
		want    interface{}
		wantStr string
	}{
		{"group", new(Group), "128", Group(128), "128"},
		{"address", new(Address), "01.04.07", Address{1, 4, 7}, "01.04.07"},
		{"address", new(Address), "1.4.7", Address{1, 4, 7}, "01.04.07"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.getter.Set(test.input)
			if err == nil {
				got := test.getter.Get()
				if test.want != got {
					t.Errorf("Wanted %v of type %T but got %v of type %T", test.want, test.want, got, got)
				}

				gotStr := test.getter.String()
				if test.wantStr != gotStr {
					t.Errorf("Wanted string %q got %q", test.wantStr, gotStr)
				}
			} else {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

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
		{"Link Record", &LinkRecord{Flags: 0xd0, Group: Group(1), Address: Address{1, 2, 3}, Data: [3]byte{4, 5, 6}}, "UC 1 01.02.03 0x04 0x05 0x06"},
		{"StandardDirectFlag", StandardDirectMessage, "SD     2:2"},
		{"ExtendedDirectFlag", ExtendedDirectMessage, "ED     2:2"},
		{"AvailableController", AvailableController, "AC"},
		{"UnavailableController", UnavailableController, "UC"},
		{"AvailableResponder", AvailableResponder, "AR"},
		{"UnavailableResponder", UnavailableResponder, "UR"},
		{"Firmware Version", FirmwareVersion(42), "42"},
		{"Broadcast Message", &Message{Address{1, 2, 3}, Address{4, 5, 6}, StandardBroadcast, commands.SetButtonPressedController, nil}, "SB     2:2 01.02.03 -> ff.ff.ff DevCat 04.05 Firmware 6 Set-button Pressed (controller)"},
		{"All-Link Broadcast Message", &Message{Address{1, 2, 3}, Address{4, 5, 14}, StandardAllLinkBroadcast, commands.AllLinkRecall, nil}, "SA     2:2 01.02.03 -> ff.ff.ff All-link recall Group(14)"},
		{"All-Link Cleanup Message", &Message{Address{1, 2, 3}, Address{4, 5, 14}, Flag(MsgTypeAllLinkCleanup, false, 2, 2), commands.AllLinkRecall, nil}, "SC     2:2 01.02.03 -> 04.05.0e Cleanup All-link recall"},
		{"Extended Direct", &Message{Address{1, 2, 3}, Address{4, 5, 6}, ExtendedDirectMessage, commands.EnterLinkingModeExt, make([]byte, 14)}, "ED     2:2 01.02.03 -> 04.05.06 Enter Linking Mode (i2cs) [00 00 00 00 00 00 00 00 00 00 00 00 00 00]"},
		{"Standard Ack", &Message{Address{1, 2, 3}, Address{4, 5, 6}, StandardDirectAck, commands.EnterLinkingMode, nil}, "SD Ack 2:2 01.02.03 -> 04.05.06 9.0"},
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
