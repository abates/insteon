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

type thermostat struct {
	*devices.Thermostat
}

func init() {
	therm := thermostat{
		Thermostat: &devices.Thermostat{},
	}

	thermCmd := &cli.Command{
		Name:        "thermostat",
		Usage:       "<device id> <command>",
		Description: "Interact with a thermostat",
		Callback:    cli.Callback(therm.init, "<device id>"),
		SubCommands: []*cli.Command{
			&cli.Command{Name: "status", Description: "get the thermostat status", Callback: cli.Callback(therm.thermStatusCmd)},
			&cli.Command{Name: "setStatus", Description: "enable/disable status reporting", Callback: cli.Callback(therm.SetStatusMessage, "<true|false>")},
		},
	}
	app.SubCommands = append(app.SubCommands, thermCmd)
}

func (therm *thermostat) init(addr insteon.Address) error {
	device, err := open(modem, addr)
	if err == nil {
		d := devices.Upgrade(device)
		if t, ok := d.(*devices.Thermostat); ok {
			*therm.Thermostat = *t
		} else {
			err = fmt.Errorf("Device at %s is a %T not a thermostat", addr, device)
		}
	}
	return err
}

func (therm *thermostat) thermStatusCmd() (err error) {
	err = therm.SetTempUnit(devices.Fahrenheit)
	//err = therm.SetTempUnit(insteon.Celsius)
	if err != nil {
		return err
	}

	status, err := therm.Status()
	if err == nil {
		fmt.Printf("")
		fmt.Printf("Temperature: %d %s\n", status.Temperature, status.Unit)
		fmt.Printf("   Humidity: %d%%\n", status.Humidity)
		fmt.Printf("   Setpoint: %d %s\n", status.Setpoint, status.Unit)
		fmt.Printf("   Deadband: %d\n", status.Deadband)
		fmt.Printf("       Mode: %s\n", status.Mode)
		fmt.Printf("      State: %s\n", status.State)
	}
	return err
}
