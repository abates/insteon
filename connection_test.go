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
	"testing"
	"time"
)

type testBus struct {
	published   *Message
	publishResp []*Message
	publishErr  error

	subscribeSrc     Address
	subscribeMatcher Matcher
	subscribeCh      <-chan *Message

	unsubscribeSrc Address
	unsubscribeCh  <-chan *Message
}

func (tb *testBus) Publish(msg *Message) (*Message, error) {
	tb.published = msg
	msg = tb.publishResp[0]
	tb.publishResp = tb.publishResp[1:]
	return msg, tb.publishErr
}

func (tb *testBus) Subscribe(src Address, matcher Matcher) <-chan *Message {
	tb.subscribeSrc = src
	tb.subscribeMatcher = matcher
	return tb.subscribeCh
}

func (tb *testBus) Unsubscribe(src Address, ch <-chan *Message) {
	tb.unsubscribeSrc = src
	tb.unsubscribeCh = ch
}

func (tb *testBus) Close() error { return nil }

func (tb *testBus) Config() ConnectionConfig { return ConnectionConfig{} }

func TestBusRun(t *testing.T) {
	tests := []struct {
		name   string
		pubSrc Address
		input  *Message
		want   bool
	}{
		{"receive msg", Address{1, 2, 3}, &Message{Src: Address{1, 2, 3}}, true},
		{"reject msg", Address{1, 2, 3}, &Message{Src: Address{4, 5, 6}}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			messages := make(chan *Message)
			bb, _ := NewBus(nil, messages)
			b := bb.(*bus)
			ch := b.Subscribe(test.pubSrc, Matches(func(*Message) bool { return true }))
			messages <- test.input
			b.Unsubscribe(test.pubSrc, ch)
			b.Close()
			if test.want {
				if len(ch) != 1 {
					t.Errorf("Wanted subcriber to receive message")
				}
			} else if len(ch) != 0 {
				t.Errorf("Wanted subscriber to not receive message")
			}

			if len(b.listeners[test.pubSrc]) > 0 {
				t.Errorf("Run did not remove subscriber")
			}
		})
	}
}

type testWriter struct {
	ch chan *Message
}

func (tw *testWriter) WriteMessage(msg *Message) error {
	tw.ch <- msg
	return nil
}

func TestPublish(t *testing.T) {
	tests := []struct {
		name      string
		timeout   time.Duration
		tries     int
		input     *Message
		resp      *Message
		wantFlags Flags
		wantErr   error
	}{
		{"normal", time.Second, 1, &Message{}, &Message{Flags: StandardDirectAck}, Flag(MsgTypeDirect, false, 0, 0), nil},
		{"nak", time.Second, 1, &Message{}, &Message{Flags: StandardDirectNak}, Flag(MsgTypeDirect, false, 0, 0), ErrNak},
		{"retries", time.Millisecond, 4, &Message{}, nil, Flag(MsgTypeDirect, false, 0, 0), ErrAckTimeout},
		{"extended", time.Millisecond, 1, &Message{Payload: []byte{1, 2, 3}}, &Message{Flags: StandardDirectAck}, Flag(MsgTypeDirect, true, 0, 0), nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			messages := make(chan *Message, 1)
			writer := make(chan *Message, test.tries)
			done := make(chan bool)
			bb, _ := NewBus(&testWriter{writer}, messages, ConnectionTTL(0), ConnectionTimeout(test.timeout), ConnectionRetry(3))
			b := bb.(*bus)
			go func() {
				_, gotErr := b.Publish(test.input)
				if test.wantErr != gotErr {
					t.Errorf("Wanted err %v got %v", test.wantErr, gotErr)
				}
				done <- true
			}()
			msg := <-writer
			if test.resp != nil {
				messages <- test.resp
			}

			if test.wantFlags != msg.Flags {
				t.Errorf("Wanted flags %v got %v", test.wantFlags, msg.Flags)
			}

			if msg.Flags.Extended() && len(msg.Payload) != 14 {
				t.Errorf("Wanted 14 byte payload, got %d", len(msg.Payload))
			}

			<-done
			b.Close()
			if len(b.listeners[test.input.Src]) != 0 {
				t.Errorf("Expected listener to be unsubscribed after publish")
			}

			if test.tries != len(writer)+1 {
				t.Errorf("Expected message to be retried %d times, got %d", test.tries, len(writer)+1)
			}
		})
	}
}

func TestConnectionOptions(t *testing.T) {
	tests := []struct {
		desc    string
		input   ConnectionOption
		want    ConnectionConfig
		wantErr string
	}{
		{"Timeout Option", ConnectionTimeout(time.Hour), ConnectionConfig{DefaultTimeout: time.Hour}, ""},
		{"TTL Option", ConnectionTTL(3), ConnectionConfig{TTL: 3}, ""},
		{"TTL Option (error)", ConnectionTTL(42), ConnectionConfig{}, "invalid ttl 42, must be in range 0-3"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := ConnectionConfig{}
			err := test.input(&got)
			if err != nil {
				if test.wantErr != err.Error() {
					t.Errorf("Wanted error %s got %v", test.wantErr, err)
				}
			} else {
				if test.want != got {
					t.Errorf("want connection %+v got %+v", test.want, got)
				}
			}
		})
	}
}

func TestConnectionIDRequest(t *testing.T) {
	tests := []struct {
		desc        string
		input       *Message
		wantVersion FirmwareVersion
		wantDevCat  DevCat
	}{
		{"Happy Path", SetButtonPressed(true, 7, 79, 42), FirmwareVersion(42), DevCat{07, 79}},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ch := make(chan *Message, 1)
			ch <- test.input
			tb := &testBus{subscribeCh: ch, publishResp: []*Message{{}}}
			gotVersion, gotDevCat, _ := IDRequest(tb, Address{})
			if gotVersion != test.wantVersion {
				t.Errorf("Want FirmwareVersion %v got %v", test.wantVersion, gotVersion)
			}
			if gotDevCat != test.wantDevCat {
				t.Errorf("Want DevCat %v got %v", test.wantDevCat, gotDevCat)
			}
			if tb.unsubscribeCh != tb.subscribeCh {
				t.Errorf("IDRequest never unsubscribed from the bus")
			}
		})
	}
}

func TestConnectionEngineVersion(t *testing.T) {
	tests := []struct {
		desc        string
		input       *Message
		publishErr  error
		wantVersion EngineVersion
		wantErr     error
	}{
		{"Regular device", &Message{Command: CmdGetEngineVersion.SubCommand(42), Flags: StandardDirectAck}, nil, EngineVersion(42), nil},
		{"I2Cs device", &Message{Command: CmdGetEngineVersion.SubCommand(0xff), Flags: StandardDirectNak}, ErrNak, VerI2Cs, ErrNotLinked},
		{"NAK", &Message{Command: CmdGetEngineVersion.SubCommand(0xfd), Flags: StandardDirectNak}, ErrNak, VerI2Cs, ErrNak},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			tb := &testBus{publishResp: []*Message{test.input}, publishErr: test.publishErr}
			gotVersion, err := GetEngineVersion(tb, Address{})
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
