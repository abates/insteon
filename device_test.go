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
	"bytes"
	"reflect"
	"testing"
	"time"
)

func TestDeviceRegistry(t *testing.T) {
	dr := &DeviceRegistry{}

	if _, found := dr.Find(Category(1)); found {
		t.Error("Expected nothing found for Category(1)")
	}

	constructorCalled := false
	dr.Register(Category(1), func(DeviceInfo, Device, time.Duration) (Device, error) {
		constructorCalled = true
		return nil, nil
	})

	if _, found := dr.Find(Category(1)); !found {
		t.Error("Expected to find Category(1)")
	}

	dr.New(DeviceInfo{DevCat: DevCat{1, 0}}, &testConnection{}, 0)
	if !constructorCalled {
		t.Errorf("Expected New() to call device constructor")
	}

	dr.Delete(Category(1))
	if _, found := dr.Find(Category(1)); found {
		t.Error("Expected nothing found for Category(1)")
	}
}

func mkPayload(buf ...byte) []byte {
	return append(buf, make([]byte, 14-len(buf))...)
}

func TestDeviceNew(t *testing.T) {
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
			device, gotErr := New(test.input, &testConnection{}, 0)
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

func TestDeviceOpen(t *testing.T) {
	tests := []struct {
		desc     string
		input    *testConnection
		wantType reflect.Type
		wantErr  error
	}{
		{"I1Device", &testConnection{engineVersion: VerI1}, reflect.TypeOf(&i1Device{}), nil},
		{"I2Device", &testConnection{engineVersion: VerI2}, reflect.TypeOf(&i2Device{}), nil},
		{"I2CsDevice", &testConnection{engineVersion: VerI2Cs}, reflect.TypeOf(&i2CsDevice{}), nil},
		{"Dimmer", &testConnection{engineVersion: VerI1, devCat: DevCat{1, 0}}, reflect.TypeOf(&dimmer{}), nil},
		{"Linkable Dimmer", &testConnection{engineVersion: VerI2Cs, devCat: DevCat{1, 0}}, reflect.TypeOf(&linkableDimmer{}), nil},
		{"Switch", &testConnection{engineVersion: VerI1, devCat: DevCat{2, 0}}, reflect.TypeOf(&switchedDevice{}), nil},
		{"Linkable Switch", &testConnection{engineVersion: VerI2Cs, devCat: DevCat{2, 0}}, reflect.TypeOf(&linkableSwitch{}), nil},
		{"ErrVersion", &testConnection{engineVersion: 4}, reflect.TypeOf(nil), ErrVersion},
		{"Not Linked", &testConnection{engineVersionErr: ErrNotLinked}, reflect.TypeOf(&i2CsDevice{}), ErrNotLinked},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			device, gotErr := Open(test.input, 0)
			gotType := reflect.TypeOf(device)

			if test.wantErr != gotErr {
				t.Errorf("want err %v got %v", test.wantErr, gotErr)
			}
			if test.wantType != gotType {
				t.Errorf("want type %v got %v", test.wantType, gotType)
			}
		})
	}
}

type commandTest struct {
	desc        string
	callback    func(Device) error
	wantCmd     Command
	wantErr     error
	wantPayload []byte
}

func testDeviceCommand(t *testing.T, constructor func(*testConnection) Device, cb func(Device) error, wantCmd Command, wantPayload []byte, wantErr error, msgs ...*Message) {
	conn := &testConnection{sendCh: make(chan *Message, 1), ackCh: make(chan *Message, 1), recvCh: make(chan *Message, 1)}
	device := constructor(conn)

	go func() {
		msg := <-conn.sendCh
		if wantErr == nil {
			if msg.Command != wantCmd {
				t.Errorf("want Command %v got %v", wantCmd, msg.Command)
			}

			if len(wantPayload) > 0 {
				if !bytes.Equal(wantPayload, msg.Payload) {
					t.Errorf("want payload %x got %x", wantPayload, msg.Payload)
				}
			}
		}

		conn.ackCh <- TestAck
		for _, msg := range msgs {
			select {
			case conn.recvCh <- msg:
			case <-time.After(time.Millisecond):
				t.Errorf("No one was waiting for the message")
			}
		}
	}()

	err := cb(device)
	if err != wantErr {
		t.Errorf("got error %v, want %v", err, wantErr)
	}

}

func testDeviceCommands(t *testing.T, constructor func(*testConnection) Device, tests []*commandTest) {
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			testDeviceCommand(t, constructor, test.callback, test.wantCmd, test.wantPayload, test.wantErr)
		})
		/*t.Run(test.desc, func(t *testing.T) {
			conn := &testConnection{sendCh: make(chan *Message, 1), ackCh: make(chan *Message, 1)}
			device := constructor(conn)

			conn.ackCh <- TestAck
			err := test.callback(device)

			if err != test.wantErr {
				t.Errorf("got error %v, want %v", err, test.wantErr)
			}

			if test.wantErr == nil {
				msg := <-conn.sendCh
				if msg.Command != test.wantCmd {
					t.Errorf("want Command %v got %v", test.wantCmd, msg.Command)
				}

				if len(test.wantPayload) > 0 {
					wantPayload := make([]byte, len(test.wantPayload))
					copy(wantPayload, test.wantPayload)
					if !bytes.Equal(wantPayload, msg.Payload) {
						t.Errorf("want payload %x got %x", wantPayload, msg.Payload)
					}
				}
			}
		})*/
	}
}
