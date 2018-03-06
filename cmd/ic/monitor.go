package main

import (
	"log"

	"github.com/abates/cli"
)

func init() {
	Commands.Register("monitor", "", "Monitor the Insteon network", monCmd)
}

func monCmd([]string, cli.NextFunc) error {
	log.Printf("Starting monitor...")
	monitor := modem.Monitor()
	for msg := range monitor.Accept() {
		log.Printf("%s", monitor.DumpMessage(msg))
	}
	return nil
}
