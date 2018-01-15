package main

import (
	"fmt"
	"os"
	"time"

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
}

func plmInfoCmd(args []string, next cli.NextFunc) (err error) {
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
		addr, err := insteon.ParseAddress(arg)
		if err == nil {
			group := insteon.Group(0x01)
			fmt.Printf("Linking to %s...", addr)
			device, err := modem.Connect(addr)
			if err == insteon.ErrNotLinked {
				err = nil
			}

			if err == nil {
				err = insteon.ForceLink(device, modem, group)
				if err == nil && crosslink {
					err = insteon.ForceLink(modem, device, group)
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
		time.Sleep(time.Second)
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
		addr, err = insteon.ParseAddress(arg)
		if err == nil {
			fmt.Printf("Unlinking from %s...", addr)
			device, err = modem.Dial(addr)

			if err == nil {
				err = insteon.Unlink(device, modem, group)
			}

			if err == nil || err == insteon.ErrNotLinked {
				err = insteon.Unlink(modem, device, group)
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

		var modemDB insteon.LinkDB
		modemDB, err = modem.LinkDB()
		if err == nil {
			var links []*insteon.LinkRecord
			links, err = modemDB.Links()
			if err == nil {
				removeable := make([]*insteon.LinkRecord, 0)
				for _, link := range links {
					if link.Address == addr {
						fmt.Printf("Cleaning up old link...")
						err = modemDB.RemoveLinks(link)
						if err == nil {
							fmt.Printf("successful\n")
						} else {
							fmt.Printf("failed: %v\n", err)
						}
					}
				}

				fmt.Printf("Cleaning up old link...")
				err = modemDB.RemoveLinks(removeable...)
				if err == nil {
					fmt.Printf("successful\n")
				} else {
					fmt.Printf("failed: %v\n", err)
				}
			}
		}

		if err != nil {
			break
		}
	}
	// TODO make this return a generic error if one or more of the links failed
	return nil
}
