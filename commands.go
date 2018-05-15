package insteon

var (
	// CmdSetButtonPressedResponder Broadcast command indicating the set button has been pressed
	CmdSetButtonPressedResponder = Command{0x01, 0x00}

	// CmdSetButtonPressedController Broadcast command indicating the set button has been pressed
	CmdSetButtonPressedController = Command{0x02, 0x00}

	// CmdAssignToAllLinkGroup Assign to ALL-Link Group
	CmdAssignToAllLinkGroup = Command{0x01, 0x00}

	// CmdDeleteFromAllLinkGroup Delete from All-Link Group
	CmdDeleteFromAllLinkGroup = Command{0x02, 0x00}

	// CmdTestPowerlinePhase is used for determining which powerline phase (A/B) to which the device is attached
	CmdTestPowerlinePhase = Command{0x03, 0x00}

	// CmdProductDataReq Product Data Request
	CmdProductDataReq = Command{0x03, 0x00}

	// CmdProductDataResp Product Data Response
	CmdProductDataResp = Command{0x03, 0x00}

	// CmdFxUsernameReq FX Username Request
	CmdFxUsernameReq = Command{0x03, 0x01}

	// CmdFxUsernameResp FX Username Response
	CmdFxUsernameResp = Command{0x03, 0x01}

	// CmdDeviceTextStringReq Device Text String Request
	CmdDeviceTextStringReq = Command{0x03, 0x02}

	// CmdDeviceTextStringResp Device Text String Response
	CmdDeviceTextStringResp = Command{0x03, 0x02}

	// CmdSetDeviceTextString sets the device text string
	CmdSetDeviceTextString = Command{0x03, 0x03}

	CmdHeartbeat = Command{0x04, 0x00}

	CmdAllLinkCleanupReport = Command{0x06, 0x00}

	// CmdExitLinkingMode Exit Linking Mode
	CmdExitLinkingMode = Command{0x08, 0x00}

	// CmdExitLinkingModeExt Exit Linking Mode
	CmdExitLinkingModeExt = Command{0x08, 0x00}

	// CmdEnterLinkingMode Enter Linking Mode
	CmdEnterLinkingMode = Command{0x09, 0x00}

	// CmdEnterLinkingModeExt Enter Linking Mode (extended command for I2CS devices)
	CmdEnterLinkingModeExt = Command{0x09, 0x00}

	// CmdEnterUnlinkingMode Enter Unlinking Mode
	CmdEnterUnlinkingMode = Command{0x0a, 0x00}

	// CmdEnterUnlinkingModeExt Enter Unlinking Mode (extended command for I2CS devices)
	CmdEnterUnlinkingModeExt = Command{0x0a, 0x00}

	// CmdGetEngineVersion Get Insteon Engine Version
	CmdGetEngineVersion = Command{0x0d, 0x00}

	// CmdPing Ping Request
	CmdPing = Command{0x0f, 0x00}

	// CmdIDRequest Send ID Request which will prompt the device to respond with a Set Button Pressed Controller/Responder
	CmdIDRequest = Command{0x10, 0x00}

	CmdAllLinkRecall = Command{0x11, 0x00}

	CmdAllLinkAlias2High = Command{0x12, 0x00}

	CmdAllLinkAlias1Low = Command{0x13, 0x00}

	CmdAllLinkAlias2Low = Command{0x14, 0x00}

	CmdAllLinkAlias3High = Command{0x15, 0x00}

	CmdAllLinkAlias3Low = Command{0x16, 0x00}

	CmdAllLinkAlias4High = Command{0x17, 0x00}

	CmdAllLinkAlias4Low = Command{0x18, 0x00}

	CmdGetOperatingFlags = Command{0x1f, 0x00}

	CmdSetOperatingFlags = Command{0x20, 0x00}

	CmdAllLinkAlias5 = Command{0x21, 0x00}

	CmdBroadCastStatusChange = Command{0x27, 0x00}

	// CmdReadWriteALDB Read/Write ALDB
	CmdReadWriteALDB = Command{0x2f, 0x00}
)

type Command [2]byte

func (cmd Command) SubCommand(command2 int) Command {
	return Command{cmd[0], byte(command2)}
}
