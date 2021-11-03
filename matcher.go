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

type Matcher interface {
	Matches(msg *Message) bool
}

type CmdMatcher Command

func (m CmdMatcher) Matches(msg *Message) bool {
	return Command(m).Matches(msg.Command)
}

type Matches func(msg *Message) bool

func (m Matches) Matches(msg *Message) bool {
	return m(msg)
}

func DuplicateMatcher(msg *Message) Matcher {
	return Matches(msg.Duplicate)
}

func SrcMatcher(src Address) Matcher {
	return Matches(func(msg *Message) bool {
		return msg.Src == src
	})
}

func AckMatcher() Matcher {
	return Matches(func(msg *Message) bool {
		return msg.Ack() || msg.Nak()
	})
}

func AllLinkMatcher() Matcher {
	return Matches(func(msg *Message) bool {
		return msg.Type().AllLink()
	})
}

func Not(matcher Matcher) Matcher {
	return Matches(func(msg *Message) bool {
		return !matcher.Matches(msg)
	})
}

func And(matchers ...Matcher) Matcher {
	return Matches(func(msg *Message) bool {
		for _, matcher := range matchers {
			if !matcher.Matches(msg) {
				return false
			}
		}
		return true
	})
}

func Or(matchers ...Matcher) Matcher {
	return Matches(func(msg *Message) bool {
		for _, matcher := range matchers {
			if matcher.Matches(msg) {
				return true
			}
		}
		return false
	})
}

// MatchAck will match the message that corresponds to the
// given ack message
func MatchAck(ack *Message) Matcher {
	return And(Not(AckMatcher()), SrcMatcher(ack.Src), CmdMatcher(ack.Command))
}
