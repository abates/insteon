package plm

import "github.com/abates/insteon"

type connection struct {
	sendCmd        Command
	matches        []Command
	sendCh         chan *insteon.PacketRequest
	upstreamSendCh chan<- *CommandRequest
	recvCh         chan []byte
	upstreamRecvCh <-chan *Packet
}

func newConnection(sendCh chan<- *CommandRequest, recvCh <-chan *Packet, sendCmd Command, recvCmds ...Command) *connection {
	conn := &connection{
		sendCmd:        sendCmd,
		matches:        recvCmds,
		sendCh:         make(chan *insteon.PacketRequest, 1),
		upstreamSendCh: sendCh,
		recvCh:         make(chan []byte, 1),
		upstreamRecvCh: recvCh,
	}

	go conn.process()

	return conn
}

func (conn *connection) process() {
	for {
		select {
		case request, open := <-conn.sendCh:
			if !open {
				close(conn.upstreamSendCh)
				close(conn.recvCh)
				return
			}
			conn.send(request)
		case packet, open := <-conn.upstreamRecvCh:
			if !open {
				close(conn.upstreamSendCh)
				close(conn.recvCh)
				return
			}
			conn.receive(packet)
		}
	}
}

func (conn *connection) send(request *insteon.PacketRequest) {
	doneCh := make(chan *CommandRequest)
	payload := request.Payload
	// PLM expects that the payload begins with the
	// destinations address so we have to slice off
	// the src address
	if conn.sendCmd == CmdSendInsteonMsg && len(payload) > 3 {
		payload = payload[3:]
	}

	conn.upstreamSendCh <- &CommandRequest{Command: conn.sendCmd, Payload: payload, DoneCh: doneCh}
	upstreamRequest := <-doneCh
	request.Err = upstreamRequest.Err
	request.DoneCh <- request
}

func (conn *connection) receive(packet *Packet) {
	if len(conn.matches) > 0 {
		for _, match := range conn.matches {
			if match == packet.Command {
				conn.recvCh <- packet.Payload
				return
			}
		}
	} else {
		conn.recvCh <- packet.Payload
	}
}
