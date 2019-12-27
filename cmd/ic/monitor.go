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
	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
)

func init() {
	app.SubCommand("monitor", cli.DescOption("Monitor the Insteon network"), cli.CallbackOption(monCmd))
}

func monCmd() (err error) {
	if monitor, ok := modem.(plm.Monitorable); ok {
		log.Printf("Starting monitor...")
		var ch <-chan *insteon.Message
		ch, err = monitor.Monitor()
		if err == nil {
			for msg := range ch {
				log.Printf("%s", msg)
			}
		}
	} else {
		log.Printf("PLM is not monitorable")
	}
	return err
}
