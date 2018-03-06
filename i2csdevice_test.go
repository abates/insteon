package insteon

import (
	"bytes"
	"reflect"
	"testing"
)

func TestI2CsDeviceFunctions(t *testing.T) {
	tests := []struct {
		function        func(*I2CsDevice) (interface{}, error)
		response        *Message
		ack             *Message
		expectedValue   interface{}
		expectedCommand CommandBytes
		expectedMatch   []*Command
		expectedPayload []byte
	}{
		{
			function:        func(device *I2CsDevice) (interface{}, error) { return nil, device.EnterLinkingMode(1) },
			expectedCommand: CommandBytes{Command1: 0x09, Command2: 0x01},
			expectedPayload: make([]byte, 14),
		},
		{
			function:        func(device *I2CsDevice) (interface{}, error) { return nil, device.EnterUnlinkingMode(1) },
			expectedCommand: CommandBytes{Command1: 0x0a, Command2: 0x01},
		},
	}

	for i, test := range tests {
		conn := &testConnection{responses: []*Message{test.response}, ackMessage: test.ack}
		address := Address([3]byte{0x01, 0x02, 0x03})
		device := NewI2CsDevice(address, conn)
		device.I2Device.I1Device.devCat = DevCat{0x00, 0x00}
		device.I2Device.I1Device.firmwareVersion = 0

		if device.String() != "I2CS Device (01.02.03)" {
			t.Errorf("tests[%d] expected %q got %q", i, "I2CS Device (01.02.03)", device.String())
		}

		value, _ := test.function(device)
		if !reflect.DeepEqual(value, test.expectedValue) {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedValue, value)
		}

		if !test.expectedCommand.Equal(conn.lastMessage.Command) {
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
