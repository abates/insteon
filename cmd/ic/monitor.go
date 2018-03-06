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
	for msg := range modem.Monitor() {
		log.Printf("%s", msg)
	}
	return nil
}
