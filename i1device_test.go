package insteon

import (
	"bytes"
	"reflect"
	"testing"
)

func TestI1DeviceFunctions(t *testing.T) {
	tests := []struct {
		function        func(*I1Device) (interface{}, error)
		response        *Message
		ack             *Message
		expectedValue   interface{}
		expectedCommand CommandBytes
		expectedMatch   []CommandBytes
		expectedPayload []byte
	}{
		// Test 0
		{
			function:        func(device *I1Device) (interface{}, error) { return nil, device.AssignToAllLinkGroup(1) },
			expectedCommand: CommandBytes{Command1: 0x01, Command2: 0x01},
		},
		// Test 1
		{
			function:        func(device *I1Device) (interface{}, error) { return nil, device.DeleteFromAllLinkGroup(1) },
			expectedCommand: CommandBytes{Command1: 0x02, Command2: 0x01},
		},
		// Test 2
		{
			function:        func(device *I1Device) (interface{}, error) { return device.ProductData() },
			response:        &Message{Payload: []byte{0, 0x04, 0x05, 0x06, 0x07, 0x08, 0, 0, 0, 0, 0, 0, 0, 0}},
			expectedCommand: CommandBytes{Command1: 0x03, Command2: 0x00},
			expectedValue:   &ProductData{ProductKey([3]byte{0x04, 0x05, 0x06}), DevCat([2]byte{0x07, 0x08})},
			expectedMatch:   []CommandBytes{CmdProductDataResp.Version(0)},
		},
		// Test 3
		{
			function:        func(device *I1Device) (interface{}, error) { return device.FXUsername() },
			response:        &Message{Payload: []byte("ABCDEFGHIJKLMN")},
			expectedCommand: CommandBytes{Command1: 0x03, Command2: 0x01},
			expectedValue:   "ABCDEFGHIJKLMN",
			expectedMatch:   []CommandBytes{CmdFxUsernameResp.Version(0)},
		},
		// Test 4
		{
			function:        func(device *I1Device) (interface{}, error) { return device.TextString() },
			response:        &Message{Payload: []byte("ABCDEFGHIJKLMN")},
			expectedCommand: CommandBytes{Command1: 0x03, Command2: 0x02},
			expectedValue:   "ABCDEFGHIJKLMN",
			expectedMatch:   []CommandBytes{CmdDeviceTextStringResp.Version(0)},
		},
		// Test 5
		{
			function: func(device *I1Device) (interface{}, error) { return nil, device.EnterLinkingMode(0) },
		},
		// Test 6
		{
			function: func(device *I1Device) (interface{}, error) { return nil, device.EnterUnlinkingMode(0) },
		},
		// Test 7
		{
			function: func(device *I1Device) (interface{}, error) { return nil, device.ExitLinkingMode() },
		},
		// Test 8
		{
			function:        func(device *I1Device) (interface{}, error) { return device.EngineVersion() },
			expectedCommand: CommandBytes{Command1: 0x0d, Command2: 0x00},
			ack:             &Message{Flags: StandardDirectAck, Command: CmdGetEngineVersion.Version(0).SubCommand(2)},
			expectedValue:   EngineVersion(2),
		},
		// Test 9
		{
			function:        func(device *I1Device) (interface{}, error) { return nil, device.Ping() },
			expectedCommand: CommandBytes{Command1: 0x0f, Command2: 0x00},
		},
		// Test 10
		{
			function:        func(device *I1Device) (interface{}, error) { return device.IDRequest() },
			response:        &Message{Command: CmdSetButtonPressedController.Version(0), Dst: Address{0x15, 0x22}},
			expectedCommand: CommandBytes{Command1: 0x10, Command2: 0x00},
			expectedMatch:   []CommandBytes{CmdSetButtonPressedController.Version(0), CmdSetButtonPressedResponder.Version(0)},
			expectedValue:   DevCat{0x15, 0x22},
		},
		// Test 11
		{
			function:        func(device *I1Device) (interface{}, error) { return nil, device.SetTextString("OPQRSTUVWXYZAB") },
			expectedCommand: CommandBytes{Command1: 0x03, Command2: 0x03},
			expectedPayload: []byte("OPQRSTUVWXYZAB"),
		},
		// Test 12
		{
			function: func(device *I1Device) (interface{}, error) { return nil, device.SetAllLinkCommandAlias(nil, nil) },
		},
		// Test 13
		{
			function: func(device *I1Device) (interface{}, error) { return nil, device.SetAllLinkCommandAliasData(nil) },
		},
		// Test 14
		{
			function: func(device *I1Device) (interface{}, error) {
				return device.BlockDataTransfer(0, 0, 0)
			},
			expectedValue: []byte(nil),
		},
		// Test 15
		{
			function: func(device *I1Device) (interface{}, error) { return device.LinkDB() },
		},
	}

	for i, test := range tests {
		conn := &testConnection{responses: []*Message{test.response}, ackMessage: test.ack}
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
			t.Logf("tests[%d] ack: %v", i, test.ack)
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedValue, value)
		}

		if test.expectedCommand != conn.lastMessage.Command {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedCommand, conn.lastMessage.Command)
		}

		if test.expectedMatch != nil {
			if !reflect.DeepEqual(conn.matchCommands, test.expectedMatch) {
				t.Errorf("tests[%d] expected %v got %v", i, test.expectedMatch, conn.matchCommands)
			}
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
