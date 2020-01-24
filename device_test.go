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

func TestDeviceRegistry(t *testing.T) {
	dr := &DeviceRegistry{}

	if _, found := dr.Find(Domain(1)); found {
		t.Error("Expected nothing found for Domain(1)")
	}

	constructorCalled := false
	dr.Register(Domain(1), func(DeviceInfo, Device, time.Duration) (Device, error) {
		constructorCalled = true
		return nil, nil
	})

	if _, found := dr.Find(Domain(1)); !found {
		t.Error("Expected to find Domain(1)")
	}

	dr.New(DeviceInfo{DevCat: DevCat{1, 0}}, &testConnection{}, 0)
	if !constructorCalled {
		t.Errorf("Expected New() to call device constructor")
	}

	dr.Delete(Domain(1))
	if _, found := dr.Find(Domain(1)); found {
		t.Error("Expected nothing found for Domain(1)")
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
		{"Dimmer", &testConnection{engineVersion: VerI1, devCat: DevCat{1, 0}}, reflect.TypeOf(&Dimmer{}), nil},
		{"Switch", &testConnection{engineVersion: VerI1, devCat: DevCat{2, 0}}, reflect.TypeOf(&Switch{}), nil},
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
