package insteon

import "sort"

var (
	// CmdSetButtonPressedResponder Broadcast command indicating the set button has been pressed
	CmdSetButtonPressedResponder = RegisterStd("Set Button Pressed (responder)", MsgTypeBroadcast, 0x01, 0x00)

	// CmdSetButtonPressedController Broadcast command indicating the set button has been pressed
	CmdSetButtonPressedController = RegisterStd("Set Button Pressed (controller)", MsgTypeBroadcast, 0x02, 0x00)

	// CmdAssignToAllLinkGroup Assign to ALL-Link Group
	CmdAssignToAllLinkGroup = RegisterStd("All Link Assign", MsgTypeDirect, 0x01, 0x00)

	// CmdDeleteFromAllLinkGroup Delete from All-Link Group
	CmdDeleteFromAllLinkGroup = RegisterStd("All Link Delete", MsgTypeDirect, 0x02, 0x00)

	// CmdProductDataReq Product Data Request
	CmdProductDataReq = RegisterStd("Product Data Req", MsgTypeDirect, 0x03, 0x00)

	// CmdProductDataResp Product Data Response
	CmdProductDataResp = RegisterExt("Product Data Resp", MsgTypeDirect, 0x03, 0x00)

	// CmdFxUsernameReq FX Username Request
	CmdFxUsernameReq = RegisterStd("FX Username Req", MsgTypeDirect, 0x03, 0x01)

	// CmdFxUsernameResp FX Username Response
	CmdFxUsernameResp = RegisterExt("FX Username Resp", MsgTypeDirect, 0x03, 0x01)

	// CmdDeviceTextStringReq Device Text String Request
	CmdDeviceTextStringReq = RegisterStd("Text String Req", MsgTypeDirect, 0x03, 0x02)

	// CmdDeviceTextStringResp Device Text String Response
	CmdDeviceTextStringResp = RegisterExt("Text String Resp", MsgTypeDirect, 0x03, 0x02)

	// CmdSetDeviceTextString sets the device text string
	CmdSetDeviceTextString = RegisterExt("Set Text String", MsgTypeDirect, 0x03, 0x03)

	// CmdExitLinkingMode Exit Linking Mode
	CmdExitLinkingMode = RegisterStd("Exit Link Mode", MsgTypeDirect, 0x08, 0x00)

	// CmdExitLinkingModeExt Exit Linking Mode
	CmdExitLinkingModeExt = RegisterExt("Exit Link Mode", MsgTypeDirect, 0x08, 0x00)

	// CmdEnterLinkingMode Enter Linking Mode
	CmdEnterLinkingMode = RegisterStd("Enter Link Mode", MsgTypeDirect, 0x09, 0x00)

	// CmdEnterLinkingModeExt Enter Linking Mode (extended command for I2CS devices)
	CmdEnterLinkingModeExt = RegisterExt("Enter Link Mode", MsgTypeDirect, 0x09, 0x00)

	// CmdEnterUnlinkingMode Enter Unlinking Mode
	CmdEnterUnlinkingMode = RegisterStd("Enter Unlink Mode", MsgTypeDirect, 0x0a, 0x00)

	// CmdEnterUnlinkingModeExt Enter Unlinking Mode (extended command for I2CS devices)
	CmdEnterUnlinkingModeExt = RegisterExt("Enter Unlink Mode", MsgTypeDirect, 0x0a, 0x00)

	// CmdGetEngineVersion Get Insteon Engine Version
	CmdGetEngineVersion = RegisterStd("Get INSTEON Ver", MsgTypeDirect, 0x0d, 0x00)

	// CmdPing Ping Request
	CmdPing = RegisterStd("Ping", MsgTypeDirect, 0x0f, 0x00)

	// CmdIDReq ID Request
	CmdIDReq = RegisterStd("ID Req", MsgTypeDirect, 0x10, 0x00)

	CmdGetOperatingFlags = RegisterStd("Get Operating Flags", MsgTypeDirect, 0x1f, 0x00)

	CmdSetOperatingFlags = RegisterStd("Set Operating Flags", MsgTypeDirect, 0x20, 0x00)

	// CmdReadWriteALDB Read/Write ALDB
	CmdReadWriteALDB = RegisterExt("Read/Write ALDB", MsgTypeDirect, 0x2f, 0x00)
)

type CommandBytes struct {
	version  FirmwareVersion
	Command1 byte
	Command2 byte
}

func (cb CommandBytes) SubCommand(command2 int) CommandBytes {
	return CommandBytes{
		version:  cb.version,
		Command1: cb.Command1,
		Command2: byte(command2),
	}
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
	name     string
	versions FirmwareIndex
}

func NewCommand(name string) *Command {
	return &Command{
		name:     name,
		versions: make(FirmwareIndex, 0),
	}
}

func (command *Command) Register(version FirmwareVersion, command1, command2 byte) {
	command.versions.Add(&CommandBytes{version: version, Command1: command1, Command2: command2})
}

func (command *Command) Version(version FirmwareVersion) CommandBytes {
	return command.versions.Find(version)
}

func RegisterExt(name string, messageType MessageType, command1, command2 byte) *Command {
	command := NewCommand(name)
	command.Register(0, command1, command2)
	return command
}

func RegisterStd(name string, messageType MessageType, command1, command2 byte) *Command {
	command := NewCommand(name)
	command.Register(0, command1, command2)
	return command
}
