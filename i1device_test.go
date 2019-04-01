package insteon

/*

func TestI1DeviceIsDevice(t *testing.T) {
	var device interface{}
	device = &I1Device{}

	if _, ok := device.(Device); !ok {
		t.Error("Expected I1Device to be linkable")
	}
}

func TestI1DeviceSend(t *testing.T) {
	tests := []struct {
		desc          string
		payload       []byte
		expectedFlags Flags
	}{
		{"SD", nil, StandardDirectMessage},
		{"ED", []byte{1, 2, 3, 4}, ExtendedDirectMessage},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
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
				t.Error("expected message to be sent upstream")
			}

			request := <-upstreamSendCh
			if request.Message.Flags != test.expectedFlags {
				t.Errorf("got %v, want flags %v", request.Message.Flags, test.expectedFlags)
			}
		})
	}
}

func TestI1DeviceReceive(t *testing.T) {
	// This test is flaky about 15% of the time.
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
		t.Error("expected message to be sent to waiting request")
	}

	var zeroTime time.Time
	if waitRequest.timeout == zeroTime {
		t.Error("expected timeout to be set")
	}
}

func TestI1DeviceReceiveAck(t *testing.T) {
	tests := []struct {
		desc           string
		input          *MessageRequest
		recvChSet      bool
		waitRequestSet bool
	}{
		{"read timeout - no recv", &MessageRequest{Ack: &Message{}, Err: ErrReadTimeout}, false, false},
		{"read timeout - recv", &MessageRequest{Ack: &Message{}, Err: ErrReadTimeout}, true, false},
		{"waitRequest", &MessageRequest{Ack: &Message{}, Err: nil}, true, true},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
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
				t.Errorf("got %v, want %v", cmdRequest.Ack, test.input.Ack)
			}

			if cmdRequest.Err != test.input.Err {
				t.Errorf("got %v, want %v", cmdRequest.Err, test.input.Err)
			}

			if test.waitRequestSet {
				if device.waitRequest == nil {
					t.Error("expected waitRequest to be set")
					select {
					case _, ok := <-requestRecvCh:
						if ok {
							t.Error("expected recv channel to be closed")
						}
					case <-time.After(time.Second):
						t.Error("timeout waiting for channel to close")
					}
				} else if device.waitRequest != cmdRequest {
					t.Errorf("got %v, want %v", device.waitRequest, cmdRequest)
				}
			}
		})
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
		t.Error("expected nil wait request")
	}

	select {
	case _, ok := <-requestRecvCh:
		if ok {
			t.Error("expected recv channel to be closed")
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for channel to close")
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
			t.Error("expected closed channel")
		}
		close(recvCh)
	}()

	device.process()
	if device.waitRequest != nil {
		t.Error("expected waitRequest to be nil")
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
		t.Error("expected queue to be empty")
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
		desc        string
		callback    func(*I1Device) error
		expectedCmd Command
		expectedErr error
	}{
		{"AssignToAllLinkGroup", func(i1 *I1Device) error { return i1.AssignToAllLinkGroup(10) }, CmdAssignToAllLinkGroup.SubCommand(10), nil},
		{"AssignToAllLinkGroup ReadTimeout", func(i1 *I1Device) error { return i1.AssignToAllLinkGroup(10) }, CmdAssignToAllLinkGroup.SubCommand(10), ErrReadTimeout},
		{"DeleteFromAllLinkGroup", func(i1 *I1Device) error { return i1.DeleteFromAllLinkGroup(10) }, CmdDeleteFromAllLinkGroup.SubCommand(10), nil},
		{"CmdPing", func(i1 *I1Device) error { return i1.Ping() }, CmdPing, nil},
		{"SetAllLinkCommandAlias NotImplemented", func(i1 *I1Device) error { return i1.SetAllLinkCommandAlias(Command{}, Command{}) }, Command{}, ErrNotImplemented},
		{"SetAllLinkCommandAlias NotImplemented 2", func(i1 *I1Device) error { return i1.SetAllLinkCommandAliasData(nil) }, Command{}, ErrNotImplemented},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			sendCh := make(chan *CommandRequest, 1)
			device := &I1Device{
				sendCh: sendCh,
			}

			if test.expectedErr != ErrNotImplemented {
				go func() {
					request := <-sendCh
					if request.Command != test.expectedCmd {
						t.Errorf("got Command %v, want %v", request.Command, test.expectedCmd)
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
				t.Errorf("got error %v, want %v", err, test.expectedErr)
			}
		})
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
}*/
