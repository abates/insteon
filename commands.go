// Copyright 2020 Andrew Bates
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
const (
	// CmdAssignToAllLinkGroup Assign to ALL-Link Group
	CmdAssignToAllLinkGroup = Command(0x000100) // Assign to All-Link Group

	// CmdDeleteFromAllLinkGroup Delete from All-Link Group
	CmdDeleteFromAllLinkGroup = Command(0x000200) // Delete from All-Link Group

	// CmdProductDataReq Product Data Request
	CmdProductDataReq = Command(0x000300) // Product Data Request

	// CmdFxUsernameReq FX Username Request
	CmdFxUsernameReq = Command(0x000301) // Fx Username Request

	// CmdDeviceTextStringReq Device Text String Request
	CmdDeviceTextStringReq = Command(0x000302) // Text String Request

	// CmdExitLinkingMode Exit Linking Mode
	CmdExitLinkingMode = Command(0x000800) // Exit Linking Mode

	// CmdEnterLinkingMode Enter Linking Mode
	CmdEnterLinkingMode = Command(0x000900) // Enter Linking Mode

	// CmdEnterUnlinkingMode Enter Unlinking Mode
	CmdEnterUnlinkingMode = Command(0x000a00) // Enter Unlinking Mode

	// CmdGetEngineVersion Get Insteon Engine Version
	CmdGetEngineVersion = Command(0x000d00) // Engine Version

	// CmdPing Ping Request
	CmdPing = Command(0x000f00) // Ping Request

	// CmdIDRequest Send ID Request which will prompt the device to respond with a Set Button Pressed Controller/Responder
	CmdIDRequest = Command(0x001000) // ID Request

	// CmdGetOperatingFlags is used to request a given operating flag
	CmdGetOperatingFlags = Command(0x001f00) // Get Operating Flags
)

// Extended Direct Commands
const (
	// CmdProductDataResp Product Data Response
	CmdProductDataResp = Command(0x010300) // Product Data Response

	// CmdFxUsernameResp FX Username Response
	CmdFxUsernameResp = Command(0x010301) // Fx Username Response

	// CmdDeviceTextStringResp Device Text String Response
	CmdDeviceTextStringResp = Command(0x010302) // Text String Response

	// CmdSetDeviceTextString sets the device text string
	CmdSetDeviceTextString = Command(0x010303) // Set Text String

	// CmdSetAllLinkCommandAlias sets the command alias to use for all-linking
	CmdSetAllLinkCommandAlias = Command(0x010304) // Set All-Link Command Alias

	// CmdSetAllLinkCommandAliasData sets the extended data to be used if the command alias is an extended command
	CmdSetAllLinkCommandAliasData = Command(0x010305) // Set All-Link Command Alias Data

	// CmdExitLinkingModeExt Exit Linking Mode
	CmdExitLinkingModeExt = Command(0x010800) // Exit Linking Mode (i2cs)

	// CmdEnterLinkingModeExt Enter Linking Mode (extended command for I2CS devices)
	CmdEnterLinkingModeExt = Command(0x010900) // Enter Linking Mode (i2cs)

	// CmdEnterUnlinkingModeExt Enter Unlinking Mode (extended command for I2CS devices)
	CmdEnterUnlinkingModeExt = Command(0x010a00) // Enter Unlinking Mode (i2cs)

	// CmdExtendedGetSet is used to get and set extended data (ha ha)
	CmdExtendedGetSet = Command(0x012e00) // Extended Get/Set

	// CmdReadWriteALDB Read/Write ALDB
	CmdReadWriteALDB = Command(0x012f00) // Read/Write ALDB
)

// All-Link Messages
const (
	// CmdAllLinkRecall is an all-link command to recall the state assigned to the entry in the all-link database
	CmdAllLinkRecall = Command(0x0c1100) // All-link recall

	// CmdAllLinkAlias2High will execute substitute Direct Command
	CmdAllLinkAlias2High = Command(0x0c1200) // All-link Alias 2 High

	// CmdAllLinkAlias1Low will execute substitute Direct Command
	CmdAllLinkAlias1Low = Command(0x0c1300) // All-link Alias 1 Low

	// CmdAllLinkAlias2Low will execute substitute Direct Command
	CmdAllLinkAlias2Low = Command(0x0c1400) // All-link Alias 2 Low

	// CmdAllLinkAlias3High will execute substitute Direct Command
	CmdAllLinkAlias3High = Command(0x0c1500) // All-link Alias 3 High

	// CmdAllLinkAlias3Low will execute substitute Direct Command
	CmdAllLinkAlias3Low = Command(0x0c1600) // All-link Alias 3 Low

	// CmdAllLinkAlias4High will execute substitute Direct Command
	CmdAllLinkAlias4High = Command(0x0c1700) // All-link Alias 4 High

	// CmdAllLinkAlias4Low will execute substitute Direct Command
	CmdAllLinkAlias4Low = Command(0x0c1800) // All-link Alias 4 Low

	// CmdAllLinkAlias5 will execute substitute Direct Command
	CmdAllLinkAlias5 = Command(0x0c2100) // All-link Alias 5
)

