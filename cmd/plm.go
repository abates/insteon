package main

import (
	"fmt"

	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
)

func init() {
	commands["plm"] = &command{
		usage:       "{info}",
		description: "Display information about the connected PLM",
		callback:    plmCmd,
	}
}

func plmCmd(args []string, plm *plm.PLM) (err error) {
	switch args[0] {
	case "info":
		err = plmInfoCmd(plm)
	default:
		err = fmt.Errorf("Unknown command %s", args[0])
	}
	return err
}

func plmInfoCmd(plm *plm.PLM) error {
	fmt.Printf("PLM Info\n")
	info, err := plm.Info()
	if err == nil {
		fmt.Printf("   Address: %s\n", info.Address)
		fmt.Printf("  Category: %02x Sub-Category: %02x\n", info.Category[0], info.Category[1])
		fmt.Printf("  Firmware: %d\n", info.Firmware)
		fmt.Printf("     Links:\n")
		var db insteon.LinkDB
		db, err = plm.LinkDB()
		for _, link := range db.Links() {
			fmt.Printf("%s\n", link)
		}
	}
	return err
}
