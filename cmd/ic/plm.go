package main

import (
	"fmt"
	"os"

	"github.com/abates/cli"
	"github.com/abates/insteon"
)

func init() {
	cmd := Commands.Register("plm", "", "Interact with the PLM", nil)
	cmd.Register("info", "", "display information (device id, link database, etc)", plmInfoCmd)
	cmd.Register("link", "<device id> ...", "Link (as a controller) the PLM to one or more devices", plmLinkCmd)
	cmd.Register("unlink", "<device id> ...", "Unlink the PLM from one or more devices", plmUnlinkCmd)
	cmd.Register("crosslink", "<device id> ...", "Crosslink the PLM to one or more devices", plmCrossLinkCmd)
	cmd.Register("alllink", "<device id> ...", "Put the PLM into linking mode for manual linking", plmAllLinkCmd)
	cmd.Register("reset", "", "Factory reset the IM", plmResetCmd)
}

func plmResetCmd(args []string, next cli.NextFunc) (err error) {
	msg := "WARNING: This will erase the modem All-Link database and reset the modem to factory defaults\nProceed? (y/n) "
	if getResponse(msg, "y", "n") == "y" {
		err = modem.Reset()
	}
	return nil
}

func plmInfoCmd(args []string, next cli.NextFunc) (err error) {
	fmt.Printf("PLM Info\n")
	info, err := modem.Info()
	if err == nil {
		fmt.Printf("   Address: %s\n", info.Address)
		fmt.Printf("  Category: %02x Sub-Category: %02x\n", info.DevCat.Category(), info.DevCat.SubCategory())
		fmt.Printf("  Firmware: %d\n", info.Firmware)
		var db insteon.LinkDB
		db, err = modem.LinkDB()
		if err == nil {
			err = printLinkDatabase(db)
		}
	}
	return err
}

func plmLinkCmd(args []string, next cli.NextFunc) error {
	return plmLink(args, false)
}

func plmCrossLinkCmd(args []string, next cli.NextFunc) error {
	return plmLink(args, true)
}

func plmLink(args []string, crosslink bool) error {
	if len(args) < 1 {
		return fmt.Errorf("at least one device id must be specified")
	}

	for _, arg := range args {
		var addr insteon.Address
		err := addr.UnmarshalText([]byte(arg))
		if err == nil {
			group := insteon.Group(0x01)
			fmt.Printf("Linking to %s...", addr)
			device, err := modem.Dial(addr)
			if err == insteon.ErrNotLinked {
				err = nil
			}

			if err == nil {
				err = insteon.ForceLink(group, modem, device)
				if err == nil && crosslink {
					err = insteon.ForceLink(group, device, modem)
				}

				if err == nil {
					fmt.Printf("done\n")
				} else {
					fmt.Printf("failed: %v\n", err)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Failed to connect to %s: %v\n", addr, err)
			}
		}
	}
	// TODO make this return a generic error if one or more of the links failed
	return nil
}

func plmAllLinkCmd(args []string, next cli.NextFunc) error {
	return modem.AddManualLink(insteon.Group(0x01))
}

func plmUnlinkCmd(args []string, next cli.NextFunc) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("at least one device id must be specified")
	}

	group := insteon.Group(0x01)

	for _, arg := range args {
		var device insteon.Device
		var addr insteon.Address
		err = addr.UnmarshalText([]byte(arg))
		if err == nil {
			fmt.Printf("Unlinking from %s...", addr)
			device, err = modem.Dial(addr)

			if err == nil {
				err = insteon.Unlink(group, device, modem)
			}

			if err == nil || err == insteon.ErrNotLinked {
				err = insteon.Unlink(group, modem, device)
			}

			if err == insteon.ErrNotLinked {
				err = nil
			}

			if err == nil {
				fmt.Printf("successful\n")
			} else {
				fmt.Printf("failed: %v\n", err)
			}
		}
	}
	// TODO make this return a generic error if one or more of the links failed
	return err
}
