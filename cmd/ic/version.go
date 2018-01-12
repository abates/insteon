package main

import (
	"fmt"

	"github.com/abates/insteon"
)

func init() {
	Commands.Register("version", "<device id>", "Retrieve the Insteon engine version", versionCmd)
}

func versionCmd(args []string, subCommand *Command) error {
	if len(args) < 1 {
		return fmt.Errorf("device id must be specified")
	}

	addr, err := insteon.ParseAddress(args[0])
	if err == nil {
		var device insteon.Device
		device, err = modem.Connect(addr)

		if err == nil {
			version := insteon.EngineVersion(0)
			version, err = device.EngineVersion()
			fmt.Printf("Device version: %s\n", version)
		}
	}
	return err
}
