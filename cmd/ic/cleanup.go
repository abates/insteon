package main

import (
	"fmt"

	"github.com/abates/insteon/plm"
)

func init() {
	commands["cleanup"] = &command{
		usage:       "<device id> ...",
		description: "Remove duplicate links from the device link DB",
		callback:    cleanupCmd,
	}
}

func cleanupCmd(args []string, plm *plm.PLM) error {
	if len(args) < 1 {
		return fmt.Errorf("at least one device id must be specified")
	}

	/*for i, arg := range args {
		addr, err := insteon.ParseAddress(arg)
		if err == nil {
			fmt.Printf("Cleaning %s...", addr)
			device, err := plm.Connect(addr)
			if err == nil {
				linkdb, err := device.LinkDB()
				if err == nil {
					err = linkdb.Cleanup()
				}
			}

			if err == nil {
				fmt.Printf("done\n")
			} else {
				fmt.Printf("failed: %v\n", err)
			}

			if i < len(args)-1 {
				time.Sleep(time.Second)
			}
		}
	}*/
	// TODO make this return a generic error if one or more of the links failed
	return nil
}
