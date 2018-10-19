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

func TestI1DeviceDoneCh(t *testing.T) {
	doneCh := make(chan *CommandResponse, 1)
	requestDoneCh := make(chan *CommandRequest, 1)
	requestRecvCh := make(chan *CommandResponse, 1)
	upstreamSendCh := make(chan *MessageRequest, 1)
	device := &I1Device{
		upstreamSendCh: upstreamSendCh,
		listenDoneCh:   doneCh,
		waitRequest:    &CommandRequest{DoneCh: requestDoneCh, RecvCh: requestRecvCh},
	}

	response := &CommandResponse{request: device.waitRequest}
	doneCh <- response
	close(doneCh)
	device.process()

	if device.waitRequest != nil {
		t.Errorf("expected nil wait request")
	}

	select {
	case _, ok := <-requestRecvCh:
		if ok {
			t.Errorf("expected recv channel to be closed")
		}
	case <-time.After(time.Second):
		t.Errorf("timeout waiting for channel to close")
	}
}

func TestI1DeviceWaitRequestTimeout(t *testing.T) {
	recvCh := make(chan *Message, 1)
	doneCh := make(chan *CommandResponse, 1)
	requestDoneCh := make(chan *CommandRequest, 1)
	requestRecvCh := make(chan *CommandResponse, 1)
	upstreamSendCh := make(chan *MessageRequest, 1)

	device := &I1Device{
		recvCh:         recvCh,
		upstreamSendCh: upstreamSendCh,
		listenDoneCh:   doneCh,
		waitRequest:    &CommandRequest{DoneCh: requestDoneCh, RecvCh: requestRecvCh},
	}

	go func() {
		_, open := <-requestRecvCh
		if open {
			t.Errorf("expected closed channel")
		}
		close(recvCh)
	}()

	device.process()
	if device.waitRequest != nil {
		t.Errorf("expected waitRequest to be nil")
	}
}

func TestI1DeviceQueueTimeout(t *testing.T) {
	recvCh := make(chan *Message, 1)
	doneCh := make(chan *CommandResponse, 1)
	requestDoneCh := make(chan *CommandRequest, 1)
	upstreamSendCh := make(chan *MessageRequest, 1)

	device := &I1Device{
		recvCh:         recvCh,
		upstreamSendCh: upstreamSendCh,
		listenDoneCh:   doneCh,
		queue:          []*CommandRequest{{DoneCh: requestDoneCh}},
	}

	go func() {
		<-requestDoneCh
		close(recvCh)
	}()

	device.process()
	if len(device.queue) > 0 {
		t.Errorf("expected queue to be empty")
	}
}

func TestI1DeviceAddress(t *testing.T) {
	expected := Address{3, 4, 5}
	device := &I1Device{address: expected}
	if device.Address() != expected {
		t.Errorf("expected %v got %v", expected, device.Address())
	}
}

func TestI1DeviceCommands(t *testing.T) {
	tests := []struct {
		callback    func(*I1Device) error
		expectedCmd Command
		expectedErr error
	}{
		{func(i1 *I1Device) error { return i1.AssignToAllLinkGroup(10) }, CmdAssignToAllLinkGroup.SubCommand(10), nil},
		{func(i1 *I1Device) error { return i1.AssignToAllLinkGroup(10) }, CmdAssignToAllLinkGroup.SubCommand(10), ErrReadTimeout},
		{func(i1 *I1Device) error { return i1.DeleteFromAllLinkGroup(10) }, CmdDeleteFromAllLinkGroup.SubCommand(10), nil},
		{func(i1 *I1Device) error { return i1.Ping() }, CmdPing, nil},
		{func(i1 *I1Device) error { return i1.SetAllLinkCommandAlias(Command{}, Command{}) }, Command{}, ErrNotImplemented},
		{func(i1 *I1Device) error { return i1.SetAllLinkCommandAliasData(nil) }, Command{}, ErrNotImplemented},
	}

	for i, test := range tests {
		sendCh := make(chan *CommandRequest, 1)
		device := &I1Device{
			sendCh: sendCh,
		}

		if test.expectedErr != ErrNotImplemented {
			go func() {
				request := <-sendCh
				if request.Command != test.expectedCmd {
					t.Errorf("tests[%d] expected Command %v got %v", i, test.expectedCmd, request.Command)
				}
				if test.expectedErr != nil {
					request.Err = test.expectedErr
				} else {
					request.Ack = &Message{Command: test.expectedCmd}
				}
				request.DoneCh <- request
			}()
		}

		err := test.callback(device)
		if err != test.expectedErr {
			t.Errorf("tests[%d] expected %v got %v", i, test.expectedErr, err)
		}
	}
}

func TestI1DeviceProductData(t *testing.T) {
	sendCh := make(chan *CommandRequest, 1)
	device := &I1Device{
		sendCh: sendCh,
	}

	expected := ProductData{ProductKey{1, 2, 3}, DevCat{4, 5}}

	go func() {
		request := <-sendCh
		request.Ack = &Message{}
		request.DoneCh <- request
		testRecv(request.RecvCh, CmdProductDataResp, &expected)
	}()

	pd, err := device.ProductData()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if *pd != expected {
		t.Errorf("Expected %v got %v", expected, *pd)
	}
}

func TestI1DeviceBlockDataTransfer(t *testing.T) {
	device := &I1Device{}
	_, err := device.BlockDataTransfer(0, 0, 0)
	if err != ErrNotImplemented {
		t.Errorf("expected %v got %v", ErrNotImplemented, err)
	}
}

func TestI1DeviceString(t *testing.T) {
	device := &I1Device{address: Address{3, 4, 5}}
	expected := "I1 Device (03.04.05)"
	if device.String() != expected {
		t.Errorf("expected %q got %q", expected, device.String())
	}
}
