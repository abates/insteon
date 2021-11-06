// Copyright 2021 Andrew Bates
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

package commands

// Standard Direct Commands
const (
	// AssignToAllLinkGroup Assign to ALL-Link Group
	AssignToAllLinkGroup = Command(0x000100) // Assign to All-Link Group

	// DeleteFromAllLinkGroup Delete from All-Link Group
	DeleteFromAllLinkGroup = Command(0x000200) // Delete from All-Link Group

	// ProductDataReq Product Data Request
	ProductDataReq = Command(0x000300) // Product Data Request

	// FxUsernameReq FX Username Request
	FxUsernameReq = Command(0x000301) // Fx Username Request

	// DeviceTextStringReq Device Text String Request
	DeviceTextStringReq = Command(0x000302) // Text String Request

	// ExitLinkingMode Exit Linking Mode
	ExitLinkingMode = Command(0x000800) // Exit Linking Mode

	// EnterLinkingMode Enter Linking Mode
	EnterLinkingMode = Command(0x000900) // Enter Linking Mode

	// EnterUnlinkingMode Enter Unlinking Mode
	EnterUnlinkingMode = Command(0x000a00) // Enter Unlinking Mode

	// GetEngineVersion Get Insteon Engine Version
	GetEngineVersion = Command(0x000d00) // Engine Version

	// Ping Ping Request
	Ping = Command(0x000f00) // Ping Request

	// IDRequest Send ID Request which will prompt the device to respond with a Set Button Pressed Controller/Responder
	IDRequest = Command(0x001000) // ID Request

	// GetOperatingFlags is used to request a given operating flag
	GetOperatingFlags = Command(0x001f00) // Get Operating Flags
)

// Extended Direct Commands
const (
	// ProductDataResp Product Data Response
	ProductDataResp = Command(0x010300) // Product Data Response

	// FxUsernameResp FX Username Response
	FxUsernameResp = Command(0x010301) // Fx Username Response

	// DeviceTextStringResp Device Text String Response
	DeviceTextStringResp = Command(0x010302) // Text String Response

	// SetDeviceTextString sets the device text string
	SetDeviceTextString = Command(0x010303) // Set Text String

	// SetAllLinkCommandAlias sets the command alias to use for all-linking
	SetAllLinkCommandAlias = Command(0x010304) // Set All-Link Command Alias

	// SetAllLinkCommandAliasData sets the extended data to be used if the command alias is an extended command
	SetAllLinkCommandAliasData = Command(0x010305) // Set All-Link Command Alias Data

	// ExitLinkingModeExt Exit Linking Mode
	ExitLinkingModeExt = Command(0x010800) // Exit Linking Mode (i2cs)

	// EnterLinkingModeExt Enter Linking Mode (extended command for I2CS devices)
	EnterLinkingModeExt = Command(0x010900) // Enter Linking Mode (i2cs)

	// EnterUnlinkingModeExt Enter Unlinking Mode (extended command for I2CS devices)
	EnterUnlinkingModeExt = Command(0x010a00) // Enter Unlinking Mode (i2cs)

	// SetOperatingFlags is used to set a given operating flag
	SetOperatingFlags = Command(0x012000) // Set Operating Flags

	// ExtendedGetSet is used to get and set extended data (ha ha)
	ExtendedGetSet = Command(0x012e00) // Extended Get/Set

	// ReadWriteALDB Read/Write ALDB
	ReadWriteALDB = Command(0x012f00) // Read/Write ALDB
)

// All-Link Messages
const (
	// AllLinkSuccessReport is an all-link command to report the number of failed cleanups
	AllLinkSuccessReport = Command(0x0c0600) // All-link Success Report

	// AllLinkRecall is an all-link command to recall the state assigned to the entry in the all-link database
	AllLinkRecall = Command(0x0c1100) // All-link recall

	// AllLinkAlias2High will execute substitute Direct Command
	AllLinkAlias2High = Command(0x0c1200) // All-link Alias 2 High

	// AllLinkAlias1Low will execute substitute Direct Command
	AllLinkAlias1Low = Command(0x0c1300) // All-link Alias 1 Low

	// AllLinkAlias2Low will execute substitute Direct Command
	AllLinkAlias2Low = Command(0x0c1400) // All-link Alias 2 Low

	// AllLinkAlias3High will execute substitute Direct Command
	AllLinkAlias3High = Command(0x0c1500) // All-link Alias 3 High

	// AllLinkAlias3Low will execute substitute Direct Command
	AllLinkAlias3Low = Command(0x0c1600) // All-link Alias 3 Low

	// AllLinkAlias4High will execute substitute Direct Command
	AllLinkAlias4High = Command(0x0c1700) // All-link Alias 4 High

	// AllLinkAlias4Low will execute substitute Direct Command
	AllLinkAlias4Low = Command(0x0c1800) // All-link Alias 4 Low

	// AllLinkAlias5 will execute substitute Direct Command
	AllLinkAlias5 = Command(0x0c2100) // All-link Alias 5
)

