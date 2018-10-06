//go:generate stringer -type=Command -linecomment=true

package insteon

const (
	// CmdSetButtonPressedResponder Broadcast command indicating the set button has been pressed
	CmdSetButtonPressedResponder Command = 0x0100 // Set-button Pressed (responder)

	// CmdSetButtonPressedController Broadcast command indicating the set button has been pressed
	CmdSetButtonPressedController Command = 0x0200 // Set-button Pressed (controller)

	// CmdAssignToAllLinkGroup Assign to ALL-Link Group
	CmdAssignToAllLinkGroup Command = 0x0100 // Assign to All-Link Group

	// CmdDeleteFromAllLinkGroup Delete from All-Link Group
	CmdDeleteFromAllLinkGroup Command = 0x0200 // Delete from All-Link Group

	// CmdTestPowerlinePhase is used for determining which powerline phase (A/B) to which the device is attached
	CmdTestPowerlinePhase Command = 0x0300 // Test Powerline Phase

	// CmdProductDataReq Product Data Request
	CmdProductDataReq Command = 0x0300 // Product Data Request

	// CmdProductDataResp Product Data Response
	CmdProductDataResp Command = 0x0300 // Product Data Response

	// CmdFxUsernameReq FX Username Request
	CmdFxUsernameReq Command = 0x0301 // Fx Username Request

	// CmdFxUsernameResp FX Username Response
	CmdFxUsernameResp Command = 0x0301 // Fx Username Response

	// CmdDeviceTextStringReq Device Text String Request
	CmdDeviceTextStringReq Command = 0x0302 // Text String Request

	// CmdDeviceTextStringResp Device Text String Response
	CmdDeviceTextStringResp Command = 0x0302 // Text String Response

	// CmdSetDeviceTextString sets the device text string
	CmdSetDeviceTextString Command = 0x0303 // Set Text String

	// CmdHeartbeat is a broadcast command that is received periodically if it
	// has been set up using the extended get/set command
	CmdHeartbeat Command = 0x0400 // Heartbeat

	// CmdExitLinkingMode Exit Linking Mode
	CmdExitLinkingMode Command = 0x0800 // Exit Linking Mode

	// CmdExitLinkingModeExt Exit Linking Mode
	CmdExitLinkingModeExt Command = 0x0800 // Exit Linking-mode (i2cs)

	// CmdEnterLinkingMode Enter Linking Mode
	CmdEnterLinkingMode Command = 0x0900 // Enter Linking-mode

	// CmdEnterLinkingModeExt Enter Linking Mode (extended command for I2CS devices)
	CmdEnterLinkingModeExt Command = 0x0900 // Enter Linking-mode (i2cs)

	// CmdEnterUnlinkingMode Enter Unlinking Mode
	CmdEnterUnlinkingMode Command = 0x0a00 // Enter Unlinking-mode

	// CmdEnterUnlinkingModeExt Enter Unlinking Mode (extended command for I2CS devices)
	CmdEnterUnlinkingModeExt Command = 0x0a00 // Enter Unlinking-mode (i2cs)

	// CmdGetEngineVersion Get Insteon Engine Version
	CmdGetEngineVersion Command = 0x0d00 // Engine Version

	// CmdPing Ping Request
	CmdPing Command = 0x0f00 // Ping Request

	// CmdIDRequest Send ID Request which will prompt the device to respond with a Set Button Pressed Controller/Responder
	CmdIDRequest Command = 0x1000 // ID Request

	// CmdAllLinkRecall is an all-link command to recall the state assigned to the entry in the all-link database
	CmdAllLinkRecall Command = 0x1100 // All-link recall

	// CmdAllLinkAlias2High will execute substitute Direct Command
	CmdAllLinkAlias2High Command = 0x1200 // All-link Alias 2 High

	// CmdAllLinkAlias1Low will execute substitute Direct Command
	CmdAllLinkAlias1Low Command = 0x1300 // All-link Alias 1 Low

	// CmdAllLinkAlias2Low will execute substitute Direct Command
	CmdAllLinkAlias2Low Command = 0x1400 // All-link Alias 2 Low

	// CmdAllLinkAlias3High will execute substitute Direct Command
	CmdAllLinkAlias3High Command = 0x1500 // All-link Alias 3 High

	// CmdAllLinkAlias3Low will execute substitute Direct Command
	CmdAllLinkAlias3Low Command = 0x1600 // All-link Alias 3 Low

	// CmdAllLinkAlias4High will execute substitute Direct Command
	CmdAllLinkAlias4High Command = 0x1700 // All-link Alias 4 High

	// CmdAllLinkAlias4Low will execute substitute Direct Command
	CmdAllLinkAlias4Low Command = 0x1800 // All-link Alias 4 Low

	// CmdGetOperatingFlags is used to request a given operating flag
	CmdGetOperatingFlags Command = 0x1f00 // Get Operating Flags

	// CmdSetOperatingFlags is used to set a given operating flag
	CmdSetOperatingFlags Command = 0x2000 // Set Operating Flags

	// CmdAllLinkAlias5 will execute substitute Direct Command
	CmdAllLinkAlias5 Command = 0x2100 // All-link Alias 5

	// CmdBroadCastStatusChange is sent by a device when its status changes
	CmdBroadCastStatusChange Command = 0x2700 // Broadcast Status Change

	// CmdReadWriteALDB Read/Write ALDB
	CmdReadWriteALDB Command = 0x2f00 // Read/Write ALDB
)

// Command is a two byte sequence in an Insteon message that indicates what action to
// take or what kind of information is contained in the message
type Command uint16

// SubCommand will return a new command where the subcommand byte is updated
// to reflect command2 from the arguments
func (cmd Command) SubCommand(command2 int) Command {
	return Command(uint16(cmd)&0xff00 | uint16(0xff&command2))
}
