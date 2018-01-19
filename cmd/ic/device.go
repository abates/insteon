package main

import (
	"fmt"

	"github.com/abates/cli"
	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
)

var device insteon.Device

func init() {
	cmd := Commands.Register("device", "<command> <device id>", "Interact with a specific device", devCmd)
	cmd.Register("info", "", "retrieve device info", devInfoCmd)
	cmd.Register("link", "", "enter linking mode", devLinkCmd)
	cmd.Register("unlink", "", "enter unlinking mode", devUnlinkCmd)
	cmd.Register("exitlink", "", "exit linking mode", devExitLinkCmd)
	cmd.Register("cleanup", "", "remove duplicate links in the all-link database", devCleanupCmd)
	cmd.Register("dump", "", "dump the device all-link database", devDumpCmd)
	cmd.Register("version", "<device id>", "Retrieve the Insteon engine version", devVersionCmd)
}

func devCmd(args []string, next cli.NextFunc) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("device id and action must be specified")
	}

	addr, err := insteon.ParseAddress(args[0])
	if err != nil {
		return fmt.Errorf("invalid device address: %v", err)
	}

	device, err = devConnect(modem, addr)
	defer device.Close()
	if err == nil {
		err = next()
	}
	return err
}

func devConnect(modem *plm.PLM, addr insteon.Address) (insteon.Device, error) {
	device, err := modem.Dial(addr)
	if err == insteon.ErrNotLinked {
		msg := fmt.Sprintf("Device %s is not linked to the PLM.  Link now? (y/n) ", addr)
		if getResponse(msg, "y", "n") == "y" {
			err = plmLinkCmd([]string{addr.String()}, nil)
		}

		if err == nil {
			device, err = modem.Dial(addr)
		}
	}
	return device, err
}

func devLinkCmd([]string, cli.NextFunc) error {
	return device.EnterLinkingMode(insteon.Group(0x01))
}

func devUnlinkCmd([]string, cli.NextFunc) error {
	return device.EnterUnlinkingMode(insteon.Group(0x01))
}

func devExitLinkCmd([]string, cli.NextFunc) error {
	return device.ExitLinkingMode()
}

func devDumpCmd([]string, cli.NextFunc) error {
	db, err := device.LinkDB()
	if err == nil {
		err = dumpLinkDatabase(db)
	}
	return err
}

func devInfoCmd([]string, cli.NextFunc) (err error) {
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

func devVersionCmd([]string, cli.NextFunc) error {
	version, err := device.EngineVersion()
	if err == nil {
		fmt.Printf("Device version: %s\n", version)
	}
	return err
}

func devCleanupCmd([]string, cli.NextFunc) error {
	/*for i, arg := range args {
		addr, err := insteon.ParseAddress(arg)
		if err == nil {
			fmt.Printf("Cleaning %s...", addr)
			device, err := plm.Connect(addr)
			if err == nil {
				linkdb, err := device.LinkDB()
				if err == nil {
					err = linkdb.Cleanup()
				}
			}

			if err == nil {
				fmt.Printf("done\n")
			} else {
				fmt.Printf("failed: %v\n", err)
			}

			if i < len(args)-1 {
				time.Sleep(time.Second)
			}
		}
	}*/
	// TODO make this return a generic error if one or more of the links failed
	return nil
}
