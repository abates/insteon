package insteon

import "testing"

func TestRecordControlFlags(t *testing.T) {
	tests := []struct {
		input      byte
		inUse      bool
		controller bool
	}{
		{0x40, false, true},
		{0x00, false, false},
		{0xc0, true, true},
		{0x80, true, false},
	}

	for i, test := range tests {
		flags := RecordControlFlags(test.input)
		if flags.InUse() != test.inUse {
			t.Errorf("tests[%d] expected %v got %v", i, test.inUse, flags.InUse())
		}

		if flags.Available() == test.inUse {
			t.Errorf("tests[%d] expected %v got %v", i, !test.inUse, flags.Available())
		}

		if flags.Controller() != test.controller {
			t.Errorf("tests[%d] expected %v got %v", i, !test.controller, flags.Controller())
		}

		if flags.Responder() == test.controller {
			t.Errorf("tests[%d] expected %v got %v", i, !test.controller, flags.Responder())
		}
	}

}

func TestLinkEqual(t *testing.T) {
	availableController := RecordControlFlags(0x40)
	availableResponder := RecordControlFlags(0x00)
	usedController := RecordControlFlags(0xc0)
	usedResponder := RecordControlFlags(0x80)

	newLink := func(flags RecordControlFlags, group Group, address Address) *Link {
		return &Link{flags, group, address, [3]byte{0x00, 0x00, 0x00}}
	}

	tests := []struct {
		link1    *Link
		link2    *Link
		expected bool
	}{
		{newLink(usedController, Group(0x01), Address{0x01, 0x02, 0x03}), newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), false},
		{newLink(usedResponder, Group(0x01), Address{0x01, 0x02, 0x03}), newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), false},
		{newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), true},
		{newLink(availableResponder, Group(0x01), Address{0x01, 0x02, 0x03}), newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), false},
		{newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x03}), newLink(availableController, Group(0x01), Address{0x01, 0x02, 0x04}), false},
	}

	for i, test := range tests {
		if test.link1.Equal(test.link2) != test.expected {
			t.Errorf("tests[%d] expected %v got %v", i, test.expected, !test.expected)
		}
	}
}
