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

func init() {
	dim := &dimmer{}

	dimCmd := app.SubCommand("dimmer", cli.UsageOption("<device id> <command>"), cli.DescOption("Interact with a specific dimmer"), cli.CallbackOption(dim.init))
	dimCmd.Arguments.Var(&dim.addr, "<device id>")

	dimCmd.SubCommand("config", cli.DescOption("retrieve dimmer configuration information"), cli.CallbackOption(dim.configCmd))
	dimCmd.SubCommand("off", cli.DescOption("turn the dimmer off"), cli.CallbackOption(dim.offCmd))
	dimCmd.SubCommand("brighten", cli.DescOption("brighten the dimmer one step"), cli.CallbackOption(dim.brightenCmd))
	dimCmd.SubCommand("dim", cli.DescOption("dim the dimmer one step"), cli.CallbackOption(dim.dimCmd))
	dimCmd.SubCommand("startBrighten", cli.CallbackOption(dim.startBrightenCmd))
	dimCmd.SubCommand("startDim", cli.CallbackOption(dim.startDimCmd))
	dimCmd.SubCommand("stopChange", cli.CallbackOption(dim.stopChangeCmd))
	dimCmd.SubCommand("status", cli.DescOption("get the switch status"), cli.CallbackOption(dim.statusCmd))

	cmd := dimCmd.SubCommand("on", cli.UsageOption("<level>"), cli.DescOption("turn the dimmer on"), cli.CallbackOption(dim.onCmd))
	cmd.Arguments.Int(&dim.level, "<level>")

	cmd = dimCmd.SubCommand("onfast", cli.UsageOption("<level>"), cli.DescOption("turn the dimmer on fast"), cli.CallbackOption(dim.onFastCmd))
	cmd.Arguments.Int(&dim.level, "<level>")

	cmd = dimCmd.SubCommand("instantChange", cli.UsageOption("<level>"), cli.DescOption("instantly set the dimmer to the desired level (0-255)"), cli.CallbackOption(dim.instantChangeCmd))
	cmd.Arguments.Int(&dim.level, "<level>")

	cmd = dimCmd.SubCommand("setstatus", cli.UsageOption("<level>"), cli.DescOption("set the dimmer switch status LED to <level> (0-31)"), cli.CallbackOption(dim.setStatusCmd))
	cmd.Arguments.Int(&dim.level, "<level>")

	cmd = dimCmd.SubCommand("onramp", cli.UsageOption("<level> <ramp>"), cli.DescOption("turn the dimmer on to the desired level (0-15) at the given ramp rate (0-15)"), cli.CallbackOption(dim.onRampCmd))
	cmd.Arguments.Int(&dim.level, "<level>")
	cmd.Arguments.Int(&dim.ramp, "<ramp>")

	cmd = dimCmd.SubCommand("offramp", cli.UsageOption("<ramp>"), cli.DescOption("turn the dimmer off at the givem ramp rate (0-31)"), cli.CallbackOption(dim.offRampCmd))
	cmd.Arguments.Int(&dim.ramp, "<ramp>")

	cmd = dimCmd.SubCommand("setramp", cli.UsageOption("<ramp>"), cli.DescOption("set default ramp rate (0-31)"), cli.CallbackOption(dim.setRampCmd))
	cmd.Arguments.Int(&dim.ramp, "<ramp>")

	cmd = dimCmd.SubCommand("setlevel", cli.UsageOption("<level>"), cli.DescOption("set default on level (0-255)"), cli.CallbackOption(dim.setOnLevelCmd))
	cmd.Arguments.Int(&dim.level, "<level>")
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
		fmt.Printf("          SNR Failures: %v\n", flags.SNRFailCount())
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

func extractErr(v interface{}, err error) error {
	return err
}

func (dim *dimmer) onCmd(string) error {
	return extractErr(dim.SendCommand(insteon.TurnLightOn(dim.level)))
}
func (dim *dimmer) offCmd(string) error { return extractErr(dim.SendCommand(insteon.TurnLightOff())) }
func (dim *dimmer) onFastCmd(string) error {
	return extractErr(dim.SendCommand(insteon.TurnLightOnFast(dim.level)))
}
func (dim *dimmer) brightenCmd(string) error { return extractErr(dim.SendCommand(insteon.Brighten())) }
func (dim *dimmer) dimCmd(string) error      { return extractErr(dim.SendCommand(insteon.Dim())) }
func (dim *dimmer) startBrightenCmd(string) error {
	return extractErr(dim.SendCommand(insteon.StartBrighten()))
}
func (dim *dimmer) startDimCmd(string) error { return extractErr(dim.SendCommand(insteon.StartDim())) }
func (dim *dimmer) stopChangeCmd(string) error {
	return extractErr(dim.SendCommand(insteon.StopChange()))
}
func (dim *dimmer) instantChangeCmd(string) error {
	return extractErr(dim.SendCommand(insteon.InstantChange(dim.level)))
}
func (dim *dimmer) setStatusCmd(string) error {
	return extractErr(dim.SendCommand(insteon.SetLightStatus(dim.level)))
}
func (dim *dimmer) onRampCmd(string) error {
	return extractErr(dim.SendCommand(insteon.LightOnAtRamp(dim.level, dim.ramp)))
}
func (dim *dimmer) offRampCmd(string) error {
	return extractErr(dim.SendCommand(insteon.LightOffAtRamp(dim.ramp)))
}
func (dim *dimmer) setRampCmd(string) error {
	return extractErr(dim.SendCommand(insteon.SetDefaultRamp(dim.ramp)))
}
func (dim *dimmer) setOnLevelCmd(string) error {
	return extractErr(dim.SendCommand(insteon.SetDefaultOnLevel(dim.level)))
}
