// Copyright 2018 Andrew Bates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package insteon

// Standard Direct Commands
var (
	// CmdAssignToAllLinkGroup Assign to ALL-Link Group
	CmdAssignToAllLinkGroup = Command{0x00, 0x01, 0x00} // Assign to All-Link Group

	// CmdDeleteFromAllLinkGroup Delete from All-Link Group
	CmdDeleteFromAllLinkGroup = Command{0x00, 0x02, 0x00} // Delete from All-Link Group

	// CmdProductDataReq Product Data Request
	CmdProductDataReq = Command{0x00, 0x03, 0x00} // Product Data Request

	// CmdFxUsernameReq FX Username Request
	CmdFxUsernameReq = Command{0x00, 0x03, 0x01} // Fx Username Request

	// CmdDeviceTextStringReq Device Text String Request
	CmdDeviceTextStringReq = Command{0x00, 0x03, 0x02} // Text String Request

	// CmdExitLinkingMode Exit Linking Mode
	CmdExitLinkingMode = Command{0x00, 0x08, 0x00} // Exit Linking Mode

	// CmdEnterLinkingMode Enter Linking Mode
	CmdEnterLinkingMode = Command{0x00, 0x09, 0x00} // Enter Linking Mode

	// CmdEnterUnlinkingMode Enter Unlinking Mode
	CmdEnterUnlinkingMode = Command{0x00, 0x0a, 0x00} // Enter Unlinking Mode

	// CmdGetEngineVersion Get Insteon Engine Version
	CmdGetEngineVersion = Command{0x00, 0x0d, 0x00} // Engine Version

	// CmdPing Ping Request
	CmdPing = Command{0x00, 0x0f, 0x00} // Ping Request

	// CmdIDRequest Send ID Request which will prompt the device to respond with a Set Button Pressed Controller/Responder
	CmdIDRequest = Command{0x00, 0x10, 0x00} // ID Request

	// CmdGetOperatingFlags is used to request a given operating flag
	CmdGetOperatingFlags = Command{0x00, 0x1f, 0x00} // Get Operating Flags

	// CmdSetOperatingFlags is used to set a given operating flag
	CmdSetOperatingFlags = Command{0x00, 0x20, 0x00} // Set Operating Flags
)

// Extended Direct Commands
var (
	// CmdProductDataResp Product Data Response
	CmdProductDataResp = Command{0x01, 0x03, 0x00} // Product Data Response

	// CmdFxUsernameResp FX Username Response
	CmdFxUsernameResp = Command{0x01, 0x03, 0x01} // Fx Username Response

	// CmdDeviceTextStringResp Device Text String Response
	CmdDeviceTextStringResp = Command{0x01, 0x03, 0x02} // Text String Response

	// CmdSetDeviceTextString sets the device text string
	CmdSetDeviceTextString = Command{0x01, 0x03, 0x03} // Set Text String

	// CmdSetAllLinkCommandAlias sets the command alias to use for all-linking
	CmdSetAllLinkCommandAlias = Command{0x01, 0x03, 0x04} // Set All-Link Command Alias

	// CmdSetAllLinkCommandAliasData sets the extended data to be used if the command alias is an extended command
	CmdSetAllLinkCommandAliasData = Command{0x01, 0x03, 0x05} // Set All-Link Command Alias Data

	// CmdExitLinkingModeExt Exit Linking Mode
	CmdExitLinkingModeExt = Command{0x01, 0x08, 0x00} // Exit Linking Mode (i2cs)

	// CmdEnterLinkingModeExt Enter Linking Mode (extended command for I2CS devices)
	CmdEnterLinkingModeExt = Command{0x01, 0x09, 0x00} // Enter Linking Mode (i2cs)

	// CmdEnterUnlinkingModeExt Enter Unlinking Mode (extended command for I2CS devices)
	CmdEnterUnlinkingModeExt = Command{0x01, 0x0a, 0x00} // Enter Unlinking Mode (i2cs)

	// CmdExtendedGetSet is used to get and set extended data (ha ha)
	CmdExtendedGetSet = Command{0x01, 0x2e, 0x00} // Extended Get/Set

	// CmdReadWriteALDB Read/Write ALDB
	CmdReadWriteALDB = Command{0x01, 0x2f, 0x00} // Read/Write ALDB
)

