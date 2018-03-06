package insteon

import (
	"fmt"
	"sort"
)

var (
	Commands = NewCommandDB()

	DefaultCategories = []Category{Category(0)}

	// CmdSetButtonPressedResponder Broadcast command indicating the set button has been pressed
	CmdSetButtonPressedResponder = Commands.RegisterStd("Set Button Pressed (responder)", DefaultCategories, MsgTypeBroadcast, 0x01, 0x00)

	// CmdSetButtonPressedController Broadcast command indicating the set button has been pressed
	CmdSetButtonPressedController = Commands.RegisterStd("Set Button Pressed (controller)", DefaultCategories, MsgTypeBroadcast, 0x02, 0x00)

	// CmdAssignToAllLinkGroup Assign to ALL-Link Group
	CmdAssignToAllLinkGroup = Commands.RegisterStd("All Link Assign", DefaultCategories, MsgTypeDirect, 0x01, 0x00)

	// CmdDeleteFromAllLinkGroup Delete from All-Link Group
	CmdDeleteFromAllLinkGroup = Commands.RegisterStd("All Link Delete", DefaultCategories, MsgTypeDirect, 0x02, 0x00)

	// CmdProductDataReq Product Data Request
	CmdProductDataReq = Commands.RegisterStd("Product Data Req", DefaultCategories, MsgTypeDirect, 0x03, 0x00)

	// CmdProductDataResp Product Data Response
	CmdProductDataResp = Commands.RegisterExt("Product Data Resp", DefaultCategories, MsgTypeDirect, 0x03, 0x00)

	// CmdFxUsernameReq FX Username Request
	CmdFxUsernameReq = Commands.RegisterStd("FX Username Req", DefaultCategories, MsgTypeDirect, 0x03, 0x01)

	// CmdFxUsernameResp FX Username Response
	CmdFxUsernameResp = Commands.RegisterExt("FX Username Resp", DefaultCategories, MsgTypeDirect, 0x03, 0x01)

	// CmdDeviceTextStringReq Device Text String Request
	CmdDeviceTextStringReq = Commands.RegisterStd("Text String Req", DefaultCategories, MsgTypeDirect, 0x03, 0x02)

	// CmdDeviceTextStringResp Device Text String Response
	CmdDeviceTextStringResp = Commands.RegisterExt("Text String Resp", DefaultCategories, MsgTypeDirect, 0x03, 0x02)

	// CmdSetDeviceTextString sets the device text string
	CmdSetDeviceTextString = Commands.RegisterExt("Set Text String", DefaultCategories, MsgTypeDirect, 0x03, 0x03)

	// CmdExitLinkingMode Exit Linking Mode
	CmdExitLinkingMode = Commands.RegisterStd("Exit Link Mode", DefaultCategories, MsgTypeDirect, 0x08, 0x00)

	// CmdExitLinkingModeExt Exit Linking Mode
	CmdExitLinkingModeExt = Commands.RegisterExt("Exit Link Mode", DefaultCategories, MsgTypeDirect, 0x08, 0x00)

	// CmdEnterLinkingMode Enter Linking Mode
	CmdEnterLinkingMode = Commands.RegisterStd("Enter Link Mode", DefaultCategories, MsgTypeDirect, 0x09, 0x00)

	// CmdEnterLinkingModeExt Enter Linking Mode (extended command for I2CS devices)
	CmdEnterLinkingModeExt = Commands.RegisterExt("Enter Link Mode", DefaultCategories, MsgTypeDirect, 0x09, 0x00)

	// CmdEnterUnlinkingMode Enter Unlinking Mode
	CmdEnterUnlinkingMode = Commands.RegisterStd("Enter Unlink Mode", DefaultCategories, MsgTypeDirect, 0x0a, 0x00)

	// CmdEnterUnlinkingModeExt Enter Unlinking Mode (extended command for I2CS devices)
	CmdEnterUnlinkingModeExt = Commands.RegisterExt("Enter Unlink Mode", DefaultCategories, MsgTypeDirect, 0x0a, 0x00)

	// CmdGetEngineVersion Get Insteon Engine Version
	CmdGetEngineVersion = Commands.RegisterStd("Get INSTEON Ver", DefaultCategories, MsgTypeDirect, 0x0d, 0x00)

	// CmdPing Ping Request
	CmdPing = Commands.RegisterStd("Ping", DefaultCategories, MsgTypeDirect, 0x0f, 0x00)

	// CmdIDReq ID Request
	CmdIDReq = Commands.RegisterStd("ID Req", DefaultCategories, MsgTypeDirect, 0x10, 0x00)

	CmdGetOperatingFlags = Commands.RegisterStd("Get Operating Flags", DefaultCategories, MsgTypeDirect, 0x1f, 0x00)

	CmdSetOperatingFlags = Commands.RegisterStd("Set Operating Flags", DefaultCategories, MsgTypeDirect, 0x20, 0x00)

	// CmdReadWriteALDB Read/Write ALDB
	CmdReadWriteALDB = Commands.RegisterExt("Read/Write ALDB", DefaultCategories, MsgTypeDirect, 0x2f, 0x00)
)

