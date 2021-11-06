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
	"errors"
	"testing"

	"github.com/abates/insteon/commands"
)

func mkPayload(buf ...byte) []byte {
	return append(buf, make([]byte, 14-len(buf))...)
}

func TestDeviceCommands(t *testing.T) {
	dimmerFactory := func(version int) func(device *BasicDevice) Device {
		return func(device *BasicDevice) Device {
			device.DeviceInfo.FirmwareVersion = FirmwareVersion(version)
			return &Dimmer{Switch: &Switch{BasicDevice: device}}
		}
	}

	switchFactory := func(device *BasicDevice) Device {
		return &Switch{BasicDevice: device}
	}

	tests := []struct {
		name        string
		setup       func(*BasicDevice) Device
		test        func(Device) error
		wantCmd     commands.Command
		wantPayload []byte
	}{
		{"switch on", switchFactory, func(device Device) error { return device.(*Switch).TurnOn(42) }, commands.LightOn.SubCommand(42), nil},
		{"switch off", switchFactory, func(device Device) error { return device.(*Switch).TurnOff() }, commands.LightOff, nil},
		{"switch backlight on", switchFactory, func(device Device) error { return device.(*Switch).SetBacklight(true) }, commands.EnableLED, make([]byte, 14)},
		{"switch backlight off", switchFactory, func(device Device) error { return device.(*Switch).SetBacklight(false) }, commands.DisableLED, make([]byte, 14)},
		{"switch enable load sense", switchFactory, func(device Device) error { return device.(*Switch).SetLoadSense(true) }, commands.EnableLoadSense, make([]byte, 14)},
		{"switch disable load sense", switchFactory, func(device Device) error { return device.(*Switch).SetLoadSense(false) }, commands.DisableLoadSense, make([]byte, 14)},

		{"dimmer brighten", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).Brighten() }, commands.LightBrighten, nil},
		{"dimmer dim", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).Dim() }, commands.LightDim, nil},
		{"dimmer start brighten", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).StartBrighten() }, commands.StartBrighten, nil},
		{"dimmer start dim", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).StartDim() }, commands.StartDim, nil},
		{"dimmer stop manual change", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).StopManualChange() }, commands.LightStopManual, nil},
		{"dimmer on fast", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).OnFast() }, commands.LightOnFast, nil},
		{"dimmer instant change", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).InstantChange(42) }, commands.LightInstantChange.SubCommand(42), nil},
		{"dimmer set status", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).SetStatus(42) }, commands.LightSetStatus.SubCommand(42), nil},
		{"dimmer on at ramp", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).OnAtRamp(0x04, 0x02) }, commands.LightOnAtRamp.SubCommand(0x42), nil},
		{"dimmer on at ramp (new)", dimmerFactory(0x43), func(device Device) error { return device.(*Dimmer).OnAtRamp(0x04, 0x02) }, commands.LightOnAtRampV67.SubCommand(0x42), nil},
		{"dimmer off at ramp", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).OffAtRamp(15) }, commands.LightOffAtRamp.SubCommand(15), nil},
		{"dimmer off at ramp (new)", dimmerFactory(0x43), func(device Device) error { return device.(*Dimmer).OffAtRamp(15) }, commands.LightOffAtRampV67.SubCommand(15), nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//td := &testDevice{}
			tw := &testWriter{}
			device := test.setup(&BasicDevice{MessageWriter: tw})

			err := test.test(device)
			if err == nil {
				if test.wantCmd != tw.written[0].Command {
					t.Errorf("Wanted command %v got %v", test.wantCmd, tw.written[0].Command)
				}

				if !bytes.Equal(test.wantPayload, tw.written[0].Payload) {
					t.Errorf("Wanted payload %x got %x", test.wantPayload, tw.written[0].Payload)
				}
			} else {
				t.Errorf("Unexpected error %v", err)
			}
		})
	}
}

func TestOpen(t *testing.T) {
	tests := []struct {
		name     string
		ackErr   error
		acks     []*Message
		read     []*Message
		wantInfo DeviceInfo
		wantErr  error
	}{
		{
			name:     "basic",
			ackErr:   nil,
			acks:     []*Message{&Message{Command: commands.Command(0x01)}},
			read:     []*Message{&Message{Dst: Address{byte(SwitchDomain), 1, 59}, Command: commands.SetButtonPressedResponder}},
			wantInfo: DeviceInfo{EngineVersion: VerI2, FirmwareVersion: FirmwareVersion(59), DevCat: DevCat{byte(SwitchDomain), 1}},
			wantErr:  nil,
		},
		{
			name:     "unlinked",
			ackErr:   ErrNak,
			acks:     []*Message{&Message{Flags: StandardDirectNak, Command: commands.Command(0x00ff)}},
			read:     []*Message{&Message{Dst: Address{byte(SwitchDomain), 1, 59}, Command: commands.SetButtonPressedResponder}},
			wantInfo: DeviceInfo{EngineVersion: VerI2Cs},
			wantErr:  ErrNotLinked,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tw := &testWriter{
				acks:   test.acks,
				ackErr: test.ackErr,
				read:   test.read,
			}

			_, gotInfo, err := Open(tw, Address{})
			if errors.Is(err, test.wantErr) {
				if test.wantInfo != gotInfo {
					t.Errorf("Wanted device info %v got %v", test.wantInfo, gotInfo)
				}
			} else {
				t.Errorf("Wanted error %v got %v", test.wantErr, err)
			}
		})
	}

}

/*func TestUpgrade(t *testing.T) {
	tests := []struct {
		name  string
		input DeviceInfo
		want  Device
	}{
		{"switch", DeviceInfo{EngineVersion: EngineVersion(1), DevCat: DevCat{byte(SwitchDomain), 0}}, &Switch{}},
		{"dimmer", DeviceInfo{EngineVersion: EngineVersion(1), DevCat: DevCat{byte(DimmerDomain), 0}}, &Dimmer{}},
		{"outlet", DeviceInfo{EngineVersion: EngineVersion(1), DevCat: DevCat{byte(SwitchDomain), 0x08}}, &Outlet{}},
		{"thermostat", DeviceInfo{EngineVersion: EngineVersion(1), DevCat: DevCat{byte(ThermostatDomain), 0}}, &Thermostat{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			device := Upgrade(&testWriter{}, test.input)
			want := reflect.TypeOf(test.want)
			got := reflect.TypeOf(device)
			if want != got {
				t.Errorf("Wanted type %s got %s", want, got)
			}
		})
	}
}*/
