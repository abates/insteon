package main

import (
	"fmt"

	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
)

var device insteon.Device

func init() {
	cmd := Commands.Register("device", "<device id> [info|link|unlink|cleanup|dump]", "Interact with a specific device", devCmd)
	cmd.Register("info", "", "retrieve device info", devInfoCmd)
	cmd.Register("link", "", "enter linking mode", devLinkCmd)
	cmd.Register("unlink", "", "enter unlinking mode", devUnlinkCmd)
	cmd.Register("exitlink", "", "exit linking mode", devExitLinkCmd)
	cmd.Register("cleanup", "", "remove duplicate links in the all-link database", devCleanupCmd)
	cmd.Register("dump", "", "dump the device all-link database", devDumpCmd)
}

func devCmd(args []string, subCommand *Command) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("device id and action must be specified")
	}

	addr, err := insteon.ParseAddress(args[0])
	if err != nil {
		return fmt.Errorf("invalid device address: %v", err)
	}

	device, err = devConnect(modem, addr)
	defer device.Close()
	if err == nil {
		err = subCommand.Run(args)
	}
	return err
}

func devConnect(modem *plm.PLM, addr insteon.Address) (insteon.Device, error) {
	device, err := modem.Connect(addr)
	if err == insteon.ErrNotLinked {
		msg := fmt.Sprintf("Device %s is not linked to the PLM.  Link now? (y/n) ", addr)
		if getResponse(msg, "y", "n") == "y" {
			err = modemLinkCmd([]string{addr.String()}, nil)
		}

		if err == nil {
			device, err = modem.Connect(addr)
		}
	}
	return device, err
}

func devLinkCmd([]string, *Command) error {
	return device.EnterLinkingMode(insteon.Group(0x01))
}

func devUnlinkCmd([]string, *Command) error {
	return device.EnterUnlinkingMode(insteon.Group(0x01))
}

func devExitLinkCmd([]string, *Command) error {
	return device.ExitLinkingMode()
}

func devCleanupCmd([]string, *Command) error {
	/*db, err := device.LinkDB()
	if err == nil {
		err = db.Cleanup()
	}
	return err*/
	return nil
}

func devDumpCmd([]string, *Command) error {
	db, err := device.LinkDB()
	if err == nil {
		err = dumpLinkDatabase(db)
	}
	return err
}

func devInfoCmd([]string, *Command) error {
	pd, err := device.ProductData()

	if err == nil {
		fmt.Printf("%s\n", device)
		fmt.Printf("  Product Key: %s\n", pd.Key)
		fmt.Printf("     Category: %s\n", pd.Category)
		var db insteon.LinkDB
		db, err = device.LinkDB()
		if err == nil {
			err = printLinkDatabase(db)
		}
	}
	return err
}
