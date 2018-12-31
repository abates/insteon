package plm

import (
	"testing"
	"time"
)

func TestPlmTimeout(t *testing.T) {
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
			t.Errorf("Expected %v got %v", ErrReadTimeout, request.Err)
		}
		close(upstreamRecvCh)
	}()

	plm.process()
}
