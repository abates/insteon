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
	"testing"

	"github.com/abates/insteon"
	"github.com/abates/insteon/commands"
)

func TestMatchers(t *testing.T) {
	ping := &insteon.Message{insteon.Address{1, 3, 3}, insteon.Address{3, 4, 5}, insteon.StandardDirectMessage, commands.Ping, nil}
	pingNak := &insteon.Message{insteon.Address{1, 2, 3}, insteon.Address{3, 4, 5}, insteon.StandardDirectNak, commands.Command(0x000f00), nil}

	tests := []struct {
		name    string
		matcher Matcher
		input   *insteon.Message
		want    bool
	}{
		{"AckMatcher (ack)", AckMatcher(), &insteon.Message{Flags: insteon.StandardDirectAck}, true},
		{"AckMatcher (ping)", AckMatcher(), ping, false},
		{"AckMatcher (nak)", AckMatcher(), &insteon.Message{Flags: insteon.StandardDirectNak}, true},
		{"AckMatcher (nak)", AckMatcher(), pingNak, true},
		{"CmdMatcher (true)", CmdMatcher(commands.Ping), ping, true},
		{"CmdMatcher (false)", CmdMatcher(commands.Ping), &insteon.Message{}, false},
		{"And (true)", And(Matches(func(*insteon.Message) bool { return true }), Matches(func(*insteon.Message) bool { return true })), &insteon.Message{}, true},
		{"And (false)", And(Matches(func(*insteon.Message) bool { return false }), Matches(func(*insteon.Message) bool { return true })), &insteon.Message{}, false},
		{"And (false 1)", And(Matches(func(*insteon.Message) bool { return false }), Matches(func(*insteon.Message) bool { return false })), &insteon.Message{}, false},
		{"Or (one)", Or(Matches(func(*insteon.Message) bool { return true }), Matches(func(*insteon.Message) bool { return true })), &insteon.Message{}, true},
		{"Or (both)", Or(Matches(func(*insteon.Message) bool { return false }), Matches(func(*insteon.Message) bool { return true })), &insteon.Message{}, true},
		{"Or (neither)", Or(Matches(func(*insteon.Message) bool { return false }), Matches(func(*insteon.Message) bool { return false })), &insteon.Message{}, false},
		{"Not (true)", Not(AckMatcher()), &insteon.Message{Flags: insteon.StandardDirectAck}, false},
		{"Not (false)", Not(AckMatcher()), &insteon.Message{}, true},
		{"Src Matcher (true)", SrcMatcher(insteon.Address{1, 2, 3}), &insteon.Message{Src: insteon.Address{1, 2, 3}}, true},
		{"Src Matcher (false)", SrcMatcher(insteon.Address{3, 4, 5}), &insteon.Message{Src: insteon.Address{1, 2, 3}}, false},
		{"All-Link Matcher (true)", AllLinkMatcher(), &insteon.Message{Flags: insteon.StandardAllLinkBroadcast}, true},
		{"All-Link Matcher (false)", AllLinkMatcher(), &insteon.Message{Flags: insteon.StandardDirectAck}, false},
		{"duplicate matcher (true)", DuplicateMatcher(&insteon.Message{Flags: insteon.Flag(insteon.MsgTypeDirect, false, 3, 3)}), &insteon.Message{Flags: insteon.Flag(insteon.MsgTypeDirect, false, 2, 3)}, true},
		{"duplicate matcher (false)", DuplicateMatcher(&insteon.Message{Flags: insteon.Flag(insteon.MsgTypeDirect, false, 3, 3)}), &insteon.Message{Flags: insteon.Flag(insteon.MsgTypeDirect, true, 2, 3)}, true},
		{
			name:    "MatchAck",
			matcher: MatchAck(&insteon.Message{Flags: insteon.Flag(insteon.MsgTypeDirectAck, false, 0, 1), Command: commands.Command(0x010203), Src: insteon.Address{3, 4, 5}}),
			input:   &insteon.Message{Flags: insteon.Flag(insteon.MsgTypeDirect, false, 0, 1), Command: commands.Command(0x010205), Src: insteon.Address{3, 4, 5}},
			want:    true,
		},
		{
			name:    "MatchAck (false)",
			matcher: MatchAck(&insteon.Message{Flags: insteon.Flag(insteon.MsgTypeDirectAck, false, 0, 1), Command: commands.Command(0x010203), Src: insteon.Address{3, 4, 5}}),
			input:   &insteon.Message{Flags: insteon.Flag(insteon.MsgTypeDirect, false, 0, 1), Command: commands.Command(0x010205), Src: insteon.Address{1, 2, 3}},
			want:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.matcher.Matches(test.input)
			if test.want != got {
				t.Errorf("Wanted match %v got %v", test.want, got)
			}
		})
	}
}
