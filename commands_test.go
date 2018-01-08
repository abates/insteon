package insteon

import (
	"testing"
)

func TestCommandEqual(t *testing.T) {
	tests := []struct {
		cmd1     *Command
		cmd2     *Command
		expected bool
	}{
		{nil, nil, true},
		{&Command{Cmd: [2]byte{0x01, 0x02}}, nil, false},
		{nil, &Command{Cmd: [2]byte{0x01, 0x02}}, false},
		{&Command{Cmd: [2]byte{0x02, 0x03}}, &Command{Cmd: [2]byte{0x01, 0x02}}, false},
		{&Command{Cmd: [2]byte{0x01, 0x02}}, &Command{Cmd: [2]byte{0x01, 0x02}}, true},
	}

	for i, test := range tests {
		if test.expected != test.cmd1.Equal(test.cmd2) {
			t.Errorf("tests[%d] expected %v got %v", i, test.expected, test.cmd1.Equal(test.cmd2))
		}
	}
}

func TestCommandRegistry(t *testing.T) {
	tests := []struct {
		name           string
		b1             byte
		b2             byte
		subcmd         int
		extended       bool
		expectedString string
	}{
		{"", 0x01, 0xff, 0, false, "01.ff"},
		{"cmd 1", 0xee, 0xff, 0, false, "cmd 1"},
		{"cmd 2", 0xee, 0x00, 0, true, "cmd 2"},
		{"cmd 3", 0xef, 0x00, 25, true, "cmd 3(25)"},
	}

	commands := CommandRegistry{
		standardCommands: make(map[[2]byte]*Command),
		extendedCommands: make(map[[2]byte]*Command),
	}

	for i, test := range tests {
		var cmd *Command

		// make sure nil is never returned
		cmd = commands.FindStd([]byte{test.b1, test.b2})
		if cmd == nil {
			t.Errorf("tests[%d] expected FindStd to return non nil", i)
		} else if cmd.generator == nil {
			t.Errorf("tests[%d] expected non-nil generator", i)
		} else if cmd.generator() == nil {
			t.Errorf("tests[%d] expected non-nil payload", i)
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
			if test.expectedString != cmd.String() {
				t.Errorf("tests[%d] expected %s got %s", i, test.expectedString, cmd.String())
			}

		} else {
			subcmd := cmd.SubCommand(test.subcmd)
			if test.expectedString != subcmd.String() {
				t.Errorf("tests[%d] expected %s got %s", i, test.expectedString, subcmd.String())
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
