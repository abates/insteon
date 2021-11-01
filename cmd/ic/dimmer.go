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
}

func init() {
	dim := &dimmer{
		Dimmer: &insteon.Dimmer{Switch: &insteon.Switch{}},
	}

	dimCmd := app.SubCommand("dimmer", cli.UsageOption("<device id> <command>"), cli.DescOption("Interact with a specific dimmer"), cli.CallbackOption(dim.init))
	dimCmd.Arguments.Var(&dim.addr, "<device id>")

	dimCmd.SubCommand("config", cli.DescOption("retrieve dimmer configuration information"), cli.CallbackOption(dim.configCmd))
	dimCmd.SubCommand("status", cli.DescOption("get the dimmer status"), cli.CallbackOption(dim.statusCmd))

	dimCmd.SubCommand("on", cli.DescOption("turn light on"), cli.UsageOption("<level>"), cli.ArgCallbackOption(dim.TurnOn))
	dimCmd.SubCommand("off", cli.DescOption("turn light off"), cli.ArgCallbackOption(dim.TurnOff))
	dimCmd.SubCommand("backlight", cli.DescOption("turn backlight on/off"), cli.UsageOption("<true|false>"), cli.ArgCallbackOption(dim.SetBacklight))
	dimCmd.SubCommand("loadsense", cli.DescOption("turn load sense on/off"), cli.UsageOption("<true|false>"), cli.ArgCallbackOption(dim.SetLoadSense))

	dimCmd.SubCommand("brighten", cli.DescOption("brighten light one step"), cli.ArgCallbackOption(dim.Brighten))
	dimCmd.SubCommand("dim", cli.DescOption("dim light one step"), cli.ArgCallbackOption(dim.Dim))
	dimCmd.SubCommand("startBrighten", cli.DescOption("start brightening the light"), cli.ArgCallbackOption(dim.StartBrighten))
	dimCmd.SubCommand("startDim", cli.DescOption("start dimming the light"), cli.ArgCallbackOption(dim.StartDim))
	dimCmd.SubCommand("stopChange", cli.DescOption("stop active change (brighten/dim)"), cli.ArgCallbackOption(dim.StopManualChange))
	dimCmd.SubCommand("onfast", cli.DescOption("turn light on fast"), cli.ArgCallbackOption(dim.OnFast))
	dimCmd.SubCommand("instantChange", cli.DescOption("light instant change"), cli.UsageOption("<level>"), cli.ArgCallbackOption(dim.InstantChange))
	dimCmd.SubCommand("setstatus", cli.DescOption("set light's status indicator"), cli.UsageOption("<level>"), cli.ArgCallbackOption(dim.SetStatus))
	dimCmd.SubCommand("onramp", cli.DescOption("turn light on to <level> at <rate>"), cli.UsageOption("<level> <rate>"), cli.ArgCallbackOption(dim.OnAtRamp))
	dimCmd.SubCommand("offramp", cli.DescOption("turn light off at <rate>"), cli.UsageOption("<rate>"), cli.ArgCallbackOption(dim.OffAtRamp))
}

func (dim *dimmer) init(string) (err error) {
	device, err := open(modem, dim.addr)
	if err == nil {
		if d, ok := device.(*insteon.Dimmer); ok {
			*dim.Dimmer.Switch = *d.Switch
		} else {
			err = fmt.Errorf("Device %s is not a dimmer", dim.addr)
		}
	}
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
