package main

import (
	"fmt"

	"github.com/abates/cli"
	"github.com/abates/insteon"
)

var sw insteon.Switch

func init() {
	cmd := Commands.Register("switch", "<command> <device id>", "Interact with a specific switch", swCmd)
	cmd.Register("on", "", "turn the switch/light on", switchOnCmd)
	cmd.Register("off", "", "turn the switch/light off", switchOffCmd)
	cmd.Register("status", "", "get the switch status", switchStatusCmd)
}

func swCmd(args []string, next cli.NextFunc) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("device id and action must be specified")
	}

	addr, err := insteon.ParseAddress(args[0])
	if err != nil {
		return fmt.Errorf("invalid device address: %v", err)
	}

	device, err := devConnect(modem, addr)
	defer device.Connection().Close()
	if err == nil {
		var ok bool
		if sw, ok = device.(insteon.Switch); ok {
			err = next()
		} else {
			err = fmt.Errorf("Device %s is not a switch", addr)
		}
	}
	return err
}

func switchOnCmd([]string, cli.NextFunc) error {
	return sw.On()
}

func switchOffCmd([]string, cli.NextFunc) error {
	return sw.Off()
}

func switchStatusCmd([]string, cli.NextFunc) error {
	level, err := sw.Status()
	if err == nil {
		if level == 0 {
			fmt.Printf("Switch is off\n")
		} else if level == 255 {
			fmt.Printf("Switch is on\n")
		} else {
			fmt.Printf("Switch is on at level %d\n", level)
		}
	}
	return err
}
