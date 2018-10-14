package insteon

import (
	"testing"
	"time"
)

func TestI1DeviceIsDevice(t *testing.T) {
	var device interface{}
	device = &I1Device{}

	if _, ok := device.(Device); !ok {
		t.Errorf("Expected I1Device to be linkable")
	}
}

func TestI1DeviceSend(t *testing.T) {
	tests := []struct {
		payload       []byte
		expectedFlags Flags
	}{
		{nil, StandardDirectMessage},
		{[]byte{1, 2, 3, 4}, ExtendedDirectMessage},
	}

	for i, test := range tests {
		sendCh := make(chan *CommandRequest, 1)
		upstreamSendCh := make(chan *MessageRequest, 1)
		device := &I1Device{
			sendCh:         sendCh,
			upstreamSendCh: upstreamSendCh,
		}

		sendCh <- &CommandRequest{Payload: test.payload}
		close(sendCh)
		device.process()
		if len(upstreamSendCh) != 1 {
			t.Errorf("tests[%d] expected message to be sent upstream", i)
		}

		request := <-upstreamSendCh
		if request.Message.Flags != test.expectedFlags {
			t.Errorf("tests[%d] expected flags to be %v got %v", i, test.expectedFlags, request.Message.Flags)
		}
	}
}

func TestI1DeviceReceive(t *testing.T) {
	recvCh := make(chan *Message, 1)
	requestRecvCh := make(chan *CommandResponse, 1)
	upstreamSendCh := make(chan *MessageRequest, 1)
	waitRequest := &CommandRequest{RecvCh: requestRecvCh}

	device := &I1Device{
		recvCh:         recvCh,
		upstreamSendCh: upstreamSendCh,
		waitRequest:    waitRequest,
	}

	recvCh <- &Message{}
	close(recvCh)
	device.process()
	if len(requestRecvCh) != 1 {
		t.Errorf("expected message to be sent to waiting request")
	}

	var zeroTime time.Time
	if waitRequest.timeout == zeroTime {
		t.Errorf("expected timeout to be set")
	}
}

func TestI1DeviceReceiveAck(t *testing.T) {
	tests := []struct {
		input          *MessageRequest
		recvChSet      bool
		waitRequestSet bool
	}{
		{&MessageRequest{Ack: &Message{}, Err: ErrReadTimeout}, false, false},
		{&MessageRequest{Ack: &Message{}, Err: ErrReadTimeout}, true, false},
		{&MessageRequest{Ack: &Message{}, Err: nil}, true, true},
	}

	for i, test := range tests {
		recvCh := make(chan *Message, 1)
		doneCh := make(chan *MessageRequest, 1)
		requestDoneCh := make(chan *CommandRequest, 1)
		requestRecvCh := make(chan *CommandResponse, 1)
		upstreamSendCh := make(chan *MessageRequest, 1)
		device := &I1Device{
			recvCh:         recvCh,
			upstreamSendCh: upstreamSendCh,
			doneCh:         doneCh,
			queue:          []*CommandRequest{{DoneCh: requestDoneCh}},
		}

		if test.recvChSet {
			device.queue[0].RecvCh = requestRecvCh
		}

		doneCh <- test.input
		close(doneCh)
		device.process()
		cmdRequest := <-requestDoneCh
		if cmdRequest.Ack != test.input.Ack {
			t.Errorf("tests[%d] expected %v got %v", i, test.input.Ack, cmdRequest.Ack)
		}

		if cmdRequest.Err != test.input.Err {
			t.Errorf("tests[%d] expected %v got %v", i, test.input.Err, cmdRequest.Err)
		}

		if test.waitRequestSet {
			if device.waitRequest == nil {
				t.Errorf("tests[%d] expected waitRequest to be set", i)
				select {
				case _, ok := <-requestRecvCh:
					if ok {
						t.Errorf("tests[%d] expected recv channel to be closed", i)
					}
				case <-time.After(time.Second):
					t.Errorf("tests[%d] timeout waiting for channel to close", i)
				}
			} else if device.waitRequest != cmdRequest {
				t.Errorf("tests[%d] expected %v got %v", i, cmdRequest, device.waitRequest)
			}
		}
	}
}
