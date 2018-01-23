package insteon

import "fmt"

var (
	// CmdSetButtonPressedResponder Broadcast command indicating the set button has been pressed
	CmdSetButtonPressedResponder = Commands.RegisterStd("Set Button Pressed (responder)", []byte{0x00}, MsgTypeBroadcast, 0x01, 0x00)

	// CmdSetButtonPressedController Broadcast command indicating the set button has been pressed
	CmdSetButtonPressedController = Commands.RegisterStd("Set Button Pressed (controller)", []byte{0x00}, MsgTypeBroadcast, 0x02, 0x00)

	// CmdAssignToAllLinkGroup Assign to ALL-Link Group
	CmdAssignToAllLinkGroup = Commands.RegisterStd("All Link Assign", []byte{0x00}, MsgTypeDirect, 0x01, 0x00)

	// CmdDeleteFromAllLinkGroup Delete from All-Link Group
	CmdDeleteFromAllLinkGroup = Commands.RegisterStd("All Link Delete", []byte{0x00}, MsgTypeDirect, 0x02, 0x00)

	// CmdProductDataReq Product Data Request
	CmdProductDataReq = Commands.RegisterStd("Product Data Req", []byte{0x00}, MsgTypeDirect, 0x03, 0x00)

	// CmdProductDataResp Product Data Response
	CmdProductDataResp = Commands.RegisterExt("Product Data Resp", []byte{0x00}, MsgTypeDirect, 0x03, 0x00, nil)

	// CmdFxUsernameReq FX Username Request
	CmdFxUsernameReq = Commands.RegisterStd("FX Username Req", []byte{0x00}, MsgTypeDirect, 0x03, 0x01)

	// CmdFxUsernameResp FX Username Response
	CmdFxUsernameResp = Commands.RegisterExt("FX Username Resp", []byte{0x00}, MsgTypeDirect, 0x03, 0x01, nil)

	// CmdDeviceTextStringReq Device Text String Request
	CmdDeviceTextStringReq = Commands.RegisterStd("Text String Req", []byte{0x00}, MsgTypeDirect, 0x03, 0x02)

	// CmdDeviceTextStringResp Device Text String Response
	CmdDeviceTextStringResp = Commands.RegisterExt("Text String Resp", []byte{0x00}, MsgTypeDirect, 0x03, 0x02, nil)

	// CmdSetDeviceTextString sets the device text string
	CmdSetDeviceTextString = Commands.RegisterExt("Set Text String", []byte{0x00}, MsgTypeDirect, 0x03, 0x03, nil)

	// CmdExitLinkingMode Exit Linking Mode
	CmdExitLinkingMode = Commands.RegisterStd("Exit Link Mode", []byte{0x00}, MsgTypeDirect, 0x08, 0x00)

	// CmdExitLinkingModeExt Exit Linking Mode
	CmdExitLinkingModeExt = Commands.RegisterExt("Exit Link Mode", []byte{0x00}, MsgTypeDirect, 0x08, 0x00, nil)

	// CmdEnterLinkingMode Enter Linking Mode
	CmdEnterLinkingMode = Commands.RegisterStd("Enter Link Mode", []byte{0x00}, MsgTypeDirect, 0x09, 0x00)

	// CmdEnterLinkingModeExt Enter Linking Mode (extended command for I2CS devices)
	CmdEnterLinkingModeExt = Commands.RegisterExt("Enter Link Mode", []byte{0x00}, MsgTypeDirect, 0x09, 0x00, nil)

	// CmdEnterUnlinkingMode Enter Unlinking Mode
	CmdEnterUnlinkingMode = Commands.RegisterStd("Enter Unlink Mode", []byte{0x00}, MsgTypeDirect, 0x0a, 0x00)

	// CmdEnterUnlinkingModeExt Enter Unlinking Mode (extended command for I2CS devices)
	CmdEnterUnlinkingModeExt = Commands.RegisterExt("Enter Unlink Mode", []byte{0x00}, MsgTypeDirect, 0x0a, 0x00, nil)

	// CmdGetEngineVersion Get Insteon Engine Version
	CmdGetEngineVersion = Commands.RegisterStd("Get INSTEON Ver", []byte{0x00}, MsgTypeDirect, 0x0d, 0x00)

	// CmdPing Ping Request
	CmdPing = Commands.RegisterStd("Ping", []byte{0x00}, MsgTypeDirect, 0x0f, 0x00)

	// CmdIDReq ID Request
	CmdIDReq = Commands.RegisterStd("ID Req", []byte{0x00}, MsgTypeDirect, 0x10, 0x00)

	CmdGetOperatingFlags = Commands.RegisterStd("Get Operating Flags", []byte{0x00}, MsgTypeDirect, 0x1f, 0x00)

	CmdSetOperatingFlags = Commands.RegisterStd("Set Operating Flags", []byte{0x00}, MsgTypeDirect, 0x20, 0x00)

	// CmdReadWriteALDB Read/Write ALDB
	CmdReadWriteALDB = Commands.RegisterExt("Read/Write ALDB", []byte{0x00}, MsgTypeDirect, 0x2f, 0x00, func() Payload { return &LinkRequest{} })
)

type commandIndex map[[2]byte]*Command

