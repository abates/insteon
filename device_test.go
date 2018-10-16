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
	"encoding"
	"testing"
	"time"
)

func TestDeviceRegistry(t *testing.T) {
	dr := &DeviceRegistry{}

	if _, found := dr.Find(Category(1)); found {
		t.Errorf("Expected nothing found for Category(1)")
	}

	dr.Register(Category(1), func(info DeviceInfo, address Address, sendCh chan<- *MessageRequest, recvCh <-chan *Message, timeout time.Duration) (Device, error) {
		return nil, nil
	})

	if _, found := dr.Find(Category(1)); !found {
		t.Errorf("Expected to find Category(1)")
	}

	dr.Delete(Category(1))
	if _, found := dr.Find(Category(1)); found {
		t.Errorf("Expected nothing found for Category(1)")
	}
}

func testRecv(sendCh <-chan *CommandRequest, ackMsg *Message, respCmd Command, payloads ...encoding.BinaryMarshaler) {
	// receive command request and send ack
	request := <-sendCh
	request.Ack = ackMsg
	request.DoneCh <- request

	doneCh := make(chan *CommandResponse, 1)

	// return subsequent traffic
	for {
		if len(payloads) > 0 {
			msg := &Message{Command: respCmd}
			msg.Payload, _ = payloads[0].MarshalBinary()
			msg.Payload = append(msg.Payload, make([]byte, 14-len(msg.Payload))...)
			request.RecvCh <- &CommandResponse{Message: msg, DoneCh: doneCh}
			payloads = payloads[1:]
		} else {
			<-doneCh
			close(request.RecvCh)
			return
		}
	}
}
