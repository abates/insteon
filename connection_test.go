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
	"io"
	"reflect"
	"sync"
	"testing"
	"time"
)

type testConnection struct {
	sync.Mutex

	addr             Address
	devCat           DevCat
	firmwareVersion  FirmwareVersion
	engineVersion    EngineVersion
	engineVersionErr error

	sendCh  chan *Message
	ackCh   chan *Message
	sendErr error

	recvCh  chan *Message
	recvErr error
}

func (tc *testConnection) Address() Address { return tc.addr }
func (tc *testConnection) EngineVersion() (EngineVersion, error) {
	return tc.engineVersion, tc.engineVersionErr
}
func (tc *testConnection) IDRequest() (FirmwareVersion, DevCat, error) {
	return tc.firmwareVersion, tc.devCat, nil
}

func (tc *testConnection) SendCommand(cmd Command, payload []byte) (Command, error) {
	msg, err := tc.Send(&Message{Command: cmd, Payload: payload})
	return msg.Command, err
}

func (tc *testConnection) Send(msg *Message) (*Message, error) {
	tc.sendCh <- msg
	if tc.sendErr != nil {
		return nil, tc.sendErr
	}
	return <-tc.ackCh, nil
}

func (tc *testConnection) Receive() (*Message, error) {
	if tc.recvErr == nil {
		return <-tc.recvCh, nil
	}
	return nil, tc.recvErr
}

func (tc *testConnection) AddListener(MessageType, ...Command) <-chan *Message { return nil }
func (tc *testConnection) RemoveListener(<-chan *Message)                      {}

func TestConnectionOptions(t *testing.T) {
	mu := &sync.Mutex{}
	tests := []struct {
		desc  string
		input ConnectionOption
		want  *connection
	}{
		{"Timeout Option", ConnectionTimeout(time.Hour), &connection{timeout: time.Hour}},
		{"Filter Option", ConnectionFilter(CmdReadWriteALDB), &connection{match: []Command{CmdReadWriteALDB}}},
		{"Mutex Option", ConnectionMutex(mu), &connection{Mutex: mu}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := &connection{}
			test.input(got)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("want connection %+v got %+v", test.want, got)
			}
		})
	}
}

func TestConnectionSend(t *testing.T) {
	tests := []struct {
		name        string
		input       *Message
		expectedErr error
	}{
		{"I1 Send", TestProductDataResponse, nil},
		{"I2 Send", TestProductDataResponse, nil},
		{"I2Cs Send", TestProductDataResponse, nil},
		{"I2Cs Send", TestProductDataResponse, ErrReadTimeout},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			txCh := make(chan *Message, 1)
			rxCh := make(chan *Message, 1)
			conn := NewConnection(txCh, rxCh, Address{}, ConnectionTimeout(time.Millisecond))
			go func() {
				<-txCh
				if test.expectedErr == nil {
					ack := *test.input
					src := ack.Src
					ack.Src = ack.Dst
					ack.Dst = src
					ack.Flags = StandardDirectAck
					if test.input.Flags.Extended() {
						ack.Flags = ExtendedDirectAck
					}
					rxCh <- &ack
				}
			}()

			_, err := conn.Send(test.input)
			if err != test.expectedErr {
				t.Errorf("Want %v got %v", test.expectedErr, err)
			}
			if closer, ok := conn.(io.Closer); ok {
				closer.Close()
			}
		})
	}
}

func TestConnectionReceive(t *testing.T) {
	tests := []struct {
		name        string
		input       *Message
		match       Command
		expectedErr error
	}{
		{"ReadTimeout 1", &Message{Command: Command{0x00, 0x00, 0x00}}, Command{0x00, 0x01, 0x01}, ErrReadTimeout},
		{"ReadTimeout 2", &Message{Command: Command{0x00, 0x01, 0xff}}, Command{0x00, 0x01, 0x01}, ErrReadTimeout},
		{"Match 1", &Message{Command: Command{0x00, 0x01, 0x01}}, Command{0x00, 0x01, 0x01}, nil},
		{"Match 2", &Message{Command: Command{0x00, 0x01, 0x01}}, Command{0x00, 0x01, 0x00}, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			txCh := make(chan *Message, 1)
			rxCh := make(chan *Message, 1)
			rxCh <- test.input
			conn := NewConnection(txCh, rxCh, Address{}, ConnectionFilter(test.match), ConnectionTimeout(time.Millisecond))
			_, err := conn.Receive()

			if test.expectedErr != err {
				t.Errorf("want %v got %v", test.expectedErr, err)
			}
			if closer, ok := conn.(io.Closer); ok {
				closer.Close()
			}
		})
	}
}

