package insteon

import "fmt"

var (
	// CmdAssignToAllLinkGroup Assign to ALL-Link Group
	CmdAssignToAllLinkGroup = Commands.RegisterStd("All Link Assign", 0x01, 0x00)

	// CmdDeleteFromAllLinkGroup Delete from All-Link Group
	CmdDeleteFromAllLinkGroup = Commands.RegisterStd("All Link Delete", 0x02, 0x00)

	// CmdProductDataReq Product Data Request
	CmdProductDataReq = Commands.RegisterStd("Product Data Req", 0x03, 0x00)

	// CmdProductDataResp Product Data Response
	CmdProductDataResp = Commands.RegisterExt("Product Data Resp", 0x03, 0x00, func() Payload { return &ProductData{} })

	// CmdFxUsernameReq FX Username Request
	CmdFxUsernameReq = Commands.RegisterStd("FX Username Req", 0x03, 0x01)

	// CmdFxUsernameResp FX Username Response
	CmdFxUsernameResp = Commands.RegisterExt("FX Username Resp", 0x03, 0x01, nil)

	// CmdDeviceTextStringReq Device Text String Request
	CmdDeviceTextStringReq = Commands.RegisterStd("Text String Req", 0x03, 0x02)

	// CmdDeviceTextStringResp Device Text String Response
	CmdDeviceTextStringResp = Commands.RegisterExt("Text String Resp", 0x03, 0x02, nil)

	// CmdSetDeviceTextString sets the device text string
	CmdSetDeviceTextString = Commands.RegisterStd("Set Text String", 0x03, 0x03)

	// CmdEnterLinkingMode Enter Linking Mode
	CmdEnterLinkingMode = Commands.RegisterStd("Enter Link Mode", 0x09, 0x00)

	// CmdExitLinkingMode Exit Linking Mode
	CmdExitLinkingMode = Commands.RegisterStd("Exit Link Mode", 0x08, 0x00)

	// CmdEnterLinkingModeExtended Enter Linking Mode (extended command for I2CS devices)
	CmdEnterLinkingModeExtended = Commands.RegisterExt("Enter Link Mode", 0x09, 0x00, nil)

	// CmdEnterUnlinkingMode Enter Unlinking Mode
	CmdEnterUnlinkingMode = Commands.RegisterStd("Enter Unlink Mode", 0x0a, 0x00)

	// CmdEnterUnlinkingModeExtended Enter Unlinking Mode (extended command for I2CS devices)
	CmdEnterUnlinkingModeExtended = Commands.RegisterStd("Enter Unlink Mode", 0x0a, 0x00)

	// CmdGetEngineVersion Get Insteon Engine Version
	CmdGetEngineVersion = Commands.RegisterStd("Get INSTEON Ver", 0x0d, 0x00)

	// CmdPing Ping Request
	CmdPing = Commands.RegisterStd("Ping", 0x0f, 0x00)

	// CmdIDReq ID Request
	CmdIDReq = Commands.RegisterStd("ID Req", 0x10, 0x00)

	// CmdReadWriteALDB Read/Write ALDB
	CmdReadWriteALDB = Commands.RegisterExt("Read/Write ALDB", 0x2f, 0x00, func() Payload { return &LinkRequest{} })
)

// CommandRegistry provides a means for registering command descriptions as well
// as payload factories for the message unmarshaller
type CommandRegistry struct {
	standardCommands map[[2]byte]*Command
	extendedCommands map[[2]byte]*Command
}

// Commands is the global CommandRegistry. This only needs to be accessed when adding
// functionality (e.g. new device types)
var Commands = CommandRegistry{
	standardCommands: make(map[[2]byte]*Command),
	extendedCommands: make(map[[2]byte]*Command),
}

// Command is a 2 byte field present in all Insteon messages
type Command struct {
	name      string
	Cmd       [2]byte
	subCmd    bool
	generator PayloadGenerator
}

// SubCommand returns a new command where the second byte matches
// the passed in value
func (c *Command) SubCommand(value int) *Command {
	return &Command{name: c.name, Cmd: [2]byte{c.Cmd[0], byte(value)}, subCmd: true, generator: c.generator}
}

// String returns the description of the command
func (c *Command) String() string {
	if c.subCmd {
		return fmt.Sprintf("%s(%d)", c.name, c.Cmd[1])
	}
	return c.name
}

// Equal will compare two commands and return true if the fields (not including
// the payload generator) are equivalent
func (c *Command) Equal(other *Command) bool {
	if c == other {
		return true
	}

	if c != nil && other != nil {
		return c.name == other.name && c.Cmd == other.Cmd && c.subCmd == other.subCmd
	}

	return false
}

func find(cmd []byte, commands map[[2]byte]*Command) *Command {
	if command, found := commands[[2]byte{cmd[0], cmd[1]}]; found {
		return command
	} else if command, found := commands[[2]byte{cmd[0], 0x00}]; found {
		return command.SubCommand(int(cmd[1]))
	}

	// fail safe so nobody is ever referring to a nil command
	name := fmt.Sprintf("UNKNOWN (%02x.%02x)", cmd[0], cmd[1])
	return &Command{name: name, Cmd: [2]byte{cmd[0], cmd[1]}, generator: func() Payload { return &BufPayload{} }}
}

// FindExt returns either the registered extended command matching the 2 bytes passed in
// or it returns a generic command so that message parsing doesn't fail
func (cf *CommandRegistry) FindExt(cmd []byte) *Command {
	return find(cmd, cf.extendedCommands)
}

// FindStd returns either the registered standard command matching the 2 bytes passed in
// or it returns a generic command so that message parsing doesn't fail
func (cf *CommandRegistry) FindStd(cmd []byte) *Command {
	return find(cmd, cf.standardCommands)
}

func register(commands map[[2]byte]*Command, name string, b1, b2 byte, generator PayloadGenerator) *Command {
	command := &Command{name: name, Cmd: [2]byte{b1, b2}, generator: generator}
	commands[command.Cmd] = command
	return command
}

// RegisterExt will register a new extended command in the ComandRegistry.
// This only needs to be called when adding new device types or extending
// functionality
func (cf *CommandRegistry) RegisterExt(name string, b1, b2 byte, generator PayloadGenerator) *Command {
	if generator == nil {
		generator = func() Payload { return &BufPayload{} }
	}
	return register(cf.extendedCommands, name, b1, b2, generator)
}

// RegisterStd will register a new standard command in the ComandRegistry
// This only needs to be called when adding new device types or extending
// functionality
func (cf *CommandRegistry) RegisterStd(name string, b1, b2 byte) *Command {
	return register(cf.standardCommands, name, b1, b2, nil)
}
