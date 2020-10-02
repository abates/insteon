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
	"bytes"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
	"github.com/abates/insteon/util"
	"github.com/tarm/serial"
)

var (
	logLevelFlag   insteon.LogLevel
	serialPortFlag string
	modem          *plm.PLM
	db             *insteon.MemDB
)

func printDevInfo(device insteon.Device) {
	/*fmt.Printf("       Device: %v\n", device)
	fmt.Printf("     Category: %v\n", device.Info().DevCat)
	fmt.Printf("     Firmware: %v\n", device.Info().FirmwareVersion)*/

	//err := util.PrintLinks(os.Stdout, device)
	err := util.PrintLinks(bytes.NewBuffer(nil), device)
	if err != nil {
		log.Printf("Failed to retrieve links from %v: %v", device, err)
	}
}

func dump(links []*insteon.LinkRecord) {
	read := make(map[insteon.Address]bool)
	for _, link := range links {
		if _, found := read[link.Address]; found {
			continue
		}
		read[link.Address] = true
		log.Printf("Querying ALDB from %s", link.Address)
		device, err := modem.Open(link.Address)
		if err == nil {
			util.Save("db.json", db)
			printDevInfo(device)
		} else {
			log.Printf("Failed to open device %s: %v", link.Address, err)
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func init() {
	flag.StringVar(&serialPortFlag, "port", "/dev/ttyUSB0", "serial port connected to a PLM")
	flag.Var(&logLevelFlag, "log", "Log Level {none|info|debug|trace}")

	var err error
	db, err = util.LoadMemDB("db.json")
	if err != nil {
		log.Fatalf("Failed to load database: %v", err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err := util.Save("db.json", db)
		if err != nil {
			log.Printf("Failed to save database: %v", err)
		}
		os.Exit(0)
	}()
}

func main() {
	flag.Parse()

	if logLevelFlag > insteon.LevelNone {
		insteon.Log.Level = logLevelFlag
	}

	c := &serial.Config{
		Name: serialPortFlag,
		Baud: 19200,
	}

	s, err := serial.OpenPort(c)

	if err != nil {
		log.Fatalf("error opening serial port: %v", err)
	}

	modem, err = plm.New(s, s, time.Second*5, plm.Database(db), plm.ConnectionOptions(insteon.ConnectionTimeout(time.Second*5)))
	if err != nil {
		log.Fatalf("failed to initialize modem: %v", err)
	}

	if links, err := modem.Links(); err == nil {
		//time.Sleep(time.Millisecond * 100)
		dump(links)
	} else {
		log.Fatalf("Failed to retrieve modem info: %v", err)
	}
}
