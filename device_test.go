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

type testDevice struct {
	conn Connection
}

func (td testDevice) Address() Address                         { return Address{} }
func (td testDevice) Dial(cmds ...Command) (Connection, error) { return td.conn, nil }

func (td testDevice) SendCommand(cmd Command, payload []byte) (ack Command, err error) {
	msg, err := td.conn.Send(&Message{Command: cmd, Payload: payload})
	if err == nil {
		ack = msg.Command
	}
	return
}

func (td testDevice) LinkDatabase() (Linkable, error) { return nil, nil }

func (td testDevice) Info() DeviceInfo { return DeviceInfo{} }

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
			device, gotErr := create(nil, DeviceInfo{EngineVersion: test.input})
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
		{"I1Device", &testConnection{acks: []*Message{TestMessageEngineVersion1, TestAck}, recv: []*Message{SetButtonPressed(false, 0, 0, 0)}}, reflect.TypeOf(&i1Device{}), nil},
		{"I2Device", &testConnection{acks: []*Message{TestMessageEngineVersion2, TestAck}, recv: []*Message{SetButtonPressed(false, 0, 0, 0)}}, reflect.TypeOf(&i2Device{}), nil},
		{"I2CsDevice", &testConnection{acks: []*Message{TestMessageEngineVersion2cs, TestAck}, recv: []*Message{SetButtonPressed(false, 0, 0, 0)}}, reflect.TypeOf(&i2CsDevice{}), nil},
		{"Dimmer", &testConnection{acks: []*Message{TestMessageEngineVersion2cs, TestAck}, recv: []*Message{SetButtonPressed(false, 1, 0, 0)}}, reflect.TypeOf(&Dimmer{}), nil},
		{"Switch", &testConnection{acks: []*Message{TestMessageEngineVersion2cs, TestAck}, recv: []*Message{SetButtonPressed(false, 2, 0, 0)}}, reflect.TypeOf(&Switch{}), nil},
		{"ErrVersion", &testConnection{acks: []*Message{TestMessageEngineVersion3, TestAck}, recv: []*Message{SetButtonPressed(false, 0, 0, 0)}}, reflect.TypeOf(nil), ErrVersion},
		{"Not Linked", &testConnection{acks: []*Message{Ack(false, 0, 255)}, recv: []*Message{SetButtonPressed(false, 0, 0, 0)}}, reflect.TypeOf(&i2CsDevice{}), ErrNotLinked},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			device, gotErr := Open(&testDialer{test.input}, Address{})
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