func TestConnectionIDRequest(t *testing.T) {
	txCh := make(chan *Message, 1)
	conn := &connection{txCh: txCh, msgCh: make(chan *Message, 2), timeout: time.Nanosecond}

	wantVersion := FirmwareVersion(42)
	wantDevCat := DevCat{07, 79}

	conn.msgCh <- TestAck
	conn.msgCh <- &Message{Dst: Address{07, 79, 42}, Command: Command{0, 1}, Flags: StandardBroadcast}

	gotVersion, gotDevCat, err := conn.IDRequest()
	<-txCh
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if gotVersion != wantVersion {
		t.Errorf("Want FirmwareVersion %v got %v", wantVersion, gotVersion)
	} else if gotDevCat != wantDevCat {
		t.Errorf("Want DevCat %v got %v", wantDevCat, gotDevCat)
	}

	// sad path
	go func() {
		conn.msgCh <- TestAck
		conn.msgCh <- TestMessagePing
		conn.msgCh <- TestMessagePing
		conn.msgCh <- TestMessagePing
		conn.msgCh <- TestMessagePing
		conn.msgCh <- TestMessagePing
	}()

	_, _, err = conn.IDRequest()
	if err != ErrReadTimeout {
		t.Errorf("Want ErrReadTimeout got %v", err)
	}
}

func TestConnectionEngineVersion(t *testing.T) {
	tests := []struct {
		desc        string
		input       *Message
		wantVersion EngineVersion
		wantErr     error
	}{
		{"Regular device", &Message{Command: CmdGetEngineVersion.SubCommand(42), Flags: StandardDirectAck}, EngineVersion(42), nil},
		{"I2Cs device", &Message{Command: CmdGetEngineVersion.SubCommand(0xff), Flags: StandardDirectNak}, VerI2Cs, ErrNotLinked},
		{"NAK", &Message{Command: CmdGetEngineVersion.SubCommand(0xfd), Flags: StandardDirectNak}, VerI2Cs, ErrNak},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			txCh := make(chan *Message, 1)
			conn := &connection{txCh: txCh, msgCh: make(chan *Message, 1), timeout: time.Nanosecond}

			conn.msgCh <- test.input

			gotVersion, err := conn.EngineVersion()
			<-txCh
			if err != test.wantErr {
				t.Errorf("want error %v got %v", test.wantErr, err)
			} else if err == nil {
				if gotVersion != test.wantVersion {
					t.Errorf("Want EngineVersion %v got %v", test.wantVersion, gotVersion)
				}
			}
		})
	}
}

func TestMsgListeners(t *testing.T) {
	ml := &msgListeners{listeners: make(map[<-chan *Message]*msgListener)}
	ch1 := ml.AddListener(MsgTypeDirect, CmdPing)
	ch2 := ml.AddListener(MsgTypeBroadcast, CmdPing)
	ch3 := ml.AddListener(MsgTypeDirect, CmdReadWriteALDB)

	if len(ml.listeners) != 3 {
		t.Errorf("Expected 3 listeners to be set")
	}

	ml.deliver(TestMessagePing)
	if len(ch1) != 1 {
		t.Errorf("Expected Ping message to be delivered to first channel")
	}

	if len(ch2) != 0 {
		t.Errorf("Expected Ping to not be delivered to second channel")
	}

	if len(ch3) != 0 {
		t.Errorf("Expected Ping to not be delivered to third channel")
	}

	ml.RemoveListener(ch1)
	ml.RemoveListener(ch2)
	ml.RemoveListener(ch3)

	if len(ml.listeners) != 0 {
		t.Errorf("Expected listeners to be empty")
	}
}
