package main

import (
	"fmt"

	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
)

func init() {
	commands["version"] = &command{
		usage:       "<device id>",
		description: "Retrieve the Insteon engine version",
		callback:    versionCmd,
	}
}

func versionCmd(args []string, plm *plm.PLM) error {
	if len(args) < 1 {
		return fmt.Errorf("device id must be specified")
	}

	addr, err := insteon.ParseAddress(args[0])
	if err == nil {
		var device insteon.Device
		device, err = plm.Connect(addr)

		if err == nil {
			version := insteon.EngineVersion(0)
			version, err = device.EngineVersion()
			fmt.Printf("Device version: %s\n", version)
		}
	}
	return err
}
