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
		callback:    plmLinkCmd,
	}

	commands["alllink"] = &command{
		description: "Put the PLM into linking mode for manual linking",
		callback:    plmAllLinkCmd,
	}
}

func plmLinkCmd(args []string, p *plm.PLM) error {
	if len(args) < 1 {
		return fmt.Errorf("at least one device id must be specified")
	}

	for _, arg := range args {
		addr, err := insteon.ParseAddress(arg)
		if err == nil {
			group := insteon.Group(0x01)
			fmt.Printf("Linking to %s...", addr)
			device, err := p.Connect(addr)
			if err == insteon.ErrNotLinked {
				err = nil
			}

			if err == nil {
				err = insteon.ForceLink(device, p, group)
				if err == nil {
					err = insteon.ForceLink(p, device, group)
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

func plmAllLinkCmd(args []string, plm *plm.PLM) error {
	return plm.AddManualLink(insteon.Group(0x01))
}
