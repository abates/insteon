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
		expectedCommand *Command
		expectedMatch   []*Command
		expectedPayload []byte
	}{
		{
			function:        func(device *I1Device) (interface{}, error) { return nil, device.AssignToAllLinkGroup(1) },
			expectedCommand: (&Command{Cmd: [2]byte{0x01, 0x00}}).SubCommand(0x01),
		},
		{
			function:        func(device *I1Device) (interface{}, error) { return nil, device.DeleteFromAllLinkGroup(1) },
			expectedCommand: (&Command{Cmd: [2]byte{0x02, 0x00}}).SubCommand(0x01),
		},
		{
			function:        func(device *I1Device) (interface{}, error) { return device.ProductData() },
			response:        &Message{Payload: &BufPayload{Buf: []byte{0, 0x04, 0x05, 0x06, 0x07, 0x08, 0, 0, 0, 0, 0, 0, 0, 0}}},
			expectedCommand: &Command{Cmd: [2]byte{0x03, 0x00}},
			expectedValue:   &ProductData{ProductKey([3]byte{0x04, 0x05, 0x06}), Category([2]byte{0x07, 0x08})},
			expectedMatch:   []*Command{CmdProductDataResp},
		},
		{
			function:        func(device *I1Device) (interface{}, error) { return device.FXUsername() },
			response:        &Message{Payload: &BufPayload{[]byte("ABCDEFGHIJKLMN")}},
			expectedCommand: &Command{Cmd: [2]byte{0x03, 0x01}},
			expectedValue:   "ABCDEFGHIJKLMN",
			expectedMatch:   []*Command{CmdFxUsernameResp},
		},
		{
			function:        func(device *I1Device) (interface{}, error) { return device.TextString() },
			response:        &Message{Payload: &BufPayload{[]byte("ABCDEFGHIJKLMN")}},
			expectedCommand: &Command{Cmd: [2]byte{0x03, 0x02}},
			expectedValue:   "ABCDEFGHIJKLMN",
			expectedMatch:   []*Command{CmdDeviceTextStringResp},
		},
		{
			function: func(device *I1Device) (interface{}, error) { return nil, device.EnterLinkingMode(0) },
		},
		{
			function: func(device *I1Device) (interface{}, error) { return nil, device.EnterUnlinkingMode(0) },
		},
		{
			function: func(device *I1Device) (interface{}, error) { return nil, device.ExitLinkingMode() },
		},
		{
			function:        func(device *I1Device) (interface{}, error) { return device.EngineVersion() },
			expectedCommand: &Command{Cmd: [2]byte{0x0d, 0x00}},
			ack:             &Message{Flags: StandardDirectAck, Command: CmdGetEngineVersion.SubCommand(2)},
			expectedValue:   EngineVersion(2),
		},
		{
			function:        func(device *I1Device) (interface{}, error) { return nil, device.Ping() },
			expectedCommand: &Command{Cmd: [2]byte{0x0f, 0x00}},
		},
		{
			function:        func(device *I1Device) (interface{}, error) { return device.IDRequest() },
			response:        &Message{Command: CmdSetButtonPressedController, Dst: Address{0x15, 0x22}},
			expectedCommand: &Command{Cmd: [2]byte{0x10, 0x00}},
			expectedMatch:   []*Command{CmdSetButtonPressedController, CmdSetButtonPressedResponder},
			expectedValue:   Category{0x15, 0x22},
		},
		{
			function:        func(device *I1Device) (interface{}, error) { return nil, device.SetTextString("OPQRSTUVWXYZAB") },
			expectedCommand: &Command{Cmd: [2]byte{0x03, 0x03}},
			expectedPayload: []byte("OPQRSTUVWXYZAB"),
		},
		{
			function: func(device *I1Device) (interface{}, error) { return nil, device.SetAllLinkCommandAlias(nil, nil) },
		},
		{
			function: func(device *I1Device) (interface{}, error) { return nil, device.SetAllLinkCommandAliasData(nil) },
		},
		{
			function: func(device *I1Device) (interface{}, error) {
				return device.BlockDataTransfer(0, 0, 0)
			},
			expectedValue: []byte(nil),
		},
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
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedValue, value)
		}

		if test.expectedCommand != nil {
			if !test.expectedCommand.Equal(conn.lastMessage.Command) {
				t.Errorf("tests[%d] expected %v got %v", i, test.expectedCommand, conn.lastMessage.Command)
			}
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
