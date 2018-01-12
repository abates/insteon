package main

import (
	"fmt"

	"github.com/abates/insteon"
)

func init() {
	Commands.Register("unlink", "<device id> ...", "Unlink the PLM from one or more devices", unlinkCmd)
}

func unlinkCmd(args []string, subCommand *Command) (err error) {
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
