package main

import (
	"fmt"

	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
)

func init() {
	commands["info"] = &command{
		usage:       "<device id>",
		description: "Display information about a specific device",
		callback:    infoCmd,
	}
}

func infoCmd(args []string, plm *plm.PLM) error {
	if len(args) < 1 {
		return fmt.Errorf("device id must be specified")
	}

	addr, err := insteon.ParseAddress(args[0])
	if err == nil {
		var device insteon.Device
		device, err = plm.Connect(addr)
		if err == insteon.ErrNotLinked {
			msg := fmt.Sprintf("Device %s is not linked to the PLM.  Link now? (y/n) ", addr)
			if getResponse(msg, "y", "n") == "y" {
				err = linkCmd([]string{args[0], "1"}, plm)
			}
		}

		if err == nil {
			fmt.Printf("Device type: %T\n", device)
			var db insteon.LinkDB
			db, err = device.LinkDB()
			for _, link := range db.Links() {
				fmt.Printf("\t%v\n", link)
			}
		}
	}
	return err
}
