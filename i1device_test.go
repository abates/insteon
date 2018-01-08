package insteon

import (
	"bytes"
	"reflect"
	"testing"
)

type I1DeviceTestConnection struct {
	command       *Command
	payload       []byte
	ack           *Message
	matchCommands []*Command
	response      *Message
	closed        bool
}

func (conn *I1DeviceTestConnection) Write(msg *Message) (ack *Message, err error) {
	conn.command = msg.Command
	if msg.Payload != nil {
		conn.payload, _ = msg.Payload.MarshalBinary()
	}
	if conn.ack == nil {
		ack = &Message{}
		buf, _ := msg.MarshalBinary()
		ack.UnmarshalBinary(buf)
		ack.Flags = StandardDirectAck
	} else {
		ack = conn.ack
	}
	return ack, nil
}

func (conn *I1DeviceTestConnection) Subscribe(match ...*Command) <-chan *Message {
	conn.matchCommands = match
	ch := make(chan *Message, 1)
	ch <- conn.response
	return ch
}

func (conn *I1DeviceTestConnection) Unsubscribe(ch <-chan *Message) {
}

func (conn *I1DeviceTestConnection) Close() error {
	conn.closed = true
	return nil
}

func TestI1DeviceFunctions(t *testing.T) {
	tests := []struct {
		function        func(*I1Device) (interface{}, error)
		response        *Message
		ack             *Message
		expectedValue   interface{}
		expectedCommand [2]byte
		expectedMatch   []*Command
		expectedPayload []byte
	}{
		{
			function:        func(device *I1Device) (interface{}, error) { return nil, device.AssignToAllLinkGroup(1) },
			expectedCommand: [2]byte{0x01, 0x01},
		},
		{
			function:        func(device *I1Device) (interface{}, error) { return nil, device.DeleteFromAllLinkGroup(1) },
			expectedCommand: [2]byte{0x02, 0x01},
		},
		{
			function:        func(device *I1Device) (interface{}, error) { return device.ProductData() },
			response:        &Message{Payload: &ProductData{ProductKey([3]byte{0x04, 0x05, 0x06}), Category([2]byte{0x07, 0x08})}},
			expectedCommand: [2]byte{0x03, 0x00},
			expectedValue:   &ProductData{ProductKey([3]byte{0x04, 0x05, 0x06}), Category([2]byte{0x07, 0x08})},
			expectedMatch:   []*Command{CmdProductDataResp},
		},
		{
			function:        func(device *I1Device) (interface{}, error) { return device.FXUsername() },
			response:        &Message{Payload: &BufPayload{[]byte("ABCDEFGHIJKLMN")}},
			expectedCommand: [2]byte{0x03, 0x01},
			expectedValue:   "ABCDEFGHIJKLMN",
			expectedMatch:   []*Command{CmdFxUsernameResp},
		},
		{
			function:        func(device *I1Device) (interface{}, error) { return device.TextString() },
			response:        &Message{Payload: &BufPayload{[]byte("ABCDEFGHIJKLMN")}},
			expectedCommand: [2]byte{0x03, 0x02},
			expectedValue:   "ABCDEFGHIJKLMN",
			expectedMatch:   []*Command{CmdDeviceTextStringResp},
		},
		{
			function:        func(device *I1Device) (interface{}, error) { return device.EngineVersion() },
			expectedCommand: [2]byte{0x0d, 0x00},
			ack:             &Message{Flags: StandardDirectAck, Command: CmdGetEngineVersion.SubCommand(2)},
			expectedValue:   EngineVersion(2),
		},
		{
			function:        func(device *I1Device) (interface{}, error) { return nil, device.Ping() },
			expectedCommand: [2]byte{0x0f, 0x00},
		},
		{
			function:        func(device *I1Device) (interface{}, error) { return nil, device.IDRequest() },
			expectedCommand: [2]byte{0x10, 0x00},
		},
		{
			function:        func(device *I1Device) (interface{}, error) { return nil, device.SetTextString("OPQRSTUVWXYZAB") },
			expectedCommand: [2]byte{0x03, 0x03},
			expectedPayload: []byte("OPQRSTUVWXYZAB"),
		},
	}

	for i, test := range tests {
		conn := &I1DeviceTestConnection{response: test.response, ack: test.ack}
		address := Address([3]byte{0x01, 0x02, 0x03})
		device := NewI1Device(address, conn)

		if device.Address() != address {
			t.Errorf("tests[%d] expected %v got %v", i, address, device.Address())
		}

		if device.String() != "I1 Device (01.02.03)" {
			t.Errorf("tests[%d] expected %q got %q", i, "I1 Device (01.02.03)", device.String())
		}

		value, _ := test.function(device)
		if !reflect.DeepEqual(value, test.expectedValue) {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedValue, value)
		}

		if test.expectedCommand != conn.command.Cmd {
			t.Errorf("tests[%d] expected 0x%04x got 0x%04x", i, test.expectedCommand, conn.command.Cmd)
		}

		if !reflect.DeepEqual(conn.matchCommands, test.expectedMatch) {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedMatch, conn.matchCommands)
		}

		if !bytes.Equal(conn.payload, test.expectedPayload) {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedPayload, conn.payload)
		}

		device.Close()
		if !conn.closed {
			t.Errorf("tests[%d] expected device.Close() to close underlying connection", i)
		}
	}
}
