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

	"github.com/abates/insteon/commands"
)

func TestMatchers(t *testing.T) {
	tests := []struct {
		name    string
		matcher Matcher
		input   *Message
		want    bool
	}{
		{"AckMatcher (ack)", AckMatcher(), TestAck, true},
		{"AckMatcher (ping)", AckMatcher(), TestPing, false},
		{"AckMatcher (nak)", AckMatcher(), TestPingNak, true},
		{"AckMatcher (nak)", AckMatcher(), TestPingNak, true},
		{"CmdMatcher (true)", CmdMatcher(commands.Ping), TestPing, true},
		{"CmdMatcher (false)", CmdMatcher(commands.Ping), &Message{}, false},
		{"And (true)", And(Matches(func(*Message) bool { return true }), Matches(func(*Message) bool { return true })), &Message{}, true},
		{"And (false)", And(Matches(func(*Message) bool { return false }), Matches(func(*Message) bool { return true })), &Message{}, false},
		{"And (false 1)", And(Matches(func(*Message) bool { return false }), Matches(func(*Message) bool { return false })), &Message{}, false},
		{"Or (one)", Or(Matches(func(*Message) bool { return true }), Matches(func(*Message) bool { return true })), &Message{}, true},
		{"Or (both)", Or(Matches(func(*Message) bool { return false }), Matches(func(*Message) bool { return true })), &Message{}, true},
		{"Or (neither)", Or(Matches(func(*Message) bool { return false }), Matches(func(*Message) bool { return false })), &Message{}, false},
		{"Not (true)", Not(AckMatcher()), TestAck, false},
		{"Not (false)", Not(AckMatcher()), &Message{}, true},
		{"Src Matcher (true)", SrcMatcher(Address{1, 2, 3}), &Message{Src: Address{1, 2, 3}}, true},
		{"Src Matcher (false)", SrcMatcher(Address{3, 4, 5}), &Message{Src: Address{1, 2, 3}}, false},
		{"All-Link Matcher (true)", AllLinkMatcher(), &Message{Flags: StandardAllLinkBroadcast}, true},
		{"All-Link Matcher (false)", AllLinkMatcher(), &Message{Flags: StandardDirectMessage}, false},
		{"duplicate matcher (true)", DuplicateMatcher(&Message{Flags: Flag(MsgTypeDirect, false, 3, 3)}), &Message{Flags: Flag(MsgTypeDirect, false, 2, 3)}, true},
		{"duplicate matcher (false)", DuplicateMatcher(&Message{Flags: Flag(MsgTypeDirect, false, 3, 3)}), &Message{Flags: Flag(MsgTypeDirect, true, 2, 3)}, true},
		{
			name:    "MatchAck",
			matcher: MatchAck(&Message{Flags: Flag(MsgTypeDirectAck, false, 0, 1), Command: commands.Command(0x010203), Src: Address{3, 4, 5}}),
			input:   &Message{Flags: Flag(MsgTypeDirect, false, 0, 1), Command: commands.Command(0x010205), Src: Address{3, 4, 5}},
			want:    true,
		},
		{
			name:    "MatchAck (false)",
			matcher: MatchAck(&Message{Flags: Flag(MsgTypeDirectAck, false, 0, 1), Command: commands.Command(0x010203), Src: Address{3, 4, 5}}),
			input:   &Message{Flags: Flag(MsgTypeDirect, false, 0, 1), Command: commands.Command(0x010205), Src: Address{1, 2, 3}},
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
