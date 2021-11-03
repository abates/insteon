package insteon

import (
	"reflect"
	"testing"
)

func TestCachePush(t *testing.T) {
	a1 := Address{0, 0, 1}
	a2 := Address{0, 0, 2}
	a3 := Address{0, 0, 3}
	a4 := Address{0, 0, 4}
	c := newCache(3, &Message{Src: a1}, &Message{Src: a2}, &Message{Src: a3})

	c.push(&Message{Src: a4})
	if c.messages[0].Src != a4 {
		t.Errorf("Wanted message with src %v got %v", a4, c.messages[0].Src)
	}

	if c.messages[1].Src != a2 {
		t.Errorf("Wanted message with src %v got %v", a2, c.messages[1].Src)
	}

	if c.messages[2].Src != a3 {
		t.Errorf("Wanted message with src %v got %v", a3, c.messages[2].Src)
	}

	c.push(&Message{Src: a1})
	if c.messages[0].Src != a4 {
		t.Errorf("Wanted message with src %v got %v", a4, c.messages[0].Src)
	}

	if c.messages[1].Src != a1 {
		t.Errorf("Wanted message with src %v got %v", a1, c.messages[1].Src)
	}

	if c.messages[2].Src != a3 {
		t.Errorf("Wanted message with src %v got %v", a3, c.messages[2].Src)
	}

	if len(c.messages) != 3 {
		t.Errorf("Wanted messages length to be 3 got %d", len(c.messages))
	}
}

func TestCacheLookup(t *testing.T) {
	tests := []struct {
		name    string
		inputI  int
		input   []*Message
		matcher Matcher
		want    bool
	}{
		{"found", 1, []*Message{&Message{Src: Address{0, 1, 2}}, &Message{Src: Address{1, 2, 3}}}, SrcMatcher(Address{1, 2, 3}), true},
		{"not found", 1, []*Message{&Message{Src: Address{0, 1, 2}}, &Message{Src: Address{1, 2, 3}}}, SrcMatcher(Address{3, 4, 5}), false},
		{"found", 0, []*Message{&Message{Src: Address{0, 1, 2}}, &Message{Src: Address{1, 2, 3}}, &Message{Src: Address{4, 1, 2}}, &Message{Src: Address{3, 5, 7}}}, SrcMatcher(Address{1, 2, 3}), true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := newCache(len(test.input), test.input...)
			c.i = test.inputI
			_, got := c.lookup(test.matcher)
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
