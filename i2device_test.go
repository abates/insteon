// Copyright 2018 Andrew Bates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package insteon

import "testing"

func TestI2DeviceIsLinkable(t *testing.T) {
	device := Device(&I2Device{})
	linkable := device.(LinkableDevice)
	if linkable == nil {
		t.Errorf("linkable should not be nil")
	}
}

func TestI2DeviceCommands(t *testing.T) {
	tests := []struct {
		callback    func(*I2Device) error
		expectedCmd Command
		expectedErr error
	}{
		{func(i2cs *I2Device) error { return i2cs.AddLink(nil) }, Command{}, ErrNotImplemented},
		{func(i2cs *I2Device) error { return i2cs.RemoveLinks(nil) }, Command{}, ErrNotImplemented},
		{func(i2cs *I2Device) error { return i2cs.EnterLinkingMode(10) }, CmdEnterLinkingMode.SubCommand(10), nil},
		{func(i2cs *I2Device) error { return i2cs.EnterUnlinkingMode(10) }, CmdEnterUnlinkingMode.SubCommand(10), nil},
		{func(i2cs *I2Device) error { return i2cs.ExitLinkingMode() }, CmdExitLinkingMode, nil},
		{func(i2cs *I2Device) error { return i2cs.WriteLink(&LinkRecord{}) }, CmdReadWriteALDB, ErrInvalidMemAddress},
		{func(i2cs *I2Device) error { return i2cs.WriteLink(&LinkRecord{memAddress: 0x01}) }, CmdReadWriteALDB, nil},
	}

	for i, test := range tests {
		sendCh := make(chan *CommandRequest, 1)
		device := &I2Device{&I1Device{sendCh: sendCh}}

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

func TestI2DeviceLinks(t *testing.T) {
	sendCh := make(chan *CommandRequest, 1)
	device := &I2Device{&I1Device{sendCh: sendCh}}

	link1 := &LinkRequest{MemAddress: 0xffff, Type: 0x02, Link: &LinkRecord{Flags: 0x01}}
	link2 := &LinkRequest{MemAddress: 0, Type: 0x02, Link: &LinkRecord{}}

	go func() {
		request := <-sendCh
		request.Ack = &Message{}
		request.DoneCh <- request
		testRecv(request.RecvCh, CmdReadWriteALDB, link1, link2)
	}()

	links, err := device.Links()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if len(links) != 1 {
		t.Errorf("Expected 1 link got %v", len(links))
	}
}

func TestI2DeviceString(t *testing.T) {
	device := &I2Device{&I1Device{address: Address{3, 4, 5}}}
	expected := "I2 Device (03.04.05)"
	if device.String() != expected {
		t.Errorf("expected %q got %q", expected, device.String())
	}
}
