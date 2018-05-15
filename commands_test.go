package insteon

import "testing"

func TestCommandSubCommand(t *testing.T) {
	cmd := Command{0x01, 0x02}
	cmd = cmd.SubCommand(3)
	if cmd[1] != 3 {
		t.Errorf("Expected 3 got %v", cmd[1])
	}
}