// Standard Broadcast Messages
const (
	// SetButtonPressedResponder Broadcast command indicating the set button has been pressed
	SetButtonPressedResponder = Command(0x080100) // Set-button Pressed (responder)

	// SetButtonPressedController Broadcast command indicating the set button has been pressed
	SetButtonPressedController = Command(0x080200) // Set-button Pressed (controller)

	// TestPowerlinePhase is used for determining which powerline phase (A/B) to which the device is attached
	TestPowerlinePhase = Command(0x080300) // Test Powerline Phase

	// Heartbeat is a broadcast command that is received periodically if it has been set up using the extended get/set command
	Heartbeat = Command(0x080400) // Heartbeat

	// BroadCastStatusChange is sent by a device when its status changes
	BroadCastStatusChange = Command(0x082700) // Broadcast Status Change
)

// Lighting Standard Direct Messages
const (
	// LightOn
	LightOn = Command(0x001100) // Light On

	// LightOnFast
	LightOnFast = Command(0x001200) // Light On Fast

	// LightOff
	LightOff = Command(0x001300) // Light Off

	// LightOffFast
	LightOffFast = Command(0x001400) // Light Off Fast

	// LightBrighten
	LightBrighten = Command(0x001500) // Brighten Light

	// LightDim
	LightDim = Command(0x001600) // Dim Light

	// LightStopManual
	LightStopManual = Command(0x001800) // Manual Light Change Stop

	// LightStatusRequest
	LightStatusRequest = Command(0x001900) // Status Request

	// LightInstantChange
	LightInstantChange = Command(0x002100) // Light Instant Change

	// LightManualOn
	LightManualOn = Command(0x002201) // Manual On

	// LightManualOff
	LightManualOff = Command(0x002301) // Manual Off

	// TapSetButtonOnce
	TapSetButtonOnce = Command(0x002501) // Set Button Tap

	// TapSetButtonTwice
	TapSetButtonTwice = Command(0x002502) // Set Button Tap Twice

	// LightSetStatus
	LightSetStatus = Command(0x002700) // Set Status

	// LightOnAtRamp
	LightOnAtRamp = Command(0x002e00) // Light On At Ramp

	// LightOnAtRampV67
	LightOnAtRampV67 = Command(0x003400) // Light On At Ramp

	// LightOffAtRamp
	LightOffAtRamp = Command(0x002f00) // Light Off At Ramp

	// LightOffAtRampV67
	LightOffAtRampV67 = Command(0x003500) // Light Off At Ramp
)

// Dimmer Convenience Commands
const (
	// StartBrighten
	StartBrighten = Command(0x001701) // Manual Start Brighten

	// StartDim
	StartDim = Command(0x001700) // Manual Start Dim

	// EnableProgramLock
	EnableProgramLock = Command(0x002000) // Enable Program Lock

	// DisableProgramLock
	DisableProgramLock = Command(0x002001) // Disable Program Lock

	// EnableTxLED
	EnableTxLED = Command(0x002002) // Enable Tx LED

	// DisableTxLED
	DisableTxLED = Command(0x002003) // Disable Tx LED

	// EnableResumeDim
	EnableResumeDim = Command(0x002004) // Enable Resume Dim

	// DisableResumeDim
	DisableResumeDim = Command(0x002005) // Disable Resume Dim

	// EnableLoadSense
	EnableLoadSense = Command(0x002006) // Enable Load Sense

	// DisableLoadSense
	DisableLoadSense = Command(0x002007) // Disable Load Sense

	// DisableLED
	DisableLED = Command(0x002008) // Disable Backlight

	// EnableLED
	EnableLED = Command(0x002009) // Enable Backlight

	// SetKeyBeep
	SetKeyBeep = Command(0x00200a) // Enable Key Beep

	// ClearKeyBeep
	ClearKeyBeep = Command(0x00200b) // Disable Key Beep
)

