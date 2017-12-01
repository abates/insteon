package main

import (
	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
)

func init() {
	commands["link"] = &command{
		usage:       "<device id>",
		description: "Link the PLM to the device",
		callback:    linkCmd,
	}
}

func linkCmd(args []string, plm *plm.PLM) error {
	return insteon.ErrNotImplemented
}
