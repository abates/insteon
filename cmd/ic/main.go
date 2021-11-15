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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/abates/cli"
	"github.com/abates/insteon"
	"github.com/abates/insteon/devices"
	"github.com/abates/insteon/plm"
	"github.com/abates/insteon/util"
	"github.com/kirsle/configdir"
	"github.com/tarm/serial"
)

var (
	modem *plm.PLM = &plm.PLM{}
	db    util.Database

	serialPortFlag string
	timeoutFlag    time.Duration
	writeDelayFlag time.Duration
	ttlFlag        int
	logFlag        bool
	debugFlag      bool
	quietFlag      bool

	app       = cli.New(os.Args[0], cli.CallbackOption(cli.Callback(run)))
	configDir string
	dbfile    string
)

func init() {
	app.SetOutput(os.Stderr)
	app.Flags.StringVar(&serialPortFlag, "port", "/dev/ttyUSB0", "serial port connected to a PLM")
	app.Flags.BoolVar(&logFlag, "log", false, "Log insteon traffic")
	app.Flags.BoolVar(&debugFlag, "debug", false, "Set debug logging")
	app.Flags.BoolVar(&debugFlag, "quietFlag", false, "Log nothing")
	app.Flags.DurationVar(&timeoutFlag, "timeout", 3*time.Second, "read/write timeout duration")
	app.Flags.DurationVar(&writeDelayFlag, "writeDelay", 0, "writeDelay duration (default of 0 indicates to compute wait time based on message length and ttl)")
	app.Flags.IntVar(&ttlFlag, "ttl", 3, "default ttl for sending Insteon messages")

	configDir = configdir.LocalConfig("go-insteon")
	dbfile = filepath.Join(configDir, "db.json")
}

func run() error {
	if debugFlag {
		plm.LogDebug.SetOutput(os.Stderr)
		devices.LogDebug.SetOutput(os.Stderr)
	}

	if quietFlag {
		devices.Log.SetOutput(ioutil.Discard)
		devices.LogDebug.SetOutput(ioutil.Discard)
		plm.Log.SetOutput(ioutil.Discard)
		plm.LogDebug.SetOutput(ioutil.Discard)
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

	db, err = util.NewFileDB(dbfile)
	if err != nil {
		log.Fatalf("Failed to load database: %v", err)
	}

	*modem = *plm.New(s, plm.Timeout(timeoutFlag), plm.WriteDelay(writeDelayFlag))
	return nil
}

func open(modem *plm.PLM, addr insteon.Address, askLink bool) (*devices.BasicDevice, error) {
	filters := []devices.Filter{}
	if logFlag {
		filters = append(filters, util.Snoop(os.Stdout, db))
	}
	filters = append(filters, devices.TTL(ttlFlag), devices.RetryFilter(3))
	device, err := db.Open(modem, addr, filters...)

	if err == devices.ErrNotLinked && askLink {
		msg := fmt.Sprintf("Device %s is not linked to the PLM.  Link now? (y/n) ", addr)
		if cli.Query(os.Stdin, os.Stdout, msg, "y", "n") == "y" {
			pc := &plmCmd{}
			err = pc.link(true, false)(insteon.Group(1), util.Addresses{addr})
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
