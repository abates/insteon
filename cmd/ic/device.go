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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/abates/cli"
	"github.com/abates/insteon"
	"github.com/abates/insteon/plm"
	"github.com/abates/insteon/util"
)

func init() {
	d := &device{}
	cmd := app.SubCommand("device", cli.UsageOption("<device id> <command>"), cli.DescOption("Interact with a specific device"), cli.CallbackOption(d.init))
	cmd.Arguments.Var(&d.addr, "<device id>")
	cmd.SubCommand("info", cli.DescOption("retrieve device info"), cli.CallbackOption(d.infoCmd))
	cmd.SubCommand("link", cli.DescOption("enter linking mode"), cli.CallbackOption(d.linkCmd))
	cmd.SubCommand("unlink", cli.DescOption("enter unlinking mode"), cli.CallbackOption(d.unlinkCmd))
	cmd.SubCommand("exitlink", cli.DescOption("exit linking mode"), cli.CallbackOption(d.exitLinkCmd))
	cmd.SubCommand("dump", cli.DescOption("dump the device all-link database"), cli.CallbackOption(d.dumpCmd))
	cmd.SubCommand("edit", cli.DescOption("edit the device all-link database"), cli.CallbackOption(d.editCmd))
	cmd.SubCommand("version", cli.UsageOption("<device id>"), cli.DescOption("Retrieve the Insteon engine version"), cli.CallbackOption(d.versionCmd))
	snd := cmd.SubCommand("send", cli.UsageOption("<cmd1>.<cmd2>"), cli.DescOption("send a standard-direct command"), cli.CallbackOption(d.sendCmd))
	snd.Arguments.Var(&d.cmd, "<cmd1>.<cmd2>")
	esnd := cmd.SubCommand("esend", cli.UsageOption("<cmd1>.<cmd2> <d1> <d2> ..."), cli.DescOption("send a extended-direct command"), cli.CallbackOption(d.sendCmd))
	esnd.Arguments.Var(&d.cmd, "<cmd1>.<cmd2>")
	esnd.Arguments.VarSlice(&d.data, "<d1> <d2> ...")
}

type cmd struct {
	insteon.Command
}

// Set satisfies the flag.Value interface
func (cmd *cmd) Set(str string) error {
	// Support non-period separated input too.
	if len(str) == 4 {
		str = strings.Join([]string{str[0:2], str[2:4]}, ".")
	}

	if len(str) != 5 {
		return fmt.Errorf("Bad command format need xx.xx or xxxx where xx represents a valid hex value.  Got: %v", str)
	}
	var c1, c2 byte
	_, err := fmt.Sscanf(str, "%2x.%2x", &c1, &c2)
	if err != nil {
		return fmt.Errorf("Bad command format need xx.xx or xxxx where xx represents a valid hex value.  Got: %v", str)
	}
	cmd.Command[1] = c1
	cmd.Command[2] = c2
	return nil
}

type data []byte

func (d *data) Set(str string) error {
	var b byte
	_, err := fmt.Sscanf(str, "%x", &b)
	if err == nil {
		*d = append(*d, b)
	}
	return err
}

func (d *data) String() string { return fmt.Sprintf("%v", *d) }

type device struct {
	insteon.Device
	addr insteon.Address
	cmd  cmd
	data data
}

func (dev *device) init() (err error) {
	dev.Device, err = connect(modem, dev.addr)
	return err
}

func connect(plm *plm.PLM, addr insteon.Address) (insteon.Device, error) {
	device, err := plm.Open(addr, insteon.ConnectionTimeout(timeoutFlag), insteon.ConnectionTTL(uint8(ttlFlag)))

	if err == insteon.ErrNotLinked {
		msg := fmt.Sprintf("Device %s is not linked to the PLM.  Link now? (y/n) ", addr)
		if cli.Query(os.Stdin, os.Stdout, msg, "y", "n") == "y" {
			pc := &plmCmd{addresses: []insteon.Address{addr}}
			err = pc.linkCmd()
		}

		if err == nil {
			device, err = plm.Open(addr, insteon.ConnectionTimeout(timeoutFlag), insteon.ConnectionTTL(uint8(ttlFlag)))
		}
	}
	return device, err
}

