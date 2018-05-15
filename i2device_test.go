package insteon

import (
	"testing"
)

func TestI2DeviceIsLinkable(t *testing.T) {
	net := &testNetwork{}
	address := Address([3]byte{0x01, 0x02, 0x03})
	var device interface{}
	device = NewI2Device(address, net)

	if _, ok := device.(LinkableDevice); !ok {
		t.Errorf("Expected I2Device to be linkable")
	}
}

func TestI2DeviceFunctions(t *testing.T) {
	tests := []struct {
		execute         func(*I2Device) error
		response        []*Message
		expectedCommand Command
		expectedErr     error
	}{
		{func(dev *I2Device) error { return extractError(dev.Links()) }, []*Message{TestAck, TestDeviceLink1, TestDeviceLink2, TestDeviceLink3}, CmdReadWriteALDB, nil},
		{func(dev *I2Device) error { return dev.EnterLinkingMode(240) }, []*Message{TestAck}, CmdEnterLinkingMode.SubCommand(240), nil},
		{func(dev *I2Device) error { return dev.EnterUnlinkingMode(240) }, []*Message{TestAck}, CmdEnterUnlinkingMode.SubCommand(240), nil},
		{func(dev *I2Device) error { return dev.ExitLinkingMode() }, []*Message{TestAck}, CmdExitLinkingMode, nil},
		{func(dev *I2Device) error { return dev.WriteLink(&LinkRecord{}) }, nil, CmdReadWriteALDB, ErrInvalidMemAddress},
		{func(dev *I2Device) error { return dev.WriteLink(&LinkRecord{memAddress: 0xffff}) }, []*Message{TestAck}, CmdReadWriteALDB, nil},
	}

	for i, test := range tests {
		net := &testNetwork{}
		device := NewI2Device(Address{}, net)
		go func() {
			for _, resp := range test.response {
				device.Notify(resp)
			}
		}()
		err := test.execute(device)

		if err == nil {
			if test.expectedErr != nil {
				t.Errorf("tests[%d] expected %v got nil", i, test.expectedErr)
			} else {
				if net.sentMessages[0].Command != test.expectedCommand {
					t.Errorf("tests[%d] expected command %v got %v", i, test.expectedCommand, net.sentMessages[0].Command)
				}
			}
		} else {
			if test.expectedErr != err {
				t.Errorf("tests[%d] expected %v got %v", i, test.expectedErr, err)
			}
		}
	}
}

func TestI2DeviceString(t *testing.T) {
	device := NewI2Device(Address{1, 2, 3}, nil)
	if device.String() != "I2 Device (01.02.03)" {
		t.Errorf("Expected %q got %q", "I2 Device (01.02.03)", device.String())
	}
}
