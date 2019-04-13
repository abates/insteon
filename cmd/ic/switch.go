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
	insteon.Switch
	addr insteon.Address
	led  bool
}

func init() {
	sw := swtch{}

	swCmd := app.SubCommand("switch", cli.UsageOption("<device id> <command>"), cli.DescOption("Interact with a specific switch"), cli.CallbackOption(sw.init))
	swCmd.Arguments.Var(&sw.addr, "<device id>")
	swCmd.SubCommand("config", cli.DescOption("retrieve switch configuration information"), cli.CallbackOption(sw.switchConfigCmd))
	swCmd.SubCommand("on", cli.DescOption("turn the switch/light on"), cli.CallbackOption(sw.switchOnCmd))
	swCmd.SubCommand("off", cli.DescOption("turn the switch/light off"), cli.CallbackOption(sw.switchOffCmd))
	swCmd.SubCommand("status", cli.DescOption("get the switch status"), cli.CallbackOption(sw.switchStatusCmd))
	cmd := swCmd.SubCommand("setled", cli.DescOption("set operating flags"), cli.UsageOption("<true|false>"), cli.CallbackOption(sw.switchSetLedCmd))
	cmd.Arguments.Bool(&sw.led, "<true|false>")
}

func (sw *swtch) init() error {
	device, err := connect(modem, sw.addr)
	if err == nil {
		if s, ok := device.(insteon.Switch); ok {
			sw.Switch = s
		} else {
			err = fmt.Errorf("Device at %s is a %T not a switch", sw.addr, device)
		}
	}
	return err
}

func (sw *swtch) switchConfigCmd() error {
	config, err := sw.SwitchConfig()
	if err == nil {
		err = printDevInfo(sw, fmt.Sprintf("  X10 Address: %02x.%02x", config.HouseCode, config.UnitCode))
	}
	return err
}

func (sw *swtch) switchOnCmd() error     { return sw.On() }
func (sw *swtch) switchOffCmd() error    { return sw.Off() }
func (sw *swtch) switchSetCmd() error    { return nil }
func (sw *swtch) switchSetLedCmd() error { return sw.SetLED(sw.led) }

func (sw *swtch) switchStatusCmd() error {
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
