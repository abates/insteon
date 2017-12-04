package insteon

import "testing"

func TestAddress(t *testing.T) {
	tests := []struct {
		input [3]byte
		str   string
	}{
		{[3]byte{0x47, 0x2d, 0x10}, "47.2d.10"},
	}

	for i, test := range tests {
		address := Address(test.input)

		if address.String() != test.str {
			t.Errorf("tests[%d] expected %q got %q", i, test.str, address.String())
		}
	}
}

func TestFlags(t *testing.T) {
	tests := []struct {
		input       byte
		messageType MessageType
		extended    bool
		maxTtl      int
		ttl         int
		str         string
	}{
		{0x0f, MsgTypeDirect, false, 3, 3, "SD     3:3"},
		{0x2f, MsgTypeDirectAck, false, 3, 3, "SD Ack 3:3"},
		{0x4f, MsgTypeAllLinkCleanup, false, 3, 3, "SC     3:3"},
		{0x6f, MsgTypeAllLinkCleanupAck, false, 3, 3, "SC Ack 3:3"},
		{0x8f, MsgTypeBroadcast, false, 3, 3, "SB     3:3"},
		{0xaf, MsgTypeDirectNak, false, 3, 3, "SD NAK 3:3"},
		{0xcf, MsgTypeAllLinkBroadcast, false, 3, 3, "SA     3:3"},
		{0xef, MsgTypeAllLinkCleanupNak, false, 3, 3, "SC NAK 3:3"},
		{0x1f, MsgTypeDirect, true, 3, 3, "ED     3:3"},
		{0x3f, MsgTypeDirectAck, true, 3, 3, "ED Ack 3:3"},
		{0x5f, MsgTypeAllLinkCleanup, true, 3, 3, "EC     3:3"},
		{0x7f, MsgTypeAllLinkCleanupAck, true, 3, 3, "EC Ack 3:3"},
		{0x9f, MsgTypeBroadcast, true, 3, 3, "EB     3:3"},
		{0xbf, MsgTypeDirectNak, true, 3, 3, "ED NAK 3:3"},
		{0xdf, MsgTypeAllLinkBroadcast, true, 3, 3, "EA     3:3"},
		{0xff, MsgTypeAllLinkCleanupNak, true, 3, 3, "EC NAK 3:3"},
	}

	for i, test := range tests {
		flags := Flags(test.input)
		if flags.Type() != test.messageType {
			t.Errorf("tests[%d] Expected Type %s got %s", i, test.messageType, flags.Type())
		}

		if test.extended {
			if flags.Standard() || !flags.Extended() {
				t.Errorf("tests[%d] Expected Extended Flag", i)
			}
		} else {
			if !flags.Standard() || flags.Extended() {
				t.Errorf("tests[%d] Expected Standard Flag", i)
			}
		}

		if flags.MaxTTL() != test.maxTtl {
			t.Errorf("tests[%d] Expected MaxTTL %d got %d", i, test.maxTtl, flags.MaxTTL())
		}

		if flags.TTL() != test.ttl {
			t.Errorf("tests[%d] Expected TTL %d got %d", i, test.ttl, flags.TTL())
		}

		if flags.String() != test.str {
			t.Errorf("tests[%d] Expected %q got %q", i, test.str, flags.String())
		}
	}

}
