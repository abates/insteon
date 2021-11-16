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

import (
	"fmt"

	"github.com/abates/cli"
	"github.com/abates/insteon"
	"github.com/abates/insteon/devices"
)

type dimmer struct {
	*devices.Dimmer
}

func init() {
	dim := &dimmer{
		Dimmer: &devices.Dimmer{Switch: &devices.Switch{}},
	}

	dimCmd := &cli.Command{
		Name:        "dimmer",
		Usage:       "<device id> <command>",
		Description: "Interact with a specific dimmer",
		Callback:    cli.Callback(dim.init, "<device id>"),
		SubCommands: []*cli.Command{
			{Name: "config", Description: "retrieve dimmer configuration information", Callback: cli.Callback(dim.configCmd)},
			{Name: "status", Description: "get the dimmer status", Callback: cli.Callback(dim.statusCmd)},
			{Name: "on", Description: "turn light on", Callback: cli.Callback(dim.TurnOn, "<level>")},
			{Name: "off", Description: "turn light off", Callback: cli.Callback(dim.TurnOff)},
			{Name: "backlight", Description: "turn backlight on/off", Callback: cli.Callback(dim.SetBacklight, "<true|false>")},
			{Name: "loadsense", Description: "turn load sense on/off", Callback: cli.Callback(dim.SetLoadSense, "<true|false>")},
			{Name: "brighten", Description: "brighten light one step", Callback: cli.Callback(dim.Brighten)},
			{Name: "dim", Description: "dim light one step", Callback: cli.Callback(dim.Dim)},
			{Name: "startBrighten", Description: "start brightening the light", Callback: cli.Callback(dim.StartBrighten)},
			{Name: "startDim", Description: "start dimming the light", Callback: cli.Callback(dim.StartDim)},
			{Name: "stopChange", Description: "stop active change (brighten/dim)", Callback: cli.Callback(dim.StopManualChange)},
			{Name: "onfast", Description: "turn light on fast", Callback: cli.Callback(dim.OnFast)},
			{Name: "instantChange", Description: "light instant change", Callback: cli.Callback(dim.InstantChange)},
			{Name: "setstatus", Description: "set light's status indicator", Callback: cli.Callback(dim.SetStatus)},
			{Name: "onramp", Description: "turn light on to <level> at <rate>", Callback: cli.Callback(dim.OnAtRamp)},
			{Name: "offramp", Description: "turn light off at <rate>", Callback: cli.Callback(dim.OffAtRamp)},
		},
	}

	app.SubCommands = append(app.SubCommands, dimCmd)
}

func (dim *dimmer) init(addr insteon.Address) (err error) {
	device, err := open(modem, addr, true)
	if err == nil {
		d := devices.Lookup(device)
		if d, ok := d.(*devices.Dimmer); ok {
			*dim.Dimmer.Switch = *d.Switch
		} else {
			err = fmt.Errorf("Device %s is not a dimmer", addr)
		}
	}
	return err
}

func (dim *dimmer) configCmd() error {
	config, err := dim.Config()
	if err == nil {
		fmt.Printf("           X10 Address: %02x.%02x\n", config.HouseCode, config.UnitCode)
		fmt.Printf("          Default Ramp: %d\n", config.Ramp)
		fmt.Printf("         Default Level: %d\n", config.OnLevel)
		fmt.Printf("                   SNR: %d\n", config.SNT)
	}

	flags, err := dim.OperatingFlags()
	if err == nil {
		fmt.Printf("          Program Lock: %v\n", flags.ProgramLock())
		fmt.Printf("             LED on Tx: %v\n", flags.TxLED())
		fmt.Printf("            Resume Dim: %v\n", flags.ResumeDim())
		fmt.Printf("                LED On: %v\n", flags.LED())
		fmt.Printf("         Load Sense On: %v\n", flags.LoadSense())
		fmt.Printf("              DB Delta: %v\n", flags.DBDelta())
		fmt.Printf("                   SNR: %v\n", flags.SNR())
		fmt.Printf("           X10 Enabled: %v\n", flags.X10Enabled())
		fmt.Printf("        Blink on Error: %v\n", flags.ErrorBlink())
		fmt.Printf("Cleanup Report Enabled: %v\n", flags.CleanupReport())
	}
	return err
}

func (dim *dimmer) statusCmd() error {
	level, err := dim.Status()
	if err == nil {
		if level == 0 {
			fmt.Printf("Dimer is off\n")
		} else if level == 255 {
			fmt.Printf("Dimer is on\n")
		} else {
			fmt.Printf("Dimer is on at level %d\n", level)
		}
	}
	return err
}
