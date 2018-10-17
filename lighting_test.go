package insteon

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func foo() Switch {
	return &i1SwitchedDevice{}
}

func TestSwitch(t *testing.T) {
	tests := []struct {
		device interface{}
	}{
		{&i1SwitchedDevice{}},
		{&i2SwitchedDevice{}},
		{&i2CsSwitchedDevice{}},
	}

	for i, test := range tests {
		if _, ok := test.device.(Switch); !ok {
			t.Errorf("tests[%d] expected Switch got %T", i, test.device)
		}
	}
}

func TestDimmer(t *testing.T) {
	tests := []struct {
		device interface{}
	}{
		{&i1DimmableDevice{}},
		{&i2DimmableDevice{}},
		{&i2DimmableDevice{}},
	}

	for i, test := range tests {
		if _, ok := test.device.(Dimmer); !ok {
			t.Errorf("tests[%d] expected Dimmer got %T", i, test.device)
		}
	}
}

func TestSwitchedDeviceFactory(t *testing.T) {
	tests := []struct {
		info     DeviceInfo
		expected interface{}
	}{
		{DeviceInfo{EngineVersion: 0}, &i1SwitchedDevice{}},
		{DeviceInfo{EngineVersion: 1}, &i2SwitchedDevice{}},
		{DeviceInfo{EngineVersion: 2}, &i2CsSwitchedDevice{}},
	}

	for i, test := range tests {
		device, _ := switchedDeviceFactory(test.info, Address{5, 6, 7}, nil, nil, time.Millisecond)
		if reflect.TypeOf(device) != reflect.TypeOf(test.expected) {
			t.Errorf("tests[%d] expected %T got %T", i, test.expected, device)
		}

		if stringer, ok := device.(fmt.Stringer); ok {
			if stringer.String() != "Switch (05.06.07)" {
				t.Errorf("expected %q got %q", "Switch (05.06.07)", stringer.String())
			}
		} else {
			t.Errorf("expected stringer")
		}
	}
}

func TestSwitchCommands(t *testing.T) {
	tests := []struct {
		firmwareVersion FirmwareVersion
		callback        func(*switchedDevice) error
		expectedCmd     Command
		expectedPayload []byte
	}{
		{1, func(sd *switchedDevice) error { return sd.SetProgramLock(true) }, CmdSetOperatingFlags.SubCommand(0), nil},
		{1, func(sd *switchedDevice) error { return sd.SetProgramLock(false) }, CmdSetOperatingFlags.SubCommand(1), nil},
		{1, func(sd *switchedDevice) error { return sd.SetTxLED(true) }, CmdSetOperatingFlags.SubCommand(2), nil},
		{1, func(sd *switchedDevice) error { return sd.SetTxLED(false) }, CmdSetOperatingFlags.SubCommand(3), nil},
		{1, func(sd *switchedDevice) error { return sd.SetResumeDim(true) }, CmdSetOperatingFlags.SubCommand(4), nil},
		{1, func(sd *switchedDevice) error { return sd.SetResumeDim(false) }, CmdSetOperatingFlags.SubCommand(5), nil},
		{1, func(sd *switchedDevice) error { return sd.SetLoadSense(false) }, CmdSetOperatingFlags.SubCommand(6), nil},
		{1, func(sd *switchedDevice) error { return sd.SetLoadSense(true) }, CmdSetOperatingFlags.SubCommand(7), nil},
		{1, func(sd *switchedDevice) error { return sd.SetLED(false) }, CmdSetOperatingFlags.SubCommand(8), nil},
		{1, func(sd *switchedDevice) error { return sd.SetLED(true) }, CmdSetOperatingFlags.SubCommand(9), nil},
	}

	for i, test := range tests {
		sender := &commandable{}
		sd := &switchedDevice{
			Commandable:     sender,
			firmwareVersion: test.firmwareVersion,
		}

		err := test.callback(sd)
		if err != nil {
			t.Errorf("tests[%d] expected nil error got %v", i, err)
		}

		if sender.sentCmd != test.expectedCmd {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedCmd, sender.sentCmd)
		}

		if !bytes.Equal(test.expectedPayload, sender.sentPayload) {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedPayload, sender.sentPayload)
		}
	}
}

