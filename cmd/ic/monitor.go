package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
)

func init() {
	commands["monitor"] = &command{
		usage:       "",
		description: "Monitor the Insteon network",
		callback:    monCmd,
	}
}

func dump(buf []byte) string {
	str := make([]string, len(buf))
	for i := range str {
		str[i] = fmt.Sprintf("%02x", buf[i])
	}
	return strings.Join(str, " ")
}

func monCmd(args []string, plm *plm.PLM) error {
	log.Printf("Starting monitor...")
	plm.Monitor(func(buf []byte, msg *insteon.Message) {
		log.Printf("%-71s %s", dump(buf), msg)
	})
	return nil
}
