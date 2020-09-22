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
	"fmt"
	"sync"
	"time"
)

const (
	NumRetries = 5

	Timeout = time.Second * 5
)

type MessageWriter interface {
	WriteMessage(*Message) error
}

type ConnectionConfig struct {
	Timeout time.Duration
	Retries int
	TTL     uint8
	BufSize int
}

type Bus interface {
	Publish(*Message) (*Message, error)
	Subscribe(src Address, matcher Matcher) <-chan *Message
	Unsubscribe(src Address, ch <-chan *Message)
	Config() ConnectionConfig
}

type Matcher interface {
	Matches(msg *Message) bool
}

type Matches func(msg *Message) bool

func (m Matches) Matches(msg *Message) bool {
	return m(msg)
}

func OrMatcher(matchers ...Matcher) Matcher {
	return Matches(func(msg *Message) bool {
		for _, matcher := range matchers {
			if matcher.Matches(msg) {
				return true
			}
		}
		return false
	})
}

type CmdMatcher Command

func (m CmdMatcher) Matches(msg *Message) bool {
	return Command(m).Command1() == msg.Command.Command1()
}

type subscriber struct {
	src     Address
	matcher Matcher
	ch      chan *Message
	readCh  <-chan *Message
}

type bus struct {
	writer      MessageWriter
	subscribe   chan *subscriber
	unsubscribe chan *subscriber
	closeCh     chan chan error
	listeners   map[Address]map[<-chan *Message]*subscriber
	config      ConnectionConfig
}

func NewBus(writer MessageWriter, messages <-chan *Message, options ...ConnectionOption) (Bus, error) {
	b := &bus{
		writer:      writer,
		subscribe:   make(chan *subscriber),
		unsubscribe: make(chan *subscriber),
		listeners:   make(map[Address]map[<-chan *Message]*subscriber),
		closeCh:     make(chan chan error),
		config: ConnectionConfig{
			Timeout: time.Second * 3,
			Retries: 3,
			TTL:     3,
			BufSize: 32,
		},
	}

	for _, option := range options {
		err := option(&b.config)
		if err != nil {
			return b, err
		}
	}

	go b.run(messages)
	return b, nil
}

func (b *bus) run(messages <-chan *Message) {
	Log.Debugf("Starting BUS")
	var workers sync.WaitGroup
	for {
		select {
		case msg := <-messages:
			Log.Debugf("Bus received %v", msg)
			for _, s := range b.listeners[msg.Src] {
				if s.matcher.Matches(msg) {
					// run this in a go routine so a wayward listener can't block up the works
					workers.Add(1)
					go func(msg *Message) {
						select {
						case s.ch <- msg:
						default:
							Log.Infof("Receive buffer full for %v listener", msg.Src)
							//case <-time.After(time.Second * 3): // this is just an arbitrarily long duration to wait, I don't think it needs to be configurable
							//Log.Infof("Timeout attempting to deliver message from %v to listener", msg.Src)
						}
						workers.Done()
					}(msg)
				}
			}

		case s := <-b.subscribe:
			m, found := b.listeners[s.src]
			if !found {
				m = make(map[<-chan *Message]*subscriber)
				b.listeners[s.src] = m
			}

			m[s.ch] = s
			Log.Debugf("Subscribed channel to %s", s.src)
		case s := <-b.unsubscribe:
			if m, found := b.listeners[s.src]; found {
				delete(m, s.readCh)
			}
		case closeCh := <-b.closeCh:
			defer func() { closeCh <- nil }()
			workers.Wait()
			return
		}
	}
}

func (b *bus) Subscribe(src Address, matcher Matcher) <-chan *Message {
	ch := make(chan *Message, b.config.BufSize)
	b.subscribe <- &subscriber{src: src, matcher: matcher, ch: ch}
	return ch
}

func (b *bus) Unsubscribe(src Address, ch <-chan *Message) {
	b.unsubscribe <- &subscriber{src: src, readCh: ch}
}

func (b *bus) Publish(msg *Message) (*Message, error) {
	if msg.Flags.Type().Direct() {
		return b.publishDirect(msg)
	} else {
		//b.publishBroadcast(msg)
	}
	return nil, nil
}

func (b *bus) Config() ConnectionConfig {
	return b.config
}

