package main

import (
	"log"

	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
)

func init() {
	commands["monitor"] = &command{
		usage:       "]",
		description: "Monitor the Insteon network",
		callback:    monCmd,
	}
}

func monCmd(args []string, plm *plm.PLM) error {
	plm.Monitor(func(msg insteon.Message) {
		log.Printf("%s", msg)
	})
	return nil
}