// All-Link Messages
var (
	// CmdAllLinkRecall is an all-link command to recall the state assigned to the entry in the all-link database
	CmdAllLinkRecall = Command{0x0c, 0x11, 0x00} // All-link recall

	// CmdAllLinkAlias2High will execute substitute Direct Command
	CmdAllLinkAlias2High = Command{0x0c, 0x12, 0x00} // All-link Alias 2 High

	// CmdAllLinkAlias1Low will execute substitute Direct Command
	CmdAllLinkAlias1Low = Command{0x0c, 0x13, 0x00} // All-link Alias 1 Low

	// CmdAllLinkAlias2Low will execute substitute Direct Command
	CmdAllLinkAlias2Low = Command{0x0c, 0x14, 0x00} // All-link Alias 2 Low

	// CmdAllLinkAlias3High will execute substitute Direct Command
	CmdAllLinkAlias3High = Command{0x0c, 0x15, 0x00} // All-link Alias 3 High

	// CmdAllLinkAlias3Low will execute substitute Direct Command
	CmdAllLinkAlias3Low = Command{0x0c, 0x16, 0x00} // All-link Alias 3 Low

	// CmdAllLinkAlias4High will execute substitute Direct Command
	CmdAllLinkAlias4High = Command{0x0c, 0x17, 0x00} // All-link Alias 4 High

	// CmdAllLinkAlias4Low will execute substitute Direct Command
	CmdAllLinkAlias4Low = Command{0x0c, 0x18, 0x00} // All-link Alias 4 Low

	// CmdAllLinkAlias5 will execute substitute Direct Command
	CmdAllLinkAlias5 = Command{0x0c, 0x21, 0x00} // All-link Alias 5
)

// Standard Broadcast Messages
var (
	// CmdSetButtonPressedResponder Broadcast command indicating the set button has been pressed
	CmdSetButtonPressedResponder = Command{0x08, 0x01, 0x00} // Set-button Pressed (responder)

	// CmdSetButtonPressedController Broadcast command indicating the set button has been pressed
	CmdSetButtonPressedController = Command{0x08, 0x02, 0x00} // Set-button Pressed (controller)

	// CmdTestPowerlinePhase is used for determining which powerline phase (A/B) to which the device is attached
	CmdTestPowerlinePhase = Command{0x08, 0x03, 0x00} // Test Powerline Phase

	// CmdHeartbeat is a broadcast command that is received periodically if it has been set up using the extended get/set command
	CmdHeartbeat = Command{0x08, 0x04, 0x00} // Heartbeat

	// CmdBroadCastStatusChange is sent by a device when its status changes
	CmdBroadCastStatusChange = Command{0x08, 0x27, 0x00} // Broadcast Status Change
)

// Lighting Standard Direct Messages
var (
	// CmdLightOn
	CmdLightOn = Command{0x00, 0x11, 0xff} // Light On

	// CmdLightOnFast
	CmdLightOnFast = Command{0x00, 0x12, 0x00} // Light On Fast

	// CmdLightOff
	CmdLightOff = Command{0x00, 0x13, 0x00} // Light Off

	// CmdLightOffFast
	CmdLightOffFast = Command{0x00, 0x14, 0x00} // Light Off Fast

	// CmdLightBrighten
	CmdLightBrighten = Command{0x00, 0x15, 0x00} // Brighten Light

	// CmdLightDim
	CmdLightDim = Command{0x00, 0x16, 0x00} // Dim Light

	// CmdLightStartManual
	CmdLightStartManual = Command{0x00, 0x17, 0x00} // Manual Light Change Start

	// CmdLightStopManual
	CmdLightStopManual = Command{0x00, 0x18, 0x00} // Manual Light Change Stop

	// CmdLightStatusRequest
	CmdLightStatusRequest = Command{0x00, 0x19, 0x00} // Status Request

	// CmdLightInstantChange
	CmdLightInstantChange = Command{0x00, 0x21, 0x00} // Light Instant Change

	// CmdLightManualOn
	CmdLightManualOn = Command{0x00, 0x22, 0x01} // Manual On

	// CmdLightManualOff
	CmdLightManualOff = Command{0x00, 0x23, 0x01} // Manual Off

	// CmdTapSetButtonOnce
	CmdTapSetButtonOnce = Command{0x00, 0x25, 0x01} // Set Button Tap

	// CmdTapSetButtonTwice
	CmdTapSetButtonTwice = Command{0x00, 0x25, 0x02} // Set Button Tap Twice

	// CmdLightSetStatus
	CmdLightSetStatus = Command{0x00, 0x27, 0x00} // Set Status

	// CmdLightOnAtRamp
	CmdLightOnAtRamp = Command{0x00, 0x2e, 0x00} // Light On At Ramp

	// CmdLightOnAtRampV67
	CmdLightOnAtRampV67 = Command{0x00, 0x34, 0x00} // Light On At Ramp

	// CmdLightOffAtRamp
	CmdLightOffAtRamp = Command{0x00, 0x2f, 0x00} // Light Off At Ramp

	// CmdLightOffAtRampV67
	CmdLightOffAtRampV67 = Command{0x00, 0x35, 0x00} // Light Off At Ramp
)

