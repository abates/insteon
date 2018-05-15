package insteon

import "testing"

func TestDimmableDeviceIsDimmer(t *testing.T) {
	net := &testNetwork{}
	address := Address([3]byte{0x01, 0x02, 0x03})
	device := dimmableLightingFactory(NewI1Device(address, net), DeviceInfo{})

	if _, ok := device.(Dimmer); !ok {
		t.Errorf("Expected DimmableDevice to be Dimmer")
	}
}

func TestSwitchedDeviceIsSwitch(t *testing.T) {
	net := &testNetwork{}
	address := Address([3]byte{0x01, 0x02, 0x03})
	device := switchedLightingFactory(NewI1Device(address, net), DeviceInfo{})

	if _, ok := device.(Switch); !ok {
		t.Errorf("Expected SwitchedDevice to be Switch")
	}
}
