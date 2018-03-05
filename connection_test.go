package insteon

import (
	"bytes"
	"testing"
	"time"
)

type testConnection struct {
	ackMessage    *Message
	lastMessage   Message
	responses     []*Message
	resp          chan *Message
	payload       []byte
	matchCommands []CommandBytes
	closed        bool
	timeout       time.Duration
}

func (conn *testConnection) FirmwareVersion() FirmwareVersion { return FirmwareVersion(0) }
func (conn *testConnection) Address() Address                 { return Address{0x00, 0x01, 0x02} }

func (conn *testConnection) Write(msg *Message) (ack *Message, err error) {
	if msg != nil {
		conn.lastMessage = *msg
		if msg.Payload != nil {
			conn.payload = msg.Payload
		}
	}

	var response *Message
	if len(conn.responses) > 0 {
		response = conn.responses[0]
		conn.responses = conn.responses[1:]
	} else {
		response = msg
	}

	go func(response *Message) {
		time.Sleep(conn.timeout)
		select {
		case conn.resp <- response:
		default:
		}
	}(response)

	if conn.ackMessage == nil {
		ack = &Message{}
		buf, _ := msg.MarshalBinary()
		ack.UnmarshalBinary(buf)
		ack.Flags = StandardDirectAck
	} else {
		ack = conn.ackMessage
	}
	return ack, nil
}

func (conn *testConnection) Subscribe(match ...CommandBytes) <-chan *Message {
	conn.matchCommands = match
	conn.resp = make(chan *Message, 1)
	return conn.resp
}

func (conn *testConnection) Unsubscribe(ch <-chan *Message) {
}

func (conn *testConnection) Close() error {
	conn.closed = true
	return nil
}

func TestI1ConnectionWrite(t *testing.T) {
	tests := []struct {
		ackMessage    *Message
		expectedError error
	}{
		{&Message{Flags: Flags(0xaf), Command: CommandBytes{Command1: 0x00, Command2: 0xfd}}, ErrUnknownCommand},
		{&Message{Flags: Flags(0xaf), Command: CommandBytes{Command1: 0x00, Command2: 0xfe}}, ErrNoLoadDetected},
		{&Message{Flags: Flags(0xaf), Command: CommandBytes{Command1: 0x00, Command2: 0xff}}, ErrNotLinked},
		{&Message{Flags: Flags(0xaf), Command: CommandBytes{Command1: 0x00, Command2: 0x0f}}, ErrUnexpectedResponse},
	}

	for i, test := range tests {
		conn := NewI1Connection(&testConnection{ackMessage: test.ackMessage})
		_, err := conn.Write(nil)
		if !IsError(test.expectedError, err) {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedError, err)
		}
	}
}

func TestI2CsConnectionWrite(t *testing.T) {
	tests := []struct {
		ackMessage    *Message
		expectedError error
	}{
		{&Message{Flags: Flags(0xaf), Command: CommandBytes{Command1: 0x00, Command2: 0xfb}}, ErrIllegalValue},
		{&Message{Flags: Flags(0xaf), Command: CommandBytes{Command1: 0x00, Command2: 0xfc}}, ErrPreNak},
		{&Message{Flags: Flags(0xaf), Command: CommandBytes{Command1: 0x00, Command2: 0xfd}}, ErrIncorrectChecksum},
		{&Message{Flags: Flags(0xaf), Command: CommandBytes{Command1: 0x00, Command2: 0xfe}}, ErrNoLoadDetected},
		{&Message{Flags: Flags(0xaf), Command: CommandBytes{Command1: 0x00, Command2: 0xff}}, ErrNotLinked},
		{&Message{Flags: Flags(0xaf), Command: CommandBytes{Command1: 0x00, Command2: 0xf0}}, ErrUnexpectedResponse},
	}

	for i, test := range tests {
		conn := NewI2CsConnection(&testConnection{ackMessage: test.ackMessage})
		_, err := conn.Write(&Message{})
		if !IsError(test.expectedError, err) {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedError, err)
		}
	}
}

func TestSendCommand(t *testing.T) {
	tests := []struct {
		command  *Command
		expected CommandBytes
	}{
		{CmdProductDataReq, CmdProductDataReq.Version(0)},
	}

	for i, test := range tests {
		conn := &testConnection{timeout: time.Millisecond}
		SendCommand(conn, test.command)
		if conn.lastMessage.Command != test.expected {
			t.Errorf("tests[%d] expected %v got %v", i, test.expected, conn.lastMessage.Command)
		}

		if conn.lastMessage.Flags != StandardDirectMessage {
			t.Errorf("tests[%d] expected %v got %v", i, StandardDirectMessage, conn.lastMessage.Flags)
		}
	}
}

func TestSendCommandAndWait(t *testing.T) {
	tests := []struct {
		command         *Command
		expectedCommand CommandBytes
		timeout         time.Duration
		ackMessage      *Message
		err             error
	}{
		{CmdProductDataReq, CmdProductDataReq.Version(0), time.Millisecond, nil, nil},
		{CmdProductDataReq, CmdProductDataReq.Version(0), Timeout * 2, nil, ErrReadTimeout},
		{CmdProductDataReq, CmdProductDataReq.Version(0), time.Millisecond, &Message{Flags: StandardDirectNak}, ErrNak},
	}

	for i, test := range tests {
		oldTimeout := Timeout
		Timeout = 10 * time.Millisecond
		conn := &testConnection{timeout: test.timeout, ackMessage: test.ackMessage}
		msg, err := SendCommandAndWait(conn, test.command, test.command)
		if err != test.err {
			t.Errorf("tests[%d] expected %v got %v", i, test.err, err)
		}

		if err == nil {
			if msg.Command != test.expectedCommand {
				t.Errorf("tests[%d] expected %v got %v", i, test.expectedCommand, msg.Command)
			}
		}
		Timeout = oldTimeout
	}
}

func TestSendExtendedCommand(t *testing.T) {
	tests := []struct {
		command         *Command
		expectedCommand CommandBytes
		payload         []byte
	}{
		{CmdProductDataResp, CmdProductDataResp.Version(0), []byte{0xfe, 0xfe, 0xaa, 0xbc}},
	}

	for i, test := range tests {
		conn := &testConnection{timeout: time.Millisecond}
		SendExtendedCommand(conn, test.command, test.payload)
		if conn.lastMessage.Command != test.expectedCommand {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedCommand, conn.lastMessage.Command)
		}

		if conn.lastMessage.Flags != ExtendedDirectMessage {
			t.Errorf("tests[%d] expected %v got %v", i, ExtendedDirectMessage, conn.lastMessage.Flags)
		}

		if !bytes.Equal(conn.lastMessage.Payload[0:len(test.payload)], test.payload) {
			t.Errorf("tests[%d] expected %v got %v", i, test.payload, conn.lastMessage.Payload)
		}
	}
}
