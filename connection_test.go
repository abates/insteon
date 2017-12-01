package insteon

import (
	"testing"
	"time"
)

type testBridge struct {
	responses []*Command
	respFlags []Flags
}

func (tb *testBridge) Send(timeout time.Duration, message *Message) error {
	return nil
}

func (tb *testBridge) Receive(timeout time.Duration) (*Message, error) {
	response := tb.responses[0]
	tb.responses = tb.responses[1:]

	respFlags := tb.respFlags[0]
	tb.respFlags = tb.respFlags[1:]

	return &Message{Flags: respFlags, Command: response}, nil
}

func testErr(t *testing.T, testNum int, expected, got error) {
	if expected == got {
		return
	}

	t.Errorf("test[%d] Expected \"%v\" got \"%v\"", testNum, expected, got)
}

func TestSendingCommands(t *testing.T) {
	tests := []struct {
		standard  bool
		command   *Command
		responses []*Command
		respFlags []Flags
		err       error
	}{
		{true, CmdPing, []*Command{CmdPing}, []Flags{StandardDirectAck}, nil},
		{true, CmdPing, []*Command{CmdPing}, []Flags{StandardDirectMessage}, ErrUnexpectedResponse},
		{true, CmdPing, []*Command{CmdIDReq}, []Flags{StandardDirectMessage}, ErrUnexpectedResponse},
		{false, CmdPing, []*Command{CmdPing}, []Flags{ExtendedDirectAck}, nil},
		{true, CmdPing, []*Command{CmdPing, CmdPing}, []Flags{StandardDirectAck, StandardDirectAck}, nil},
	}

	for i, test := range tests {
		tb := &testBridge{
			responses: test.responses,
			respFlags: test.respFlags,
		}
		conn := NewDeviceConnection(time.Second, Address([3]byte{0x00, 0x00, 0x00}), tb)
		if len(test.responses) > 1 {
			if test.standard {
				_, err := conn.SendStandardCommandAndWait(test.command)
				testErr(t, i, test.err, err)
			} else {
				t.Errorf("Can't wait for an extended response")
			}
		} else {
			if test.standard {
				err := conn.SendStandardCommand(test.command)
				testErr(t, i, test.err, err)
			} else {
				err := conn.SendExtendedCommand(test.command, &BufPayload{})
				testErr(t, i, test.err, err)
			}
		}
	}
}
