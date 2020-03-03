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
	sent    []*Message
	acks    []*Message
	sendErr error
	recv    []*Message
	recvErr error
}

func (tc *testConnection) Close() error { return nil }

func (tc *testConnection) Send(msg *Message) (*Message, error) {
	sent := &Message{}
	*sent = *msg
	tc.sent = append(tc.sent, sent)
	if tc.sendErr != nil {
		return nil, tc.sendErr
	}
	ack := tc.acks[0]
	tc.acks = tc.acks[1:]
	err := error(nil)
	if ack.Nak() {
		err = ErrNak
	}
	return ack, err
}

func (tc *testConnection) Receive() (*Message, error) {
	if len(tc.recv) > 0 {
		msg := tc.recv[0]
		tc.recv = tc.recv[1:]
		return msg, tc.recvErr
	}

	return nil, tc.recvErr
}

type testDialer struct {
	conn *testConnection
}

func (td testDialer) Dial(dst Address, cmds ...Command) (conn Connection, err error) {
	return td.conn, nil
}

type testDeviceDialer struct {
	conn *testConnection
}

func (td testDeviceDialer) Dial(cmds ...Command) (conn Connection, err error) {
	return td.conn, nil
}

type testReader struct {
	msgs []*Message
}

func (tr *testReader) ReadMessage() (*Message, error) {
	if len(tr.msgs) > 0 {
		msg := tr.msgs[0]
		tr.msgs = tr.msgs[1:]
		return msg, nil
	}
	return nil, io.EOF
}

type testWriter struct {
	msgs []*Message
	err  error
}

func (tr *testWriter) WriteMessage(msg *Message) error {
	tr.msgs = append(tr.msgs, msg)
	return tr.err
}

type testBridge struct {
	MessageReader
	MessageWriter
}

func TestDemuxDispatch(t *testing.T) {
	ch1 := make(chan *Message, 1)
	ch2 := make(chan *Message, 2)

	addr1 := Address{1, 2, 3}
	addr2 := Address{3, 4, 5}
	mux := &demux{
		connections: map[Address][]*connection{
			addr1:    {{recvCh: ch1}},
			Wildcard: {{recvCh: ch2}},
		},
	}
	mux.Dispatch(&Message{Src: addr1})
	mux.Dispatch(&Message{Src: addr2})

	if len(ch1) != 1 {
		t.Errorf("Expected one message to be dispatched to connection got %d", len(ch1))
	}

	if len(ch2) != 2 {
		t.Errorf("Expected two messages to be dispatched to wildcard connection got %d", len(ch2))
	}
}

func TestConnectionOptions(t *testing.T) {
	tests := []struct {
		desc    string
		input   ConnectionOption
		want    *connection
		wantErr string
	}{
		{"Timeout Option", ConnectionTimeout(time.Hour), &connection{timeout: time.Hour}, ""},
		{"TTL Option", ConnectionTTL(3), &connection{ttl: 3}, ""},
		{"TTL Option (error)", ConnectionTTL(42), nil, "invalid ttl 42, must be in range 0-3"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := &connection{}
			err := test.input(got)
			if err != nil {
				if test.wantErr != err.Error() {
					t.Errorf("Wanted error %s got %v", test.wantErr, err)
				}
			} else {
				if !reflect.DeepEqual(test.want, got) {
					t.Errorf("want connection %+v got %+v", test.want, got)
				}
			}
		})
	}
}

func TestDemuxDialClose(t *testing.T) {
	demux := NewDemux(&testWriter{}).(*demux)
	conn, err := demux.Dial(Address{1, 2, 3})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	} else {
		if connections, found := demux.connections[Address{1, 2, 3}]; found {
			if len(connections) != 1 {
				t.Errorf("Expected exactly one connection, got %d", len(connections))
			} else if connections[0] != conn {
				t.Errorf("Connections should be the same")
			} else {
				conn.Close()
				if len(demux.connections[Address{1, 2, 3}]) != 0 {
					t.Errorf("Expected connections slice to be empty, found %d", len(demux.connections[Address{1, 2, 3}]))
				}
			}
		} else {
			t.Errorf("Expected connection to be assigned by address")
		}
	}

}

