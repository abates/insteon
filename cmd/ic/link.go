package main

import (
	"fmt"
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

func linkCmd(args []string, plm *plm.PLM) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("at least one device id must be specified")
	}

	for i, arg := range args {
		var addr insteon.Address
		addr, err = insteon.ParseAddress(arg)
		if err == nil {
			group := insteon.Group(0x01)
			fmt.Printf("Linking to %s...", addr)
			err = plm.EnterLinkingMode(group)
			// wait for the message to propogate the network
			time.Sleep(2 * time.Second)
			if err == nil {
				bridge := plm.Dial(addr)
				device := insteon.Device(insteon.NewI2CsDevice(addr, bridge))
				err = device.EnterLinkingMode(group)
				if err == nil {
					fmt.Printf("successful\n")
				} else {
					fmt.Printf("failed: %v\n", err)
				}

				if i < len(args)-1 {
					time.Sleep(time.Second)
				}
			}
		}
	}
	return err
}
