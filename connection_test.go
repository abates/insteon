package insteon

import (
	"testing"
)

type testBridge struct {
	responses []*Command
	respFlags []Flags
}

func (tb *testBridge) Send(payload Payload) error {
	return nil
}

func (tb *testBridge) Receive() (Payload, error) {
	response := tb.responses[0]
	tb.responses = tb.responses[1:]

	respFlags := tb.respFlags[0]
	tb.respFlags = tb.respFlags[1:]

	return &Message{Flags: respFlags, Command: response}, nil
}

func testErr(t *testing.T, testNum int, expected, got error) {
	if IsError(expected, got) {
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
		conn := NewI1Connection(Address([3]byte{0x00, 0x00, 0x00}), tb)
		if len(test.responses) > 1 {
			if test.standard {
				_, err := SendStandardCommandAndWait(conn, test.command)
				testErr(t, i, test.err, err)
			} else {
				t.Errorf("Can't wait for an extended response")
			}
		} else {
			if test.standard {
				_, err := SendStandardCommand(conn, test.command)
				testErr(t, i, test.err, err)
			} else {
				_, err := SendExtendedCommand(conn, test.command, &BufPayload{})
				testErr(t, i, test.err, err)
			}
		}
	}
}
