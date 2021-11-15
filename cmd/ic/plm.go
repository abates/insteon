// Copyright 2018 Andrew Bates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"

	"github.com/abates/cli"
	"github.com/abates/insteon"
	"github.com/abates/insteon/devices"
	"github.com/abates/insteon/plm"
	"github.com/abates/insteon/util"
)

type addresses []insteon.Address

func (a *addresses) Set(str string) error {
	addr := insteon.Address(0)
	err := addr.Set(str)
	if err == nil {
		*a = append(*a, addr)
	}
	return err
}

type plmCmd struct {
	*plm.PLM
	group     int
	addresses addresses
	flag      string
}

func init() {
	linkCmd := func(name, desc string, cb interface{}) *cli.Command {
		return &cli.Command{
			Name:        name,
			Usage:       "<group> <device id>...",
			Description: desc,
			Callback:    cli.Callback(cb, "<group id>", "<device id> <device id> ..."),
		}
	}

	p := &plmCmd{PLM: modem}

	pc := &cli.Command{
		Name:        "plm",
		Description: "Interact with the PLM",
		SubCommands: []*cli.Command{
			{
				Name:        "setflag",
				Usage:       "<flag> (L - Auto Link, M - Monitor, A - Auto LED, D - Deadman mode)",
				Description: "set a config flag on the PLM",
				Callback:    cli.Callback(p.flagCmd(true), "[L|M|A|D]"),
			},
			{
				Name:        "clearflag",
				Usage:       "<flag> (L - Auto Link, M - Monitor, A - Auto LED, D - Deadman mode)",
				Description: "clear a config flag on the PLM",
				Callback:    cli.Callback(p.flagCmd(false), "[L|M|A|D]"),
			},
			{Name: "edit", Description: "edit the PLM all-link database", Callback: cli.Callback(p.editCmd)},
			{Name: "info", Description: "display information (device id, link database, etc)", Callback: cli.Callback(p.infoCmd)},
			{Name: "reset", Description: "Factory reset the IM", Callback: cli.Callback(p.resetCmd)},
			{
				Name:        "link",
				Description: "Link the PLM to a device",
				SubCommands: []*cli.Command{
					linkCmd("controller", "Link (as a controller) the PLM to one or more devices.", p.link(true, false)),
					linkCmd("responder", "Link (as a responder) the PLM to one or more devices.", p.link(false, true)),
				},
			},
			linkCmd("crosslink", "Crosslink the PLM to one or more devices", p.link(true, true)),
			linkCmd("unlink", "Unlink the PLM from one or more devices", p.unlinkCmd),
			{
				Name:        "alllink",
				Usage:       "<group>",
				Description: "Put the PLM into linking mode for manual linking",
				Callback:    cli.Callback(p.EnterLinkingMode, "<group id>"),
			},
		},
	}
	app.SubCommands = append(app.SubCommands, pc)
}

func (p *plmCmd) editCmd() error {
	return editLinks(modem)
}

func (p *plmCmd) resetCmd() (err error) {
	msg := "WARNING: This will erase the modem All-Link database and reset the modem to factory defaults\nProceed? (y/n) "
	if cli.Query(os.Stdin, os.Stdout, msg, "y", "n") == "y" {
		err = modem.Reset()
	}
	return err
}

func (p *plmCmd) infoCmd() (err error) {
	fmt.Printf("PLM Info\n")
	info, err := modem.Info()
	if err == nil {
		fmt.Printf("      Address: %s\n", info.Address)
		fmt.Printf("     Category: %02x Sub-Category: %02x\n", info.DevCat.Domain(), info.DevCat.Category())
		fmt.Printf("     Firmware: %d\n", info.Firmware)
		var config plm.Config
		config, err = modem.Config()
		if err == nil {
			fmt.Printf("       Config: %v\n", config)
			err = util.PrintLinkDatabase(os.Stdout, modem)
		}
	}
	return err
}

func (p *plmCmd) link(controller, responder bool) func(group insteon.Group, addresses util.Addresses) error {
	return func(group insteon.Group, addresses util.Addresses) error {
		for _, addr := range addresses {
			fmt.Printf("Linking to %s...", addr)
			device, err := open(modem, addr, false)
			if err == devices.ErrNotLinked {
				err = nil
			}

			if err == nil {
				if controller {
					err = util.ForceLink(group, modem, device)
				}

				if err == nil && responder {
					err = util.ForceLink(group, device, modem)
				}
				return err

				if err == nil {
					fmt.Printf("done\n")
				} else {
					fmt.Printf("failed: %v\n", err)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Failed to connect to %s: %v\n", addr, err)
			}
		}
		return nil
	}
}

func (p *plmCmd) unlinkCmd(addresses util.Addresses) (err error) {
	group := insteon.Group(p.group)

	for _, addr := range p.addresses {
		var device *devices.BasicDevice
		device, err = open(modem, addr, false)

		if err == nil {
			fmt.Printf("Unlinking from %s:%s...", device, addr)
			if err == nil {
				err = util.Unlink(group, device, modem)
			}

			if err == nil || err == devices.ErrNotLinked {
				err = util.Unlink(group, modem, device)
			}
			return err
		} else if err == devices.ErrNotLinked {
			err = nil
		}

		if err == nil {
			fmt.Printf("successful\n")
		} else {
			fmt.Printf("failed: %v\n", err)
			break
		}
	}
	return err
}

func (p *plmCmd) flagCmd(set bool) func(string) error {
	return func(flag string) error {
		config, err := modem.Config()
		if err != nil {
			return err
		}

		switch flag {
		case "L":
			if set {
				config.SetAutomaticLinking()
			} else {
				config.ClearAutomaticLinking()
			}
		case "M":
			if set {
				config.SetMonitorMode()
			} else {
				config.ClearMonitorMode()
			}
		case "A":
			if set {
				config.SetAutomaticLED()
			} else {
				config.ClearAutomaticLED()
			}
		case "D":
			if set {
				config.SetDeadmanMode()
			} else {
				config.ClearDeadmanMode()
			}
		default:
			return fmt.Errorf("Unrecognized flag %q", p.flag)
		}

		return modem.SetConfig(config)
	}
}
