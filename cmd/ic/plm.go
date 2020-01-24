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
	addresses addresses
}

func init() {
	p := &plmCmd{PLM: modem}

	pc := app.SubCommand("plm", cli.DescOption("Interact with the PLM"))
	pc.SubCommand("info", cli.DescOption("display information (device id, link database, etc)"), cli.CallbackOption(p.infoCmd))
	pc.SubCommand("reset", cli.DescOption("Factory reset the IM"), cli.CallbackOption(p.resetCmd))

	cmd := pc.SubCommand("link", cli.UsageOption("<device id>,..."), cli.DescOption("Link (as a controller) the PLM to one or more devices. Device IDs must be comma separated"), cli.CallbackOption(p.linkCmd))
	cmd.Arguments.VarSlice((*addrList)(&p.addresses), "<device id>,...")

	cmd = pc.SubCommand("unlink", cli.UsageOption("<device id>,..."), cli.DescOption("Unlink the PLM from one or more devices. Device IDs must be comma separated"), cli.CallbackOption(p.unlinkCmd))
	cmd.Arguments.VarSlice((*addrList)(&p.addresses), "<device id>,...")

	cmd = pc.SubCommand("crosslink", cli.UsageOption("<device id>,..."), cli.DescOption("Crosslink the PLM to one or more devices. Device IDs must be comma separated"), cli.CallbackOption(p.crossLinkCmd))
	cmd.Arguments.VarSlice((*addrList)(&p.addresses), "<device id>,...")

	cmd = pc.SubCommand("alllink", cli.UsageOption("<device id>,..."), cli.DescOption("Put the PLM into linking mode for manual linking. Device IDs must be comma separated"), cli.CallbackOption(p.allLinkCmd))
	cmd.Arguments.VarSlice((*addrList)(&p.addresses), "<device id>,...")
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
		fmt.Printf("  Category: %02x Sub-Category: %02x\n", info.DevCat.Category(), info.DevCat.SubCategory())
		fmt.Printf("  Firmware: %d\n", info.Firmware)
		err = isLinkable(modem, func(linkable insteon.Linkable) error {
			return util.PrintLinks(os.Stdout, linkable)
		})
	}
	return err
}

func (p *plmCmd) linkCmd(string) error      { return p.link(false) }
func (p *plmCmd) crossLinkCmd(string) error { return p.link(true) }

func (p *plmCmd) link(crosslink bool) error {
	return isLinkable(modem, func(lmodem insteon.Linkable) (err error) {
		for _, addr := range p.addresses {
			group := insteon.Group(0x01)
			fmt.Printf("Linking to %s...", addr)
			device, err := modem.Open(addr, insteon.ConnectionTimeout(timeoutFlag), insteon.ConnectionTTL(uint8(ttlFlag)))
			if err == insteon.ErrNotLinked {
				err = nil
			}

			if err == nil {
				err = isLinkable(device, func(ldevice insteon.Linkable) (err error) {
					err = util.ForceLink(group, lmodem, ldevice)
					if err == nil && crosslink {
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
				break
			}
		}
		return err
	})
}

func (p *plmCmd) allLinkCmd(string) error {
	return isLinkable(modem, func(linkable insteon.Linkable) error {
		return linkable.EnterLinkingMode(insteon.Group(0x01))
	})
}

func (p *plmCmd) unlinkCmd(string) (err error) {
	group := insteon.Group(0x01)

	return isLinkable(modem, func(lmodem insteon.Linkable) (err error) {
		for _, addr := range p.addresses {
			var device insteon.Device
			fmt.Printf("Unlinking from %s...", addr)
			device, err = modem.Open(addr, insteon.ConnectionTimeout(timeoutFlag), insteon.ConnectionTTL(uint8(ttlFlag)))

			if err == nil {
				err = isLinkable(device, func(ldevice insteon.Linkable) (err error) {
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
