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
	"strings"

	"github.com/abates/cli"
	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
	"github.com/abates/insteon/util"
)

type addrList []insteon.Address

func (al *addrList) Set(str string) error {
	for _, v := range strings.Split(str, ",") {
		addr := insteon.Address{}
		err := addr.Set(v)
		if err != nil {
			return err
		}
		*al = append(*al, addr)
	}
	return nil
}

func (al *addrList) String() string {
	list := make([]string, len(*al))
	for i, addr := range *al {
		list[i] = addr.String()
	}
	return strings.Join(list, ",")
}

type addresses []insteon.Address

func (a *addresses) Set(str string) error {
	addr := insteon.Address{}
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
}

func init() {
	p := &plmCmd{PLM: modem}

	pc := app.SubCommand("plm", cli.DescOption("Interact with the PLM"))
	pc.SubCommand("edit", cli.DescOption("edit the PLM all-link database"), cli.CallbackOption(p.editCmd))
	pc.SubCommand("info", cli.DescOption("display information (device id, link database, etc)"), cli.CallbackOption(p.infoCmd))
	pc.SubCommand("reset", cli.DescOption("Factory reset the IM"), cli.CallbackOption(p.resetCmd))

	linkCmd := pc.SubCommand("link", cli.DescOption("Link the PLM to a device"))
	cmd := linkCmd.SubCommand("controller", cli.UsageOption("<group> <device id>,..."), cli.DescOption("Link (as a controller) the PLM to one or more devices. Device IDs must be comma separated"), cli.CallbackOption(p.controllerLinkCmd))
	cmd.Arguments.Int(&p.group, "<group id>")
	cmd.Arguments.VarSlice((*addrList)(&p.addresses), "<device id>,...")

	cmd = linkCmd.SubCommand("responder", cli.UsageOption("<group> <device id>,..."), cli.DescOption("Link (as a responder) the PLM to one or more devices. Device IDs must be comma separated"), cli.CallbackOption(p.responderLinkCmd))
	cmd.Arguments.Int(&p.group, "<group id>")
	cmd.Arguments.VarSlice((*addrList)(&p.addresses), "<device id>,...")

	cmd = pc.SubCommand("unlink", cli.UsageOption("<group> <device id>,..."), cli.DescOption("Unlink the PLM from one or more devices. Device IDs must be comma separated"), cli.CallbackOption(p.unlinkCmd))
	cmd.Arguments.Int(&p.group, "<group id>")
	cmd.Arguments.VarSlice((*addrList)(&p.addresses), "<device id>,...")

	cmd = pc.SubCommand("crosslink", cli.UsageOption("<group> <device id>,..."), cli.DescOption("Crosslink the PLM to one or more devices. Device IDs must be comma separated"), cli.CallbackOption(p.crossLinkCmd))
	cmd.Arguments.Int(&p.group, "<group id>")
	cmd.Arguments.VarSlice((*addrList)(&p.addresses), "<device id>,...")

	cmd = pc.SubCommand("alllink", cli.UsageOption("<group> <device id>,..."), cli.DescOption("Put the PLM into linking mode for manual linking. Device IDs must be comma separated"), cli.CallbackOption(p.allLinkCmd))
	cmd.Arguments.Int(&p.group, "<group id>")
	cmd.Arguments.VarSlice((*addrList)(&p.addresses), "<device id>,...")
}

func (p *plmCmd) editCmd(string) error {
	return editLinks(modem)
}

func (p *plmCmd) resetCmd(string) (err error) {
	msg := "WARNING: This will erase the modem All-Link database and reset the modem to factory defaults\nProceed? (y/n) "
	if cli.Query(os.Stdin, os.Stdout, msg, "y", "n") == "y" {
		err = modem.Reset()
	}
	return err
}

func (p *plmCmd) infoCmd(string) (err error) {
	fmt.Printf("PLM Info\n")
	info, err := modem.Info()
	if err == nil {
		fmt.Printf("   Address: %s\n", info.Address)
		fmt.Printf("  Category: %02x Sub-Category: %02x\n", info.DevCat.Domain(), info.DevCat.Category())
		fmt.Printf("  Firmware: %d\n", info.Firmware)
		err = util.PrintLinks(os.Stdout, modem)
	}
	return err
}

func (p *plmCmd) controllerLinkCmd(string) error { return p.link(true, false) }
func (p *plmCmd) responderLinkCmd(string) error  { return p.link(false, true) }
func (p *plmCmd) crossLinkCmd(string) error      { return p.link(true, true) }

func (p *plmCmd) link(controller, responder bool) error {
	return util.IfLinkable(modem, func(lmodem insteon.Linkable) (err error) {
		for _, addr := range p.addresses {
			group := insteon.Group(p.group)
			fmt.Printf("Linking to %s...", addr)
			device, err := modem.Open(addr)
			if err == insteon.ErrNotLinked {
				err = nil
			}

			if err == nil {
				err = util.IfLinkable(device, func(ldevice insteon.Linkable) (err error) {
					if controller {
						err = util.ForceLink(group, lmodem, ldevice)
					}

					if err == nil && responder {
						err = util.ForceLink(group, ldevice, lmodem)
					}
					return err
				})

				if err == nil {
					fmt.Printf("done\n")
				} else {
					fmt.Printf("failed: %v\n", err)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Failed to connect to %s: %v\n", addr, err)
			}
		}
		return err
	})
}

func (p *plmCmd) allLinkCmd(string) error {
	return util.IfLinkable(modem, func(linkable insteon.Linkable) error {
		return linkable.EnterLinkingMode(insteon.Group(p.group))
	})
}

func (p *plmCmd) unlinkCmd(string) (err error) {
	group := insteon.Group(p.group)

	return util.IfLinkable(modem, func(lmodem insteon.Linkable) (err error) {
		for _, addr := range p.addresses {
			var device insteon.Device
			device, err = modem.Open(addr)

			if err == nil {
				fmt.Printf("Unlinking from %s:%s...", device, addr)
				err = util.IfLinkable(device, func(ldevice insteon.Linkable) (err error) {
					if err == nil {
						err = util.Unlink(group, ldevice, lmodem)
					}

					if err == nil || err == insteon.ErrNotLinked {
						err = util.Unlink(group, lmodem, ldevice)
					}
					return err
				})
			} else if err == insteon.ErrNotLinked {
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
	})
}
