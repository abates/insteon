package insteon

import (
	"bytes"
	"reflect"
	"testing"
)

func TestI2DeviceFunctions(t *testing.T) {
	tests := []struct {
		function        func(*I2Device) (interface{}, error)
		response        *Message
		ack             *Message
		expectedValue   interface{}
		expectedCommand *Command
		expectedMatch   []*Command
		expectedPayload []byte
	}{
		{
			function:        func(device *I2Device) (interface{}, error) { return nil, device.EnterLinkingMode(1) },
			expectedCommand: (&Command{Cmd: [2]byte{0x09, 0x00}}).SubCommand(0x01),
		},
		{
			function:        func(device *I2Device) (interface{}, error) { return nil, device.EnterUnlinkingMode(1) },
			expectedCommand: (&Command{Cmd: [2]byte{0x0a, 0x00}}).SubCommand(0x01),
		},
		{
			function:        func(device *I2Device) (interface{}, error) { return nil, device.ExitLinkingMode() },
			expectedCommand: &Command{Cmd: [2]byte{0x08, 0x00}},
		},
	}

	for i, test := range tests {
		conn := &testConnection{responses: []*Message{test.response}, ackMessage: test.ack}
		address := Address([3]byte{0x01, 0x02, 0x03})
		device := NewI2Device(NewI1Device(address, conn))

		if device.String() != "I2 Device (01.02.03)" {
			t.Errorf("tests[%d] expected %q got %q", i, "I2 Device (01.02.03)", device.String())
		}

		db, _ := device.LinkDB()
		if db == nil {
			t.Errorf("tests[%d] expected non-nil link database", i)
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
