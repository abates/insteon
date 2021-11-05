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
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/abates/cli"
	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
	"github.com/abates/insteon/util"
	"github.com/kirsle/configdir"
	"github.com/tarm/serial"
)

var (
	modem *plm.PLM = &plm.PLM{}
	db    util.Database

	logLevelFlag   insteon.LogLevel
	serialPortFlag string
	timeoutFlag    time.Duration
	writeDelayFlag time.Duration
	ttlFlag        int

	app       = cli.New(os.Args[0], cli.CallbackOption(cli.Callback(run)))
	configDir string
	dbfile    string
)

func init() {
	app.SetOutput(os.Stderr)
	app.Flags.StringVar(&serialPortFlag, "port", "/dev/ttyUSB0", "serial port connected to a PLM")
	app.Flags.Var(&logLevelFlag, "log", "Log Level {none|info|debug|trace}")
	app.Flags.DurationVar(&timeoutFlag, "timeout", 3*time.Second, "read/write timeout duration")
	app.Flags.DurationVar(&writeDelayFlag, "writeDelay", 0, "writeDelay duration (default of 0 indicates to compute wait time based on message length and ttl)")
	app.Flags.IntVar(&ttlFlag, "ttl", 3, "default ttl for sending Insteon messages")

	configDir = configdir.LocalConfig("go-insteon")
	dbfile = filepath.Join(configDir, "db.json")
}

func run() error {
	if logLevelFlag > insteon.LevelNone {
		insteon.SetLogLevel(logLevelFlag, os.Stderr)
	}

	err := configdir.MakePath(configDir)
	if err != nil {
		return err
	}

	c := &serial.Config{
		Name: serialPortFlag,
		Baud: 19200,
	}

	s, err := serial.OpenPort(c)

	if err != nil {
		return fmt.Errorf("error opening serial port: %v", err)
	}

	db = util.NewMemDB()
	err = util.LoadDB(dbfile, db)
	if err != nil {
		log.Fatalf("Failed to load database: %v", err)
	}

	*modem = *plm.New(s, plm.Timeout(timeoutFlag), plm.WriteDelay(writeDelayFlag))
	return nil
}

func open(modem *plm.PLM, addr insteon.Address) (*insteon.BasicDevice, error) {
	device, err := util.Open(insteon.TTL(ttlFlag).Filter(modem), addr, db, dbfile)

	if err == insteon.ErrNotLinked {
		msg := fmt.Sprintf("Device %s is not linked to the PLM.  Link now? (y/n) ", addr)
		if cli.Query(os.Stdin, os.Stdout, msg, "y", "n") == "y" {
			pc := &plmCmd{group: 0x01, addresses: []insteon.Address{addr}}
			err = pc.link(true, false)()
		}
	}
	return device, err
}

func main() {
	_, err := app.Run(os.Args[1:])
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
