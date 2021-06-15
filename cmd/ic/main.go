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
	"github.com/abates/insteon/db"
	"github.com/abates/insteon/plm"
	"github.com/kirsle/configdir"
	"github.com/tarm/serial"
)

var (
	modem          *plm.PLM
	logLevelFlag   insteon.LogLevel
	serialPortFlag string
	timeoutFlag    time.Duration
	writeDelayFlag time.Duration
	ttlFlag        uint
	app            = cli.New(os.Args[0], cli.CallbackOption(run))
)

func init() {
	app.SetOutput(os.Stderr)
	app.Flags.StringVar(&serialPortFlag, "port", "/dev/ttyUSB0", "serial port connected to a PLM")
	app.Flags.Var(&logLevelFlag, "log", "Log Level {none|info|debug|trace}")
	app.Flags.DurationVar(&timeoutFlag, "timeout", 3*time.Second, "read/write timeout duration")
	app.Flags.DurationVar(&writeDelayFlag, "writeDelay", 0, "writeDelay duration (default of 0 indicates to compute wait time based on message length and ttl)")
	app.Flags.UintVar(&ttlFlag, "ttl", 3, "default ttl for sending Insteon messages")
}

func run(string) error {
	if logLevelFlag > insteon.LevelNone {
		insteon.Log.Level = logLevelFlag
	}

	dir := configdir.LocalConfig("go-insteon")
	dbfile := filepath.Join(dir, "db.json")
	err := configdir.MakePath(dir)
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

	database := db.NewMemDB()
	err = db.Load(dbfile, database)
	if err != nil {
		log.Fatalf("Failed to load database: %v", err)
	}

	modem, err = plm.New(s, s, timeoutFlag, plm.Database(database), plm.WriteDelay(writeDelayFlag), plm.ConnectionOptions(insteon.ConnectionTimeout(timeoutFlag), insteon.ConnectionTTL(uint8(ttlFlag))))

	if err != nil {
		return fmt.Errorf("error opening plm: %v", err)
	}

	return nil
}

func main() {
	app.Parse(os.Args[1:])
	err := app.Run()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
