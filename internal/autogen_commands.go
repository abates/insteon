package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"text/template"
	"time"
)

type Command struct {
	Name    string
	Comment string
	String  string
	Byte1   string
	Byte2   string
}

type CommandGroup struct {
	Name     string
	Byte0    string
	Commands []Command
}

var commands = []CommandGroup{
	{
		Name:  "Standard Direct Commands",
		Byte0: "0x00",
		Commands: []Command{
			{"CmdAssignToAllLinkGroup", "Assign to ALL-Link Group", "Assign to All-Link Group", "0x01", "0x00"},
			{"CmdDeleteFromAllLinkGroup", "Delete from All-Link Group", "Delete from All-Link Group", "0x02", "0x00"},
			{"CmdProductDataReq", "Product Data Request", "Product Data Request", "0x03", "0x00"},
			{"CmdFxUsernameReq", "FX Username Request", "Fx Username Request", "0x03", "0x01"},
			{"CmdDeviceTextStringReq", "Device Text String Request", "Text String Request", "0x03", "0x02"},
			{"CmdExitLinkingMode", "Exit Linking Mode", "Exit Linking Mode", "0x08", "0x00"},
			{"CmdEnterLinkingMode", "Enter Linking Mode", "Enter Linking Mode", "0x09", "0x00"},
			{"CmdEnterUnlinkingMode", "Enter Unlinking Mode", "Enter Unlinking Mode", "0x0a", "0x00"},
			{"CmdGetEngineVersion", "Get Insteon Engine Version", "Engine Version", "0x0d", "0x00"},
			{"CmdPing", "Ping Request", "Ping Request", "0x0f", "0x00"},
			{"CmdIDRequest", "Send ID Request which will prompt the device to respond with a Set Button Pressed Controller/Responder", "ID Request", "0x10", "0x00"},
			{"CmdGetOperatingFlags", "is used to request a given operating flag", "Get Operating Flags", "0x1f", "0x00"},
			{"CmdSetOperatingFlags", "is used to set a given operating flag", "Set Operating Flags", "0x20", "0x00"},
		},
	},
	{
		Name:  "Extended Direct Commands",
		Byte0: "0x01",
		Commands: []Command{
			{"CmdProductDataResp", "Product Data Response", "Product Data Response", "0x03", "0x00"},
			{"CmdFxUsernameResp", "FX Username Response", "Fx Username Response", "0x03", "0x01"},
			{"CmdDeviceTextStringResp", "Device Text String Response", "Text String Response", "0x03", "0x02"},
			{"CmdSetDeviceTextString", "sets the device text string", "Set Text String", "0x03", "0x03"},
			{"CmdSetAllLinkCommandAlias", "sets the command alias to use for all-linking", "Set All-Link Command Alias", "0x03", "0x04"},
			{"CmdSetAllLinkCommandAliasData", "sets the extended data to be used if the command alias is an extended command", "Set All-Link Command Alias Data", "0x03", "0x05"},
			{"CmdExitLinkingModeExt", "Exit Linking Mode", "Exit Linking Mode (i2cs)", "0x08", "0x00"},
			{"CmdEnterLinkingModeExt", "Enter Linking Mode (extended command for I2CS devices)", "Enter Linking Mode (i2cs)", "0x09", "0x00"},
			{"CmdEnterUnlinkingModeExt", "Enter Unlinking Mode (extended command for I2CS devices)", "Enter Unlinking Mode (i2cs)", "0x0a", "0x00"},
			{"CmdExtendedGetSet", "is used to get and set extended data (ha ha)", "Extended Get/Set", "0x2e", "0x00"},
			{"CmdReadWriteALDB", "Read/Write ALDB", "Read/Write ALDB", "0x2f", "0x00"},
		},
	},
	{
		Name:  "All-Link Messages",
		Byte0: "0x0c",
		Commands: []Command{
			{"CmdAllLinkRecall", "is an all-link command to recall the state assigned to the entry in the all-link database", "All-link recall", "0x11", "0x00"},
			{"CmdAllLinkAlias2High", "will execute substitute Direct Command", "All-link Alias 2 High", "0x12", "0x00"},
			{"CmdAllLinkAlias1Low", "will execute substitute Direct Command", "All-link Alias 1 Low", "0x13", "0x00"},
			{"CmdAllLinkAlias2Low", "will execute substitute Direct Command", "All-link Alias 2 Low", "0x14", "0x00"},
			{"CmdAllLinkAlias3High", "will execute substitute Direct Command", "All-link Alias 3 High", "0x15", "0x00"},
			{"CmdAllLinkAlias3Low", "will execute substitute Direct Command", "All-link Alias 3 Low", "0x16", "0x00"},
			{"CmdAllLinkAlias4High", "will execute substitute Direct Command", "All-link Alias 4 High", "0x17", "0x00"},
			{"CmdAllLinkAlias4Low", "will execute substitute Direct Command", "All-link Alias 4 Low", "0x18", "0x00"},
			{"CmdAllLinkAlias5", "will execute substitute Direct Command", "All-link Alias 5", "0x21", "0x00"},
		},
	},
	{
		Name:  "Standard Broadcast Messages",
		Byte0: "0x08",
		Commands: []Command{
			{"CmdSetButtonPressedResponder", "Broadcast command indicating the set button has been pressed", "Set-button Pressed (responder)", "0x01", "0x00"},
			{"CmdSetButtonPressedController", "Broadcast command indicating the set button has been pressed", "Set-button Pressed (controller)", "0x02", "0x00"},
			{"CmdTestPowerlinePhase", "is used for determining which powerline phase (A/B) to which the device is attached", "Test Powerline Phase", "0x03", "0x00"},
			{"CmdHeartbeat", "is a broadcast command that is received periodically if it has been set up using the extended get/set command", "Heartbeat", "0x04", "0x00"},
			{"CmdBroadCastStatusChange", "is sent by a device when its status changes", "Broadcast Status Change", "0x27", "0x00"},
		},
	},
	{
		Name:  "Lighting Standard Direct Messages",
		Byte0: "0x00",
		Commands: []Command{
			{"CmdLightOn", "", "Light On", "0x11", "0xff"},
			{"CmdLightOnFast", "", "Light On Fast", "0x12", "0x00"},
			{"CmdLightOff", "", "Light Off", "0x13", "0x00"},
			{"CmdLightOffFast", "", "Light Off Fast", "0x14", "0x00"},
			{"CmdLightBrighten", "", "Brighten Light", "0x15", "0x00"},
			{"CmdLightDim", "", "Dim Light", "0x16", "0x00"},
			{"CmdLightStartManual", "", "Manual Light Change Start", "0x17", "0x00"},
			{"CmdLightStopManual", "", "Manual Light Change Stop", "0x18", "0x00"},
			{"CmdLightStatusRequest", "", "Status Request", "0x19", "0x00"},
			{"CmdLightInstantChange", "", "Light Instant Change", "0x21", "0x00"},
			{"CmdLightManualOn", "", "Manual On", "0x22", "0x01"},
			{"CmdLightManualOff", "", "Manual Off", "0x23", "0x01"},
			{"CmdTapSetButtonOnce", "", "Set Button Tap", "0x25", "0x01"},
			{"CmdTapSetButtonTwice", "", "Set Button Tap Twice", "0x25", "0x02"},
			{"CmdLightSetStatus", "", "Set Status", "0x27", "0x00"},
			{"CmdLightOnAtRamp", "", "Light On At Ramp", "0x2e", "0x00"},
			{"CmdLightOnAtRampV67", "", "Light On At Ramp", "0x34", "0x00"},
			{"CmdLightOffAtRamp", "", "Light Off At Ramp", "0x2f", "0x00"},
			{"CmdLightOffAtRampV67", "", "Light Off At Ramp", "0x35", "0x00"},
		},
	},
}

