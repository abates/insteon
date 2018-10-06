package insteon

import (
	"testing"
	"time"
)

func newTestConnection(dst Address) (*connection, chan *MessageRequest, chan *Message) {
	sendCh := make(chan *MessageRequest, 10)
	recvCh := make(chan *Message, 10)
	return newConnection(sendCh, recvCh, dst, 1, time.Millisecond), sendCh, recvCh
}

// TODO need to rewrite this test because it sucks and is full
// of race conditions
func TestConnectionProcess(t *testing.T) {
	/*oldTimeout := Timeout
	Timeout = time.Millisecond
	conn, _, recvCh := newTestConnection(testSrcAddr)
	close(recvCh)
	select {
	case _, open := <-conn.recvCh:
		if open {
			t.Errorf("Expected recvCh to be closed")
		}
	case <-time.After(Timeout):
		t.Errorf("Expected recvCh to be closed")
	}

	conn, sendCh, _ := newTestConnection(testSrcAddr)
	close(conn.sendCh)

	select {
	case _, open := <-sendCh:
		if open {
			t.Errorf("Expected sendCh to be closed")
		}
	case <-time.After(Timeout):
		t.Errorf("Expected sendCh to be closed")
	}

	conn, sendCh, recvCh = newTestConnection(testSrcAddr)
	doneCh := make(chan bool, 1)
	request := &MessageRequest{timeout: time.Now().Add(Timeout), DoneCh: doneCh}
	conn.queue = append(conn.queue, request)
	<-doneCh
	if request.Err != ErrReadTimeout {
		t.Errorf("Expected %v got %v", ErrReadTimeout, request.Err)
	}

	if len(conn.queue) != 0 {
		t.Errorf("Expected queue to be zero, got %d", len(conn.queue))
	}
	Timeout = oldTimeout*/
}

func TestConnectionReceiveAck(t *testing.T) {
	tests := []struct {
		version     EngineVersion
		returnedAck *Message
		expectedErr error
	}{
		{VerI1, TestMessageUnknownCommandNak, ErrUnknownCommand},
		{VerI1, TestMessageNoLoadDetected, ErrNoLoadDetected},
		{VerI1, TestMessageNotLinked, ErrNotLinked},
		{VerI1, &Message{Src: testDstAddr, Flags: StandardDirectNak, Command: 0x0001}, ErrUnexpectedResponse},
		{VerI2Cs, TestMessageIllegalValue, ErrIllegalValue},
		{VerI2Cs, TestMessagePreNak, ErrPreNak},
		{VerI2Cs, TestMessageIncorrectChecksum, ErrIncorrectChecksum},
		{VerI2Cs, TestMessageNoLoadDetectedI2Cs, ErrNoLoadDetected},
		{VerI2Cs, TestMessageNotLinkedI2Cs, ErrNotLinked},
		{VerI2Cs, &Message{Src: testDstAddr, Flags: StandardDirectNak, Command: 0x0001}, ErrUnexpectedResponse},
	}

	for i, test := range tests {
		conn := &connection{
			addr:    testDstAddr,
			version: test.version,
		}
		doneCh := make(chan *MessageRequest, 1)
		request := &MessageRequest{Message: &Message{Command: test.returnedAck.Command & 0xff00}, DoneCh: doneCh}
		conn.queue = append(conn.queue, request)
		conn.receive(test.returnedAck)

		if !IsError(request.Err, test.expectedErr) {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedErr, request.Err)
		} else if request.Ack != test.returnedAck {
			t.Errorf("tests[%d] expected %v got %v", i, test.returnedAck, request.Ack)
		}
	}
}

func TestConnectionReceiveMatch(t *testing.T) {
	tests := []struct {
		input    *Message
		match    Command
		expected int
	}{
		{&Message{Src: testDstAddr, Command: 0x0000}, 0x0101, 0},
		{&Message{Src: testDstAddr, Command: 0x01ff}, 0x0101, 0},
		{&Message{Src: testDstAddr, Command: 0x0101}, 0x0101, 1},
		{&Message{Src: testDstAddr, Command: 0x0101}, 0x0100, 1},
	}

	for i, test := range tests {
		conn := &connection{
			addr:   testDstAddr,
			match:  []Command{test.match},
			recvCh: make(chan *Message, 1),
		}

		conn.receive(test.input)

		if test.expected != len(conn.recvCh) {
			t.Errorf("tests[%d] Expected %d packets to be received. Got %d", i, test.expected, len(conn.recvCh))
		}
	}
}

func TestConnectionReceive(t *testing.T) {
	tests := []struct {
		input    *Message
		expected int
	}{
		{&Message{Src: testSrcAddr}, 0},
		{&Message{Src: testDstAddr}, 1},
	}

	for i, test := range tests {
		conn := &connection{
			addr:   testDstAddr,
			recvCh: make(chan *Message, 1),
		}
		conn.receive(test.input)

		if test.expected != len(conn.recvCh) {
			t.Errorf("tests[%d] Expected %d packets to be received. Got %d", i, test.expected, len(conn.recvCh))
		}
	}
}

func TestConnectionSend(t *testing.T) {
	upstreamSendCh := make(chan *MessageRequest, 1)
	conn := &connection{
		addr:           testDstAddr,
		upstreamSendCh: upstreamSendCh,
	}

	doneCh := make(chan *MessageRequest, 1)
	request := &MessageRequest{Message: &Message{}, DoneCh: doneCh}
	conn.queue = append(conn.queue, request)
	go func() {
		request := <-upstreamSendCh
		request.Err = ErrReadTimeout
		request.DoneCh <- request
	}()

	conn.send()

	<-doneCh
	if request.Message.Dst != testDstAddr {
		t.Errorf("Expected destination to be %v got %v", testSrcAddr, request.Message.Dst)
	}

	if request.Err != ErrReadTimeout {
		t.Errorf("Expected %v got %v", ErrReadTimeout, request.Err)
	}
}
