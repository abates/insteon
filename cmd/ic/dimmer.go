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
)

type dimmer struct {
	*insteon.Dimmer
	addr insteon.Address

	level int
	ramp  int
}

var dimmerCommands = []Command{
	Cmd("brighten", "brighten light one step", insteon.CmdLightBrighten),
	Cmd("dim", "dim light one step", insteon.CmdLightDim),
	Cmd("startBrighten", "start brightening the light", insteon.CmdStartBrighten),
	Cmd("startDim", "start dimming the light", insteon.CmdStartDim),
	Cmd("stopChange", "stop active change (brighten/dim)", insteon.CmdLightStopManual),
	Cmd("onfast", "turn light on fast", insteon.CmdLightOnFast),
	IntCmd("instantChange", "light instant change", "<level>", insteon.InstantChange),
	IntCmd("setstatus", "set light's status indicator", "<level>", insteon.SetLightStatus),
	TwintCmd("onramp", "turn light on to <level> at <rate>", "<level> <rate>", insteon.LightOnAtRamp),
	IntCmd("offramp", "turn light off at <rate>", "<rate>", insteon.LightOffAtRamp),
}

func init() {
	dim := &dimmer{}

	dimCmd := app.SubCommand("dimmer", cli.UsageOption("<device id> <command>"), cli.DescOption("Interact with a specific dimmer"), cli.CallbackOption(dim.init))
	dimCmd.Arguments.Var(&dim.addr, "<device id>")

	dimCmd.SubCommand("config", cli.DescOption("retrieve dimmer configuration information"), cli.CallbackOption(dim.configCmd))
	dimCmd.SubCommand("status", cli.DescOption("get the switch status"), cli.CallbackOption(dim.statusCmd))

	for _, commands := range [][]Command{switchCommands, dimmerCommands} {
		for _, cmd := range commands {
			cb := func(cmd Command) func(string) error {
				return func(string) error { return dim.runCmd(cmd) }
			}(cmd)
			c := dimCmd.SubCommand(cmd.Name(), cli.DescOption(cmd.Desc()), cli.UsageOption(cmd.Usage()), cli.CallbackOption(cb))
			cmd.Setup(&c.Arguments)
		}
	}
}

func (dim *dimmer) init(string) (err error) {
	device, err := connect(modem, dim.addr)
	if err == nil {
		if d, ok := device.(*insteon.Dimmer); ok {
			dim.Dimmer = d
		} else {
			err = fmt.Errorf("Device %s is not a dimmer", dim.addr)
		}
	}
	return err
}

func (dim *dimmer) runCmd(cmd Command) (err error) {
	_, err = dim.SendCommand(cmd.Command())

	return err
}

func (dim *dimmer) configCmd(string) error {
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

func (dim *dimmer) statusCmd(string) error {
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
