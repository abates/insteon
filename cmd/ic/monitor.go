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
	"log"

	"github.com/abates/cli"
)

func init() {
	app.SubCommand("monitor", cli.DescOption("Monitor the Insteon network"), cli.CallbackOption(monCmd))
}

func monCmd(string) (err error) {
	log.Printf("Starting monitor...")

	config, err := modem.Config()
	if err != nil {
		return err
	}

	config.SetMonitorMode()
	err = modem.SetConfig(config)
	if err != nil {
		return err
	}

	/*ch := modem.Subscribe(insteon.Wildcard, insteon.Matches(func(*insteon.Message) bool { return true }))
	// TODO: Catch signals here for cleanup
	for msg := range ch {
		log.Printf("%s", msg)
	}
	config.ClearMonitorMode()
	err = modem.SetConfig(config)*/
	return err
}
