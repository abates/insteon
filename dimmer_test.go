package insteon

import (
	"bytes"
	"testing"
	"time"
)

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

func TestDimmableDeviceConfig(t *testing.T) {
	want := DimmerConfig{31, 42, 15, 27, 4}
	payload, _ := want.MarshalBinary()
	msg := &Message{Command: CmdExtendedGetSet, Payload: make([]byte, 14)}
	copy(msg.Payload, payload)

	conn := &testConnection{recv: []*Message{msg}, acks: []*Message{TestAck}}
	dd := NewDimmer(DeviceInfo{FirmwareVersion: 67}, conn, time.Millisecond)
	got, err := dd.DimmerConfig()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if got != want {
		t.Errorf("Want config %v got %v", want, got)
	}
}

func TestDimmerSendCommand(t *testing.T) {
	tests := []struct {
		name    string
		v       FirmwareVersion
		sendCmd Command
		wantCmd Command
	}{
		{"On", 0x01, CmdLightOnAtRamp, CmdLightOnAtRamp},
		{"On (v67)", 0x43, CmdLightOnAtRamp, CmdLightOnAtRampV67},
		{"Off", 0x01, CmdLightOffAtRamp, CmdLightOffAtRamp},
		{"Off (v67)", 0x43, CmdLightOffAtRamp, CmdLightOffAtRampV67},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conn := &testConnection{acks: []*Message{TestAck}}
			dimmer := &Dimmer{Switch: &Switch{Device: conn}, firmwareVersion: test.v}
			dimmer.SendCommand(test.sendCmd, nil)
			gotCmd := conn.sent[0].Command
			if test.wantCmd != gotCmd {
				t.Errorf("Wanted command %v got %v", test.wantCmd, gotCmd)
			}
		})
	}
}
