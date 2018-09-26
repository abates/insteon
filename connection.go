package insteon

import (
	"errors"
	"time"
)

var (
	ErrClosed = errors.New("Connection Closed")
)

type connection struct {
	addr    Address
	match   []Command
	version EngineVersion
	timeout time.Duration

	sendCh         chan *MessageRequest
	upstreamSendCh chan<- *MessageRequest
	recvCh         chan *Message
	upstreamRecvCh <-chan *Message

	queue []*MessageRequest
}

func newConnection(upstreamSendCh chan<- *MessageRequest, upstreamRecvCh <-chan *Message, addr Address, version EngineVersion, timeout time.Duration, match ...Command) *connection {
	conn := &connection{
		addr:    addr,
		match:   match,
		version: version,
		timeout: timeout,

		sendCh:         make(chan *MessageRequest, 1),
		upstreamSendCh: upstreamSendCh,
		recvCh:         make(chan *Message, 1),
		upstreamRecvCh: upstreamRecvCh,
	}

	go conn.process()
	return conn
}

func (conn *connection) process() {
	for {
		select {
		case msg, open := <-conn.upstreamRecvCh:
			if !open {
				close(conn.recvCh)
				close(conn.upstreamSendCh)
				return
			}
			conn.receive(msg)
		case request, open := <-conn.sendCh:
			if !open {
				close(conn.recvCh)
				close(conn.upstreamSendCh)
				return
			}
			conn.queue = append(conn.queue, request)
			if len(conn.queue) == 1 {
				conn.send()
			}
		case <-time.After(conn.timeout):
			// prevent head of line blocking for a lost/nonexistant Ack
			if len(conn.queue) > 0 && conn.queue[0].timeout.Before(time.Now()) {
				conn.queue[0].Err = ErrReadTimeout
				conn.queue[0].DoneCh <- conn.queue[0]
				close(conn.queue[0].DoneCh)
				conn.queue = conn.queue[1:]
				conn.send()
			}
		}
	}
}

func (conn *connection) receiveAck(msg *Message) {
	if len(conn.queue) > 0 {
		request := conn.queue[0]
		if msg.Command[0] == request.Message.Command[0] {
			conn.queue[0].Ack = msg
			if msg.Flags.Type() == MsgTypeDirectNak {
				if VerI1 <= conn.version && conn.version <= VerI2 {
					switch msg.Command[1] {
					case 0xfd:
						request.Err = ErrUnknownCommand
					case 0xfe:
						request.Err = ErrNoLoadDetected
					case 0xff:
						request.Err = ErrNotLinked
					default:
						request.Err = NewTraceError(ErrUnexpectedResponse)
					}
				} else if conn.version == VerI2Cs {
					switch msg.Command[1] {
					case 0xfb:
						request.Err = ErrIllegalValue
					case 0xfc:
						request.Err = ErrPreNak
					case 0xfd:
						request.Err = ErrIncorrectChecksum
					case 0xfe:
						request.Err = ErrNoLoadDetected
					case 0xff:
						request.Err = ErrNotLinked
					default:
						request.Err = NewTraceError(ErrUnexpectedResponse)
					}
				}
			}

			conn.queue[0].DoneCh <- conn.queue[0]
			close(conn.queue[0].DoneCh)

			conn.queue = conn.queue[1:]
			conn.send()
		}
	}
}

func (conn *connection) receiveMatch(msg *Message) {
	for _, m := range conn.match {
		if msg.Command == m || msg.Command[0] == m[0] && m[1] == 0x00 {
			conn.recvCh <- msg
		}
	}
}

func (conn *connection) receive(msg *Message) {
	if msg.Src == conn.addr {
		if msg.Ack() || msg.Nak() {
			conn.receiveAck(msg)
		} else if len(conn.match) > 0 {
			conn.receiveMatch(msg)
		} else {
			conn.recvCh <- msg
		}
	}
}

func (conn *connection) send() {
	if len(conn.queue) > 0 {
		request := conn.queue[0]
		request.timeout = time.Now().Add(conn.timeout)
		request.Message.Dst = conn.addr

		oldCh := request.DoneCh
		doneCh := make(chan *MessageRequest, 1)
		request.DoneCh = doneCh
		conn.upstreamSendCh <- request
		<-doneCh
		request.DoneCh = oldCh

		if request.Err != nil {
			conn.queue = conn.queue[1:]
			request.DoneCh <- request
			close(request.DoneCh)
		}
	}
}

/*func (conn *connection) SendMessage(msg *Message) (*Message, error) {
	doneCh := make(chan bool, 1)
	request := &MessageRequest{Message: msg, DoneCh: doneCh}
	conn.sendCh <- request
	<-doneCh
	return request.Ack, request.Err
}

func (conn *connection) ReceiveMessage() (msg *Message, err error) {
	select {
	case msg = <-conn.recvCh:
	case <-time.After(conn.timeout):
		err = ErrReadTimeout
	}
	return
}

func (conn *connection) Close() error {
	close(conn.sendCh)
	return nil
}*/
