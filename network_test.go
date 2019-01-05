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

import (
	"reflect"
	"testing"
	"time"
)

func newTestNetwork(bufSize int) (*Network, chan *PacketRequest, chan []byte) {
	sendCh := make(chan *PacketRequest, bufSize)
	recvCh := make(chan []byte, bufSize)
	return New(sendCh, recvCh, time.Millisecond), sendCh, recvCh
}

/*func TestNetworkProcess(t *testing.T) {
	connection := make(chan *Message, 1)
	network, _, recvCh := newTestNetwork(0)
	network.connectCh <- connection

	buf, _ := TestMessagePingAck.MarshalBinary()
	recvCh <- buf
	select {
	case <-connection:
	case <-time.After(time.Millisecond):
		t.Error("Expected connection to receive a message")
	}

	network.disconnectCh <- connection
	network.Close()

	if len(network.connections) != 0 {
		t.Errorf("Expected connnection queue to be empty, got %d", len(network.connections))
	}
}*/

func TestNetworkReceive(t *testing.T) {
	tests := []struct {
		desc            string
		input           *Message
		expectedUpdates []string
	}{
		{"SetButtonPressed", TestMessageSetButtonPressedController, []string{"FirmwareVersion", "DevCat"}},
		{"EngineVersion", TestMessageEngineVersionAck, []string{"EngineVersion"}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			recvCh := make(chan []byte, 1)
			testDb := newTestProductDB()
			connection := make(chan *Message, 1)

			network := &Network{
				recvCh:      recvCh,
				DB:          testDb,
				connections: []chan<- *Message{connection},
			}

			buf, _ := test.input.MarshalBinary()
			recvCh <- buf
			close(recvCh)
			network.process()

			for _, update := range test.expectedUpdates {
				if !testDb.WasUpdated(update) {
					t.Errorf("expected %v to be updated in the database", update)
				}
			}

			if len(connection) != 1 {
				t.Error("expected connection to have received the message")
			}
		})
	}
}

func TestNetworkSendMessage(t *testing.T) {
	tests := []struct {
		desc       string
		input      *Message
		timeout    bool
		err        error
		deviceInfo *DeviceInfo
		bufUpdated bool
	}{
		{"VerI1", TestProductDataResponse, false, nil, &DeviceInfo{EngineVersion: VerI1}, false},
		{"VerI2Cs", TestProductDataResponse, false, nil, &DeviceInfo{EngineVersion: VerI2Cs}, true},
		{"ErrReadTimeout", TestProductDataResponse, false, ErrReadTimeout, nil, false},
		{"ErrSendTimeout", TestProductDataResponse, true, ErrSendTimeout, nil, false},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			sendCh := make(chan *PacketRequest, 1)
			testDb := newTestProductDB()
			testDb.deviceInfo = test.deviceInfo
			network := &Network{
				DB:      testDb,
				sendCh:  sendCh,
				timeout: time.Millisecond,
			}

			go func() {
				if !test.timeout {
					request := <-sendCh
					if test.bufUpdated && request.Payload[len(request.Payload)-1] == 0x00 {
						t.Error("expected checksum to be set")
					}
					request.Err = test.err
					request.DoneCh <- request
				}
			}()

			err := network.sendMessage(test.input)
			if err != test.err {
				t.Errorf("got error %v, want %v", err, test.err)
			}
		})
	}
}

func TestNetworkEngineVersion(t *testing.T) {
	tests := []struct {
		desc            string
		returnedAck     *Message
		returnedErr     error
		expectedVersion EngineVersion
	}{
		{"v1", TestMessageEngineVersionAck, nil, 1},
		{"v2", TestMessageEngineVersionAck, nil, 2},
		{"v3", TestMessageEngineVersionAck, nil, 3},
		{"err", TestMessageEngineVersionAck, ErrReadTimeout, 0},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			sendCh := make(chan *PacketRequest, 1)
			recvCh := make(chan []byte, 1)
			network := New(sendCh, recvCh, time.Millisecond)

			go func() {
				request := <-sendCh
				if test.returnedErr == nil {
					ack := *test.returnedAck
					ack.Src = testDstAddr
					ack.Command = Command{0x00, ack.Command[1], byte(test.expectedVersion)}
					buf, _ := ack.MarshalBinary()
					recvCh <- buf
				} else {
					request.Err = test.returnedErr
				}
				request.DoneCh <- request
			}()

			version, err := network.EngineVersion(testDstAddr)

			if err != test.returnedErr {
				t.Errorf("got error %v, want %v", err, test.returnedErr)
			}

			if version != test.expectedVersion {
				t.Errorf("got version %v, want %v", version, test.expectedVersion)
			}
			network.Close()
		})
	}
}

