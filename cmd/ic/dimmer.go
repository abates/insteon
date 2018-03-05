package main

import (
	"fmt"
	"strconv"

	"github.com/abates/cli"
	"github.com/abates/insteon"
)

var dimmer insteon.Dimmer

func init() {
	cmd := Commands.Register("dimmer", "<command> <device id>", "Interact with a specific dimmer", dimmerCmd)
	cmd.Register("on", "<level>", "turn the dimmer on", dimmerOnCmd)
	cmd.Register("off", "", "turn the dimmer off", switchOffCmd)
	cmd.Register("onfast", "<level>", "turn the dimmer on fast", dimmerOnFastCmd)
	cmd.Register("brighten", "", "brighten the dimmer one step", dimmerBrightenCmd)
	cmd.Register("dim", "", "dim the dimmer one step", dimmerDimCmd)
	cmd.Register("startBrighten", "", "", dimmerStartBrightenCmd)
	cmd.Register("startDim", "", "", dimmerStartDimCmd)
	cmd.Register("stopChange", "", "", dimmerStopChangeCmd)
	cmd.Register("instantChange", "<level>", "", dimmerInstantChangeCmd)
	cmd.Register("status", "", "get the switch status", switchStatusCmd)
	cmd.Register("setstatus", "<level>", "", dimmerSetStatusCmd)
	cmd.Register("onramp", "<level> <ramp>", "", dimmerOnRampCmd)
	cmd.Register("offramp", "<ramp>", "", dimmerOffRampCmd)
}

func dimmerCmd(args []string, next cli.NextFunc) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("device id and action must be specified")
	}

	addr, err := insteon.ParseAddress(args[0])
	if err != nil {
		return fmt.Errorf("invalid device address: %v", err)
	}

	device, err := devConnect(modem, addr)
	if closeable, ok := device.(insteon.Closeable); ok {
		defer closeable.Close()
	}
	if err == nil {
		var ok bool
		if dimmer, ok = device.(insteon.Dimmer); ok {
			sw = dimmer
			err = next()
		} else {
			err = fmt.Errorf("Device %s is not a dimmer", addr)
		}
	}
	return err
}

func dimmerOnCmd(args []string, next cli.NextFunc) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("no on level given")
	}
	level, err := strconv.Atoi(args[1])
	if err == nil {
		err = dimmer.OnLevel(level)
	}
	return err
}

func dimmerOnFastCmd(args []string, next cli.NextFunc) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("no on level given")
	}
	level, err := strconv.Atoi(args[1])
	if err == nil {
		err = dimmer.OnFast(level)
	}
	return err
}

func dimmerBrightenCmd(args []string, next cli.NextFunc) (err error) {
	return dimmer.Brighten()
}

func dimmerDimCmd(args []string, next cli.NextFunc) (err error) {
	return dimmer.Dim()
}

func dimmerStartBrightenCmd(args []string, next cli.NextFunc) (err error) {
	return dimmer.StartBrighten()
}

func dimmerStartDimCmd(args []string, next cli.NextFunc) (err error) {
	return dimmer.StartDim()
}

func dimmerStopChangeCmd(args []string, next cli.NextFunc) (err error) {
	return dimmer.StopChange()
}

func dimmerInstantChangeCmd(args []string, next cli.NextFunc) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("no on level given")
	}
	level, err := strconv.Atoi(args[1])
	if err == nil {
		err = dimmer.InstantChange(level)
	}
	return err
}

func dimmerSetStatusCmd(args []string, next cli.NextFunc) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("no on level given")
	}
	level, err := strconv.Atoi(args[1])
	if err == nil {
		err = dimmer.SetStatus(level)
	}
	return err
}

func dimmerOnRampCmd(args []string, next cli.NextFunc) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("no on level given")
	}

	if len(args) < 3 {
		return fmt.Errorf("no ramp rate given")
	}
	level, err := strconv.Atoi(args[1])
	if err == nil {
		var ramp int
		ramp, err = strconv.Atoi(args[2])
		if err == nil {
			err = dimmer.OnAtRamp(level, ramp)
		}
	}
	return err
}

func dimmerOffRampCmd(args []string, next cli.NextFunc) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("no ramp rate given")
	}
	ramp, err := strconv.Atoi(args[1])
	if err == nil {
		err = dimmer.OffAtRamp(ramp)
	}
	return err
}
