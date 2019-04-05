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
	"time"
)

func TestI2DeviceIsLinkable(t *testing.T) {
	device := Device(&i2Device{})
	linkable := device.(Linkable)
	if linkable == nil {
		t.Error("linkable should not be nil")
	}
}

func TestI2DeviceCommands(t *testing.T) {
	tests := []*commandTest{
		{"AddLink", func(d Device) error { return d.(*i2Device).AddLink(nil) }, Command{}, ErrNotImplemented, nil},
		{"RemoveLinks", func(d Device) error { return d.(*i2Device).RemoveLinks(nil) }, Command{}, ErrNotImplemented, nil},
		{"ExitLinkingMode", func(d Device) error { return d.(*i2Device).ExitLinkingMode() }, CmdExitLinkingMode, nil, nil},
		{"WriteLink - error", func(d Device) error { return d.(*i2Device).WriteLink(&LinkRecord{}) }, CmdReadWriteALDB, ErrInvalidMemAddress, nil},
		{"WriteLink", func(d Device) error { return d.(*i2Device).WriteLink(&LinkRecord{memAddress: 0x01}) }, CmdReadWriteALDB, nil, nil},
	}

	testDeviceCommands(t, func(conn *testConnection) Device { return newI2Device(conn, time.Millisecond) }, tests)
}

func TestI2DeviceEnterLinkingMode(t *testing.T) {
	constructor := func(conn *testConnection) Device { return newI2Device(conn, time.Millisecond) }
	callback := func(d Device) error { return d.(*i2Device).EnterLinkingMode(10) }
	// happy path
	testDeviceCommand(t, constructor, callback, CmdEnterLinkingMode.SubCommand(10), nil, nil, &Message{Flags: StandardBroadcast, Command: CmdSetButtonPressedResponder})

	// sad path
	testDeviceCommand(t, constructor, callback, CmdEnterLinkingMode.SubCommand(10), nil, ErrReadTimeout)
}

func TestI2DeviceEnterUnlinkingMode(t *testing.T) {
	constructor := func(conn *testConnection) Device { return newI2Device(conn, time.Millisecond) }
	callback := func(d Device) error { return d.(*i2Device).EnterUnlinkingMode(10) }
	// happy path
	testDeviceCommand(t, constructor, callback, CmdEnterUnlinkingMode.SubCommand(10), nil, nil, &Message{Flags: StandardBroadcast, Command: CmdSetButtonPressedResponder})

	// sad path
	testDeviceCommand(t, constructor, callback, CmdEnterUnlinkingModeExt.SubCommand(10), nil, ErrReadTimeout)
}

func i2DeviceLinks(conn *testConnection) []*linkRequest {
	linkRequests := []*linkRequest{
		{MemAddress: 0xffff, Type: 0x02, Link: &LinkRecord{Flags: 0x01}},
		{MemAddress: 0, Type: 0x02, Link: &LinkRecord{}},
	}
	conn.recvCh = make(chan *Message, len(linkRequests))

	msgs := []*Message{}
	for _, lr := range linkRequests {
		msg := &Message{Command: CmdReadWriteALDB, Flags: ExtendedDirectMessage, Payload: make([]byte, 14)}
		buf, _ := lr.MarshalBinary()
		copy(msg.Payload, buf)
		msgs = append(msgs, msg)
	}

	for _, msg := range msgs {
		conn.recvCh <- msg
	}

	return linkRequests
}

func TestI2DeviceLinks(t *testing.T) {
	conn := &testConnection{sendCh: make(chan *Message, 1), ackCh: make(chan *Message, 1)}
	device := newI2Device(conn, time.Nanosecond)

	i2DeviceLinks(conn)
	conn.ackCh <- TestAck

	links, err := device.Links()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if len(links) != 1 {
		t.Errorf("want 1 link got %v", len(links))
	}
	<-conn.sendCh

	// test sad path
	conn.ackCh <- TestAck
	go func() {
		time.Sleep(time.Millisecond)
		conn.recvCh <- TestMessagePing
	}()

	_, err = device.Links()
	if err != ErrReadTimeout {
		t.Errorf("want ErrReadTimeout got %v", err)
	}
}

func TestI2AppendLink(t *testing.T) {
	conn := &testConnection{sendCh: make(chan *Message, 1), ackCh: make(chan *Message, 2)}
	device := newI2Device(conn, time.Nanosecond)

	go func() {
		err := device.AppendLink(&LinkRecord{})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}()

	i2DeviceLinks(conn)
	// Ack the ALDB request for links
	conn.ackCh <- TestAck
	// Ack the ALDB for write link
	conn.ackCh <- TestAck

	// receive the ALDB request
	<-conn.sendCh

	// receive the write link request
	<-conn.sendCh
}
