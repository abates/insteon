package plm

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestPacketAckNak(t *testing.T) {
	tests := []struct {
		cmd   Command
		input byte
		ack   bool
		nak   bool
	}{
		{0x60, 0x06, true, false},
		{0x60, 0x15, false, true},
		{0x01, 0x06, false, false},
		{0x01, 0x15, false, false},
	}

	for i, test := range tests {
		p := &Packet{Command: test.cmd, Ack: test.input}
		if p.ACK() != test.ack {
			t.Errorf("tests[%d] expected ack to be %v got %v", i, test.ack, p.ACK())
		}

		if p.NAK() != test.nak {
			t.Errorf("tests[%d] expected nak to be %v got %v", i, test.nak, p.NAK())
		}
	}
}

func TestPacketMarshalUnmarshalBinary(t *testing.T) {
	tests := []struct {
		input       []byte
		expected    *Packet
		expectedErr error
	}{
		{[]byte{0x00}, &Packet{}, ErrNoSync},
		{[]byte{0x02, byte(CmdSendInsteonMsg), 0x01, 0x02, 0x03, 0x04, 0x06}, &Packet{Ack: 0x06, Command: CmdSendInsteonMsg, Payload: []byte{0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04}}, nil},
	}

	for i, test := range tests {
		packet := &Packet{}
		err := packet.UnmarshalBinary(test.input)
		if err == test.expectedErr {
			if err == nil {
				if !reflect.DeepEqual(packet, test.expected) {
					t.Errorf("tests[%d] expected %v got %v", i, test.expected, packet)
				}
			}
		} else {
			t.Errorf("tests[%d] expected error %v got %v", i, test.expectedErr, err)
		}
	}
}

func TestPacketMarshalBinary(t *testing.T) {
	tests := []struct {
		input    *Packet
		expected []byte
	}{
		{&Packet{Ack: 0x06, Command: CmdSendInsteonMsg, Payload: []byte{0x01, 0x02, 0x03, 0x04}}, []byte{0x02, byte(CmdSendInsteonMsg), 0x01, 0x02, 0x03, 0x04}},
	}

	for i, test := range tests {
		buf, _ := test.input.MarshalBinary()
		if !bytes.Equal(test.expected, buf) {
			t.Errorf("tests[%d] expected %v got %v", i, test.expected, buf)
		}
	}
}

func TestPacketString(t *testing.T) {
	tests := []struct {
		input    *Packet
		expected string
	}{
		{&Packet{Command: CmdGetInfo, Ack: 0x06}, fmt.Sprintf("%v ACK", CmdGetInfo)},
		{&Packet{Command: CmdGetInfo, Ack: 0x15}, fmt.Sprintf("%v NAK", CmdGetInfo)},
	}

	for i, test := range tests {
		str := test.input.String()
		if str != test.expected {
			t.Errorf("tests[%d] expected %q got %q", i, test.expected, str)
		}
	}
}

func TestPacketFormat(t *testing.T) {
	tests := []struct {
		format   string
		input    *Packet
		expected string
	}{
		{"%x", &Packet{Command: 0x01, Ack: 0x06}, "0106"},
		{"%x", &Packet{Command: 0x01, Payload: []byte{2, 3, 4, 5, 15}, Ack: 0x06}, "01020304050f06"},
		{"%X", &Packet{Command: 0x01, Payload: []byte{2, 3, 4, 5, 15}, Ack: 0x06}, "01020304050F06"},
		{"%s", &Packet{Command: CmdGetInfo, Ack: 0x06}, fmt.Sprintf("%v ACK", CmdGetInfo)},
		{"%v", &Packet{Command: CmdGetInfo, Payload: []byte{2, 3, 4, 5, 15}, Ack: 0x06}, fmt.Sprintf("%v 02 03 04 05 0f ACK", CmdGetInfo)},
		{"%v", &Packet{Command: CmdGetInfo, Payload: []byte{2, 3, 4, 5, 15}, Ack: 0x15}, fmt.Sprintf("%v 02 03 04 05 0f NAK", CmdGetInfo)},
		{"%q", &Packet{Command: CmdGetInfo, Payload: []byte{2, 3, 4, 5, 15}, Ack: 0x15}, fmt.Sprintf("%q", fmt.Sprintf("%v NAK", CmdGetInfo))},
		{"%d", &Packet{Command: CmdGetInfo, Ack: 0x06}, fmt.Sprintf("%%!d(packet=%v ACK)", CmdGetInfo)},
	}

	for i, test := range tests {
		str := fmt.Sprintf(test.format, test.input)
		if str != test.expected {
			t.Errorf("tests[%d] expected %q got %q", i, test.expected, str)
		}
	}
}