// Standard Broadcast Messages
const (
	// CmdSetButtonPressedResponder Broadcast command indicating the set button has been pressed
	CmdSetButtonPressedResponder = Command(0x080100) // Set-button Pressed (responder)

	// CmdSetButtonPressedController Broadcast command indicating the set button has been pressed
	CmdSetButtonPressedController = Command(0x080200) // Set-button Pressed (controller)

	// CmdTestPowerlinePhase is used for determining which powerline phase (A/B) to which the device is attached
	CmdTestPowerlinePhase = Command(0x080300) // Test Powerline Phase

	// CmdHeartbeat is a broadcast command that is received periodically if it has been set up using the extended get/set command
	CmdHeartbeat = Command(0x080400) // Heartbeat

	// CmdBroadCastStatusChange is sent by a device when its status changes
	CmdBroadCastStatusChange = Command(0x082700) // Broadcast Status Change
)

// Lighting Standard Direct Messages
const (
	// CmdLightOn
	CmdLightOn = Command(0x001100) // Light On

	// CmdLightOnFast
	CmdLightOnFast = Command(0x001200) // Light On Fast

	// CmdLightOff
	CmdLightOff = Command(0x001300) // Light Off

	// CmdLightOffFast
	CmdLightOffFast = Command(0x001400) // Light Off Fast

	// CmdLightBrighten
	CmdLightBrighten = Command(0x001500) // Brighten Light

	// CmdLightDim
	CmdLightDim = Command(0x001600) // Dim Light

	// CmdLightStopManual
	CmdLightStopManual = Command(0x001800) // Manual Light Change Stop

	// CmdLightStatusRequest
	CmdLightStatusRequest = Command(0x001900) // Status Request

	// CmdLightInstantChange
	CmdLightInstantChange = Command(0x002100) // Light Instant Change

	// CmdLightManualOn
	CmdLightManualOn = Command(0x002201) // Manual On

	// CmdLightManualOff
	CmdLightManualOff = Command(0x002301) // Manual Off

	// CmdTapSetButtonOnce
	CmdTapSetButtonOnce = Command(0x002501) // Set Button Tap

	// CmdTapSetButtonTwice
	CmdTapSetButtonTwice = Command(0x002502) // Set Button Tap Twice

	// CmdLightSetStatus
	CmdLightSetStatus = Command(0x002700) // Set Status

	// CmdLightOnAtRamp
	CmdLightOnAtRamp = Command(0x002e00) // Light On At Ramp

	// CmdLightOnAtRampV67
	CmdLightOnAtRampV67 = Command(0x003400) // Light On At Ramp

	// CmdLightOffAtRamp
	CmdLightOffAtRamp = Command(0x002f00) // Light Off At Ramp

	// CmdLightOffAtRampV67
	CmdLightOffAtRampV67 = Command(0x003500) // Light Off At Ramp
)

// Dimmer Convenience Commands
const (
	// CmdStartBrighten
	CmdStartBrighten = Command(0x001701) // Manual Start Brighten

	// CmdStartDim
	CmdStartDim = Command(0x001700) // Manual Start Dim

	// CmdEnableProgramLock
	CmdEnableProgramLock = Command(0x002000) // Enable Program Lock

	// CmdDisableProgramLock
	CmdDisableProgramLock = Command(0x002001) // Disable Program Lock

	// CmdEnableTxLED
	CmdEnableTxLED = Command(0x002002) // Enable Tx LED

	// CmdDisableTxLED
	CmdDisableTxLED = Command(0x002003) // Disable Tx LED

	// CmdEnableResumeDim
	CmdEnableResumeDim = Command(0x002004) // Enable Resume Dim

	// CmdDisableResumeDim
	CmdDisableResumeDim = Command(0x002005) // Disable Resume Dim

	// CmdEnableLoadSense
	CmdEnableLoadSense = Command(0x002006) // Enable Load Sense

	// CmdDisableLoadSense
	CmdDisableLoadSense = Command(0x002007) // Disable Load Sense

	// CmdDisableLED
	CmdDisableLED = Command(0x002008) // Disable Backlight

	// CmdEnableLED
	CmdEnableLED = Command(0x002009) // Enable Backlight

	// CmdSetKeyBeep
	CmdSetKeyBeep = Command(0x00200a) // Enable Key Beep

	// CmdClearKeyBeep
	CmdClearKeyBeep = Command(0x00200b) // Disable Key Beep
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
	CmdStartBrighten:              "Manual Start Brighten",
	CmdStartDim:                   "Manual Start Dim",
	CmdEnableProgramLock:          "Enable Program Lock",
	CmdDisableProgramLock:         "Disable Program Lock",
	CmdEnableTxLED:                "Enable Tx LED",
	CmdDisableTxLED:               "Disable Tx LED",
	CmdEnableResumeDim:            "Enable Resume Dim",
	CmdDisableResumeDim:           "Disable Resume Dim",
	CmdEnableLoadSense:            "Enable Load Sense",
	CmdDisableLoadSense:           "Disable Load Sense",
	CmdDisableLED:                 "Disable Backlight",
	CmdEnableLED:                  "Enable Backlight",
	CmdSetKeyBeep:                 "Enable Key Beep",
	CmdClearKeyBeep:               "Disable Key Beep",
}
