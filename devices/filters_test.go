package devices

import (
	"reflect"
	"testing"

	"github.com/abates/insteon"
	"github.com/abates/insteon/commands"
)

func TestCacheReadWrite(t *testing.T) {
	m1 := &insteon.Message{Src: insteon.Address{1, 2, 3}}
	m2 := &insteon.Message{Src: insteon.Address{9, 8, 7}}
	tw := &testWriter{
		read: []*insteon.Message{m2},
	}

	c := NewCache(3)
	mw := c.Filter(tw)
	mw.Write(m1)

	if c.Messages[0] != m1 {
		t.Errorf("Expected message %v got %v", m1, c.Messages[0])
	}

	g2, _ := mw.Read()
	if g2 != m2 {
		t.Errorf("Expected message %v got %v", m2, g2)
	}

	if c.Messages[1] != m2 {
		t.Errorf("Expected message %v got %v", m2, c.Messages[1])
	}
}

func TestCachePush(t *testing.T) {
	a1 := insteon.Address{0, 0, 1}
	a2 := insteon.Address{0, 0, 2}
	a3 := insteon.Address{0, 0, 3}
	a4 := insteon.Address{0, 0, 4}
	c := NewCache(3, &insteon.Message{Src: a1}, &insteon.Message{Src: a2}, &insteon.Message{Src: a3})

	c.push(&insteon.Message{Src: a4})
	if c.Messages[0].Src != a4 {
		t.Errorf("Wanted message with src %v got %v", a4, c.Messages[0].Src)
	}

	if c.Messages[1].Src != a2 {
		t.Errorf("Wanted message with src %v got %v", a2, c.Messages[1].Src)
	}

	if c.Messages[2].Src != a3 {
		t.Errorf("Wanted message with src %v got %v", a3, c.Messages[2].Src)
	}

	c.push(&insteon.Message{Src: a1})
	c.push(nil)
	if c.Messages[0].Src != a4 {
		t.Errorf("Wanted message with src %v got %v", a4, c.Messages[0].Src)
	}

	if c.Messages[1].Src != a1 {
		t.Errorf("Wanted message with src %v got %v", a1, c.Messages[1].Src)
	}

	if c.Messages[2].Src != a3 {
		t.Errorf("Wanted message with src %v got %v", a3, c.Messages[2].Src)
	}

	if len(c.Messages) != 3 {
		t.Errorf("Wanted messages length to be 3 got %d", len(c.Messages))
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
		{
			name:    "empty cache",
			inputI:  1,
			input:   nil,
			matcher: SrcMatcher(insteon.Address{1, 2, 3}),
			want:    false,
		},
		{
			name:   "found",
			inputI: 1,
			input: []*insteon.Message{
				&insteon.Message{Src: insteon.Address{0, 1, 2}},
				&insteon.Message{Src: insteon.Address{1, 2, 3}},
			},
			matcher: SrcMatcher(insteon.Address{1, 2, 3}),
			want:    true,
		},
		{
			name:   "not found",
			inputI: 1,
			input: []*insteon.Message{
				&insteon.Message{Src: insteon.Address{0, 1, 2}},
				&insteon.Message{Src: insteon.Address{1, 2, 3}},
			},
			matcher: SrcMatcher(insteon.Address{3, 4, 5}),
			want:    false,
		},
		{
			name:   "found",
			inputI: 0,
			input: []*insteon.Message{
				&insteon.Message{Src: insteon.Address{0, 1, 2}},
				&insteon.Message{Src: insteon.Address{1, 2, 3}},
				&insteon.Message{Src: insteon.Address{4, 1, 2}},
				&insteon.Message{Src: insteon.Address{3, 5, 7}},
			},
			matcher: SrcMatcher(insteon.Address{1, 2, 3}),
			want:    true,
		},
		{
			name:   "ack",
			inputI: 0,
			input: []*insteon.Message{
				&insteon.Message{Dst: insteon.Address{0, 1, 2}, Command: commands.ReadWriteALDB, Flags: insteon.Flag(insteon.MsgTypeDirect, true, 2, 2), Payload: make([]byte, 14)},
			},
			matcher: MatchAck(&insteon.Message{Src: insteon.Address{0, 1, 2}, Command: commands.ReadWriteALDB, Flags: insteon.Flag(insteon.MsgTypeDirectAck, false, 2, 2)}),
			want:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := NewCache(len(test.input), test.input...)
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
