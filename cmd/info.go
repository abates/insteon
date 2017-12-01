package main

import (
	"fmt"

	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
)

func init() {
	commands["info"] = &command{usage: "<device id>", callback: info}
}

func info(args []string, plm plm.PLM) error {
	if len(args) < 1 {
		return fmt.Errorf("device id must be specified")
	}

	addr, err := insteon.ParseAddress(args[0])
	if err == nil {
		var device insteon.Device
		device, err = plm.Connect(addr)
		if err == nil {
			fmt.Printf("Device type: %T\n", device)
			for _, link := range device.Links() {
				fmt.Printf("\t%v\n", link)
			}
		}
	} else if err == insteon.ErrNotLinked {
		err = nil
		fmt.Printf("Device %s is not linked to the PLM.  Link it now? (y/n) ", addr)
	}
	return err
}
