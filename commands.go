package insteon

import "fmt"

var Commands CommandRegistry

type CommandRegistry struct {
	standardCommands map[[2]byte]*Command
	extendedCommands map[[2]byte]*Command
}

type Command struct {
	name      string
	cmd       [2]byte
	subCmd    bool
	generator PayloadGenerator
}

func (c *Command) SubCommand(value int) *Command {
	return &Command{name: c.name, cmd: [2]byte{c.cmd[0], byte(value)}, subCmd: true, generator: c.generator}
}

func (c *Command) String() string {
	if c.subCmd {
		return fmt.Sprintf("%s(%d)", c.name, c.cmd[1])
	}
	return c.name
}

func (cf *CommandRegistry) FindExt(cmd []byte) *Command {
	return cf.extendedCommands[[2]byte{cmd[0], cmd[1]}]
}

func (cf *CommandRegistry) FindStd(cmd []byte) *Command {
	return cf.standardCommands[[2]byte{cmd[0], cmd[1]}]
}

func (cf *CommandRegistry) RegisterExt(name string, b1, b2 byte, generator PayloadGenerator) *Command {
	command := &Command{name: name, cmd: [2]byte{b1, b2}, generator: generator}
	if cf.extendedCommands == nil {
		cf.extendedCommands = make(map[[2]byte]*Command)
	}
	cf.extendedCommands[command.cmd] = command
	return command
}

func (cf *CommandRegistry) RegisterStd(name string, b1, b2 byte) *Command {
	command := &Command{name: name, cmd: [2]byte{b1, b2}}
	if cf.standardCommands == nil {
		cf.standardCommands = make(map[[2]byte]*Command)
	}
	cf.standardCommands[command.cmd] = command
	return command
}
