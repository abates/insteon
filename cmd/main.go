package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

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
	usage    string
	flags    *flag.FlagSet
	callback func(args []string, plm plm.PLM) error
}

var (
	logLevelFlag   LogLevelFlag
	serialPortFlag string

	commands = make(map[string]*command)
)

func info(args []string, plm plm.PLM) error {
	if len(args) < 1 {
		return fmt.Errorf("device id must be specified")
	}

	addr, err := insteon.ParseAddress(args[0])
	if err == nil {
		//addr := insteon.Address([3]byte{0x47, 0x1b, 0xd4})
		//addr := insteon.Address([3]byte{0x48, 0x0f, 0xfe})
		//addr := insteon.Address([3]byte{0x48, 0xd5, 0xf0})
		var device insteon.Device
		device, err = plm.Connect(addr)
		if err == nil {
			fmt.Printf("Device type: %T\n", device)
			for _, link := range device.Links() {
				fmt.Printf("\t%v\n", link)
			}
		}
	}
	return err
}

func init() {
	flag.StringVar(&serialPortFlag, "port", "/dev/ttyUSB0", "serial port connected to a PLM")
	flag.Var(&logLevelFlag, "log", "Log Level {none|info|debug|trace}")
	commands["info"] = &command{usage: "<device id>", callback: info}
}

func run(args []string, command func([]string, plm.PLM) error) error {
	c := &serial.Config{
		Name: serialPortFlag,
		Baud: 19200,
	}

	s, err := serial.OpenPort(c)
	if err == nil {
		defer s.Close()

		plm := plm.New(s)
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
	nameFmt := fmt.Sprintf("%%%ds %%s\n", maxNameLen+5)

	sort.Strings(commandNames)

	fmt.Fprintf(os.Stderr, "Usage: %s [global options] <command> [command options]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nGlobal Options:\n")
	flag.PrintDefaults()

	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	for _, commandName := range commandNames {
		command := commands[commandName]
		fmt.Fprintf(os.Stderr, nameFmt, commandName, command.usage)
		if command.flags != nil {
			command.flags.PrintDefaults()
		}
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
		fmt.Fprintf(os.Stderr, "Unknown command %s", cmdName)
		os.Exit(3)
	}

	if command.flags != nil {
		command.flags.Parse(args)
		args = command.flags.Args()
	}

	if logLevelFlag > insteon.LevelNone {
		insteon.Log = insteon.StderrLogger
		insteon.Log.Level(insteon.LogLevel(logLevelFlag))
	}

	err := run(args, command.callback)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
}
