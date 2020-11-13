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
)

func mkPayload(buf ...byte) []byte {
	return append(buf, make([]byte, 14-len(buf))...)
}

type testPubSub struct {
	published   []*Message
	publishResp []*Message
	publishErr  error

	rxCh          <-chan *Message
	subscribedCh  <-chan *Message
	unsubscribeCh <-chan *Message
}

func (tps *testPubSub) Publish(msg *Message) (*Message, error) {
	tps.published = append(tps.published, msg)
	msg = tps.publishResp[0]
	tps.publishResp = tps.publishResp[1:]
	return msg, tps.publishErr
}

func (tps *testPubSub) Subscribe(matcher Matcher) <-chan *Message {
	ch := make(chan *Message, cap(tps.rxCh))
	tps.subscribedCh = ch
	go func() {
		for msg := range tps.rxCh {
			if matcher.Matches(msg) {
				ch <- msg
			}
		}
	}()
	return ch
}

func (tps *testPubSub) Unsubscribe(ch <-chan *Message) { tps.unsubscribeCh = ch }

func TestDeviceCreate(t *testing.T) {
	tests := []struct {
		desc     string
		input    EngineVersion
		wantType reflect.Type
		wantErr  error
	}{
		{"I1Device", VerI1, reflect.TypeOf(&i1Device{}), nil},
		{"I2Device", VerI2, reflect.TypeOf(&i2Device{}), nil},
		{"I2CsDevice", VerI2Cs, reflect.TypeOf(&i2CsDevice{}), nil},
		{"ErrVersion", 4, reflect.TypeOf(nil), ErrVersion},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			device, gotErr := Create(&testBus{}, DeviceInfo{EngineVersion: test.input})
			gotType := reflect.TypeOf(device)

			if test.wantErr != gotErr {
				t.Errorf("want err %v got %v", test.wantErr, gotErr)
			} else if gotErr == nil {
				if test.wantType != gotType {
					t.Errorf("want type %v got %v", test.wantType, gotType)
				}
			}
		})
	}
}
