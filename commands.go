package insteon

import "fmt"

var (
	CmdAssignToAllLinkGroup    = Commands.RegisterStd("All Link Assign", 0x01, 0x00)
	CmdDeleteFromAllLinkGroup  = Commands.RegisterStd("All Link Delete", 0x02, 0x00)
	CmdProductDataReq          = Commands.RegisterStd("Product Data Req", 0x03, 0x00)
	CmdProductDataResp         = Commands.RegisterExt("Product Data Resp", 0x03, 0x00, func() Payload { return &ProductData{} })
	CmdFxUsernameReq           = Commands.RegisterStd("FX Username Req", 0x03, 0x01)
	CmdFxUsernameResp          = Commands.RegisterExt("FX Username Resp", 0x03, 0x01, nil)
	CmdDeviceTextStringReq     = Commands.RegisterStd("Text String Req", 0x03, 0x02)
	CmdDeviceTextStringResp    = Commands.RegisterExt("Text String Resp", 0x03, 0x02, nil)
	CmdEnterLinkingMode        = Commands.RegisterStd("Enter Link Mode", 0x09, 0x00)
	CmdEnterUnlinkingMode      = Commands.RegisterStd("Enter Unlink Mode", 0x0a, 0x00)
	CmdGetInsteonEngineVersion = Commands.RegisterStd("Get INSTEON Ver", 0x0d, 0x00)
	CmdPing                    = Commands.RegisterStd("Ping", 0x0f, 0x00)
	CmdIDReq                   = Commands.RegisterStd("ID Req", 0x10, 0x00)
	CmdReadWriteALDB           = Commands.RegisterExt("Read/Write ALDB", 0x2f, 0x00, func() Payload { return &LinkRequest{} })
)

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
	if cmd, found := cf.standardCommands[[2]byte{cmd[0], cmd[1]}]; found {
		return cmd
	}

	// fail safe so nobody is ever referring to a nil command
	return &Command{name: "", cmd: [2]byte{cmd[0], cmd[1]}}
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
