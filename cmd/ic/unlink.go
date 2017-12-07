package main

import (
	"fmt"
	"os"
	"time"

	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
)

func init() {
	commands["unlink"] = &command{
		usage:       "<device id> ...",
		description: "Unlink the PLM from one or more devices",
		callback:    unlinkCmd,
	}
}

func unlinkCmd(args []string, plm *plm.PLM) error {
	if len(args) < 1 {
		return fmt.Errorf("at least one device id must be specified")
	}

	for i, arg := range args {
		addr, err := insteon.ParseAddress(arg)
		if err == nil {
			fmt.Printf("Unlinking from %s...", addr)
			device, err := plm.Connect(addr)
			if err == nil {
				err = insteon.Unlink(plm, device)
				if err == nil {
					fmt.Printf("successful\n")
				} else {
					fmt.Printf("failed: %v\n", err)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Failed to unlink from %s: %v", addr, err)
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
