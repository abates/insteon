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

func TestI2CsDeviceCommands(t *testing.T) {
	tests := []struct {
		desc        string
		callback    func(*I2CsDevice) error
		expectedCmd Command
		expectedErr error
	}{
		{"EnterLinkingMode", func(i2cs *I2CsDevice) error { return i2cs.EnterLinkingMode(15) }, CmdEnterLinkingModeExt.SubCommand(15), nil},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			sendCh := make(chan *CommandRequest, 1)
			device := &I2CsDevice{&I2Device{&I1Device{sendCh: sendCh}}}

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

func TestI2CsSendCommand(t *testing.T) {
	sendCh := make(chan *CommandRequest, 1)
	device := &I2CsDevice{&I2Device{&I1Device{sendCh: sendCh}}}
	go func() {
		request := <-sendCh
		request.Ack = &Message{}
		request.DoneCh <- request
		if len(request.Payload) != 14 {
			t.Error("Expected payload to be set")
		}
	}()
	device.SendCommand(CmdSetOperatingFlags, nil)
}

func TestI2CsDeviceString(t *testing.T) {
	device := &I2CsDevice{&I2Device{&I1Device{address: Address{3, 4, 5}}}}
	expected := "I2CS Device (03.04.05)"
	if device.String() != expected {
		t.Errorf("expected %q got %q", expected, device.String())
	}
}
