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
	cmd.Register("edit", "", "edit the device all-link database", devEditCmd)
	cmd.Register("version", "<device id>", "Retrieve the Insteon engine version", devVersionCmd)
}

func devCmd(args []string, next cli.NextFunc) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("device id and action must be specified")
	}

	var addr insteon.Address
	err = addr.UnmarshalText([]byte(args[0]))
	if err != nil {
		return fmt.Errorf("invalid device address: %v", err)
	}

	device, err = devConnect(modem, addr)
	if err == nil {
		err = next()
	}
	return err
}

func devConnect(plm *plm.PLM, addr insteon.Address) (insteon.Device, error) {
	device, err := plm.Open(addr)
	if err == insteon.ErrNotLinked {
		msg := fmt.Sprintf("Device %s is not linked to the PLM.  Link now? (y/n) ", addr)
		if getResponse(msg, "y", "n") == "y" {
			err = plmLinkCmd([]string{addr.String()}, nil)
		}

		if err == nil {
			device, err = plm.Open(addr)
		}
	}
	return device, err
}

func devLink(cb func(linkable insteon.LinkableDevice) error) error {
	if linkable, ok := device.(insteon.LinkableDevice); ok {
		return cb(linkable)
	}
	return fmt.Errorf("%v is not a linkable device", device)
}

func devLinkCmd([]string, cli.NextFunc) error {
	return devLink(func(linkable insteon.LinkableDevice) error {
		return linkable.EnterLinkingMode(insteon.Group(0x01))
	})
}

func devUnlinkCmd([]string, cli.NextFunc) error {
	return devLink(func(linkable insteon.LinkableDevice) error {
		return linkable.EnterUnlinkingMode(insteon.Group(0x01))
	})
}

func devExitLinkCmd([]string, cli.NextFunc) error {
	return devLink(func(linkable insteon.LinkableDevice) error {
		return linkable.ExitLinkingMode()
	})
}

func devDumpCmd([]string, cli.NextFunc) error {
	return devLink(func(linkable insteon.LinkableDevice) error {
		err := dumpLinkDatabase(linkable)
		return err
	})
}

func devInfoCmd([]string, cli.NextFunc) (err error) {
	return printDevInfo(device, "")
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

		err = devLink(func(linkable insteon.LinkableDevice) error {
			return printLinkDatabase(linkable)
		})
	}
	return err
}

func devVersionCmd([]string, cli.NextFunc) error {
	firmware, _, err := device.IDRequest()
	if err == nil {
		fmt.Printf("Device version: %s\n", firmware)
	}
	return err
}

func devEditCmd([]string, cli.NextFunc) error {
	return devLink(func(linkable insteon.LinkableDevice) error {
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
			cmd := exec.Command(EDITOR, tmpfile.Name())
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Start()
			err = cmd.Wait()
			if err == nil {
				input, err := ioutil.ReadFile(tmpfile.Name())
				if err == nil && !bytes.Equal(buf.Bytes(), input) {
					i := 0
					for _, line := range bytes.Split(input, []byte("\n")) {
						line = bytes.TrimSpace(line)
						if len(line) == 0 || bytes.Index(line, []byte("#")) == 0 {
							continue
						}
						if i < len(dbLinks) {
							err = dbLinks[i].UnmarshalText(line)
							if err == nil {
								fmt.Printf("Writing %s...", dbLinks[i])
								err = linkable.WriteLink(dbLinks[i])
								if err == nil {
									fmt.Printf("done\n")
								} else {
									fmt.Printf("%v\n", err)
								}
							} else {
								fmt.Printf("Skipping invalid line %q: %v\n", string(line), err)
							}
						} else {
							link := &insteon.LinkRecord{}
							err = link.UnmarshalText(line)
							if err == nil {
								fmt.Printf("Adding line %q...", string(line))
								err = linkable.AppendLink(link)
								if err == nil {
									fmt.Printf("done\n")
								} else {
									fmt.Printf("%v\n", err)
								}
							}
						}
						i++
					}
				}
			}
		}
		return err
	})
}

func devCleanupCmd([]string, cli.NextFunc) error {
	return nil
}
