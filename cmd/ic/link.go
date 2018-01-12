package main

import (
	"fmt"
	"os"
	"time"

	"github.com/abates/insteon"
)

func init() {
	Commands.Register("link", "<device id> ...", "Link the PLM to one or more devices", modemLinkCmd)
	Commands.Register("alllink", "<device id> ...", "Put the PLM into linking mode for manual linking", modemAllLinkCmd)
}

func modemLinkCmd(args []string, subCommand *Command) error {
	if len(args) < 1 {
		return fmt.Errorf("at least one device id must be specified")
	}

	for _, arg := range args {
		addr, err := insteon.ParseAddress(arg)
		if err == nil {
			group := insteon.Group(0x01)
			fmt.Printf("Linking to %s...", addr)
			device, err := modem.Connect(addr)
			if err == insteon.ErrNotLinked {
				err = nil
			}

			if err == nil {
				err = insteon.ForceLink(device, modem, group)
				if err == nil {
					err = insteon.ForceLink(modem, device, group)
				}

				if err == nil {
					fmt.Printf("done\n")
				} else {
					fmt.Printf("failed: %v\n", err)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Failed to connect to %s: %v\n", addr, err)
			}
		}
		time.Sleep(time.Second)
	}
	// TODO make this return a generic error if one or more of the links failed
	return nil
}

func modemAllLinkCmd(args []string, subCommand *Command) error {
	return modem.AddManualLink(insteon.Group(0x01))
}
