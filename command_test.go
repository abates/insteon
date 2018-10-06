package insteon

import "testing"

func TestCommandSubCommand(t *testing.T) {
	cmd := Command(0x0102)
	cmd = cmd.SubCommand(3)
	if cmd&0x00ff != 3 {
		t.Errorf("Expected 3 got %v", cmd&0x00ff)
	}
}
