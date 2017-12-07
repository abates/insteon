package main

import (
	"fmt"
	"os"
	"time"

	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
)

func init() {
	commands["link"] = &command{
		usage:       "<device id> ...",
		description: "Link the PLM to one or more devices",
		callback:    linkCmd,
	}
}

func linkCmd(args []string, plm *plm.PLM) error {
	if len(args) < 1 {
		return fmt.Errorf("at least one device id must be specified")
	}

	for i, arg := range args {
		addr, err := insteon.ParseAddress(arg)
		if err == nil {
			group := insteon.Group(0x01)
			fmt.Printf("Linking to %s...", addr)
			device, err := plm.Connect(addr)
			if err == nil {
				err = insteon.ForceCreateLink(plm, device, group)
				if err == nil {
					fmt.Printf("successful\n")
				} else {
					fmt.Printf("failed: %v\n", err)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Failed to link to %s: %v", addr, err)
			}

			if i < len(args)-1 {
				time.Sleep(time.Second)
			}
		}
	}
	plm.CancelLinkingMode()
	// TODO make this return a generic error if one or more of the links failed
	return nil
}