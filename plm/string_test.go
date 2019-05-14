package plm

import (
	"fmt"
	"testing"

	"github.com/abates/insteon"
)

func TestDeviceString(t *testing.T) {
	tests := []struct {
		desc   string
		device fmt.Stringer
		want   string
	}{
		{"Config (auto linking)", Config(0x80), "L..."},
		{"Config (monitor mode)", Config(0x40), ".M.."},
		{"Config (auto led)", Config(0x20), "..A."},
		{"Config (deadman)", Config(0x10), "...D"},
		{"Config (all)", Config(0xF0), "LMAD"},
		{"Version", Version(42), "42"},
		{"Info", &Info{insteon.Address{1, 2, 3}, insteon.DevCat{4, 5}, Version(6)}, "01.02.03 category 04.05 version 6"},
		{"manageRecordRequest", &manageRecordRequest{0x25, insteon.ControllerLink(1, insteon.Address{4, 5, 6})}, "25 UC 1 04.05.06 0x00 0x00 0x00"},
		{"allLinkReq", &allLinkReq{0x42, 22}, "42 22"},
		{"ACK packet", &Packet{Command: CmdGetInfo, Ack: 0x06}, fmt.Sprintf("%v ACK", CmdGetInfo)},
		{"NAK packet", &Packet{Command: CmdGetInfo, Ack: 0x15}, fmt.Sprintf("%v NAK", CmdGetInfo)},
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
