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

func TestI2DeviceIsLinkable(t *testing.T) {
	device := Device(&i2Device{})
	linkable := device.(Linkable)
	if linkable == nil {
		t.Error("linkable should not be nil")
	}
}

func TestI2DeviceCommands(t *testing.T) {
	tests := []*commandTest{
		{"ExitLinkingMode", func(d Device) error { return d.(*i2Device).ExitLinkingMode() }, CmdExitLinkingMode, nil, nil},
	}

	testDeviceCommands(t, func(conn *testConnection) Device { return newI2Device(conn, time.Millisecond) }, tests)
}

func TestI2DeviceEnterLinkingMode(t *testing.T) {
	constructor := func(conn *testConnection) Device { return newI2Device(conn, time.Millisecond) }
	callback := func(d Device) error { return d.(*i2Device).EnterLinkingMode(10) }
	// happy path
	testDeviceCommand(t, constructor, callback, CmdEnterLinkingMode.SubCommand(10), nil, nil, &Message{Flags: StandardBroadcast, Command: CmdSetButtonPressedResponder})

	// sad path
	testDeviceCommand(t, constructor, callback, CmdEnterLinkingMode.SubCommand(10), nil, ErrReadTimeout)
}

func TestI2DeviceEnterUnlinkingMode(t *testing.T) {
	constructor := func(conn *testConnection) Device { return newI2Device(conn, time.Millisecond) }
	callback := func(d Device) error { return d.(*i2Device).EnterUnlinkingMode(10) }
	// happy path
	testDeviceCommand(t, constructor, callback, CmdEnterUnlinkingMode.SubCommand(10), nil, nil, &Message{Flags: StandardBroadcast, Command: CmdSetButtonPressedResponder})

	// sad path
	testDeviceCommand(t, constructor, callback, CmdEnterUnlinkingModeExt.SubCommand(10), nil, ErrReadTimeout)
}
