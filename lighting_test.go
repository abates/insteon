package insteon

import (
	"bytes"
	"encoding"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestSwitchConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input             []byte
		expectedErr       error
		expectedHouseCode int
		expectedUnitCode  int
	}{
		{mkPayload(0, 0, 0, 0, 4, 5), nil, 4, 5},
		{nil, ErrBufferTooShort, 0, 0},
	}

	for i, test := range tests {
		config := &SwitchConfig{}
		err := config.UnmarshalBinary(test.input)
		if err != test.expectedErr {
			t.Errorf("tests[%d] expected %v got %v", err, test.expectedErr, err)
		} else if err == nil {
			if test.expectedHouseCode != config.HouseCode {
				t.Errorf("tests[%d] expected %d got %d", i, test.expectedHouseCode, config.HouseCode)
			}

			if test.expectedUnitCode != config.UnitCode {
				t.Errorf("tests[%d] expected %d got %d", i, test.expectedUnitCode, config.UnitCode)
			}

			buf, _ := config.MarshalBinary()
			if !bytes.Equal(test.input, buf) {
				t.Errorf("tests[%d] expected %v got %v", i, test.input, buf)
			}
		}
	}
}

func TestDimmerConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input             []byte
		expectedErr       error
		expectedHouseCode int
		expectedUnitCode  int
		expectedRamp      int
		expectedOnLevel   int
		expectedSNT       int
	}{
		{mkPayload(0, 0, 0, 0, 4, 5, 6, 7, 8), nil, 4, 5, 6, 7, 8},
		{nil, ErrBufferTooShort, 0, 0, 0, 0, 0},
	}

	for i, test := range tests {
		config := &DimmerConfig{}
		err := config.UnmarshalBinary(test.input)
		if err != test.expectedErr {
			t.Errorf("tests[%d] expected %v got %v", err, test.expectedErr, err)
		} else if err == nil {
			if test.expectedHouseCode != config.HouseCode {
				t.Errorf("tests[%d] expected %d got %d", i, test.expectedHouseCode, config.HouseCode)
			}

			if test.expectedUnitCode != config.UnitCode {
				t.Errorf("tests[%d] expected %d got %d", i, test.expectedUnitCode, config.UnitCode)
			}

			if test.expectedRamp != config.Ramp {
				t.Errorf("tests[%d] expected %d got %d", i, test.expectedRamp, config.Ramp)
			}

			if test.expectedOnLevel != config.OnLevel {
				t.Errorf("tests[%d] expected %d got %d", i, test.expectedOnLevel, config.OnLevel)
			}

			if test.expectedSNT != config.SNT {
				t.Errorf("tests[%d] expected %d got %d", i, test.expectedSNT, config.SNT)
			}

			buf, _ := config.MarshalBinary()
			if !bytes.Equal(test.input, buf) {
				t.Errorf("tests[%d] expected %v got %v", i, test.input, buf)
			}
		}
	}
}

func TestLightFlags(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    LightFlags
		test     func(flags LightFlags) bool
		expected bool
	}{
		{LightFlags{1, 0, 0, 0, 0}, func(flags LightFlags) bool { return flags.ProgramLock() }, true},
		{LightFlags{0, 0, 0, 0, 0}, func(flags LightFlags) bool { return flags.ProgramLock() }, false},
		{LightFlags{2, 0, 0, 0, 0}, func(flags LightFlags) bool { return flags.TxLED() }, true},
		{LightFlags{0, 0, 0, 0, 0}, func(flags LightFlags) bool { return flags.TxLED() }, false},
		{LightFlags{4, 0, 0, 0, 0}, func(flags LightFlags) bool { return flags.ResumeDim() }, true},
		{LightFlags{0, 0, 0, 0, 0}, func(flags LightFlags) bool { return flags.ResumeDim() }, false},
		{LightFlags{8, 0, 0, 0, 0}, func(flags LightFlags) bool { return flags.LED() }, true},
		{LightFlags{0, 0, 0, 0, 0}, func(flags LightFlags) bool { return flags.LED() }, false},
		{LightFlags{16, 0, 0, 0, 0}, func(flags LightFlags) bool { return flags.LoadSense() }, true},
		{LightFlags{0, 0, 0, 0, 0}, func(flags LightFlags) bool { return flags.LoadSense() }, false},
		{LightFlags{16, 22, 0, 0, 0}, func(flags LightFlags) bool { return flags.DBDelta() == 22 }, true},
		{LightFlags{16, 22, 42, 0, 0}, func(flags LightFlags) bool { return flags.SNR() == 42 }, true},
		{LightFlags{16, 22, 42, 31, 0}, func(flags LightFlags) bool { return flags.SNRFailCount() == 31 }, true},
		{LightFlags{16, 22, 42, 31, 0}, func(flags LightFlags) bool { return flags.X10Enabled() }, true},
		{LightFlags{16, 22, 42, 31, 1}, func(flags LightFlags) bool { return flags.X10Enabled() }, false},
		{LightFlags{16, 22, 42, 31, 2}, func(flags LightFlags) bool { return flags.ErrorBlink() }, true},
		{LightFlags{16, 22, 42, 31, 1}, func(flags LightFlags) bool { return flags.ErrorBlink() }, false},
		{LightFlags{16, 22, 42, 31, 4}, func(flags LightFlags) bool { return flags.CleanupReport() }, true},
		{LightFlags{16, 22, 42, 31, 1}, func(flags LightFlags) bool { return flags.CleanupReport() }, false},
	}

	for i, test := range tests {
		if test.test(test.input) != test.expected {
			t.Errorf("tests[%d] expected %v got %v", i, test.expected, test.test(test.input))
		}
	}
}

