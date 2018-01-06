package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

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

type command struct {
	usage       string
	description string
	flags       *flag.FlagSet
	callback    func(args []string, plm *plm.PLM) error
}

var (
	logLevelFlag   LogLevelFlag
	serialPortFlag string
	timeoutFlag    time.Duration

	commands = make(map[string]*command)
)

func init() {
	flag.StringVar(&serialPortFlag, "port", "/dev/ttyUSB0", "serial port connected to a PLM")
	flag.Var(&logLevelFlag, "log", "Log Level {none|info|debug|trace}")
	flag.DurationVar(&timeoutFlag, "timeout", 5*time.Second, "read/write timeout duration")
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

func run(args []string, command func([]string, *plm.PLM) error) error {
	c := &serial.Config{
		Name: serialPortFlag,
		Baud: 19200,
	}

	s, err := serial.OpenPort(c)
	if err == nil {
		defer s.Close()

		plm := plm.New(s, timeoutFlag)
		defer plm.Close()
		if logLevelFlag == insteon.LevelTrace {
			plm.StartMonitor()
			defer plm.StopMonitor()
		}

		err = command(args, plm)
	}
	return err
}

func usage() {
	maxNameLen := 0
	var commandNames []string
	for name, _ := range commands {
		if len(name) > maxNameLen {
			maxNameLen = len(name)
		}
		commandNames = append(commandNames, name)
	}
	maxNameLen += 5
	nameFmt := fmt.Sprintf("%%%ds %%s\n", maxNameLen)

	sort.Strings(commandNames)

	fmt.Fprintf(os.Stderr, "Usage: %s [global options] <command> [command options]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nGlobal Options:\n")
	flag.PrintDefaults()

	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	for _, commandName := range commandNames {
		command := commands[commandName]
		fmt.Fprintf(os.Stderr, nameFmt, commandName, command.usage)
		if command.description != "" {
			fmt.Fprintf(os.Stderr, "%s %s\n", strings.Repeat(" ", maxNameLen), command.description)
		}
		if command.flags != nil {
			command.flags.PrintDefaults()
		}
		fmt.Fprintf(os.Stderr, "\n")
	}
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if len(flag.Args()) < 1 {
		usage()
		os.Exit(2)
	}

	args := flag.Args()
	cmdName := args[0]
	args = args[1:]
	command := commands[cmdName]
	if command == nil {
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", cmdName)
		os.Exit(3)
	}

	if command.flags != nil {
		command.flags.Parse(args)
		args = command.flags.Args()
	}

	if logLevelFlag > insteon.LevelNone {
		insteon.Log.Level(insteon.LogLevel(logLevelFlag))
	}

	err := run(args, command.callback)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
