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

	commands := CommandRegistry{
		standardCommands: make(map[[2]byte]*Command),
		extendedCommands: make(map[[2]byte]*Command),
	}

	for i, test := range tests {
		var cmd *Command

		// make sure nil is never returned
		if commands.FindStd([]byte{test.b1, test.b2}) == nil {
			t.Errorf("tests[%d] expected FindStd to return non nil", i)
		}

		if test.extended {
			cmd = commands.RegisterExt(test.name, test.b1, test.b2, nil)
			if cmd != commands.FindExt([]byte{test.b1, test.b2}) {
				t.Errorf("tests[%d] expected %v got %v", i, cmd, commands.FindExt([]byte{test.b1, test.b2}))
			}
		} else {
			cmd = commands.RegisterStd(test.name, test.b1, test.b2)
			if cmd != commands.FindStd([]byte{test.b1, test.b2}) {
				t.Errorf("tests[%d] expected %v got %v", i, cmd, commands.FindStd([]byte{test.b1, test.b2}))
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

			if test.extended {
				if !subcmd.Equal(commands.FindExt(subcmd.Cmd[:])) {
					t.Errorf("tests[%d] expected %#v got %#v", i, subcmd, commands.FindExt(subcmd.Cmd[:]))
				}
			} else {
				if !subcmd.Equal(commands.FindStd(subcmd.Cmd[:])) {
					t.Errorf("tests[%d] expected %+v got %+v", i, subcmd, commands.FindStd(subcmd.Cmd[:]))
				}
			}
		}
	}
}
