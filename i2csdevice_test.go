package insteon

import (
	"testing"
)

func TestI2CsDeviceFunctions(t *testing.T) {
	tests := []struct {
		execute         func(*I2CsDevice)
		response        []*Message
		expectedCommand Command
	}{
		{func(dev *I2CsDevice) { dev.EnterLinkingMode(240) }, []*Message{TestAck}, CmdEnterLinkingModeExt.SubCommand(240)},
	}

	for i, test := range tests {
		net := &testNetwork{}
		device := NewI2CsDevice(Address{}, net)
		go func() {
			for _, resp := range test.response {
				device.Notify(resp)
			}
		}()
		test.execute(device)

		if net.sentMessages[0].Command != test.expectedCommand {
			t.Errorf("tests[%d] expected command %v got %v", i, test.expectedCommand, net.sentMessages[0].Command)
		}
	}
}

func TestI2CSDeviceString(t *testing.T) {
	device := NewI2CsDevice(Address{1, 2, 3}, nil)
	if device.String() != "I2CS Device (01.02.03)" {
		t.Errorf("Expected %q got %q", "I2CS Device (01.02.03)", device.String())
	}
}
