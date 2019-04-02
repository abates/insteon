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

package network

/*
import (
	"reflect"
	"testing"
	"time"

	"github.com/abates/insteon"
)

var (
	testSrcAddr = insteon.Address{1, 2, 3}
	testDstAddr = insteon.Address{3, 4, 5}

	TestMessageSetButtonPressedController = &insteon.Message{testDstAddr, testSrcAddr, insteon.StandardBroadcast, insteon.Command{0x00, 0x02, 0xff}, nil}
	TestMessageEngineVersionAck           = &insteon.Message{testDstAddr, testSrcAddr, insteon.StandardDirectAck, insteon.Command{0x00, 0x0d, 0x01}, nil}

	TestProductDataResponse = &insteon.Message{testDstAddr, testSrcAddr, insteon.ExtendedDirectMessage, insteon.CmdProductDataResp, []byte{0, 1, 2, 3, 4, 5, 0xff, 0xff, 0, 0, 0, 0, 0, 0}}
)

type testBridge struct {
	rx      chan []byte
	send    chan []byte
	sendErr error
}

func (tb *testBridge) Send(buf []byte) error {
	tb.send <- buf
	return tb.sendErr
}

func (tb *testBridge) Receive() <-chan []byte {
	return tb.rx
}

func newTestNetwork() (*Network, *testBridge) {
	bridge := &testBridge{
		rx:   make(chan []byte),
		send: make(chan []byte, 1),
	}
	return New(bridge, time.Millisecond), bridge
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
}

type testConnection struct {
	rxCh chan *insteon.Message
}

func (tc *testConnection) Send(msg *insteon.Message) (*insteon.Message, error) { return nil, nil }
func (tc *testConnection) Push(msg *insteon.Message)                           { tc.rxCh <- msg }
func (tc *testConnection) Receive() (*insteon.Message, error) {
	return <-tc.rxCh, nil
}

func TestNetworkReceive(t *testing.T) {

	tests := []struct {
		desc            string
		input           *insteon.Message
		expectedUpdates []string
	}{
		{"SetButtonPressed", TestMessageSetButtonPressedController, []string{"FirmwareVersion", "DevCat"}},
		{"EngineVersion", TestMessageEngineVersionAck, []string{"EngineVersion"}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			testDb := newTestProductDB()
			conn := &testConnection{
				rxCh: make(chan *insteon.Message, 1),
			}

			bridge := &testBridge{rx: make(chan []byte)}

			network := &Network{
				bridge:      bridge,
				DB:          testDb,
				connections: []insteon.Connection{conn},
				closeCh:     make(chan chan error, 1),
			}

			go network.process()
			buf, _ := test.input.MarshalBinary()
			bridge.rx <- buf
			closeCh := make(chan error)
			network.closeCh <- closeCh
			<-closeCh

			for _, update := range test.expectedUpdates {
				if !testDb.WasUpdated(update) {
					t.Errorf("expected %v to be updated in the database", update)
				}
			}

			if len(conn.rxCh) != 1 {
				t.Error("expected connection to have received the message")
			}
		})
	}
}

func TestNetworkEngineVersion(t *testing.T) {
	tests := []struct {
		desc            string
		returnedAck     *insteon.Message
		returnedErr     error
		expectedVersion insteon.EngineVersion
	}{
		{"v1", TestMessageEngineVersionAck, nil, 1},
		{"v2", TestMessageEngineVersionAck, nil, 2},
		{"v3", TestMessageEngineVersionAck, nil, 3},
		{"err", TestMessageEngineVersionAck, insteon.ErrReadTimeout, 0},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			network, bridge := newTestNetwork()

			go func() {
				if test.returnedErr == nil {
					<-bridge.send
					ack := *test.returnedAck
					ack.Src = testDstAddr
					ack.Command = insteon.Command{0x00, ack.Command[1], byte(test.expectedVersion)}
					buf, _ := ack.MarshalBinary()
					bridge.rx <- buf
				}
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
		expectedDevCat   insteon.DevCat
		expectedFirmware insteon.FirmwareVersion
	}{
		{"1", false, insteon.ErrReadTimeout, insteon.ErrReadTimeout, insteon.DevCat{0, 0}, 0},
		{"2", false, nil, nil, insteon.DevCat{1, 2}, 3},
		{"3", false, nil, nil, insteon.DevCat{2, 3}, 4},
		{"4", true, nil, insteon.ErrReadTimeout, insteon.DevCat{0, 0}, 0},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			network, bridge := newTestNetwork()

			go func() {
				buf := <-bridge.send
				if test.returnedErr == nil {
					// the test has to send an ACK, since the device would ack the set button pressed
					// command before sending a broadcast response
					ack := &insteon.Message{}
					ack.UnmarshalBinary(buf)
					src := ack.Dst
					ack.Dst = ack.Src
					ack.Src = src
					ack.Flags = insteon.StandardDirectAck
					buf, _ := ack.MarshalBinary()
					bridge.rx <- buf

					if !test.timeout {
						// send the broadcast
						msg := *TestMessageSetButtonPressedController
						msg.Src = src
						msg.Dst = insteon.Address{test.expectedDevCat[0], test.expectedDevCat[1], byte(test.expectedFirmware)}
						buf, _ := msg.MarshalBinary()
						bridge.rx <- buf
					}
				}
			}()

			info, err := network.IDRequest(testDstAddr)

			if err != test.expectedErr {
				t.Errorf("got error %v, want %v", err, test.expectedErr)
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
		deviceInfo    *insteon.DeviceInfo
		engineVersion byte
		sendError     error
		expectedErr   error
		expected      interface{}
	}{
		{"1", &insteon.DeviceInfo{EngineVersion: insteon.VerI1}, 0, nil, nil, &insteon.I1Device{}},
		{"2", &insteon.DeviceInfo{EngineVersion: insteon.VerI2}, 0, nil, nil, &insteon.I2Device{}},
		{"3", &insteon.DeviceInfo{EngineVersion: insteon.VerI2Cs}, 0, nil, nil, &insteon.I2CsDevice{}},
		{"4", nil, 0, nil, nil, &insteon.I1Device{}},
		{"5", nil, 1, nil, nil, &insteon.I2Device{}},
		{"6", nil, 2, nil, nil, &insteon.I2CsDevice{}},
		{"7", nil, 3, nil, insteon.ErrVersion, nil},
		{"8", nil, 0, insteon.ErrNotLinked, insteon.ErrNotLinked, &insteon.I2CsDevice{}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			testDb := newTestProductDB()
			network, bridge := newTestNetwork()
			network.DB = testDb

			if test.deviceInfo == nil {
				bridge.sendErr = test.sendError
				go func() {
					<-bridge.send
					if test.sendError == nil {
						msg := *TestMessageEngineVersionAck
						msg.Src = insteon.Address{1, 2, 3}
						msg.Command = insteon.Command{0x00, msg.Command[1], byte(test.engineVersion)}
						buf, _ := msg.MarshalBinary()
						bridge.rx <- buf
					}
				}()
			} else {
				testDb.deviceInfo = test.deviceInfo
			}

			device, err := network.Dial(insteon.Address{1, 2, 3})

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
	network, _ := newTestNetwork()
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
