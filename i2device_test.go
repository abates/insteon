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
)

func TestI2DeviceLinkCommands(t *testing.T) {
	tests := []struct {
		name string
		run  func(*i2Device)
		want Command
	}{
		{"EnterLinkingMode", func(d *i2Device) { d.EnterLinkingMode(40) }, CmdEnterLinkingMode.SubCommand(40)},
		{"EnterUnlinkingMode", func(d *i2Device) { d.EnterUnlinkingMode(41) }, CmdEnterUnlinkingMode.SubCommand(41)},
		{"ExitLinkingMode", func(d *i2Device) { d.ExitLinkingMode() }, CmdExitLinkingMode},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := &testBus{publishResp: []*Message{TestAck}}
			device := &i2Device{i1Device: newI1Device(b, DeviceInfo{})}
			test.run(device)
			if test.want != b.published.Command {
				t.Errorf("Wanted command %v got %v", test.want, b.published.Command)
			}
		})
	}
}

func TestI2DeviceLinkDatabase(t *testing.T) {
	want := &i2Device{}
	got, _ := want.LinkDatabase()
	if want != got {
		t.Errorf("Expected LinkDatabase to return the device")
	}
}
