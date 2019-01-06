package plm

import (
	"testing"
	"time"
)

func TestPlmTimeout(t *testing.T) {
	t.Parallel()
	upstreamRecvCh := make(chan []byte, 1)
	doneCh := make(chan *PacketRequest, 1)
	request := &PacketRequest{DoneCh: doneCh}
	plm := &PLM{
		timeout:        time.Millisecond,
		upstreamRecvCh: upstreamRecvCh,
		upstreamSendCh: make(chan []byte),
		queue:          []*PacketRequest{request},
	}
	go func() {
		<-doneCh
		if request.Err != ErrReadTimeout {
			t.Errorf("got error %v, want %v", request.Err, ErrReadTimeout)
		}
		close(upstreamRecvCh)
	}()

	plm.process()
}

func TestPlmOption(t *testing.T) {
	t.Parallel()
	want := 1234 * time.Millisecond

	without := New(&Port{}, 5*time.Second)
	if without.writeDelay == want {
		t.Errorf("writeDelay is %v, expected anything else", without.writeDelay)
	}

	with := New(&Port{}, 5*time.Second, WriteDelay(want))
	if with.writeDelay != want {
		t.Errorf("writeDelay is %v, want %v", without.writeDelay, want)
	}

}
