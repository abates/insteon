package insteon

import (
	"fmt"
	"testing"
)

func TestCommandRegistry(t *testing.T) {
	tests := []struct {
		name     string
		b1       byte
		b2       byte
		subcmd   int
		extended bool
	}{
		{"cmd 1", 0xee, 0xff, 0, false},
		{"cmd 2", 0xee, 0x00, 0, true},
		{"cmd 3", 0xef, 0x00, 25, true},
	}

	for i, test := range tests {
		var cmd *Command
		if test.extended {
			cmd = Commands.RegisterExt(test.name, test.b1, test.b2, nil)
			if cmd != Commands.FindExt([]byte{test.b1, test.b2}) {
				t.Errorf("tests[%d] expected %v got %v", i, cmd, Commands.FindExt([]byte{test.b1, test.b2}))
			}
		} else {
			cmd = Commands.RegisterStd(test.name, test.b1, test.b2)
			if cmd != Commands.FindStd([]byte{test.b1, test.b2}) {
				t.Errorf("tests[%d] expected %v got %v", i, cmd, Commands.FindStd([]byte{test.b1, test.b2}))
			}
		}

		if test.subcmd == 0 {
			if test.name != cmd.String() {
				t.Errorf("tests[%d] expected %s got %s", i, test.name, cmd.String())
			}
		} else {
			subcmd := cmd.SubCommand(test.subcmd)
			testStr := fmt.Sprintf("%s(%d)", test.name, test.subcmd)
			if testStr != subcmd.String() {
				t.Errorf("tests[%d] expected %s got %s", i, testStr, subcmd.String())
			}
		}
	}
}
