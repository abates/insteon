package insteon

import (
	"reflect"
	"testing"
)

func TestRingPush(t *testing.T) {
	a1 := Address{0, 0, 1}
	a2 := Address{0, 0, 2}
	a3 := Address{0, 0, 3}
	a4 := Address{0, 0, 4}
	r := &ring{
		messages: []*Message{
			&Message{Src: a1},
			&Message{Src: a2},
			&Message{Src: a3},
		},
		i: 2,
		j: 2,
	}

	r.push(&Message{Src: a4})
	if r.messages[0].Src != a4 {
		t.Errorf("Wanted message with src %v got %v", a4, r.messages[0].Src)
	}

	if r.messages[1].Src != a2 {
		t.Errorf("Wanted message with src %v got %v", a2, r.messages[1].Src)
	}

	if r.messages[2].Src != a3 {
		t.Errorf("Wanted message with src %v got %v", a3, r.messages[2].Src)
	}

	r.push(&Message{Src: a1})
	if r.messages[0].Src != a4 {
		t.Errorf("Wanted message with src %v got %v", a4, r.messages[0].Src)
	}

	if r.messages[1].Src != a1 {
		t.Errorf("Wanted message with src %v got %v", a1, r.messages[1].Src)
	}

	if r.messages[2].Src != a3 {
		t.Errorf("Wanted message with src %v got %v", a3, r.messages[2].Src)
	}

	if len(r.messages) != 3 {
		t.Errorf("Wanted messages length to be 3 got %d", len(r.messages))
	}
}

func TestRingMatches(t *testing.T) {
	tests := []struct {
		name    string
		input   []*Message
		matcher Matcher
		want    bool
	}{
		{"found", []*Message{&Message{Src: Address{0, 1, 2}}, &Message{Src: Address{1, 2, 3}}}, SrcMatcher(Address{1, 2, 3}), true},
		{"not found", []*Message{&Message{Src: Address{0, 1, 2}}, &Message{Src: Address{1, 2, 3}}}, SrcMatcher(Address{3, 4, 5}), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := &ring{messages: make([]*Message, len(test.input))}
			for _, msg := range test.input {
				r.push(msg)
			}
			_, got := r.matches(test.matcher)
			if test.want != got {
				t.Errorf("Wanted %v got %v", test.want, got)
			}
		})
	}
}

func TestTTLFilter(t *testing.T) {
	tests := []struct {
		name    string
		wantTTL uint8
	}{
		{"ttl 1", 1},
		{"ttl 2", 2},
		{"ttl 3", 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tw := &testWriter{}
			f := TTL(int(test.wantTTL))(tw)
			f.Write(&Message{})
			if tw.written[0].TTL() != test.wantTTL {
				t.Errorf("Wanted ttl to be %d got %d", test.wantTTL, tw.written[0].TTL())
			}

			if tw.written[0].MaxTTL() != test.wantTTL {
				t.Errorf("Wanted max ttl to be %d got %d", test.wantTTL, tw.written[0].MaxTTL())
			}
		})
	}
}

func TestFilterDuplicates(t *testing.T) {
	tests := []struct {
		name  string
		input []*Message
		want  []*Message
	}{
		{
			name:  "no duplicates",
			input: []*Message{&Message{Src: Address{1, 2, 3}}, &Message{Src: Address{1, 2, 4}}, &Message{Src: Address{1, 2, 5}}},
			want:  []*Message{&Message{Src: Address{1, 2, 3}}, &Message{Src: Address{1, 2, 4}}, &Message{Src: Address{1, 2, 5}}},
		},
		{
			name: "duplicates",
			input: []*Message{
				&Message{Flags: Flag(MsgTypeDirect, false, 3, 3), Src: Address{1, 2, 3}},
				&Message{Flags: Flag(MsgTypeDirect, false, 2, 3), Src: Address{1, 2, 3}},
				&Message{Src: Address{1, 2, 4}},
				&Message{Src: Address{1, 2, 5}},
			},
			want: []*Message{&Message{Flags: Flag(MsgTypeDirect, false, 3, 3), Src: Address{1, 2, 3}}, &Message{Src: Address{1, 2, 4}}, &Message{Src: Address{1, 2, 5}}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tw := &testWriter{
				read: test.input,
			}
			f := FilterDuplicates()(tw)
			got := []*Message{}
			for msg, err := f.Read(); err == nil; msg, err = f.Read() {
				got = append(got, msg)
			}

			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("Wanted messages %v got %v", test.want, got)
			}
		})
	}
}
