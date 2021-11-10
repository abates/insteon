// Copyright 2021 Andrew Bates
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
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/abates/insteon"
	"github.com/abates/insteon/devices"
	"github.com/abates/insteon/plm"
	"github.com/abates/insteon/util"
	"github.com/kirsle/configdir"
	"github.com/tarm/serial"
)

func main() {
	configDir := configdir.LocalConfig("go-insteon")
	dbfile := filepath.Join(configDir, "db.json")

	debugFlag := false
	serialPortFlag := ""

	flag.BoolVar(&debugFlag, "debug", false, "turn on debug log")
	flag.StringVar(&serialPortFlag, "port", "/dev/ttyUSB0", "serial port connected to a PLM")
	flag.Parse()

	if debugFlag {
		plm.LogDebug.SetOutput(os.Stderr)
		devices.LogDebug.SetOutput(os.Stderr)
	}

	c := &serial.Config{
		Name: serialPortFlag,
		Baud: 19200,
	}

	s, err := serial.OpenPort(c)

	if err != nil {
		log.Fatalf("error opening serial port: %v", err)
	}

	db, err := util.NewFileDB(dbfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize device database: %v\n", err)
		os.Exit(1)
	}

	rxReader, rxWriter := io.Pipe()
	txReader, txWriter := io.Pipe()

	tx := io.TeeReader(s, txWriter)
	rx := io.TeeReader(os.Stdin, rxWriter)

	go func() {
		mon := util.Snoop(os.Stderr, db).Filter(plm.Snoop(rxReader, txReader))
		for _, err = mon.Read(); err == nil || errors.Is(err, insteon.ErrReadTimeout); _, err = mon.Read() {
		}
	}()

	go io.Copy(os.Stdout, tx)
	io.Copy(s, rx)
}
