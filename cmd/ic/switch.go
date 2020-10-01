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

type swtch struct {
	*insteon.Switch
	addr insteon.Address
	led  bool
}

var switchCommands = []Command{
	IntCmd("on", "turn light on", "<level>", insteon.TurnLightOn),
	Cmd("off", "turn light off", insteon.CmdLightOff),
	BoolCmd("backlight", "turn backlight on/off", "<true|false>", insteon.Backlight),
}

func init() {
	sw := swtch{}

	swCmd := app.SubCommand("switch", cli.UsageOption("<device id> <command>"), cli.DescOption("Interact with a specific switch"), cli.CallbackOption(sw.init))
	swCmd.Arguments.Var(&sw.addr, "<device id>")

	swCmd.SubCommand("config", cli.DescOption("retrieve switch configuration information"), cli.CallbackOption(sw.switchConfigCmd))
	swCmd.SubCommand("status", cli.DescOption("get the switch status"), cli.CallbackOption(sw.switchStatusCmd))

	for _, cmd := range switchCommands {
		cb := func(cmd Command) func(string) error {
			return func(string) error { return sw.runCmd(cmd) }
		}(cmd)
		c := swCmd.SubCommand(cmd.Name(), cli.DescOption(cmd.Desc()), cli.UsageOption(cmd.Usage()), cli.CallbackOption(cb))
		cmd.Setup(&c.Arguments)
	}
}

func (sw *swtch) init(string) error {
	device, err := connect(modem, sw.addr)
	if err == nil {
		if s, ok := device.(*insteon.Switch); ok {
			sw.Switch = s
		} else {
			err = fmt.Errorf("Device at %s is a %T not a switch", sw.addr, device)
		}
	}
	return err
}

func (sw *swtch) runCmd(cmd Command) error {
	_, err := sw.SendCommand(cmd.Command())
	return err
}

func (sw *swtch) switchConfigCmd(string) error {
	config, err := sw.Config()
	if err == nil {
		err = printDevInfo(sw, fmt.Sprintf("  X10 Address: %02x.%02x", config.HouseCode, config.UnitCode))
	}
	return err
}

func (sw *swtch) switchStatusCmd(string) error {
	level, err := sw.Status()
	if err == nil {
		if level == 0 {
			fmt.Printf("Switch is off\n")
		} else {
			fmt.Printf("Switch is on\n")
		}
	}
	return err
}
