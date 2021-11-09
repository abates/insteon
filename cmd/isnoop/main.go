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
	"github.com/abates/insteon/plm"
	"github.com/abates/insteon/util"
	"github.com/creack/pty"
	"github.com/kirsle/configdir"
	"github.com/tarm/serial"
)

func main() {
	configDir := configdir.LocalConfig("go-insteon")
	dbfile := filepath.Join(configDir, "db.json")

	serialPortFlag := ""
	flag.StringVar(&serialPortFlag, "port", "/dev/ttyUSB0", "serial port connected to a PLM")
	flag.Parse()

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

	p, f, err := pty.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start PTY: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "Connect intercepted application to %s\n", f.Name())

	txReader, txWriter := io.Pipe()
	rxReader, rxWriter := io.Pipe()

	rx := io.TeeReader(s, rxWriter)
	tx := io.TeeReader(p, txWriter)

	go func() {
		mon := util.Snoop(os.Stdout, db).Filter(plm.Snoop(rxReader, txReader))
		for _, err = mon.Read(); err == nil || errors.Is(err, insteon.ErrReadTimeout); _, err = mon.Read() {
		}
	}()

	go io.Copy(p, rx)
	io.Copy(s, tx)
}
