package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/abates/cli"
	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
	"github.com/tarm/serial"
)

type LogLevelFlag insteon.LogLevel

func (llf *LogLevelFlag) Set(s string) (err error) {
	switch s {
	case "none":
	case "info":
		(*llf) = insteon.LevelInfo
	case "debug":
		(*llf) = insteon.LevelDebug
	case "trace":
		(*llf) = insteon.LevelTrace
	default:
		err = fmt.Errorf("valid values {none|info|debug|trace}")
	}
	return err
}

func (llf *LogLevelFlag) Get() interface{} {
	return insteon.LogLevel(*llf)
}

func (llf *LogLevelFlag) String() string {
	return insteon.LogLevel(*llf).String()
}

var (
	modem          *plm.PLM
	network        insteon.Network
	logLevelFlag   LogLevelFlag
	serialPortFlag string
	timeoutFlag    time.Duration

	Commands = cli.New(os.Args[0], "", "", run)
)

func init() {
	Commands.SetOutput(os.Stderr)
	Commands.Flags.StringVar(&serialPortFlag, "port", "/dev/ttyUSB0", "serial port connected to a PLM")
	Commands.Flags.Var(&logLevelFlag, "log", "Log Level {none|info|debug|trace}")
	Commands.Flags.DurationVar(&timeoutFlag, "timeout", 5*time.Second, "read/write timeout duration")
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
		insteon.Log.Level(insteon.LogLevel(logLevelFlag))
	}

	c := &serial.Config{
		Name: serialPortFlag,
		Baud: 19200,
	}

	s, err := serial.OpenPort(c)
	if err == nil {
		defer s.Close()

		modem = plm.New(s, timeoutFlag)
		network = modem.Network()
		defer modem.Close()
		if logLevelFlag == insteon.LevelTrace {
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
