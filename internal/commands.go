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

package main

type command struct {
	Name    string
	Comment string
	String  string
	Byte1   byte
	Byte2   byte
	Notes   string
}

type commandGroup struct {
	Name        string
	Byte0       byte
	Convenience bool
	Commands    []command
}

var commands = []commandGroup{
	{
		Name:        "Standard Direct Commands",
		Byte0:       0x00,
		Convenience: false,
		Commands: []command{
			{"AssignToAllLinkGroup", "Assign to ALL-Link Group", "Assign to All-Link Group", 0x01, 0x00, ""},
			{"DeleteFromAllLinkGroup", "Delete from All-Link Group", "Delete from All-Link Group", 0x02, 0x00, ""},
			{"ProductDataReq", "Product Data Request", "Product Data Request", 0x03, 0x00, ""},
			{"FxUsernameReq", "FX Username Request", "Fx Username Request", 0x03, 0x01, ""},
			{"DeviceTextStringReq", "Device Text String Request", "Text String Request", 0x03, 0x02, ""},
			{"ExitLinkingMode", "Exit Linking Mode", "Exit Linking Mode", 0x08, 0x00, ""},
			{"EnterLinkingMode", "Enter Linking Mode", "Enter Linking Mode", 0x09, 0x00, ""},
			{"EnterUnlinkingMode", "Enter Unlinking Mode", "Enter Unlinking Mode", 0x0a, 0x00, ""},
			{"GetEngineVersion", "Get Insteon Engine Version", "Engine Version", 0x0d, 0x00, ""},
			{"Ping", "Ping Request", "Ping Request", 0x0f, 0x00, ""},
			{"IDRequest", "Send ID Request which will prompt the device to respond with a Set Button Pressed Controller/Responder", "ID Request", 0x10, 0x00, ""},
			{"GetOperatingFlags", "is used to request a given operating flag", "Get Operating Flags", 0x1f, 0x00, ""},
		},
	},
	{
		Name:        "Extended Direct Commands",
		Byte0:       0x01,
		Convenience: false,
		Commands: []command{
			{"ProductDataResp", "Product Data Response", "Product Data Response", 0x03, 0x00, ""},
			{"FxUsernameResp", "FX Username Response", "Fx Username Response", 0x03, 0x01, ""},
			{"DeviceTextStringResp", "Device Text String Response", "Text String Response", 0x03, 0x02, ""},
			{"SetDeviceTextString", "sets the device text string", "Set Text String", 0x03, 0x03, ""},
			{"SetAllLinkCommandAlias", "sets the command alias to use for all-linking", "Set All-Link Command Alias", 0x03, 0x04, ""},
			{"SetAllLinkCommandAliasData", "sets the extended data to be used if the command alias is an extended command", "Set All-Link Command Alias Data", 0x03, 0x05, ""},
			{"ExitLinkingModeExt", "Exit Linking Mode", "Exit Linking Mode (i2cs)", 0x08, 0x00, "Insteon version 2 with checksum devices only respond to extended linking commands"},
			{"EnterLinkingModeExt", "Enter Linking Mode (extended command for I2CS devices)", "Enter Linking Mode (i2cs)", 0x09, 0x00, "Insteon version 2 with checksum devices only respond to extended linking commands"},
			{"EnterUnlinkingModeExt", "Enter Unlinking Mode (extended command for I2CS devices)", "Enter Unlinking Mode (i2cs)", 0x0a, 0x00, "Insteon version 2 with checksum devices only respond to extended linking commands"},
			{"SetOperatingFlags", "is used to set a given operating flag", "Set Operating Flags", 0x20, 0x00, ""},
			{"ExtendedGetSet", "is used to get and set extended data (ha ha)", "Extended Get/Set", 0x2e, 0x00, ""},
			{"ReadWriteALDB", "Read/Write ALDB", "Read/Write ALDB", 0x2f, 0x00, ""},
		},
	},
	{
		Name:        "All-Link Messages",
		Byte0:       0x0c,
		Convenience: false,
		Commands: []command{
			{"AllLinkSuccessReport", "is an all-link command to report the number of failed cleanups", "All-link Success Report", 0x06, 0x00, ""},
			{"AllLinkRecall", "is an all-link command to recall the state assigned to the entry in the all-link database", "All-link recall", 0x11, 0x00, ""},
			{"AllLinkAlias2High", "will execute substitute Direct Command", "All-link Alias 2 High", 0x12, 0x00, ""},
			{"AllLinkAlias1Low", "will execute substitute Direct Command", "All-link Alias 1 Low", 0x13, 0x00, ""},
			{"AllLinkAlias2Low", "will execute substitute Direct Command", "All-link Alias 2 Low", 0x14, 0x00, ""},
			{"AllLinkAlias3High", "will execute substitute Direct Command", "All-link Alias 3 High", 0x15, 0x00, ""},
			{"AllLinkAlias3Low", "will execute substitute Direct Command", "All-link Alias 3 Low", 0x16, 0x00, ""},
			{"AllLinkAlias4High", "will execute substitute Direct Command", "All-link Alias 4 High", 0x17, 0x00, ""},
			{"AllLinkAlias4Low", "will execute substitute Direct Command", "All-link Alias 4 Low", 0x18, 0x00, ""},
			{"AllLinkAlias5", "will execute substitute Direct Command", "All-link Alias 5", 0x21, 0x00, ""},
		},
	},
	{
		Name:        "Standard Broadcast Messages",
		Byte0:       0x08,
		Convenience: false,
		Commands: []command{
			{"SetButtonPressedResponder", "Broadcast command indicating the set button has been pressed", "Set-button Pressed (responder)", 0x01, 0x00, ""},
			{"SetButtonPressedController", "Broadcast command indicating the set button has been pressed", "Set-button Pressed (controller)", 0x02, 0x00, ""},
			{"TestPowerlinePhase", "is used for determining which powerline phase (A/B) to which the device is attached", "Test Powerline Phase", 0x03, 0x00, ""},
			{"Heartbeat", "is a broadcast command that is received periodically if it has been set up using the extended get/set command", "Heartbeat", 0x04, 0x00, ""},
			{"BroadCastStatusChange", "is sent by a device when its status changes", "Broadcast Status Change", 0x27, 0x00, ""},
		},
	},
	{
		Name:        "Lighting Standard Direct Messages",
		Byte0:       0x00,
		Convenience: false,
		Commands: []command{
			{"LightOn", "", "Light On", 0x11, 0x00, ""},
			{"LightOnFast", "", "Light On Fast", 0x12, 0x00, ""},
			{"LightOff", "", "Light Off", 0x13, 0x00, ""},
			{"LightOffFast", "", "Light Off Fast", 0x14, 0x00, ""},
			{"LightBrighten", "", "Brighten Light", 0x15, 0x00, ""},
			{"LightDim", "", "Dim Light", 0x16, 0x00, ""},
			{"LightStopManual", "", "Manual Light Change Stop", 0x18, 0x00, ""},
			{"LightStatusRequest", "", "Status Request", 0x19, 0x00, ""},
			{"LightInstantChange", "", "Light Instant Change", 0x21, 0x00, ""},
			{"LightManualOn", "", "Manual On", 0x22, 0x01, ""},
			{"LightManualOff", "", "Manual Off", 0x23, 0x01, ""},
			{"TapSetButtonOnce", "", "Set Button Tap", 0x25, 0x01, ""},
			{"TapSetButtonTwice", "", "Set Button Tap Twice", 0x25, 0x02, ""},
			{"LightSetStatus", "", "Set Status", 0x27, 0x00, ""},
			{"LightOnAtRamp", "", "Light On At Ramp", 0x2e, 0x00, "This command is for dimmers with firmware version less than version 67"},
			{"LightOnAtRampV67", "", "Light On At Ramp", 0x34, 0x00, "Dimmers running firmware version 67 and higher"},
			{"LightOffAtRamp", "", "Light Off At Ramp", 0x2f, 0x00, ""},
			{"LightOffAtRampV67", "", "Light Off At Ramp", 0x35, 0x00, ""},
		},
	},
	{
		Name:        "Dimmer Convenience Commands",
		Byte0:       0x00,
		Convenience: true,
		Commands: []command{
			{"StartBrighten", "", "Manual Start Brighten", 0x17, 0x01, ""},
			{"StartDim", "", "Manual Start Dim", 0x17, 0x00, ""},
			{"EnableProgramLock", "", "Enable Program Lock", 0x20, 0x00, ""},
			{"DisableProgramLock", "", "Disable Program Lock", 0x20, 0x01, ""},
			{"EnableTxLED", "", "Enable Tx LED", 0x20, 0x02, ""},
			{"DisableTxLED", "", "Disable Tx LED", 0x20, 0x03, ""},
			{"EnableResumeDim", "", "Enable Resume Dim", 0x20, 0x04, ""},
			{"DisableResumeDim", "", "Disable Resume Dim", 0x20, 0x05, ""},
			{"EnableLoadSense", "", "Enable Load Sense", 0x20, 0x06, ""},
			{"DisableLoadSense", "", "Disable Load Sense", 0x20, 0x07, ""},
			{"DisableLED", "", "Disable Backlight", 0x20, 0x08, ""},
			{"EnableLED", "", "Enable Backlight", 0x20, 0x09, ""},
			{"SetKeyBeep", "", "Enable Key Beep", 0x20, 0x0a, ""},
			{"ClearKeyBeep", "", "Disable Key Beep", 0x20, 0x0b, ""},
		},
	},
	{
		Name:        "Thermostat Standard Direct Messages",
		Byte0:       0x00,
		Convenience: false,
		Commands: []command{
			{"DecreaseTemp", "Decrease Temperature", "Decrease Temp", 0x68, 0x00, ""},
			{"IncreaseTemp", "Increase Temperature", "Increase Temp", 0x69, 0x00, ""},
			{"GetZoneInfo", "Get Zone Information", "Get Zone Info", 0x6a, 0x00, ""},
			{"GetThermostatMode", "Get Mode", "Get Mode", 0x6b, 0x02, ""},
			{"GetAmbientTemp", "Get Ambient Temperature", "Get Ambient Temp", 0x6b, 0x03, ""},
			{"SetHeat", "Set Heat", "Set Heat", 0x6b, 0x04, ""},
			{"SetCool", "Set Cool", "Set Cool", 0x6b, 0x05, ""},
			{"SetModeAuto", "Set Mode Auto", "Set Auto", 0x6b, 0x06, ""},
			{"SetFan", "Turn Fan On", "Turn Fan On", 0x6b, 0x07, ""},
			{"ClearFan", "Turn Fan Off", "Turn Fan Off", 0x6b, 0x08, ""},
			{"ThermOff", "Turn Thermostat Off", "Turn Thermostat Off", 0x6b, 0x09, ""},
			{"SetProgramHeat", "Set mode to Program Heat", "Set Program Heat", 0x6b, 0x0a, ""},
			{"SetProgramCool", "Set mode to Program Cool", "Set Program Cool", 0x6b, 0x0b, ""},
			{"SetProgramAuto", "Set mode to Program Auto", "Set Program Auto", 0x6b, 0x0c, ""},
			{"GetEquipmentState", "Get Equipment State", "Get State", 0x6b, 0x0d, ""},
			{"SetEquipmentState", "Set Equipment State", "Set State", 0x6b, 0x0e, ""},
			{"GetTempUnits", "Get Temperature Units", "Get Temp Units", 0x6b, 0x0f, ""},
			{"SetFahrenheit", "Set Units to Fahrenheit", "Set Units Fahrenheit", 0x6b, 0x10, ""},
			{"SetCelsius", "Set Units to Celsius", "Set Units Celsius", 0x6b, 0x11, ""},
			{"GetFanOnSpeed", "Get Fan On-Speed", "Get Fan On-Speed", 0x6b, 0x12, ""},
			{"SetFanOnLow", "Set Fan On-Speed to Low", "Set Fan-Speed Low", 0x6b, 0x13, ""},
			{"SetFanOnMed", "Set Fan On-Speed to Med", "Set Fan-Speed Med", 0x6b, 0x14, ""},
			{"SetFanOnHigh", "Set Fan On-Speed to High", "Set Fan-Speed High", 0x6b, 0x15, ""},
			{"EnableStatusMessage", "Enable Status Change Messages", "Enable Status Change", 0x6b, 0x16, ""},
			{"DisableStatusMessage", "Disable Status Change Messages", "Disable Status Change", 0x6b, 0x17, ""},
			{"SetCoolSetpoint", "Set Cool Set-Point", "Set Cool Set-Point", 0x6c, 0x00, ""},
			{"SetHeatSetpoint", "Set Heat Set-Point", "Set Heat Set-Point", 0x6d, 0x00, ""},
		},
	},
	{
		Name:        "Thermostat Extended Direct Messages",
		Byte0:       0x01,
		Convenience: false,
		Commands: []command{
			{"ZoneTempUp", "Increase Zone Temp", "Increase Zone Temp", 0x68, 0x00, ""},
			{"ZoneTempDown", "Decrease Zone Temp", "Decrease Zone Temp", 0x69, 0x00, ""},
			{"SetZoneCoolSetpoint", "Set Zone Cooling Set Point", "Set Zone Cool Set-Point", 0x6c, 0x00, ""},
			{"SetZoneHeatSetpoint", "Set Zone Heating Set Point", "Set Zone Heat Set-Point", 0x6d, 0x00, ""},
		},
	},
}

func init() {
	autogenCommands["commands"] = autogenCommand{
		templates: []autogenTemplate{
			{
				input:  "internal/commands.go.tmpl",
				output: "commands/commands.go",
				data:   func() interface{} { return commands },
			},
			{
				input:  "internal/COMMANDS.md.tmpl",
				output: "COMMANDS.md",
				data:   func() interface{} { return commands },
			},
		},
	}
}