var cmdStrings = map[Command]string{
	CmdAssignToAllLinkGroup:       "Assign to All-Link Group",
	CmdDeleteFromAllLinkGroup:     "Delete from All-Link Group",
	CmdProductDataReq:             "Product Data Request",
	CmdFxUsernameReq:              "Fx Username Request",
	CmdDeviceTextStringReq:        "Text String Request",
	CmdExitLinkingMode:            "Exit Linking Mode",
	CmdEnterLinkingMode:           "Enter Linking Mode",
	CmdEnterUnlinkingMode:         "Enter Unlinking Mode",
	CmdGetEngineVersion:           "Engine Version",
	CmdPing:                       "Ping Request",
	CmdIDRequest:                  "ID Request",
	CmdGetOperatingFlags:          "Get Operating Flags",
	CmdSetOperatingFlags:          "Set Operating Flags",
	CmdProductDataResp:            "Product Data Response",
	CmdFxUsernameResp:             "Fx Username Response",
	CmdDeviceTextStringResp:       "Text String Response",
	CmdSetDeviceTextString:        "Set Text String",
	CmdSetAllLinkCommandAlias:     "Set All-Link Command Alias",
	CmdSetAllLinkCommandAliasData: "Set All-Link Command Alias Data",
	CmdExitLinkingModeExt:         "Exit Linking Mode (i2cs)",
	CmdEnterLinkingModeExt:        "Enter Linking Mode (i2cs)",
	CmdEnterUnlinkingModeExt:      "Enter Unlinking Mode (i2cs)",
	CmdExtendedGetSet:             "Extended Get/Set",
	CmdReadWriteALDB:              "Read/Write ALDB",
	CmdAllLinkRecall:              "All-link recall",
	CmdAllLinkAlias2High:          "All-link Alias 2 High",
	CmdAllLinkAlias1Low:           "All-link Alias 1 Low",
	CmdAllLinkAlias2Low:           "All-link Alias 2 Low",
	CmdAllLinkAlias3High:          "All-link Alias 3 High",
	CmdAllLinkAlias3Low:           "All-link Alias 3 Low",
	CmdAllLinkAlias4High:          "All-link Alias 4 High",
	CmdAllLinkAlias4Low:           "All-link Alias 4 Low",
	CmdAllLinkAlias5:              "All-link Alias 5",
	CmdSetButtonPressedResponder:  "Set-button Pressed (responder)",
	CmdSetButtonPressedController: "Set-button Pressed (controller)",
	CmdTestPowerlinePhase:         "Test Powerline Phase",
	CmdHeartbeat:                  "Heartbeat",
	CmdBroadCastStatusChange:      "Broadcast Status Change",
	CmdLightOn:                    "Light On",
	CmdLightOnFast:                "Light On Fast",
	CmdLightOff:                   "Light Off",
	CmdLightOffFast:               "Light Off Fast",
	CmdLightBrighten:              "Brighten Light",
	CmdLightDim:                   "Dim Light",
	CmdLightStartManual:           "Manual Light Change Start",
	CmdLightStopManual:            "Manual Light Change Stop",
	CmdLightStatusRequest:         "Status Request",
	CmdLightInstantChange:         "Light Instant Change",
	CmdLightManualOn:              "Manual On",
	CmdLightManualOff:             "Manual Off",
	CmdTapSetButtonOnce:           "Set Button Tap",
	CmdTapSetButtonTwice:          "Set Button Tap Twice",
	CmdLightSetStatus:             "Set Status",
	CmdLightOnAtRamp:              "Light On At Ramp",
	CmdLightOnAtRampV67:           "Light On At Ramp",
	CmdLightOffAtRamp:             "Light Off At Ramp",
	CmdLightOffAtRampV67:          "Light Off At Ramp",
}
