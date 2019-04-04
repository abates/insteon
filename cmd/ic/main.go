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
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/abates/cli"
	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
	"github.com/tarm/serial"
)

var (
	modem          *plm.PLM
	logLevelFlag   insteon.LogLevel
	serialPortFlag string
	timeoutFlag    time.Duration
	writeDelayFlag time.Duration

	Commands = cli.New(os.Args[0], "", "", run)
)

func init() {
	Commands.SetOutput(os.Stderr)
	Commands.Flags.StringVar(&serialPortFlag, "port", "/dev/ttyUSB0", "serial port connected to a PLM")
	Commands.Flags.Var(&logLevelFlag, "log", "Log Level {none|info|debug|trace}")
	Commands.Flags.DurationVar(&timeoutFlag, "timeout", 3*time.Second, "read/write timeout duration")
	Commands.Flags.DurationVar(&writeDelayFlag, "writeDelay", 500*time.Millisecond, "writeDelay duration")
}

func getResponse(message string, acceptable ...string) (resp string) {
	accept := make(map[string]bool, len(acceptable))
	for _, a := range acceptable {
		accept[a] = true
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(message)
		resp, _ = reader.ReadString('\n')
		resp = strings.ToLower(strings.TrimSpace(resp))
		if accept[resp] {
			break
		}
		fmt.Printf("Invalid input\n")
	}
	return resp
}

func run(args []string, next cli.NextFunc) error {
	if logLevelFlag > insteon.LevelNone {
		insteon.Log.Level(logLevelFlag)
	}

	// allow simulation from stdin
	var s io.ReadWriter
	var err error
	if serialPortFlag == "-" {
		s = &simReaderWriter{}
	} else {
		c := &serial.Config{
			Name: serialPortFlag,
			Baud: 19200,
		}

		s, err = serial.OpenPort(c)
	}

	if err == nil {
		if closer, ok := s.(io.Closer); ok {
			defer closer.Close()
		}

		modem = plm.New(plm.NewPort(s, timeoutFlag), timeoutFlag, plm.WriteDelay(writeDelayFlag))
		defer modem.Close()
		if logLevelFlag == insteon.LevelTrace {
			var config *plm.Config
			config, err = modem.Config()
			config.SetMonitorMode()
			err = modem.SetConfig(config)
			//modem.StartMonitor()
			//defer modem.StopMonitor()
		}
		return next()
	}
	return err
}

func main() {
	err := Commands.Run(os.Args[1:])
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
