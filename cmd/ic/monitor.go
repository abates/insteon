package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/abates/insteon"
)

func init() {
	Commands.Register("monitor", "", "Monitor the Insteon network", monCmd)
}

func dump(buf []byte) string {
	str := make([]string, len(buf))
	for i := range str {
		str[i] = fmt.Sprintf("%02x", buf[i])
	}
	return strings.Join(str, " ")
}

func monCmd(args []string, subCommand *Command) error {
	log.Printf("Starting monitor...")
	modem.Monitor(func(buf []byte, msg *insteon.Message) {
		log.Printf("%-71s %s", dump(buf), msg)
	})
	return nil
}
