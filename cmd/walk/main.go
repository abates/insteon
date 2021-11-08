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
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
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

func printDevInfo(device *devices.BasicDevice) {
	err := util.PrintLinkDatabase(io.Discard, device)
	if errors.Is(err, insteon.ErrReadTimeout) {
		// try again
		err = util.PrintLinkDatabase(io.Discard, device)
	}

	if err != nil {
		log.Printf("Failed to retrieve links from %v: %v", device, err)
	}
}

func dump(links []insteon.LinkRecord) {
	read := make(map[insteon.Address]bool)
	for _, link := range links {
		if _, found := read[link.Address]; found {
			continue
		}
		read[link.Address] = true
		log.Printf("Querying ALDB from %s", link.Address)
		device, err := util.Open(devices.TTL(ttlFlag).Filter(modem), link.Address, db, dbfile)
		if errors.Is(err, insteon.ErrReadTimeout) {
			// retry
			device, err = util.Open(devices.TTL(ttlFlag).Filter(modem), link.Address, db, dbfile)
		}

		if err == nil {
			util.SaveDB(dbfile, db)
			printDevInfo(device)
		} else {
			log.Printf("Failed to open device %s: %v", link.Address, err)
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func init() {
	dir := configdir.LocalConfig("go-insteon")
	dbfile = filepath.Join(dir, "db.json")
	err := configdir.MakePath(dir)
	if err != nil {
		log.Fatalf("Failed to create config file path: %v", err)
	}

	flag.StringVar(&serialPortFlag, "port", "/dev/ttyUSB0", "serial port connected to a PLM")
	flag.BoolVar(&debugFlag, "debug", false, "Debug logging")
	flag.IntVar(&ttlFlag, "ttl", 3, "default ttl for sending Insteon messages")

	db = util.NewMemDB()
	err = util.LoadDB(dbfile, db)
	if err != nil {
		log.Fatalf("Failed to load database: %v", err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err := util.SaveDB(dbfile, db)
		if err != nil {
			log.Printf("Failed to save database: %v", err)
		}
		os.Exit(0)
	}()
}

func main() {
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

	if links, err := modem.Links(); err == nil {
		util.PrintLinks(os.Stdout, links)
		//time.Sleep(time.Millisecond * 100)
		dump(links)
	} else {
		log.Fatalf("Failed to retrieve modem info: %v", err)
	}
}
