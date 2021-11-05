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
	"os"

	"github.com/abates/cli"
	"github.com/abates/insteon"
	"github.com/abates/insteon/util"
)

type linkableDevice interface {
	insteon.Device
	insteon.Linkable
}

type device struct {
	*insteon.BasicDevice
}

func init() {
	d := &device{
		BasicDevice: &insteon.BasicDevice{},
	}

	cmd := &cli.Command{
		Name:        "device",
		Usage:       "<device id> <command>",
		Description: "Interact with a specific device",
		Callback:    cli.Callback(d.init, "<device id>"),
		SubCommands: []*cli.Command{
			&cli.Command{Name: "info", Description: "retrieve device info", Callback: cli.Callback(d.infoCmd)},
			&cli.Command{Name: "link", Description: "enter linking mode", Callback: cli.Callback(d.EnterLinkingMode, "<group>")},
			&cli.Command{Name: "unlink", Description: "enter unlinking mode", Callback: cli.Callback(d.EnterUnlinkingMode, "<group>")},
			&cli.Command{Name: "exitlink", Description: "exit linking mode", Callback: cli.Callback(d.ExitLinkingMode)},
			&cli.Command{Name: "dump", Description: "dump the device all-link database", Callback: cli.Callback(d.dumpCmd)},
			&cli.Command{Name: "edit", Description: "edit the device all-link database", Callback: cli.Callback(d.editCmd)},
			&cli.Command{Name: "version", Description: "Retrieve the Insteon engine version", Callback: cli.Callback(d.versionCmd)},
			&cli.Command{Name: "send", Description: "send an arbitrary standard-direct command", Callback: cli.Callback(d.sendCmd, "<cmd1>.<cmd2>")},
			&cli.Command{Name: "esend", Description: "send an arbitrary extended-direct command", Callback: cli.Callback(d.esendCmd, "<cmd1>.<cmd2>", "<d1> <d2> ...")},
		},
	}
	app.SubCommands = append(app.SubCommands, cmd)
}

func (dev *device) init(address insteon.Address) error {
	device, err := open(modem, address)
	if err == nil {
		dev.BasicDevice.MessageWriter = device.MessageWriter
		dev.BasicDevice.DeviceInfo = device.DeviceInfo
	}
	return err
}

func (dev *device) dumpCmd() error {
	return util.DumpLinkDatabase(os.Stdout, dev)
}

func (dev *device) infoCmd() (err error) {
	return printDevInfo(dev, "")
}

func printDevInfo(device linkableDevice, extra string) (err error) {
	fmt.Printf("       Device: %v\n", device)
	if err == nil {
		fmt.Printf("       Engine: %v\n", device.Info().EngineVersion)
		fmt.Printf("     Category: %v\n", device.Info().DevCat)
		fmt.Printf("     Firmware: %v\n", device.Info().FirmwareVersion)

		if extra != "" {
			fmt.Printf("%s\n", extra)
		}

		err = util.PrintLinkDatabase(os.Stdout, device)
	}
	return err
}

func (dev *device) versionCmd() error {
	fmt.Printf("Device version: %s\n", dev.Info().FirmwareVersion)
	return nil
}

func (dev *device) editCmd(string) error {
	return editLinks(dev)
}

func editLinks(linkable insteon.Linkable) error {
	dbLinks, _ := linkable.Links()
	if len(dbLinks) == 0 {
		return fmt.Errorf("No links to edit")
	}

	inputLinksText := []byte(util.LinksToText(dbLinks))
	outputLinksText, err := cli.Edit(inputLinksText)
	if err == nil && !bytes.Equal(inputLinksText, outputLinksText) {
		dbLinks, err = util.TextToLinks(string(outputLinksText))
		if err == nil {
			err = linkable.WriteLinks(dbLinks...)
		}
	}
	return err
}

func (dev *device) esendCmd(cmd *cmdVar, data dataVar) error {
	return dev.SendCommand(cmd.Command, data)
}

func (dev *device) sendCmd(cmd cmdVar) error {
	return dev.SendCommand(cmd.Command, nil)
}
