package insteon

import (
	"bytes"
	"testing"
	"time"
)

func TestSwitchConfig(t *testing.T) {
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

func TestSwitchedDeviceConfig(t *testing.T) {
	want := SwitchConfig{31, 42}
	payload, _ := want.MarshalBinary()
	msg := &Message{Command: CmdExtendedGetSet, Payload: make([]byte, 14)}
	copy(msg.Payload, payload)

	conn := &testConnection{recv: []*Message{msg}, acks: []*Message{TestAck}}
	sd := NewSwitch(DeviceInfo{}, conn, time.Millisecond)

	got, err := sd.SwitchConfig()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if got != want {
		t.Errorf("Want config %v got %v", want, got)
	}
}

func TestSwitchedDeviceOperatingFlags(t *testing.T) {
	cmds := []Command{
		CmdGetOperatingFlags.SubCommand(3),
		CmdGetOperatingFlags.SubCommand(4),
		CmdGetOperatingFlags.SubCommand(5),
		CmdGetOperatingFlags.SubCommand(6),
		CmdGetOperatingFlags.SubCommand(7),
	}
	conn := &testConnection{}
	for _, cmd := range cmds {
		conn.acks = append(conn.acks, &Message{Command: cmd})
	}

	sd := NewSwitch(DeviceInfo{}, conn, time.Nanosecond)
	want := LightFlags{3, 4, 5, 6, 7}
	got, _ := sd.OperatingFlags()

	if want != got {
		t.Errorf("want flags %v got %v", want, got)
	}
}
