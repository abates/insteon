package insteon

import (
	"testing"
	"time"
)

func TestPropagationDelay(t *testing.T) {
	tests := []struct {
		name          string
		inputTTL      uint8
		inputExtended bool
		want          time.Duration
	}{
		{"ttl 0, standard", 0, false, time.Second * 12 / 60},
		{"ttl 0, extended", 0, true, time.Second * 26 / 60},
		{"ttl 2, standard", 2, false, time.Second * 36 / 60},
		{"ttl 2, extended", 2, true, time.Second * 78 / 60},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := PropagationDelay(test.inputTTL, test.inputExtended)
			if test.want != got {
				t.Errorf("Wanted delay %v got %v", test.want, got)
			}
		})
	}
}

func TestReadWithTimeout(t *testing.T) {
	fillChan := func(messages ...*Message) <-chan *Message {
		ch := make(chan *Message, len(messages))
		for _, msg := range messages {
			ch <- msg
		}
		return ch
	}

	tests := []struct {
		name         string
		inputCh      <-chan *Message
		inputTimeout time.Duration
		want         error
	}{
		{"success", fillChan(&Message{}), time.Second, nil},
		{"success", fillChan(), time.Microsecond, ErrReadTimeout},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, got := ReadWithTimeout(test.inputCh, test.inputTimeout)
			if test.want != got {
				t.Errorf("Wanted error %v got %v", test.want, got)
			}
		})
	}
}