func (b *bus) Close() error {
	ch := make(chan error)
	b.closeCh <- ch
	return <-ch
}

func (b *bus) publishDirect(msg *Message) (ack *Message, err error) {
	msg.Flags = Flag(MsgTypeDirect, len(msg.Payload) > 0, b.config.TTL, b.config.TTL)

	if len(msg.Payload) > 0 && len(msg.Payload) < 14 {
		tmp := make([]byte, 14)
		copy(tmp, msg.Payload)
		msg.Payload = tmp
	}

	rx := b.Subscribe(msg.Dst, CmdMatcher(msg.Command))
	defer b.Unsubscribe(msg.Dst, rx)

	err = Retry(b.config.Retries, func() error {
		Log.Debugf("Publishing %s", msg)
		err = b.writer.WriteMessage(msg)
		if err == nil {
			select {
			case ack = <-rx:
				if ack.Nak() {
					err = ErrNak
				}
			case <-time.After(b.config.Timeout):
				err = ErrReadTimeout
			}
		}
		return err
	})

	if err == ErrReadTimeout {
		err = ErrAckTimeout
	}
	return ack, err
}

// ConnectionOption provides a means to customize the connection config
type ConnectionOption func(*ConnectionConfig) error

// ConnectionBufSize sets the receive buffer size for incoming messages
func ConnectionBufSize(size int) ConnectionOption {
	return func(config *ConnectionConfig) error {
		if size < 0 {
			return fmt.Errorf("Buffer size cannot be less than zero")
		}
		config.BufSize = size
		return nil
	}
}

// ConnectionTTL will set the connection's time to live flag
func ConnectionTTL(ttl uint8) ConnectionOption {
	return func(config *ConnectionConfig) error {
		if ttl < 0 || ttl > 3 {
			return fmt.Errorf("invalid ttl %d, must be in range 0-3", ttl)
		}
		config.TTL = ttl
		return nil
	}
}

// ConnectionTimeout is a ConnectionOption that will set the connection's read
// timeout
func ConnectionTimeout(timeout time.Duration) ConnectionOption {
	return func(config *ConnectionConfig) error {
		config.Timeout = timeout
		return nil
	}
}

func ConnectionRetry(retries int) ConnectionOption {
	return func(config *ConnectionConfig) error {
		config.Retries = retries
		return nil
	}
}

func Retry(retries int, cb func() error) (err error) {
	tries := 0
	for {
		tries++
		err = cb()
		if err == nil || err != ErrReadTimeout {
			break
		}
		retries--
		if retries > 0 {
			Log.Infof("Read Timeout, retrying")
		} else {
			Log.Infof("Retry count exceeded (%d)", tries)
			break
		}
	}
	return err
}

func IDRequest(bus Bus, dst Address) (version FirmwareVersion, devCat DevCat, err error) {
	rx := bus.Subscribe(dst, OrMatcher(CmdMatcher(CmdSetButtonPressedResponder), CmdMatcher(CmdSetButtonPressedController)))
	defer bus.Unsubscribe(dst, rx)

	err = Retry(bus.Config().Retries, func() error {
		_, err := bus.Publish(&Message{Dst: dst, Flags: StandardDirectMessage, Command: CmdIDRequest})
		if err == nil {
			select {
			case msg := <-rx:
				version = FirmwareVersion(msg.Dst[2])
				devCat = DevCat{msg.Dst[0], msg.Dst[1]}
			case <-time.After(bus.Config().Timeout):
				err = ErrReadTimeout
			}
		}
		return err
	})
	return
}

func GetEngineVersion(bus Bus, dst Address) (version EngineVersion, err error) {
	ack, err := bus.Publish(&Message{Dst: dst, Flags: StandardDirectMessage, Command: CmdGetEngineVersion})
	if err == nil {
		Log.Debugf("Device %v responded with an engine version %d", dst, ack.Command.Command2())
		version = EngineVersion(ack.Command.Command2())
	} else if err == ErrNak {
		// This only happens if the device is an I2Cs device and
		// is not linked to the queryier
		if ack.Command.Command2() == 0xff {
			Log.Debugf("Device %v is an unlinked I2Cs device", dst)
			version = VerI2Cs
			err = ErrNotLinked
		} else {
			err = ErrNak
		}
	}
	return
}
