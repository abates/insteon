package main

import (
	"fmt"
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
			if err != nil {
				fmt.Printf("failed: %v\n", err)
				continue
			}

			err = insteon.Unlink(plm, device)
			if err != nil {
				fmt.Printf("failed: %v\n", err)
				continue
			}

			fmt.Printf("successful\n")

			if i < len(args)-1 {
				time.Sleep(time.Second)
			}
		}

		plmDB, err := plm.LinkDB()
		if err == nil {
			for _, link := range plmDB.Links() {
				if link.Address == addr {
					err = plmDB.RemoveLink(link)
				}
			}
		}
	}
	plm.CancelLinkingMode()
	// TODO make this return a generic error if one or more of the links failed
	return nil
}
