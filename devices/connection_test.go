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

package devices

import (
	"io"
	"testing"

	"github.com/abates/insteon"
	"github.com/abates/insteon/commands"
)

type testWriter struct {
	read    []*insteon.Message
	readErr error
	written []*insteon.Message
	acks    []*insteon.Message
	ackErr  error
}

func (tw *testWriter) Write(msg *insteon.Message) (*insteon.Message, error) {
	tw.written = append(tw.written, msg)
	if len(tw.acks) > 0 {
		ack := tw.acks[0]
		tw.acks = tw.acks[1:]
		return ack, tw.ackErr
	}
	return &insteon.Message{Src: msg.Dst, Dst: msg.Src, Flags: insteon.StandardDirectAck}, tw.ackErr
}

func (tw *testWriter) Read() (*insteon.Message, error) {
	if len(tw.read) > 0 {
		msg := tw.read[0]
		tw.read = tw.read[1:]
		return msg, tw.readErr
	}
	return nil, io.EOF
}

func TestConnectionIDRequest(t *testing.T) {
	setButtonPressed := func(controller bool, domain, category, firmware byte) *insteon.Message {
		sbp := commands.SetButtonPressedResponder
		if controller {
			sbp = commands.SetButtonPressedController
		}
		addr := insteon.Address(0)
		addr.Put([]byte{domain, category, firmware})
		return &insteon.Message{insteon.Address(0x030405), addr, insteon.StandardBroadcast, sbp, nil}
	}

	tests := []struct {
		desc        string
		input       *insteon.Message
		wantVersion insteon.FirmwareVersion
		wantDevCat  insteon.DevCat
	}{
		{"Happy Path", setButtonPressed(true, 7, 79, 42), insteon.FirmwareVersion(42), insteon.DevCat{07, 79}},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			tw := &testWriter{read: []*insteon.Message{test.input}}
			gotVersion, gotDevCat, _ := IDRequest(tw, insteon.Address(0))
			if gotVersion != test.wantVersion {
				t.Errorf("Want FirmwareVersion %v got %v", test.wantVersion, gotVersion)
			}
			if gotDevCat != test.wantDevCat {
				t.Errorf("Want DevCat %v got %v", test.wantDevCat, gotDevCat)
			}
		})
	}
}

func TestConnectionEngineVersion(t *testing.T) {
	tests := []struct {
		desc        string
		input       *insteon.Message
		ackErr      error
		wantVersion insteon.EngineVersion
		wantErr     error
	}{
		{"Regular device", &insteon.Message{Command: commands.GetEngineVersion.SubCommand(42), Flags: insteon.StandardDirectAck}, nil, insteon.EngineVersion(42), nil},
		{"I2Cs device", &insteon.Message{Command: commands.GetEngineVersion.SubCommand(0xff), Flags: insteon.StandardDirectNak}, ErrNak, insteon.VerI2Cs, ErrNotLinked},
		{"NAK", &insteon.Message{Command: commands.GetEngineVersion.SubCommand(0xfd), Flags: insteon.StandardDirectNak}, ErrNak, insteon.VerI2Cs, ErrNak},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			tw := &testWriter{acks: []*insteon.Message{test.input}, ackErr: test.ackErr}
			gotVersion, err := GetEngineVersion(tw, insteon.Address(0))
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

func TestRead(t *testing.T) {
	tests := []struct {
		name    string
		input   []commands.Command
		matcher Matcher
		want    []commands.Command
	}{
		{
			name:    "Simple",
			input:   []commands.Command{commands.Command(1), commands.Command(2), commands.Command(3), commands.Command(4)},
			matcher: Matches(func(msg *insteon.Message) bool { return int(msg.Command)%2 == 0 }),
			want:    []commands.Command{commands.Command(2), commands.Command(4)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := []commands.Command{}
			input := []*insteon.Message{}
			for _, c := range test.input {
				input = append(input, &insteon.Message{Command: c})
			}
			tw := &testWriter{read: input}
			for m, err := Read(tw, test.matcher); err == nil; m, err = Read(tw, test.matcher) {
				got = append(got, m.Command)
			}

			if len(test.want) != len(got) {
				t.Errorf("Wanted to receive %d messages got %d", len(test.want), len(got))
			} else {
				for i, w := range test.want {
					if w != got[i] {
						t.Errorf("Wanted command %d to be %v got %v", i, w, got[i])
					}
				}
			}
		})
	}
}

/*func TestRetry(t *testing.T) {
	tests := []struct {
		name    string
		errors  []error
		retries int
		want    error
	}{
		{"happy path", []error{nil}, 1, nil},
		{"retry success", []error{ErrReadTimeout, nil}, 2, nil},
		{"retry timeout", []error{ErrReadTimeout, ErrReadTimeout}, 2, ErrReadTimeout},
		{"third time's a charm", []error{ErrReadTimeout, ErrReadTimeout, nil}, 3, nil},
		{"third time sometimes fails too", []error{ErrReadTimeout, ErrReadTimeout, ErrReadTimeout}, 3, ErrReadTimeout},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			i := 0
			cb := func() error {
				i++
				return test.errors[i-1]
			}
			got := retry(test.retries, cb)
			if !errors.Is(test.want, got) {
				t.Errorf("Wanted error %v got %v", test.want, got)
			}
		})
	}
}*/
