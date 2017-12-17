package main

import (
	"fmt"

	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
)

func init() {
	commands["device"] = &command{
		usage:       "<device id> [info|link|unlink|cleanup|dump]",
		description: "Interact with a specific device",
		callback:    devCmd,
	}
}

func devCmd(args []string, plm *plm.PLM) error {
	if len(args) < 2 {
		return fmt.Errorf("device id and action must be specified")
	}

	addr, err := insteon.ParseAddress(args[0])
	if err != nil {
		return fmt.Errorf("invalid device address: %v", err)
	}

	device, err := devConnect(plm, addr)
	defer device.Close()
	if err != nil {
		return err
	}

	switch args[1] {
	case "info":
		err = devInfoCmd(device)
	case "link":
		err = devLinkCmd(device)
	case "unlink":
		err = devUnlinkCmd(device)
	case "exitlink":
		err = devExitLinkCmd(device)
	case "cleanup":
		err = devCleanupCmd(device)
	case "dump":
		err = devDumpCmd(device)
	default:
		err = fmt.Errorf("Unknown command %q", args[1])
	}
	return err
}

func devConnect(plm *plm.PLM, addr insteon.Address) (insteon.Device, error) {
	device, err := plm.Connect(addr)
	if err == insteon.ErrNotLinked {
		msg := fmt.Sprintf("Device %s is not linked to the PLM.  Link now? (y/n) ", addr)
		if getResponse(msg, "y", "n") == "y" {
			err = plmLinkCmd([]string{addr.String()}, plm)
		}

		if err == nil {
			device, err = plm.Connect(addr)
		}
	}
	return device, err
}

func devLinkCmd(device insteon.Device) error {
	return device.EnterLinkingMode(insteon.Group(0x01))
}

func devUnlinkCmd(device insteon.Device) error {
	return device.EnterUnlinkingMode(insteon.Group(0x01))
}

func devExitLinkCmd(device insteon.Device) error {
	return device.ExitLinkingMode()
}

func devCleanupCmd(device insteon.Device) error {
	/*db, err := device.LinkDB()
	if err == nil {
		err = db.Cleanup()
	}
	return err*/
	return nil
}

func devDumpCmd(device insteon.Device) error {
	db, err := device.LinkDB()
	if err == nil {
		err = dumpLinkDatabase(db)
	}
	return err
}

func devInfoCmd(device insteon.Device) error {
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