func TestDimmerCommands(t *testing.T) {
	tests := []struct {
		firmwareVersion FirmwareVersion
		callback        func(*dimmableDevice) error
		expectedCmd     Command
		expectedPayload []byte
	}{
		{1, func(dd *dimmableDevice) error { return dd.OnLevel(10) }, CmdLightOn.SubCommand(10), nil},
		{1, func(dd *dimmableDevice) error { return dd.OnFast(10) }, CmdLightOnFast.SubCommand(10), nil},
		{1, func(dd *dimmableDevice) error { return dd.Brighten() }, CmdLightBrighten, nil},
		{1, func(dd *dimmableDevice) error { return dd.Dim() }, CmdLightDim, nil},
		{1, func(dd *dimmableDevice) error { return dd.StartBrighten() }, CmdLightStartManual.SubCommand(1), nil},
		{1, func(dd *dimmableDevice) error { return dd.StartDim() }, CmdLightStartManual.SubCommand(0), nil},
		{1, func(dd *dimmableDevice) error { return dd.StopChange() }, CmdLightStopManual, nil},
		{1, func(dd *dimmableDevice) error { return dd.InstantChange(15) }, CmdLightInstantChange.SubCommand(15), nil},
		{1, func(dd *dimmableDevice) error { return dd.SetStatus(31) }, CmdLightSetStatus.SubCommand(31), nil},
		{1, func(dd *dimmableDevice) error { return dd.OnAtRamp(0x03, 0x07) }, CmdLightOnAtRamp.SubCommand(0x37), nil},
		{67, func(dd *dimmableDevice) error { return dd.OnAtRamp(0x03, 0x07) }, CmdLightOnAtRampV67.SubCommand(0x37), nil},
		{1, func(dd *dimmableDevice) error { return dd.OffAtRamp(0xfa) }, CmdLightOffAtRamp.SubCommand(0x0a), nil},
		{67, func(dd *dimmableDevice) error { return dd.OffAtRamp(0xfa) }, CmdLightOffAtRampV67.SubCommand(0x0a), nil},
		{1, func(dd *dimmableDevice) error { return dd.SetDefaultRamp(3) }, CmdExtendedGetSet, []byte{1, 5, 3}},
		{1, func(dd *dimmableDevice) error { return dd.SetDefaultOnLevel(7) }, CmdExtendedGetSet, []byte{1, 6, 7}},
	}

	for i, test := range tests {
		sender := &commandable{}
		dd := &dimmableDevice{
			Commandable:     sender,
			firmwareVersion: test.firmwareVersion,
		}

		err := test.callback(dd)
		if err != nil {
			t.Errorf("tests[%d] expected nil error got %v", i, err)
		}

		if sender.sentCmd != test.expectedCmd {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedCmd, sender.sentCmd)
		}

		if !bytes.Equal(test.expectedPayload, sender.sentPayload) {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedPayload, sender.sentPayload)
		}
	}
}

func TestDimmableDeviceFactory(t *testing.T) {
	tests := []struct {
		info     DeviceInfo
		expected interface{}
	}{
		{DeviceInfo{EngineVersion: 0}, &i1DimmableDevice{}},
		{DeviceInfo{EngineVersion: 1}, &i2DimmableDevice{}},
		{DeviceInfo{EngineVersion: 2}, &i2CsDimmableDevice{}},
	}

	for i, test := range tests {
		device, _ := dimmableDeviceFactory(test.info, Address{3, 4, 5}, nil, nil, time.Millisecond)
		if reflect.TypeOf(device) != reflect.TypeOf(test.expected) {
			t.Errorf("tests[%d] expected %T got %T", i, test.expected, device)
		}

		if stringer, ok := device.(fmt.Stringer); ok {
			if stringer.String() != "Dimmable Light (03.04.05)" {
				t.Errorf("expected %q got %q", "Dimmable Light (03.04.05)", stringer.String())
			}
		} else {
			t.Errorf("expected stringer")
		}
	}
}
