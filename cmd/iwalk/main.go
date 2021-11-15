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
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/abates/insteon"
	"github.com/abates/insteon/devices"
	"github.com/abates/insteon/plm"
	"github.com/abates/insteon/util"
	"github.com/kirsle/configdir"
	"github.com/tarm/serial"
)

var (
	debugFlag      bool
	serialPortFlag string
	ttlFlag        int
	modem          *plm.PLM
	db             util.Database
	dbfile         string
)

func main() {
	dir := configdir.LocalConfig("go-insteon")
	dbfile = filepath.Join(dir, "db.json")
	err := configdir.MakePath(dir)
	if err != nil {
		log.Fatalf("Failed to create config file path: %v", err)
	}

	flag.StringVar(&serialPortFlag, "port", "/dev/ttyUSB0", "serial port connected to a PLM")
	flag.BoolVar(&debugFlag, "debug", false, "Debug logging")
	flag.IntVar(&ttlFlag, "ttl", 3, "default ttl for sending Insteon messages")

	db, err = util.NewFileDB(dbfile)
	if err != nil {
		log.Fatalf("Failed to load database: %v", err)
	}

	flag.Parse()

	if debugFlag {
		devices.LogDebug.SetOutput(os.Stderr)
		plm.LogDebug.SetOutput(os.Stderr)
	}

	c := &serial.Config{
		Name: serialPortFlag,
		Baud: 19200,
	}

	s, err := serial.OpenPort(c)

	if err != nil {
		log.Fatalf("error opening serial port: %v", err)
	}

	modem = plm.New(s, plm.Timeout(time.Second*5))
	err = modem.IterateDevices(func(addr insteon.Address) {
		log.Printf("Querying ALDB from %s", addr)
		device, err := db.Open(modem, addr, devices.RetryFilter(3), devices.TTL(ttlFlag))
		if err == nil {
			err := util.PrintLinkDatabase(io.Discard, device)
			if err != nil {
				log.Printf("Failed to retrieve links from %v: %v", device, err)
			}
		} else {
			log.Printf("Failed to open device %s: %v", addr, err)
		}
		time.Sleep(time.Millisecond * 100)
	})

	if err != nil {
		log.Fatalf("Failed to retrieve modem info: %v", err)
	}
}