const cmdsTemplate = `
{{$byte0 := .Byte0}}// {{.Name}}
var({{range .Commands}}
// {{ .Name }} {{ .Comment }}
{{.Name}} = Command{ {{$byte0}}, {{.Byte1}}, {{.Byte2}} } // {{.String}}
{{end}})
`

const strTemplate = `
	var cmdStrings = map[Command]string { {{range .}}
		{{.Name}}: "{{.String}}",{{end}}
	}
`

var gopath string
var unitPkgPath string

func owner() string { return "Andrew Bates" }

func main() {
	cmdStrings := []Command{}
	licenseText, _ := ioutil.ReadFile("internal/license.tmpl")
	funcs := template.FuncMap{"now": time.Now, "owner": owner}
	license := template.Must(template.New("license").Funcs(funcs).Parse(string(licenseText)))
	form := template.Must(template.New("format").Funcs(funcs).Parse(cmdsTemplate))
	str := template.Must(template.New("format").Funcs(funcs).Parse(strTemplate))

	f, err := os.Create("commands.go")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	buf := bytes.NewBuffer(make([]byte, 0))
	err = license.Execute(buf, struct{}{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(buf, "\npackage insteon\n")

	for _, group := range commands {
		form.Execute(buf, group)
		if err != nil {
			log.Fatal(err)
		}
		cmdStrings = append(cmdStrings, group.Commands...)
	}

	str.Execute(buf, cmdStrings)

	b, err := format.Source(buf.Bytes())
	if err != nil {
		f.Write(buf.Bytes()) // This is here to debug bad format
		log.Fatalf("error formatting: %s", err)
	}

	f.Write(b)
}
