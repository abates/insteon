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
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

type testSender struct {
	sndMsg []*Message
	sndErr error
	sndCb  func()
}

func (tsr *testSender) Send(msg *Message) error {
	tsr.sndMsg = append(tsr.sndMsg, msg)
	if tsr.sndCb != nil {
		tsr.sndCb()
	}
	return tsr.sndErr
}

type testConnection struct {
	sync.Mutex

	addr             Address
	devCat           DevCat
	firmwareVersion  FirmwareVersion
	engineVersion    EngineVersion
	engineVersionErr error

	sent       []*Message
	acks       []*Message
	dispatched []*Message
	sendErr    error

	recv    []*Message
	recvErr error
}

func (tc *testConnection) Address() Address { return tc.addr }

func (tc *testConnection) LinkDatabase() (Linkable, error) { return nil, nil }

func (tc *testConnection) AddHandler(ch chan<- *Message, cmds ...Command) {
	go func() {
		for len(tc.recv) > 0 {
			if tc.recvErr == nil {
				ch <- tc.recv[0]
				tc.recv = tc.recv[1:]
			}
		}
	}()
}

func (tc *testConnection) Dispatch(msg *Message) {
	tc.dispatched = append(tc.dispatched, msg)
}

func (tc *testConnection) RemoveHandler(chan<- *Message, ...Command) {}

func (tc *testConnection) EngineVersion() (EngineVersion, error) {
	return tc.engineVersion, tc.engineVersionErr
}
func (tc *testConnection) IDRequest() (FirmwareVersion, DevCat, error) {
	return tc.firmwareVersion, tc.devCat, nil
}

func (tc *testConnection) SendCommand(cmd Command, payload []byte) error {
	_, err := tc.Send(&Message{Command: cmd, Payload: payload})
	return err
}

func (tc *testConnection) Send(msg *Message) (*Message, error) {
	sent := &Message{}
	*sent = *msg
	tc.sent = append(tc.sent, sent)
	if tc.sendErr != nil {
		return nil, tc.sendErr
	}
	ack := tc.acks[0]
	tc.acks = tc.acks[1:]
	return ack, nil
}

func (tc *testConnection) Receive() (*Message, error) {
	if tc.recvErr == nil {
		msg := tc.recv[0]
		tc.recv = tc.recv[1:]
		return msg, nil
	}
	return nil, tc.recvErr
}

func TestDemuxDispatch(t *testing.T) {
	conn := &testConnection{}
	wildcard := &testConnection{}
	addr := Address{1, 2, 3}
	mux := &Demux{connections: map[Address][]Connection{addr: {conn}, Wildcard: {wildcard}}}
	mux.Dispatch(&Message{Src: addr})
	if len(conn.dispatched) != 1 {
		t.Errorf("Expected message to be dispatched to connection")
	}

	if len(wildcard.dispatched) != 1 {
		t.Errorf("Expected message to be dispatched to wildcard connection")
	}

	mux.Dispatch(&Message{Src: Address{3, 4, 5}})
	if len(conn.dispatched) != 1 {
		t.Errorf("Connection unexpectedly received message")
	}

	if len(wildcard.dispatched) != 2 {
		t.Errorf("Expected message to be dispatched to wildcard connection")
	}
}

func TestConnectionOptions(t *testing.T) {
	tests := []struct {
		desc  string
		input ConnectionOption
		want  *connection
	}{
		{"Timeout Option", ConnectionTimeout(time.Hour), &connection{timeout: time.Hour}},
		{"TTL Option", ConnectionTTL(3), &connection{ttl: 3}},
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
		response    *Message
		expectedErr error
	}{
		{"Send Ack", TestPing, TestPingAck, nil},
		{"Send Nak", TestProductDataResponse, TestProductDataResponse, ErrAckTimeout},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sender := &testSender{sndErr: test.expectedErr}
			conn, err := newConnection(sender, Address{}, ConnectionTimeout(time.Millisecond))
			sender.sndCb = func() { conn.Dispatch(test.response) }
			if err != nil {
				t.Errorf("Unexpected error from NewConnection(): %v", err)
			}

			_, err = conn.Send(test.input)
			if err != test.expectedErr {
				t.Errorf("Want %v got %v", test.expectedErr, err)
			}
		})
	}
}

func TestNewConnectionTTL(t *testing.T) {
	tests := []struct {
		ttl     uint8
		wantErr string
	}{
		{0, ""},
		{1, ""},
		{2, ""},
		{3, ""},
		{4, "invalid ttl 4, must be in range 0-3"},
		{254, "invalid ttl 254, must be in range 0-3"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("ttl %d", tt.ttl), func(t *testing.T) {
			sender := &testSender{}
			_, err := NewConnection(sender, Address{}, ConnectionTTL(tt.ttl))

			// TODO: consider switching to cmp package
			var got string
			if err != nil {
				got = fmt.Sprintf("%v", err)
			}
			if got != tt.wantErr {
				t.Errorf("got error %q, want %q", got, tt.wantErr)
			}
		})
	}
}

func TestConnectionIDRequest(t *testing.T) {
	sender := &testSender{}
	conn, _ := newConnection(sender, Address{}, ConnectionTimeout(time.Millisecond))
	sender.sndCb = func() {
		conn.Dispatch(TestAck)
		conn.Dispatch(&Message{Dst: Address{07, 79, 42}, Command: Command{0, 1}, Flags: StandardBroadcast})
	}

	wantVersion := FirmwareVersion(42)
	wantDevCat := DevCat{07, 79}

	gotVersion, gotDevCat, err := conn.IDRequest()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if gotVersion != wantVersion {
		t.Errorf("Want FirmwareVersion %v got %v", wantVersion, gotVersion)
	} else if gotDevCat != wantDevCat {
		t.Errorf("Want DevCat %v got %v", wantDevCat, gotDevCat)
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
			sender := &testSender{}
			conn, _ := newConnection(sender, Address{}, ConnectionTimeout(time.Millisecond))
			sender.sndCb = func() { conn.Dispatch(test.input) }

			gotVersion, err := conn.EngineVersion()
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
