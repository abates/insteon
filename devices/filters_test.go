package devices

import (
	"reflect"
	"testing"

	"github.com/abates/insteon"
)

func TestCacheReadWrite(t *testing.T) {
	m1 := &insteon.Message{Src: insteon.Address{1, 2, 3}}
	m2 := &insteon.Message{Src: insteon.Address{9, 8, 7}}
	tw := &testWriter{
		read: []*insteon.Message{m2},
	}

	c := newCache(3)
	mw := c.Filter(tw)
	mw.Write(m1)

	if c.messages[0] != m1 {
		t.Errorf("Expected message %v got %v", m1, c.messages[0])
	}

	g2, _ := mw.Read()
	if g2 != m2 {
		t.Errorf("Expected message %v got %v", m2, g2)
	}

	if c.messages[1] != m2 {
		t.Errorf("Expected message %v got %v", m2, c.messages[1])
	}
}

func TestCachePush(t *testing.T) {
	a1 := insteon.Address{0, 0, 1}
	a2 := insteon.Address{0, 0, 2}
	a3 := insteon.Address{0, 0, 3}
	a4 := insteon.Address{0, 0, 4}
	c := newCache(3, &insteon.Message{Src: a1}, &insteon.Message{Src: a2}, &insteon.Message{Src: a3})

	c.push(&insteon.Message{Src: a4})
	if c.messages[0].Src != a4 {
		t.Errorf("Wanted message with src %v got %v", a4, c.messages[0].Src)
	}

	if c.messages[1].Src != a2 {
		t.Errorf("Wanted message with src %v got %v", a2, c.messages[1].Src)
	}

	if c.messages[2].Src != a3 {
		t.Errorf("Wanted message with src %v got %v", a3, c.messages[2].Src)
	}

	c.push(&insteon.Message{Src: a1})
	c.push(nil)
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
		input   []*insteon.Message
		matcher Matcher
		want    bool
	}{
		{"empty cache", 1, nil, SrcMatcher(insteon.Address{1, 2, 3}), false},
		{"found", 1, []*insteon.Message{&insteon.Message{Src: insteon.Address{0, 1, 2}}, &insteon.Message{Src: insteon.Address{1, 2, 3}}}, SrcMatcher(insteon.Address{1, 2, 3}), true},
		{"not found", 1, []*insteon.Message{&insteon.Message{Src: insteon.Address{0, 1, 2}}, &insteon.Message{Src: insteon.Address{1, 2, 3}}}, SrcMatcher(insteon.Address{3, 4, 5}), false},
		{"found", 0, []*insteon.Message{&insteon.Message{Src: insteon.Address{0, 1, 2}}, &insteon.Message{Src: insteon.Address{1, 2, 3}}, &insteon.Message{Src: insteon.Address{4, 1, 2}}, &insteon.Message{Src: insteon.Address{3, 5, 7}}}, SrcMatcher(insteon.Address{1, 2, 3}), true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := newCache(len(test.input), test.input...)
			c.i = test.inputI
			_, got := c.Lookup(test.matcher)
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
			f := TTL(int(test.wantTTL)).Filter(tw)
			f.Write(&insteon.Message{})
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
		input []*insteon.Message
		want  []*insteon.Message
	}{
		{
			name:  "no duplicates",
			input: []*insteon.Message{&insteon.Message{Src: insteon.Address{1, 2, 3}}, &insteon.Message{Src: insteon.Address{1, 2, 4}}, &insteon.Message{Src: insteon.Address{1, 2, 5}}},
			want:  []*insteon.Message{&insteon.Message{Src: insteon.Address{1, 2, 3}}, &insteon.Message{Src: insteon.Address{1, 2, 4}}, &insteon.Message{Src: insteon.Address{1, 2, 5}}},
		},
		{
			name: "duplicates",
			input: []*insteon.Message{
				&insteon.Message{Flags: insteon.Flag(insteon.MsgTypeDirect, false, 3, 3), Src: insteon.Address{1, 2, 3}},
				&insteon.Message{Flags: insteon.Flag(insteon.MsgTypeDirect, false, 2, 3), Src: insteon.Address{1, 2, 3}},
				&insteon.Message{Src: insteon.Address{1, 2, 4}},
				&insteon.Message{Src: insteon.Address{1, 2, 5}},
			},
			want: []*insteon.Message{&insteon.Message{Flags: insteon.Flag(insteon.MsgTypeDirect, false, 3, 3), Src: insteon.Address{1, 2, 3}}, &insteon.Message{Src: insteon.Address{1, 2, 4}}, &insteon.Message{Src: insteon.Address{1, 2, 5}}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tw := &testWriter{
				read: test.input,
			}
			f := FilterDuplicates().Filter(tw)
			got := []*insteon.Message{}
			for msg, err := f.Read(); err == nil; msg, err = f.Read() {
				got = append(got, msg)
			}

			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("Wanted messages %v got %v", test.want, got)
			}
		})
	}
}
