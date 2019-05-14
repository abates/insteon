package insteon

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

func TestDimmerFactory(t *testing.T) {
	tests := []struct {
		desc  string
		input Switch
		want  reflect.Type
	}{
		{"Dimmer", &switchedDevice{}, reflect.TypeOf(&dimmer{})},
		{"Linkable Dimmer", &linkableSwitch{}, reflect.TypeOf(&linkableDimmer{})},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := reflect.TypeOf(NewDimmer(test.input, 0, 0))
			if test.want != got {
				t.Errorf("want type %v got %v", test.want, got)
			}
		})
	}
}

func TestDimmerConfig(t *testing.T) {
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

func TestDimmerCommands(t *testing.T) {
	tests := []*commandTest{
		{"OnLevel", func(d Device) error { return d.(Dimmer).OnLevel(10) }, CmdLightOn.SubCommand(10), nil, nil},
		{"OnFast", func(d Device) error { return d.(Dimmer).OnFast(10) }, CmdLightOnFast.SubCommand(10), nil, nil},
		{"Brighten", func(d Device) error { return d.(Dimmer).Brighten() }, CmdLightBrighten, nil, nil},
		{"Dim", func(d Device) error { return d.(Dimmer).Dim() }, CmdLightDim, nil, nil},
		{"StartBrighten", func(d Device) error { return d.(Dimmer).StartBrighten() }, CmdLightStartManual.SubCommand(1), nil, nil},
		{"StartTime", func(d Device) error { return d.(Dimmer).StartDim() }, CmdLightStartManual.SubCommand(0), nil, nil},
		{"StopChange", func(d Device) error { return d.(Dimmer).StopChange() }, CmdLightStopManual, nil, nil},
		{"InstantChange", func(d Device) error { return d.(Dimmer).InstantChange(15) }, CmdLightInstantChange.SubCommand(15), nil, nil},
		{"SetStatus", func(d Device) error { return d.(Dimmer).SetStatus(31) }, CmdLightSetStatus.SubCommand(31), nil, nil},
		{"OnAtRamp", func(d Device) error { return d.(Dimmer).OnAtRamp(0x03, 0x07) }, CmdLightOnAtRamp.SubCommand(0x37), nil, nil},
		{"OffAtRamp", func(d Device) error { return d.(Dimmer).OffAtRamp(0xfa) }, CmdLightOffAtRamp.SubCommand(0x0a), nil, nil},
		{"SetDefaultRamp", func(d Device) error { return d.(Dimmer).SetDefaultRamp(3) }, CmdExtendedGetSet, nil, []byte{1, 5, 3}},
		{"SetDefaultOnLevel", func(d Device) error { return d.(Dimmer).SetDefaultOnLevel(7) }, CmdExtendedGetSet, nil, []byte{1, 6, 7}},
	}

	testDeviceCommands(t, func(conn *testConnection) Device {
		return NewDimmer(NewSwitch(conn, time.Nanosecond), time.Nanosecond, 0)
	}, tests)

	tests = []*commandTest{
		{"OnAtRamp v67", func(d Device) error { return d.(Dimmer).OnAtRamp(0x03, 0x07) }, CmdLightOnAtRampV67.SubCommand(0x37), nil, nil},
		{"OffAtRamp v67", func(d Device) error { return d.(Dimmer).OffAtRamp(0xfa) }, CmdLightOffAtRampV67.SubCommand(0x0a), nil, nil},
	}
	testDeviceCommands(t, func(conn *testConnection) Device {
		return NewDimmer(NewSwitch(conn, time.Nanosecond), time.Nanosecond, 67)
	}, tests)

}

func TestDimmableDeviceConfig(t *testing.T) {
	conn := &testConnection{recvCh: make(chan *Message, 1), sendCh: make(chan *Message, 1), ackCh: make(chan *Message, 1)}
	dd := NewDimmer(NewSwitch(conn, time.Millisecond), time.Millisecond, 67)
	want := DimmerConfig{31, 42, 15, 27, 4}
	payload, _ := want.MarshalBinary()
	msg := &Message{Command: CmdExtendedGetSet, Payload: make([]byte, 14)}
	copy(msg.Payload, payload)
	conn.recvCh <- msg
	conn.ackCh <- TestAck

	got, err := dd.DimmerConfig()
	<-conn.sendCh
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if got != want {
		t.Errorf("Want config %v got %v", want, got)
	}
}
