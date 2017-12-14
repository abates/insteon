package main

import (
	"fmt"

	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
)

func init() {
	commands["plm"] = &command{
		description: "Display information about the connected PLM",
		callback:    plmCmd,
	}
}

func plmCmd(args []string, plm *plm.PLM) (err error) {
	fmt.Printf("PLM Info\n")
	info, err := plm.Info()
	if err == nil {
		fmt.Printf("   Address: %s\n", info.Address)
		fmt.Printf("  Category: %02x Sub-Category: %02x\n", info.Category[0], info.Category[1])
		fmt.Printf("  Firmware: %d\n", info.Firmware)
		var db insteon.LinkDB
		db, err = plm.LinkDB()
		if err == nil {
			printLinkDatabase(db)
		}
	}
	return err
}
