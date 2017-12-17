package insteon

import (
	"testing"
	"time"
)

type testConnection struct {
	message *Message
	resp    chan *Message
	timeout time.Duration
}

func (*testConnection) Close() error { return nil }

func (tc *testConnection) Subscribe(...*Command) <-chan *Message {
	tc.resp = make(chan *Message, 1)
	return tc.resp
}

func (*testConnection) Unsubscribe(<-chan *Message) {
}

func (tc *testConnection) Write(message *Message) (*Message, error) {
	tc.message = message
	go func() {
		time.Sleep(tc.timeout)
		select {
		case tc.resp <- message:
		default:
		}
	}()
	return nil, nil
}

func TestSendStandardCommand(t *testing.T) {
	tests := []struct {
		command *Command
	}{
		{CmdProductDataReq},
	}

	for i, test := range tests {
		conn := &testConnection{timeout: time.Millisecond}
		SendStandardCommand(conn, test.command)
		if conn.message.Command != test.command {
			t.Errorf("tests[%d] expected %v got %v", i, test.command, conn.message.Command)
		}

		if conn.message.Flags != StandardDirectMessage {
			t.Errorf("tests[%d] expected %v got %v", StandardDirectMessage, conn.message.Flags)
		}
	}
}

func TestSendStandardCommandAndWait(t *testing.T) {
	tests := []struct {
		command *Command
		timeout time.Duration
		err     error
	}{
		{CmdProductDataReq, time.Millisecond, nil},
		{CmdProductDataReq, Timeout * 2, ErrAckTimeout},
	}

	for i, test := range tests {
		oldTimeout := Timeout
		Timeout = 10 * time.Millisecond
		conn := &testConnection{timeout: test.timeout}
		msg, err := SendStandardCommandAndWait(conn, test.command, test.command)
		if err != test.err {
			t.Errorf("tests[%d] expected %v got %v", test.err, err)
		}

		if err == nil {
			if msg.Command != test.command {
				t.Errorf("tests[%d] expected %v got %v", i, test.command, msg.Command)
			}
		}
		Timeout = oldTimeout
	}
}

func TestSendExtendedCommand(t *testing.T) {
	tests := []struct {
		command *Command
		payload Payload
	}{
		{CmdProductDataResp, &BufPayload{[]byte{0xfe, 0xfe, 0xaa, 0xbc}}},
	}

	for i, test := range tests {
		conn := &testConnection{timeout: time.Millisecond}
		SendExtendedCommand(conn, test.command, test.payload)
		if conn.message.Command != test.command {
			t.Errorf("tests[%d] expected %v got %v", i, test.command, conn.message.Command)
		}

		if conn.message.Flags != ExtendedDirectMessage {
			t.Errorf("tests[%d] expected %v got %v", i, ExtendedDirectMessage, conn.message.Flags)
		}

		if conn.message.Payload != test.payload {
			t.Errorf("tests[%d] expected %v got %v", i, test.payload, conn.message.Payload)
		}
	}
}
