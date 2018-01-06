package main

import (
	"fmt"

	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
)

func init() {
	commands["unlink"] = &command{
		usage:       "<device id> ...",
		description: "Unlink the PLM from one or more devices",
		callback:    unlinkCmd,
	}
}

func unlinkCmd(args []string, p *plm.PLM) (err error) {
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
			device, err = p.Dial(addr)

			if err == nil {
				err = insteon.Unlink(device, p, group)
			}

			if err == nil || err == insteon.ErrNotLinked {
				err = insteon.Unlink(p, device, group)
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

		var plmDB insteon.LinkDB
		plmDB, err = p.LinkDB()
		if err == nil {
			var links []*insteon.Link
			links, err = plmDB.Links()
			if err == nil {
				removeable := make([]*insteon.Link, 0)
				for _, link := range links {
					if link.Address == addr {
						removeable = append(removeable, link)
					}
				}

				fmt.Printf("Cleaning up old link...")
				err = plmDB.RemoveLinks(removeable...)
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