func TestNetworkIDRequest(t *testing.T) {
	tests := []struct {
		desc             string
		timeout          bool
		returnedErr      error
		expectedErr      error
		expectedDevCat   DevCat
		expectedFirmware FirmwareVersion
	}{
		{"1", false, ErrReadTimeout, ErrReadTimeout, DevCat{0, 0}, 0},
		{"2", false, nil, nil, DevCat{1, 2}, 3},
		{"3", false, nil, nil, DevCat{2, 3}, 4},
		{"4", true, nil, ErrReadTimeout, DevCat{0, 0}, 0},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			sendCh := make(chan *PacketRequest, 1)
			recvCh := make(chan []byte, 1)
			network := New(sendCh, recvCh, time.Millisecond)

			go func() {
				request := <-sendCh
				if test.returnedErr == nil {
					// the test has to send an ACK, since the device would ack the set button pressed
					// command before sending a broadcast response
					ack := &Message{}
					ack.UnmarshalBinary(request.Payload)
					src := ack.Dst
					ack.Dst = ack.Src
					ack.Src = src
					ack.Flags = StandardDirectAck
					buf, _ := ack.MarshalBinary()
					recvCh <- buf

					if !test.timeout {
						// send the broadcast
						msg := *TestMessageSetButtonPressedController
						msg.Src = src
						msg.Dst = Address{test.expectedDevCat[0], test.expectedDevCat[1], byte(test.expectedFirmware)}
						buf, _ = msg.MarshalBinary()
						recvCh <- buf
					}
				} else {
					request.Err = test.returnedErr
				}
				request.DoneCh <- request
			}()

			info, err := network.IDRequest(testDstAddr)

			if err != test.expectedErr {
				t.Errorf("got error %v, want %v", err, test.returnedErr)
			}

			if info.FirmwareVersion != test.expectedFirmware {
				t.Errorf("got FirmwareVersion %v, want %v", info.FirmwareVersion, test.expectedFirmware)
			}

			if info.DevCat != test.expectedDevCat {
				t.Errorf("got DevCat %v, want %v", info.DevCat, test.expectedDevCat)
			}
			network.Close()
		})
	}
}

func TestNetworkDial(t *testing.T) {
	tests := []struct {
		desc          string
		deviceInfo    *DeviceInfo
		engineVersion byte
		sendError     error
		expectedErr   error
		expected      interface{}
	}{
		{"1", &DeviceInfo{EngineVersion: VerI1}, 0, nil, nil, &I1Device{}},
		{"2", &DeviceInfo{EngineVersion: VerI2}, 0, nil, nil, &I2Device{}},
		{"3", &DeviceInfo{EngineVersion: VerI2Cs}, 0, nil, nil, &I2CsDevice{}},
		{"4", nil, 0, nil, nil, &I1Device{}},
		{"5", nil, 1, nil, nil, &I2Device{}},
		{"6", nil, 2, nil, nil, &I2CsDevice{}},
		{"7", nil, 3, nil, ErrVersion, nil},
		{"8", nil, 0, ErrNotLinked, ErrNotLinked, &I2CsDevice{}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			testDb := newTestProductDB()
			network, sendCh, recvCh := newTestNetwork(1)
			network.DB = testDb

			if test.deviceInfo == nil {
				go func() {
					request := <-sendCh
					if test.sendError == nil {
						msg := *TestMessageEngineVersionAck
						msg.Src = Address{1, 2, 3}
						msg.Command = Command{0x00, msg.Command[1], byte(test.engineVersion)}
						buf, _ := msg.MarshalBinary()
						recvCh <- buf
						request.DoneCh <- request
					} else {
						request.Err = test.sendError
						request.DoneCh <- request
					}
				}()
			} else {
				testDb.deviceInfo = test.deviceInfo
			}

			device, err := network.Dial(Address{1, 2, 3})

			if err != test.expectedErr {
				t.Errorf("got error %v, want %v", err, test.expectedErr)
			} else if reflect.TypeOf(device) != reflect.TypeOf(test.expected) {
				t.Fatalf("got type %T, want type %T", device, test.expected)
			}

			network.Close()
		})
	}
}

func TestNetworkClose(t *testing.T) {
	network, _, _ := newTestNetwork(1)
	network.Close()

	select {
	case _, open := <-network.closeCh:
		if open {
			t.Error("Expected closeCh to be closed")
		}
	default:
		t.Error("Expected read from closeCh to indicate a closed channel")
	}
}

/*
func TestNetworkConnect(t *testing.T) {
	tests := []struct {
		deviceInfo    *DeviceInfo
		engineVersion EngineVersion
		dst           Address
		expected      Device
	}{
		{&DeviceInfo{DevCat: DevCat{42, 2}}, VerI1, Address{}, &I2Device{}},
		{nil, VerI1, Address{42, 2, 3}, &I2Device{}},
	}

	for _, test := range tests {
		var category Category
		testDb := newTestProductDB()
		bridge := &testBridge{}

		if test.deviceInfo == nil {
			msg := *TestMessageEngineVersionAck
			msg.Command[1] = byte(test.engineVersion)

			msg = *TestMessageSetButtonPressedController
			msg.Dst = test.dst
			category = Category(test.dst[0])
		} else {
			testDb.deviceInfo = test.deviceInfo
			category = test.deviceInfo.DevCat.Category()
		}
		Devices.Register(category, func(Device, DeviceInfo) Device { return test.expected })

		network := &NetworkImpl{
			db:     testDb,
			bridge: bridge,
		}

		device, _ := network.Connect(Address{1, 2, 3})

		if device != test.expected {
			t.Fatalf("expected %v got %v", test.expected, device)
		}
		Devices.Delete(category)
	}
}*/
