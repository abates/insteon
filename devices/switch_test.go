package devices

import (
	"bytes"
	"testing"

	"github.com/abates/insteon"
	"github.com/abates/insteon/commands"
)

func TestSwitchConfig(t *testing.T) {
	tests := []struct {
		input             []byte
		expectedErr       error
		expectedHouseCode int
		expectedUnitCode  int
	}{
		{mkPayload(0, 0, 0, 0, 4, 5), nil, 4, 5},
		{nil, insteon.ErrBufferTooShort, 0, 0},
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

func TestLightFlags(t *testing.T) {
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
		{LightFlags{0, 0, 0, 16, 0}, func(flags LightFlags) bool { return flags.LED() }, true},
		{LightFlags{0, 0, 0, 0, 0}, func(flags LightFlags) bool { return flags.LED() }, false},
		{LightFlags{0, 0, 0, 0, 32}, func(flags LightFlags) bool { return flags.LoadSense() }, true},
		{LightFlags{0, 0, 0, 0, 0}, func(flags LightFlags) bool { return flags.LoadSense() }, false},
		{LightFlags{16, 22, 0, 0, 0}, func(flags LightFlags) bool { return flags.DBDelta() == 22 }, true},
		{LightFlags{16, 22, 42, 0, 0}, func(flags LightFlags) bool { return flags.SNR() == 42 }, true},
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

func TestSwitchedDeviceConfig(t *testing.T) {
	want := SwitchConfig{31, 42}
	payload, _ := want.MarshalBinary()
	msg := &insteon.Message{Command: commands.ExtendedGetSet, Payload: make([]byte, 14)}
	copy(msg.Payload, payload)

	tw := &testWriter{
		read: []*insteon.Message{msg},
	}
	sd := NewSwitch(&BasicDevice{MessageWriter: tw, DeviceInfo: DeviceInfo{}})

	got, err := sd.Config()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if got != want {
		t.Errorf("Want config %v got %v", want, got)
	}
}

func TestSwitchedDeviceOperatingFlags(t *testing.T) {
	cmds := []commands.Command{
		commands.GetOperatingFlags.SubCommand(3),
		commands.GetOperatingFlags.SubCommand(4),
		commands.GetOperatingFlags.SubCommand(5),
		commands.GetOperatingFlags.SubCommand(6),
		commands.GetOperatingFlags.SubCommand(7),
	}
	tw := &testWriter{}
	for _, cmd := range cmds {
		tw.acks = append(tw.acks, &insteon.Message{Command: cmd})
	}

	sd := NewSwitch(&BasicDevice{MessageWriter: tw, DeviceInfo: DeviceInfo{}})
	want := LightFlags{3, 4, 5, 6, 7}
	got, _ := sd.OperatingFlags()

	if want != got {
		t.Errorf("want flags %v got %v", want, got)
	}
}

func TestSwitchStatus(t *testing.T) {
	want := 43
	tw := &testWriter{
		acks: []*insteon.Message{{Command: commands.LightStatusRequest.SubCommand(want)}},
	}
	sw := &Switch{BasicDevice: &BasicDevice{MessageWriter: tw}}
	got, _ := sw.Status()
	if want != got {
		t.Errorf("Wanted level %d got %d", want, got)
	}
}
