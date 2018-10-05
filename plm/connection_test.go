package plm

import (
	"testing"
	"time"

	"github.com/abates/insteon"
)

func testClosedChannels(sendCh chan *CommandRequest, conn *connection) (string, bool) {
	select {
	case _, open := <-sendCh:
		if open {
			return "Expected upstreamSendCh to be closed", false
		}
	case <-time.After(time.Second):
		return "Timed out waiting for upstreamSendCh to be closed", false
	}

	select {
	case _, open := <-conn.recvCh:
		if open {
			return "Expected recvCh to be closed", false
		}
	case <-time.After(time.Second):
		return "Timed out waiting for recvCh to be closed", false
	}
	return "", true
}

func TestConnectionSend(t *testing.T) {
	sendCh := make(chan *CommandRequest, 1)
	conn := newConnection(sendCh, nil, CmdSendInsteonMsg)

	doneCh := make(chan *insteon.PacketRequest, 1)
	conn.sendCh <- &insteon.PacketRequest{DoneCh: doneCh}
	request := <-sendCh
	if request.Command != CmdSendInsteonMsg {
		t.Errorf("Expected %v to be sent, but got %v instead", CmdSendInsteonMsg, request.Command)
		request.DoneCh <- request
	} else {
		request.Err = ErrReadTimeout
		request.DoneCh <- request
		packetRequest := <-doneCh
		if packetRequest.Err != ErrReadTimeout {
			t.Errorf("Expected %v but got %v", ErrReadTimeout, packetRequest.Err)
		}
	}

	close(conn.sendCh)
	if msg, passed := testClosedChannels(sendCh, conn); !passed {
		t.Errorf("%v", msg)
	}
}

func TestConnectionReceive(t *testing.T) {
	tests := []struct {
		input    *Packet
		match    []Command
		expected bool
	}{
		{&Packet{Command: CmdStdMsgReceived}, []Command{CmdStdMsgReceived}, true},
		{&Packet{Command: CmdNak}, []Command{CmdStdMsgReceived}, false},
		{&Packet{Command: CmdNak}, nil, true},
	}

	for i, test := range tests {
		sendCh := make(chan *CommandRequest, 1)
		recvCh := make(chan *Packet, 1)
		conn := &connection{
			upstreamRecvCh: recvCh,
			matches:        test.match,
			upstreamSendCh: sendCh,
			recvCh:         make(chan []byte, 1),
		}

		recvCh <- test.input
		close(recvCh)
		conn.process()

		if test.expected && len(conn.recvCh) == 0 {
			t.Errorf("tests[%d] expected packet to be delivered", i)
		} else {
			<-conn.recvCh
		}

		if msg, passed := testClosedChannels(sendCh, conn); !passed {
			t.Errorf("%v", msg)
		}
	}
}
