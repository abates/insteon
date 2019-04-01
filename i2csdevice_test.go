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

func TestChecksum(t *testing.T) {
	tests := []struct {
		desc     string
		input    []byte
		expected byte
	}{
		{"1", []byte{0x2E, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xd1},
		{"2", []byte{0x2F, 0x00, 0x00, 0x00, 0x0F, 0xFF, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xC2},
		{"3", []byte{0x2F, 0x00, 0x01, 0x01, 0x0F, 0xFF, 0x00, 0xA2, 0x00, 0x19, 0x70, 0x1A, 0xFF, 0x1F, 0x01}, 0x5D},
		{"4", []byte{0x2F, 0x00, 0x00, 0x00, 0x0F, 0xF7, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xCA},
		{"5", []byte{0x2F, 0x00, 0x01, 0x01, 0x0F, 0xF7, 0x00, 0xE2, 0x01, 0x19, 0x70, 0x1A, 0xFF, 0x1F, 0x01}, 0x24},
		{"6", []byte{0x2F, 0x00, 0x00, 0x00, 0x0F, 0xEF, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xD2},
		{"7", []byte{0x2F, 0x00, 0x01, 0x01, 0x0F, 0xEF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xD1},
		{"8", []byte{0x2F, 0x00, 0x01, 0x02, 0x0F, 0xFF, 0x08, 0xE2, 0x01, 0x08, 0xB6, 0xEA, 0x00, 0x1B, 0x01}, 0x11},
		{"9", []byte{0x09, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0xF6},
		{"A", []byte{0x2f, 0x00, 0x00, 0x02, 0x0f, 0xff, 0x08, 0xe2, 0x01, 0x08, 0xb6, 0xea, 0x00, 0x1b, 0x01}, 0x12},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := checksum(Command{0x00, test.input[0], test.input[1]}, test.input[2:])
			if got != test.expected {
				t.Errorf("got checksum %02x, want %02x", got, test.expected)
			}
		})
	}
}

/*
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
}*/