type messageTypeIndex struct {
	standardCommands commandIndex
	extendedCommands commandIndex
}

// CommandRegistry provides a means for registering command descriptions as well
// as payload factories for the message unmarshaller
type CommandRegistry struct {
	commands map[byte]map[MessageType]*messageTypeIndex
}

var (
	// Commands is the global CommandRegistry. This only needs to be accessed when adding
	// functionality (e.g. new device types)
	Commands = newCommandRegistry()
)

func newCommandRegistry() CommandRegistry {
	commands := CommandRegistry{
		commands: make(map[byte]map[MessageType]*messageTypeIndex),
	}

	commands.commands[0] = make(map[MessageType]*messageTypeIndex)

	for _, mt := range []MessageType{MsgTypeDirect, MsgTypeBroadcast, MsgTypeAllLinkBroadcast} {
		commands.commands[0][mt] = &messageTypeIndex{
			standardCommands: make(commandIndex),
			extendedCommands: make(commandIndex),
		}
	}

	return commands
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
	name := c.name
	if name == "" {
		name = fmt.Sprintf("%02x.%02x", c.Cmd[0], c.Cmd[1])
	}

	if c.subCmd {
		return fmt.Sprintf("%s(%d)", name, c.Cmd[1])
	}
	return name
}

// Equal will compare two commands and return true if the fields (not including
// the payload generator) are equivalent
func (c *Command) Equal(other *Command) bool {
	if c == other {
		return true
	}

	if c != nil && other != nil {
		return c.Cmd == other.Cmd && c.subCmd == other.subCmd
	}

	return false
}

func find(mti *messageTypeIndex, extended bool, cmd []byte) *Command {
	ci := mti.standardCommands
	if extended {
		ci = mti.extendedCommands
	}

	if command, found := ci[[2]byte{cmd[0], cmd[1]}]; found {
		return command
	} else if command, found := ci[[2]byte{cmd[0], 0x00}]; found {
		return command.SubCommand(int(cmd[1]))
	}
	return nil
}

func (cr *CommandRegistry) Find(devCat byte, messageType MessageType, extended bool, cmd []byte) (command *Command) {
	if messageType == MsgTypeDirectAck || messageType == MsgTypeDirectNak {
		messageType = MsgTypeDirect
	}

	index := cr.commands[devCat]
	if index == nil {
		// try all devCat 0x00
		command = find(cr.commands[0x00][messageType], extended, cmd)
	} else {
		command = find(index[messageType], extended, cmd)
		if command == nil {
			// try all devCat 0x00
			command = find(cr.commands[0x00][messageType], extended, cmd)
		}
	}

	// fail safe so nobody is ever referring to a nil command
	if command == nil {
		name := fmt.Sprintf("UNKNOWN (%02x.%02x)", cmd[0], cmd[1])
		command = &Command{name: name, Cmd: [2]byte{cmd[0], cmd[1]}, generator: func() Payload { return &BufPayload{} }}
	}

	return command
}

// FindExt returns either the registered extended command matching the 2 bytes passed in
// or it returns a generic command so that message parsing doesn't fail
func (cr *CommandRegistry) FindExt(devCat byte, messageType MessageType, cmd []byte) *Command {
	return cr.Find(devCat, messageType, true, cmd)
}

// FindStd returns either the registered standard command matching the 2 bytes passed in
// or it returns a generic command so that message parsing doesn't fail
func (cr *CommandRegistry) FindStd(devCat byte, messageType MessageType, cmd []byte) *Command {
	return cr.Find(devCat, messageType, false, cmd)
}

func (cr *CommandRegistry) Register(name string, devCats []byte, messageType MessageType, extended bool, b1, b2 byte, generator PayloadGenerator) *Command {
	command := &Command{name: name, Cmd: [2]byte{b1, b2}, generator: generator}
	for _, devCat := range devCats {
		index := cr.commands[devCat]
		if index == nil {
			index = make(map[MessageType]*messageTypeIndex)
			for _, mt := range []MessageType{MsgTypeDirect, MsgTypeBroadcast, MsgTypeAllLinkBroadcast} {
				index[mt] = &messageTypeIndex{
					standardCommands: make(commandIndex),
					extendedCommands: make(commandIndex),
				}
			}
		}

		mti := index[messageType]
		if extended {
			mti.extendedCommands[command.Cmd] = command
		} else {
			mti.standardCommands[command.Cmd] = command
		}
	}
	return command
}

// RegisterExt will register a new extended command in the ComandRegistry.
// This only needs to be called when adding new device types or extending
// functionality
func (cr *CommandRegistry) RegisterExt(name string, devCats []byte, messageType MessageType, b1, b2 byte, generator PayloadGenerator) *Command {
	if generator == nil {
		generator = func() Payload { return &BufPayload{} }
	}
	return cr.Register(name, devCats, messageType, true, b1, b2, generator)
}

// RegisterStd will register a new standard command in the ComandRegistry
// This only needs to be called when adding new device types or extending
// functionality
func (cr *CommandRegistry) RegisterStd(name string, devCats []byte, messageType MessageType, b1, b2 byte) *Command {
	return cr.Register(name, devCats, messageType, false, b1, b2, nil)
}
