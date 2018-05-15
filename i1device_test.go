package insteon

import (
	"testing"
)

func TestI1DeviceIsDevice(t *testing.T) {
	net := &testNetwork{}
	address := Address([3]byte{0x01, 0x02, 0x03})
	var device interface{}
	device = NewI1Device(address, net)

	if _, ok := device.(Device); !ok {
		t.Errorf("Expected I1Device to be linkable")
	}
}

func TestI1DeviceNotify(t *testing.T) {
	tests := []struct {
		input *Message
	}{
		{&Message{EngineVersion(1), Address{1, 2, 3}, Address{4, 5, 6}, Flags(0x20), Command{1, 2}, nil}},
		{&Message{EngineVersion(1), Address{1, 2, 3}, Address{4, 5, 6}, Flags(0xa0), Command{1, 2}, nil}},
		{&Message{EngineVersion(1), Address{1, 2, 3}, Address{4, 5, 6}, Flags(0x1a), Command{0x03, 0x00}, make([]byte, 14)}},
	}

	for i, test := range tests {
		device := NewI1Device(Address{}, &testNetwork{})
		device.ackCh = make(chan *Message, 1)
		device.productDataCh = make(chan *Message, 1)
		device.Notify(test.input)
		if (test.input.Ack() || test.input.Nak()) && len(device.ackCh) < 1 {
			t.Errorf("tests[%d] expected message to be delivered to the ack channel", i)
		}

		if test.input.Command[0] == 0x03 && len(device.productDataCh) < 1 {
			t.Errorf("tests[%d] expected message to be delivered to the product data channel", i)
		}
	}
}

func TestI1DeviceSendCommand(t *testing.T) {
	tests := []struct {
		inputCommand Command
		inputPayload []byte
	}{
		{CmdPing, nil},
		{CmdPing, make([]byte, 14)},
	}

	for i, test := range tests {
		net := &testNetwork{}
		device := NewI1Device(Address{1, 2, 3}, net)
		device.ackCh = make(chan *Message, 1)
		device.ackCh <- &Message{}

		device.SendCommand(test.inputCommand, test.inputPayload)

		if len(net.sentMessages) < 1 {
			t.Errorf("tests[%d] expected message to be sent, but it wasn't", i)
		} else {
			msg := net.sentMessages[0]
			if msg.Command != test.inputCommand {
				t.Errorf("tests[%d] expected Command %v got %v", i, test.inputCommand, msg.Command)
				continue
			}

			if len(test.inputPayload) == 0 {
				if msg.Flags != StandardDirectMessage {
					t.Errorf("tests[%d] expected StandardDirectMessage got %v", i, msg.Flags)
				}
			} else {
				if msg.Flags != ExtendedDirectMessage {
					t.Errorf("tests[%d] expected ExtendedDirectMessage got %v", i, msg.Flags)
				}
			}
		}
	}
}

func TestI1DeviceAddress(t *testing.T) {
	address := Address{36, 25, 36}
	i1device := &I1Device{address: address}

	if address != i1device.Address() {
		t.Errorf("Expected %v got %v", address, i1device.Address())
	}
}

func TestI1DeviceCommands(t *testing.T) {
	tests := []struct {
		execute         func(*I1Device)
		response        []*Message
		expectedCommand Command
	}{
		{func(dev *I1Device) { dev.AssignToAllLinkGroup(240) }, []*Message{TestAck}, CmdAssignToAllLinkGroup.SubCommand(240)},
		{func(dev *I1Device) { dev.DeleteFromAllLinkGroup(240) }, []*Message{TestAck}, CmdDeleteFromAllLinkGroup.SubCommand(240)},
		{func(dev *I1Device) { dev.ProductData() }, []*Message{TestAck, TestProductDataResponse}, CmdProductDataReq},
		{func(dev *I1Device) { dev.Ping() }, []*Message{TestAck}, CmdPing},
	}

	for i, test := range tests {
		net := &testNetwork{}
		device := NewI1Device(Address{}, net)
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

func TestI1DeviceString(t *testing.T) {
	device := NewI1Device(Address{1, 2, 3}, nil)
	if device.String() != "I1 Device (01.02.03)" {
		t.Errorf("Expected %q got %q", "I1 Device (01.02.03)", device.String())
	}
}