func TestSwitchIsASwitch(t *testing.T) {
	t.Parallel()
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

func TestSwitchProcess(t *testing.T) {
	t.Parallel()
	downstreamCh := make(chan *Message, 1)
	recvCh := make(chan *Message, 1)
	sd := &switchedDevice{downstreamRecvCh: downstreamCh, recvCh: recvCh}
	recvCh <- &Message{}
	if len(recvCh) != 1 {
		t.Errorf("expected 1 message in the queue got %v", len(recvCh))
	}
	close(recvCh)
	sd.process()

	if len(recvCh) != 0 {
		t.Errorf("expected empty queue got %v", len(recvCh))
	}

	if len(downstreamCh) != 1 {
		t.Errorf("expected 1 message in the downstream queue got %v", len(downstreamCh))
	}
}

func TestSwitchedDeviceFactory(t *testing.T) {
	t.Parallel()
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
			t.Error("expected stringer")
		}
	}
}

func TestSwitchCommands(t *testing.T) {
	t.Parallel()
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
		{1, func(sd *switchedDevice) error { return sd.On() }, CmdLightOn, nil},
		{1, func(sd *switchedDevice) error { return sd.Off() }, CmdLightOff, nil},
		{1, func(sd *switchedDevice) error { return extractError(sd.Status()) }, CmdLightStatusRequest, nil},
		{1, func(sd *switchedDevice) error { return sd.SetX10Address(7, 8, 9) }, CmdExtendedGetSet, []byte{7, 4, 8, 9}},
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

		if sender.sentCmds[0] != test.expectedCmd {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedCmd, sender.sentCmds[0])
		}

		if !bytes.Equal(test.expectedPayload, sender.sentPayloads[0]) {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedPayload, sender.sentPayloads[0])
		}
	}
}

func TestSwitchedDeviceConfig(t *testing.T) {
	t.Parallel()
	sender := &commandable{
		recvCmd: CmdExtendedGetSet,
	}
	sd := &switchedDevice{Commandable: sender}

	expected := SwitchConfig{31, 42}

	sender.recvPayloads = []encoding.BinaryMarshaler{&expected}

	config, err := sd.SwitchConfig()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if config != expected {
		t.Errorf("Expected %v got %v", expected, config)
	}
}

func TestSwitchedDeviceOperatingFlags(t *testing.T) {
	t.Parallel()
	sender := &commandable{
		respCmds: []Command{
			CmdGetOperatingFlags.SubCommand(3),
			CmdGetOperatingFlags.SubCommand(4),
			CmdGetOperatingFlags.SubCommand(5),
			CmdGetOperatingFlags.SubCommand(6),
			CmdGetOperatingFlags.SubCommand(7),
		},
	}

	sd := &switchedDevice{Commandable: sender}

	expected := LightFlags{3, 4, 5, 6, 7}
	flags, _ := sd.OperatingFlags()

	if flags != expected {
		t.Errorf("expected %v got %v", expected, flags)
	}
}

func TestDimmerIsADimmer(t *testing.T) {
	t.Parallel()
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

func TestDimmerProcess(t *testing.T) {
	t.Parallel()
	downstreamCh := make(chan *Message, 1)
	recvCh := make(chan *Message, 1)
	dd := &dimmableDevice{downstreamRecvCh: downstreamCh, recvCh: recvCh}
	recvCh <- &Message{}
	if len(recvCh) != 1 {
		t.Errorf("expected 1 message in the queue got %v", len(recvCh))
	}
	close(recvCh)
	dd.process()

	if len(recvCh) != 0 {
		t.Errorf("expected empty queue got %v", len(recvCh))
	}

	if len(downstreamCh) != 1 {
		t.Errorf("expected 1 message in the downstream queue got %v", len(downstreamCh))
	}
}

func TestDimmerCommands(t *testing.T) {
	t.Parallel()
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

		if sender.sentCmds[0] != test.expectedCmd {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedCmd, sender.sentCmds[0])
		}

		if !bytes.Equal(test.expectedPayload, sender.sentPayloads[0]) {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedPayload, sender.sentPayloads[0])
		}
	}
}

func TestDimmableDeviceConfig(t *testing.T) {
	t.Parallel()
	sender := &commandable{
		recvCmd: CmdExtendedGetSet,
	}
	dd := &dimmableDevice{Commandable: sender}

	expected := DimmerConfig{31, 42, 15, 27, 4}

	sender.recvPayloads = []encoding.BinaryMarshaler{&expected}

	config, err := dd.DimmerConfig()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if config != expected {
		t.Errorf("Expected %v got %v", expected, config)
	}
}

func TestDimmableDeviceFactory(t *testing.T) {
	t.Parallel()
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
			t.Error("expected stringer")
		}
	}
}