// Thermostat Standard Direct Messages
const (
	// DecreaseTemp Decrease Temperature
	DecreaseTemp = Command(0x006800) // Decrease Temp

	// IncreaseTemp Increase Temperature
	IncreaseTemp = Command(0x006900) // Increase Temp

	// GetZoneInfo Get Zone Information
	GetZoneInfo = Command(0x006a00) // Get Zone Info

	// GetThermostatMode Get Mode
	GetThermostatMode = Command(0x006b02) // Get Mode

	// GetAmbientTemp Get Ambient Temperature
	GetAmbientTemp = Command(0x006b03) // Get Ambient Temp

	// SetHeat Set Heat
	SetHeat = Command(0x006b04) // Set Heat

	// SetCool Set Cool
	SetCool = Command(0x006b05) // Set Cool

	// SetModeAuto Set Mode Auto
	SetModeAuto = Command(0x006b06) // Set Auto

	// SetFan Turn Fan On
	SetFan = Command(0x006b07) // Turn Fan On

	// ClearFan Turn Fan Off
	ClearFan = Command(0x006b08) // Turn Fan Off

	// ThermOff Turn Thermostat Off
	ThermOff = Command(0x006b09) // Turn Thermostat Off

	// SetProgramHeat Set mode to Program Heat
	SetProgramHeat = Command(0x006b0a) // Set Program Heat

	// SetProgramCool Set mode to Program Cool
	SetProgramCool = Command(0x006b0b) // Set Program Cool

	// SetProgramAuto Set mode to Program Auto
	SetProgramAuto = Command(0x006b0c) // Set Program Auto

	// GetEquipmentState Get Equipment State
	GetEquipmentState = Command(0x006b0d) // Get State

	// SetEquipmentState Set Equipment State
	SetEquipmentState = Command(0x006b0e) // Set State

	// GetTempUnits Get Temperature Units
	GetTempUnits = Command(0x006b0f) // Get Temp Units

	// SetFahrenheit Set Units to Fahrenheit
	SetFahrenheit = Command(0x006b10) // Set Units Fahrenheit

	// SetCelsius Set Units to Celsius
	SetCelsius = Command(0x006b11) // Set Units Celsius

	// GetFanOnSpeed Get Fan On-Speed
	GetFanOnSpeed = Command(0x006b12) // Get Fan On-Speed

	// SetFanOnLow Set Fan On-Speed to Low
	SetFanOnLow = Command(0x006b13) // Set Fan-Speed Low

	// SetFanOnMed Set Fan On-Speed to Med
	SetFanOnMed = Command(0x006b14) // Set Fan-Speed Med

	// SetFanOnHigh Set Fan On-Speed to High
	SetFanOnHigh = Command(0x006b15) // Set Fan-Speed High

	// EnableStatusMessage Enable Status Change Messages
	EnableStatusMessage = Command(0x006b16) // Enable Status Change

	// DisableStatusMessage Disable Status Change Messages
	DisableStatusMessage = Command(0x006b17) // Disable Status Change

	// SetCoolSetpoint Set Cool Set-Point
	SetCoolSetpoint = Command(0x006c00) // Set Cool Set-Point

	// SetHeatSetpoint Set Heat Set-Point
	SetHeatSetpoint = Command(0x006d00) // Set Heat Set-Point
)

// Thermostat Extended Direct Messages
const (
	// ZoneTempUp Increase Zone Temp
	ZoneTempUp = Command(0x016800) // Increase Zone Temp

	// ZoneTempDown Decrease Zone Temp
	ZoneTempDown = Command(0x016900) // Decrease Zone Temp

	// SetZoneCoolSetpoint Set Zone Cooling Set Point
	SetZoneCoolSetpoint = Command(0x016c00) // Set Zone Cool Set-Point

	// SetZoneHeatSetpoint Set Zone Heating Set Point
	SetZoneHeatSetpoint = Command(0x016d00) // Set Zone Heat Set-Point
)

