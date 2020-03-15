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
)

func init() {
	app.SubCommand("monitor", cli.DescOption("Monitor the Insteon network"), cli.CallbackOption(monCmd))
}

func monCmd(string) (err error) {
	log.Printf("Starting monitor...")
	err = modem.Monitor(func(conn insteon.Connection) {
		var msg *insteon.Message
		for err == nil {
			msg, err = conn.Receive()
			if err == insteon.ErrReadTimeout {
				err = nil
			} else if err == nil {
				log.Printf("%s", msg)
			}
		}
	})
	return err
}
