package insteon

import (
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
		device, _ := switchedDeviceFactory(test.info, Address{1, 2, 3}, nil, nil, time.Millisecond)
		if reflect.TypeOf(device) != reflect.TypeOf(test.expected) {
			t.Errorf("tests[%d] expected %T got %T", i, test.expected, device)
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
		device, _ := dimmableDeviceFactory(test.info, Address{1, 2, 3}, nil, nil, time.Millisecond)
		if reflect.TypeOf(device) != reflect.TypeOf(test.expected) {
			t.Errorf("tests[%d] expected %T got %T", i, test.expected, device)
		}
	}
}