var cmdStrings = map[Command]string{
	AssignToAllLinkGroup:       "Assign to All-Link Group",
	DeleteFromAllLinkGroup:     "Delete from All-Link Group",
	ProductDataReq:             "Product Data Request",
	FxUsernameReq:              "Fx Username Request",
	DeviceTextStringReq:        "Text String Request",
	ExitLinkingMode:            "Exit Linking Mode",
	EnterLinkingMode:           "Enter Linking Mode",
	EnterUnlinkingMode:         "Enter Unlinking Mode",
	GetEngineVersion:           "Engine Version",
	Ping:                       "Ping Request",
	IDRequest:                  "ID Request",
	GetOperatingFlags:          "Get Operating Flags",
	ProductDataResp:            "Product Data Response",
	FxUsernameResp:             "Fx Username Response",
	DeviceTextStringResp:       "Text String Response",
	SetDeviceTextString:        "Set Text String",
	SetAllLinkCommandAlias:     "Set All-Link Command Alias",
	SetAllLinkCommandAliasData: "Set All-Link Command Alias Data",
	ExitLinkingModeExt:         "Exit Linking Mode (i2cs)",
	EnterLinkingModeExt:        "Enter Linking Mode (i2cs)",
	EnterUnlinkingModeExt:      "Enter Unlinking Mode (i2cs)",
	SetOperatingFlags:          "Set Operating Flags",
	ExtendedGetSet:             "Extended Get/Set",
	ReadWriteALDB:              "Read/Write ALDB",
	AllLinkSuccessReport:       "All-link Success Report",
	AllLinkRecall:              "All-link recall",
	AllLinkAlias2High:          "All-link Alias 2 High",
	AllLinkAlias1Low:           "All-link Alias 1 Low",
	AllLinkAlias2Low:           "All-link Alias 2 Low",
	AllLinkAlias3High:          "All-link Alias 3 High",
	AllLinkAlias3Low:           "All-link Alias 3 Low",
	AllLinkAlias4High:          "All-link Alias 4 High",
	AllLinkAlias4Low:           "All-link Alias 4 Low",
	AllLinkAlias5:              "All-link Alias 5",
	SetButtonPressedResponder:  "Set-button Pressed (responder)",
	SetButtonPressedController: "Set-button Pressed (controller)",
	TestPowerlinePhase:         "Test Powerline Phase",
	Heartbeat:                  "Heartbeat",
	BroadCastStatusChange:      "Broadcast Status Change",
	LightOn:                    "Light On",
	LightOnFast:                "Light On Fast",
	LightOff:                   "Light Off",
	LightOffFast:               "Light Off Fast",
	LightBrighten:              "Brighten Light",
	LightDim:                   "Dim Light",
	LightStopManual:            "Manual Light Change Stop",
	LightStatusRequest:         "Status Request",
	LightInstantChange:         "Light Instant Change",
	LightManualOn:              "Manual On",
	LightManualOff:             "Manual Off",
	TapSetButtonOnce:           "Set Button Tap",
	TapSetButtonTwice:          "Set Button Tap Twice",
	LightSetStatus:             "Set Status",
	LightOnAtRamp:              "Light On At Ramp",
	LightOnAtRampV67:           "Light On At Ramp",
	LightOffAtRamp:             "Light Off At Ramp",
	LightOffAtRampV67:          "Light Off At Ramp",
	StartBrighten:              "Manual Start Brighten",
	StartDim:                   "Manual Start Dim",
	EnableProgramLock:          "Enable Program Lock",
	DisableProgramLock:         "Disable Program Lock",
	EnableTxLED:                "Enable Tx LED",
	DisableTxLED:               "Disable Tx LED",
	EnableResumeDim:            "Enable Resume Dim",
	DisableResumeDim:           "Disable Resume Dim",
	EnableLoadSense:            "Enable Load Sense",
	DisableLoadSense:           "Disable Load Sense",
	DisableLED:                 "Disable Backlight",
	EnableLED:                  "Enable Backlight",
	SetKeyBeep:                 "Enable Key Beep",
	ClearKeyBeep:               "Disable Key Beep",
	DecreaseTemp:               "Decrease Temp",
	IncreaseTemp:               "Increase Temp",
	GetZoneInfo:                "Get Zone Info",
	GetThermostatMode:          "Get Mode",
	GetAmbientTemp:             "Get Ambient Temp",
	SetHeat:                    "Set Heat",
	SetCool:                    "Set Cool",
	SetModeAuto:                "Set Auto",
	SetFan:                     "Turn Fan On",
	ClearFan:                   "Turn Fan Off",
	ThermOff:                   "Turn Thermostat Off",
	SetProgramHeat:             "Set Program Heat",
	SetProgramCool:             "Set Program Cool",
	SetProgramAuto:             "Set Program Auto",
	GetEquipmentState:          "Get State",
	SetEquipmentState:          "Set State",
	GetTempUnits:               "Get Temp Units",
	SetFahrenheit:              "Set Units Fahrenheit",
	SetCelsius:                 "Set Units Celsius",
	GetFanOnSpeed:              "Get Fan On-Speed",
	SetFanOnLow:                "Set Fan-Speed Low",
	SetFanOnMed:                "Set Fan-Speed Med",
	SetFanOnHigh:               "Set Fan-Speed High",
	EnableStatusMessage:        "Enable Status Change",
	DisableStatusMessage:       "Disable Status Change",
	SetCoolSetpoint:            "Set Cool Set-Point",
	SetHeatSetpoint:            "Set Heat Set-Point",
	ZoneTempUp:                 "Increase Zone Temp",
	ZoneTempDown:               "Decrease Zone Temp",
	SetZoneCoolSetpoint:        "Set Zone Cool Set-Point",
	SetZoneHeatSetpoint:        "Set Zone Heat Set-Point",
}
