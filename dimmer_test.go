package insteon

import (
	"bytes"
	"testing"
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

	ch := make(chan *Message, 1)
	ch <- msg

	tw := &testWriter{}
	tw.read = []*Message{msg}
	dd := NewDimmer(&BasicDevice{MessageWriter: tw, DeviceInfo: DeviceInfo{FirmwareVersion: 67}})
	got, err := dd.Config()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if got != want {
		t.Errorf("Want config %v got %v", want, got)
	}
}