func isLinkable(thing interface{}, cb func(linkable insteon.Linkable) error) error {
	if linkable, ok := thing.(insteon.Linkable); ok {
		return cb(linkable)
	}
	return fmt.Errorf("%v is not linkable", thing)
}

func (dev *device) linkCmd() error {
	return isLinkable(dev.Device, func(linkable insteon.Linkable) error {
		return linkable.EnterLinkingMode(insteon.Group(0x01))
	})
}

func (dev *device) unlinkCmd() error {
	return isLinkable(dev.Device, func(linkable insteon.Linkable) error {
		return linkable.EnterUnlinkingMode(insteon.Group(0x01))
	})
}

func (dev *device) exitLinkCmd() error {
	return isLinkable(dev.Device, func(linkable insteon.Linkable) error {
		return linkable.ExitLinkingMode()
	})
}

func (dev *device) dumpCmd() error {
	return isLinkable(dev.Device, func(linkable insteon.Linkable) error {
		err := dumpLinkDatabase(linkable)
		return err
	})
}

func (dev *device) infoCmd() (err error) {
	return printDevInfo(dev.Device, "")
}

func printDevInfo(device insteon.Device, extra string) (err error) {
	fmt.Printf("       Device: %v\n", device)
	firmware, devCat, err := device.IDRequest()

	if err == nil {
		fmt.Printf("     Category: %v\n", devCat)
		fmt.Printf("     Firmware: %v\n", firmware)

		if extra != "" {
			fmt.Printf("%s\n", extra)
		}

		err = isLinkable(device, func(linkable insteon.Linkable) error {
			return util.PrintLinks(os.Stdout, linkable)
		})
	}
	return err
}

func (dev *device) versionCmd() error {
	firmware, _, err := dev.IDRequest()
	if err == nil {
		fmt.Printf("Device version: %s\n", firmware)
	}
	return err
}

func (dev *device) editCmd() error {
	return isLinkable(dev.Device, func(linkable insteon.Linkable) error {
		dbLinks, _ := linkable.Links()
		if len(dbLinks) == 0 {
			return fmt.Errorf("No links to edit")
		}

		tmpfile, err := ioutil.TempFile("", "insteon_")
		if err != nil {
			return err
		}
		defer os.Remove(tmpfile.Name())

		buf := bytes.NewBuffer(nil)
		fmt.Fprintf(buf, "#\n")
		fmt.Fprintf(buf, "# Lines beginning with a # are ignored\n")
		fmt.Fprintf(buf, "# DO NOT delete lines, this will cause the entries to\n")
		fmt.Fprintf(buf, "# shift up and then the last entry will be in the database twice\n")
		fmt.Fprintf(buf, "# To delete a record simply mark it 'Available' by changing the\n")
		fmt.Fprintf(buf, "# first letter of the Flags to 'A'\n")
		fmt.Fprintf(buf, "#\n")
		fmt.Fprintf(buf, "# Flags Group Address    Data\n")
		for _, link := range dbLinks {
			output, _ := link.MarshalText()
			fmt.Fprintf(buf, "  %s\n", string(output))
		}

		tmpfile.Write(buf.Bytes())

		if err = tmpfile.Close(); err == nil {
			cmd := exec.Command(editor, tmpfile.Name())
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Start()
			err = cmd.Wait()
			if err == nil {
				dbLinks = nil
				input, err := ioutil.ReadFile(tmpfile.Name())
				if err == nil && !bytes.Equal(buf.Bytes(), input) {
					for _, line := range bytes.Split(input, []byte("\n")) {
						line = bytes.TrimSpace(line)
						if len(line) == 0 || bytes.Index(line, []byte("#")) == 0 {
							continue
						}
						link := &insteon.LinkRecord{}
						err = link.UnmarshalText(line)
						if err == nil {
							dbLinks = append(dbLinks, link)
						} else {
							fmt.Printf("Skipping invalid line %q: %v\n", string(line), err)
						}
					}
					linkable.WriteLinks(dbLinks...)
				}
			}
		}
		return err
	})
}

func (dev *device) sendCmd() error {
	_, err := dev.SendCommand(dev.cmd.Command, dev.data)
	return err
}
