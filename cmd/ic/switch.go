package main

import (
	"fmt"

	"github.com/abates/cli"
	"github.com/abates/insteon"
)

var sw insteon.Switch

func init() {
	cmd := Commands.Register("switch", "<command> <device id>", "Interact with a specific switch", swCmd)
	cmd.Register("config", "", "retrieve switch configuration information", switchConfigCmd)
	cmd.Register("on", "", "turn the switch/light on", switchOnCmd)
	cmd.Register("off", "", "turn the switch/light off", switchOffCmd)
	cmd.Register("status", "", "get the switch status", switchStatusCmd)
}

func swCmd(args []string, next cli.NextFunc) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("device id and action must be specified")
	}

	var addr insteon.Address
	err = addr.UnmarshalText([]byte(args[0]))
	if err != nil {
		return fmt.Errorf("invalid device address: %v", err)
	}

	device, err := devConnect(network, addr)
	if err == nil {
		var ok bool
		if sw, ok = device.(insteon.Switch); ok {
			err = next()
		} else {
			err = fmt.Errorf("Device at %s is a %T not a switch", addr, device)
		}
	}
	return err
}

func switchConfigCmd([]string, cli.NextFunc) error {
	config, err := sw.SwitchConfig()
	if err == nil {
		err = printDevInfo(sw.(insteon.Device), fmt.Sprintf("  X10 Address: %02x.%02x", config.HouseCode, config.UnitCode))
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
