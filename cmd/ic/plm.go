package main

import (
	"fmt"

	"github.com/abates/insteon"
)

func init() {
	Commands.Register("plm", "", "Display information about the connected PLM", plmCmd)
}

func plmCmd(args []string, subCommand *Command) (err error) {
	fmt.Printf("PLM Info\n")
	info, err := modem.Info()
	if err == nil {
		fmt.Printf("   Address: %s\n", info.Address)
		fmt.Printf("  Category: %02x Sub-Category: %02x\n", info.Category[0], info.Category[1])
		fmt.Printf("  Firmware: %d\n", info.Firmware)
		var db insteon.LinkDB
		db, err = modem.LinkDB()
		if err == nil {
			err = printLinkDatabase(db)
		}
	}
	return err
}