type CommandBytes struct {
	version    FirmwareVersion
	name       string
	subCommand bool
	Command1   byte
	Command2   byte
}

func (cb CommandBytes) Equal(other CommandBytes) bool {
	return cb.Command1 == other.Command1 && cb.Command2 == other.Command2
}

func (cb CommandBytes) SubCommand(command2 int) CommandBytes {
	return CommandBytes{
		version:    cb.version,
		name:       cb.name,
		subCommand: true,
		Command1:   cb.Command1,
		Command2:   byte(command2),
	}
}

func (cb CommandBytes) String() (str string) {
	if cb.name == "" {
		str = fmt.Sprintf("Unknown Command (%02x.%02x)", cb.Command1, cb.Command2)
	} else if cb.subCommand {
		str = fmt.Sprintf("%s(%d)", cb.name, cb.Command2)
	} else {
		str = cb.name
	}
	return str
}

type FirmwareIndex []*CommandBytes

func (fi FirmwareIndex) Len() int {
	return len(fi)
}

func (fi FirmwareIndex) Less(i, j int) bool {
	return fi[i].version < fi[j].version
}

func (fi FirmwareIndex) Swap(i, j int) {
	tmp := fi[i]
	fi[i] = fi[j]
	fi[j] = tmp
}

func (fi FirmwareIndex) Find(version FirmwareVersion) (command CommandBytes) {
	for i := len(fi) - 1; i >= 0; i-- {
		if fi[i].version <= version {
			command = *fi[i]
			break
		}
	}
	return command
}

func (fi *FirmwareIndex) Add(commandBytes *CommandBytes) {
	*fi = append(*fi, commandBytes)
	sort.Sort(*fi)
}

type Command struct {
	name       string
	versions   FirmwareIndex
	categories []Category
}

func NewCommand(name string, categories []Category) *Command {
	return &Command{
		name:       name,
		versions:   make(FirmwareIndex, 0),
		categories: categories,
	}
}

func (command *Command) Version(version FirmwareVersion) CommandBytes {
	return command.versions.Find(version)
}

func (command *Command) String() string {
	return command.name
}

type CommandIndex map[[2]byte]*CommandBytes

func (ci CommandIndex) Find(commandBytes CommandBytes) (command CommandBytes) {
	key := [2]byte{commandBytes.Command1, commandBytes.Command2}
	if _, found := ci[key]; found {
		command = *ci[key]
	} else {
		key[1] = 0x00
		if _, found := ci[key]; found {
			command = ci[key].SubCommand(int(commandBytes.Command2))
		}
	}
	return command
}

func (ci CommandIndex) Register(commandBytes *CommandBytes) {
	ci[[2]byte{commandBytes.Command1, commandBytes.Command2}] = commandBytes
}

type DevCatIndex map[Category]CommandIndex

func (dci DevCatIndex) Find(category Category, commandBytes CommandBytes) (command CommandBytes) {
	if index, found := dci[category]; found {
		command = index.Find(commandBytes)
	}

	if command.Command1 != commandBytes.Command1 && dci[Category(0)] != nil {
		command = dci[Category(0)].Find(commandBytes)
	}
	return command
}

func (dci DevCatIndex) Register(category Category, commandBytes *CommandBytes) {
	index := dci[category]
	if index == nil {
		index = make(CommandIndex)
		dci[category] = index
	}
	index.Register(commandBytes)
}

type CommandDB struct {
	standardCommands DevCatIndex
	extendedCommands DevCatIndex
}

func NewCommandDB() *CommandDB {
	return &CommandDB{
		standardCommands: make(DevCatIndex),
		extendedCommands: make(DevCatIndex),
	}
}

func (cdb *CommandDB) RegisterExt(name string, categories []Category, messageType MessageType, command1, command2 byte) *Command {
	command := NewCommand(name, categories)
	cdb.RegisterVersionExt(command, 0, command1, command2)
	return command
}

func (cdb *CommandDB) RegisterStd(name string, categories []Category, messageType MessageType, command1, command2 byte) *Command {
	command := NewCommand(name, categories)
	cdb.RegisterVersionStd(command, 0, command1, command2)
	return command
}

func (cdb *CommandDB) RegisterVersionStd(command *Command, version FirmwareVersion, command1, command2 byte) {
	commandBytes := &CommandBytes{name: command.name, version: version, Command1: command1, Command2: command2}
	command.versions.Add(commandBytes)
	for _, category := range command.categories {
		cdb.standardCommands.Register(category, commandBytes)
	}
}

func (cdb *CommandDB) RegisterVersionExt(command *Command, version FirmwareVersion, command1, command2 byte) {
	commandBytes := &CommandBytes{name: command.name, version: version, Command1: command1, Command2: command2}
	command.versions.Add(commandBytes)
	for _, category := range command.categories {
		cdb.extendedCommands.Register(category, commandBytes)
	}
}

func (cdb *CommandDB) FindStd(devCat DevCat, commandBytes CommandBytes) CommandBytes {
	return cdb.standardCommands.Find(devCat.Category(), commandBytes)
}

func (cdb *CommandDB) FindExt(devCat DevCat, commandBytes CommandBytes) CommandBytes {
	return cdb.extendedCommands.Find(devCat.Category(), commandBytes)
}
