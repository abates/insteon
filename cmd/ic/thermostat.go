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

type thermostat struct {
	*insteon.Thermostat
	addr insteon.Address
}

func init() {
	therm := thermostat{
		Thermostat: &insteon.Thermostat{},
	}

	thermCmd := app.SubCommand("thermostat", cli.UsageOption("<device id> <command>"), cli.DescOption("Interact with a thermostat"), cli.CallbackOption(therm.init))
	thermCmd.Arguments.Var(&therm.addr, "<device id>")

	thermCmd.SubCommand("status", cli.DescOption("get the thermostat status"), cli.CallbackOption(therm.thermStatusCmd))
	thermCmd.SubCommand("setStatus", cli.DescOption("enable/disable status reporting"), cli.UsageOption("<true|false>"), cli.ArgCallbackOption(therm.SetStatusMessage))
}

func (therm *thermostat) init(string) error {
	device, err := connect(modem, therm.addr)
	if err == nil {
		if t, ok := device.(*insteon.Thermostat); ok {
			*therm.Thermostat = *t
		} else {
			err = fmt.Errorf("Device at %s is a %T not a thermostat", therm.addr, device)
		}
	}
	return err
}

func (therm *thermostat) thermStatusCmd(string) (err error) {
	err = therm.SetTempUnit(insteon.Fahrenheit)
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
