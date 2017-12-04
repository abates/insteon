package main

import (
	"fmt"
	"strconv"

	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
)

func init() {
	commands["link"] = &command{
		usage:       "<device id> <group id>",
		description: "Link the PLM to the device",
		callback:    linkCmd,
	}
}

func linkCmd(args []string, plm *plm.PLM) error {
	if len(args) < 2 {
		return fmt.Errorf("device id and group must be specified")
	}

	addr, err := insteon.ParseAddress(args[0])
	if err == nil {
		var groupId int
		groupId, err = strconv.Atoi(args[1])
		if groupId < 0 || groupId > 255 {
			err = fmt.Errorf("Group must be between 0 and 255 (inclusive)")
		}
		if err == nil {
			device := plm.StandardConnect(addr)
			err = insteon.CreateLink(device, plm, insteon.Group(byte(groupId)))
		}
	}
	return err
}