func TestConnectionSend(t *testing.T) {
	tests := []struct {
		name      string
		inputAddr Address
		input     *Message
		inputAck  *Message
		wantMsg   *Message
		wantErr   error
	}{
		{"Standard Length Message", Address{1, 2, 3}, &Message{}, &Message{Flags: StandardDirectAck}, &Message{Dst: Address{1, 2, 3}, Flags: StandardDirectMessage}, nil},
		{"Extended Length Message", Address{1, 2, 3}, &Message{Payload: []byte{1, 2}}, &Message{Flags: StandardDirectAck}, &Message{Dst: Address{1, 2, 3}, Flags: ExtendedDirectMessage, Payload: []byte{1, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}, nil},
		{"Nak", Address{1, 2, 3}, &Message{}, &Message{Flags: StandardDirectNak}, nil, ErrNak},
		{"Ack Timeout", Address{1, 2, 3}, &Message{}, nil, nil, ErrAckTimeout},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writer := &testWriter{}
			reader := make(chan *Message, 1)
			if test.inputAck != nil {
				reader <- test.inputAck
			}

			conn, err := newConnection(writer, test.inputAddr, nil, ConnectionTimeout(time.Microsecond))
			conn.retries = 0
			conn.recvCh = reader
			if err != nil {
				t.Errorf("Unexpected error from NewConnection(): %v", err)
			}

			_, gotErr := conn.Send(test.input)
			if test.wantErr != gotErr {
				t.Errorf("Want error %v got %v", test.wantErr, gotErr)
			} else if gotErr == nil {
				gotMsg := writer.msgs[0]
				if !reflect.DeepEqual(test.wantMsg, gotMsg) {
					t.Errorf("Wanted message %v got %v", test.wantMsg, gotMsg)
				}
			}
		})
	}
}

func TestConnectionDispatch(t *testing.T) {
	tests := []struct {
		name     string
		match    []Command
		input    *Message
		wantRecv bool
	}{
		{"match everything", nil, &Message{}, true},
		{"match one command", []Command{{1, 2, 3}}, &Message{Command: Command{1, 2, 3}}, true},
		{"no match", []Command{{1, 2, 3}}, &Message{Command: Command{4, 5, 6}}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conn := &connection{
				match:  test.match,
				recvCh: make(chan *Message, 1),
			}
			conn.dispatch(test.input)
			if test.wantRecv {
				if len(conn.recvCh) != 1 {
					t.Errorf("Wanted message to be delivered")
				}
			} else {
				if len(conn.recvCh) != 0 {
					t.Errorf("Message should not have been delivered")
				}
			}
		})
	}
}

func TestConnectionIDRequest(t *testing.T) {
	tests := []struct {
		desc        string
		inputAck    *Message
		input       *Message
		wantVersion FirmwareVersion
		wantDevCat  DevCat
		wantErr     error
	}{
		{"Happy Path", &Message{Command: CmdIDRequest, Flags: StandardDirectAck}, SetButtonPressed(true, 7, 79, 42), FirmwareVersion(42), DevCat{07, 79}, nil},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			conn := &testConnection{acks: []*Message{test.inputAck}, recv: []*Message{test.input}}
			gotVersion, gotDevCat, gotErr := IDRequest(testDialer{conn}, Address{})
			if test.wantErr != gotErr {
				t.Errorf("Wanted error %v got %v", test.wantErr, gotErr)
			} else if gotErr == nil {
				if gotVersion != test.wantVersion {
					t.Errorf("Want FirmwareVersion %v got %v", test.wantVersion, gotVersion)
				}
				if gotDevCat != test.wantDevCat {
					t.Errorf("Want DevCat %v got %v", test.wantDevCat, gotDevCat)
				}
			}
		})
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
			conn := &testConnection{acks: []*Message{test.input}}

			gotVersion, err := GetEngineVersion(testDialer{conn}, Address{})
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
