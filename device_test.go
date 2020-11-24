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
)

func mkPayload(buf ...byte) []byte {
	return append(buf, make([]byte, 14-len(buf))...)
}

type testDevice struct {
	sendCmd     Command
	sendPayload []byte
	sendAck     Command
	sendErr     error
}

func (td *testDevice) SendCommand(cmd Command, payload []byte) (err error) {
	_, err = td.Send(cmd, payload)
	return err
}

func (td *testDevice) Send(cmd Command, payload []byte) (ack Command, err error) {
	td.sendCmd = cmd
	td.sendPayload = payload
	return td.sendAck, td.sendErr
}

func (td *testDevice) Publish(*Message) (*Message, error) {
	return nil, ErrNotSupported
}

func (td *testDevice) Subscribe(matcher Matcher) <-chan *Message {
	return nil
}

func (td *testDevice) Unsubscribe(ch <-chan *Message) {
}

func (td *testDevice) LinkDatabase() (Linkable, error) {
	return nil, ErrNotSupported
}

func (td *testDevice) Info() DeviceInfo {
	return DeviceInfo{}
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

func TestDeviceCommands(t *testing.T) {
	dimmerFactory := func(version int) func(td *testDevice) Device {
		return func(td *testDevice) Device {
			return &Dimmer{Switch: &Switch{Device: td}, info: DeviceInfo{FirmwareVersion: FirmwareVersion(version)}}
		}
	}

	switchFactory := func(td *testDevice) Device {
		return &Switch{Device: td}
	}

	tests := []struct {
		name        string
		setup       func(*testDevice) Device
		test        func(Device) error
		wantCmd     Command
		wantPayload []byte
	}{
		{"switch on", switchFactory, func(device Device) error { return device.(*Switch).TurnOn(42) }, CmdLightOn.SubCommand(42), nil},
		{"switch off", switchFactory, func(device Device) error { return device.(*Switch).TurnOff() }, CmdLightOff, nil},
		{"switch backlight on", switchFactory, func(device Device) error { return device.(*Switch).SetBacklight(true) }, CmdEnableLED, make([]byte, 14)},
		{"switch backlight off", switchFactory, func(device Device) error { return device.(*Switch).SetBacklight(false) }, CmdDisableLED, make([]byte, 14)},

		{"dimmer brighten", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).Brighten() }, CmdLightBrighten, nil},
		{"dimmer dim", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).Dim() }, CmdLightDim, nil},
		{"dimmer start brighten", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).StartBrighten() }, CmdStartBrighten, nil},
		{"dimmer start dim", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).StartDim() }, CmdStartDim, nil},
		{"dimmer stop manual change", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).StopManualChange() }, CmdLightStopManual, nil},
		{"dimmer on fast", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).OnFast() }, CmdLightOnFast, nil},
		{"dimmer instant change", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).InstantChange(42) }, CmdLightInstantChange.SubCommand(42), nil},
		{"dimmer set status", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).SetStatus(42) }, CmdLightSetStatus.SubCommand(42), nil},
		{"dimmer on at ramp", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).OnAtRamp(0x04, 0x02) }, CmdLightOnAtRamp.SubCommand(0x42), nil},
		{"dimmer on at ramp (new)", dimmerFactory(0x43), func(device Device) error { return device.(*Dimmer).OnAtRamp(0x04, 0x02) }, CmdLightOnAtRampV67.SubCommand(0x42), nil},
		{"dimmer off at ramp", dimmerFactory(0), func(device Device) error { return device.(*Dimmer).OffAtRamp(15) }, CmdLightOffAtRamp.SubCommand(15), nil},
		{"dimmer off at ramp (new)", dimmerFactory(0x43), func(device Device) error { return device.(*Dimmer).OffAtRamp(15) }, CmdLightOffAtRampV67.SubCommand(15), nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			td := &testDevice{}
			device := test.setup(td)
			err := test.test(device)
			if err == nil {
				if test.wantCmd != td.sendCmd {
					t.Errorf("Wanted command %v got %v", test.wantCmd, td.sendCmd)
				}

				if !bytes.Equal(test.wantPayload, td.sendPayload) {
					t.Errorf("Wanted payload %x got %x", test.wantPayload, td.sendPayload)
				}
			} else {
				t.Errorf("Unexpected error %v", err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name  string
		input DeviceInfo
		want  reflect.Type
	}{
		{"switch", DeviceInfo{EngineVersion: EngineVersion(1), DevCat: DevCat{byte(SwitchDomain), 0}}, reflect.TypeOf(&Switch{})},
		{"dimmer", DeviceInfo{EngineVersion: EngineVersion(1), DevCat: DevCat{byte(DimmerDomain), 0}}, reflect.TypeOf(&Dimmer{})},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			device, err := New(&testBus{}, test.input)
			if err == nil {
				got := reflect.TypeOf(device)
				if test.want != got {
					t.Errorf("Wanted type %s got %s", test.want, got)
				}
			} else {
				t.Errorf("Unexpected error %v", err)
			}
		})
	}
}
