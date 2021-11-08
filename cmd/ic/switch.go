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

type swtch struct {
	*devices.Switch
}

func init() {
	sw := swtch{}

	swCmd := &cli.Command{
		Name:        "switch",
		Usage:       "<device id> <command>",
		Description: "Interact with a specific switch",
		Callback:    cli.Callback(sw.init, "<device id>"),
		SubCommands: []*cli.Command{
			&cli.Command{Name: "config", Description: "retrieve switch configuration information", Callback: cli.Callback(sw.switchConfigCmd)},
			&cli.Command{Name: "status", Description: "get the switch status", Callback: cli.Callback(sw.switchStatusCmd)},
			&cli.Command{Name: "on", Description: "turn light on", Callback: cli.Callback(sw.TurnOn)},
			&cli.Command{Name: "off", Description: "turn light off", Callback: cli.Callback(sw.TurnOff)},
			&cli.Command{Name: "backlight", Description: "turn backlight on/off", Callback: cli.Callback(sw.SetBacklight, "<true|false>")},
			&cli.Command{Name: "loadsense", Description: "turn load sense on/off", Callback: cli.Callback(sw.SetLoadSense, "<true|false>")},
		},
	}
	app.SubCommands = append(app.SubCommands, swCmd)
}

func (sw *swtch) init(addr insteon.Address) error {
	device, err := open(modem, addr)
	if err == nil {
		d := devices.Lookup(device)
		if s, ok := d.(*devices.Switch); ok {
			sw.Switch = s
		} else {
			err = fmt.Errorf("Device at %s is a %T not a switch", addr, device)
		}
	}
	return err
}

func (sw *swtch) switchConfigCmd() error {
	config, err := sw.Config()
	if err == nil {
		err = printDevInfo(sw, fmt.Sprintf("  X10 Address: %02x.%02x", config.HouseCode, config.UnitCode))
	}
	return err
}

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
