package insteon

import (
	"time"
)

// I1Device provides remote communication to version 1 engines
type I1Device struct {
	address         Address
	devCat          DevCat
	firmwareVersion FirmwareVersion
	timeout         time.Duration
	waitRequest     *CommandRequest
	queue           []*CommandRequest

	upstreamSendCh chan<- *MessageRequest
	sendCh         chan *CommandRequest
	recvCh         <-chan *Message
	doneCh         chan *MessageRequest
}

// NewI1Device will construct an I1Device for the given address
func NewI1Device(address Address, sendCh chan<- *MessageRequest, recvCh <-chan *Message, timeout time.Duration) Device {
	i1 := &I1Device{
		address:         address,
		devCat:          DevCat{0xff, 0xff},
		firmwareVersion: FirmwareVersion(0x00),
		timeout:         timeout,

		upstreamSendCh: sendCh,
		sendCh:         make(chan *CommandRequest, 1),
		recvCh:         recvCh,
		doneCh:         make(chan *MessageRequest, 1),
	}

	go i1.process()
	return i1
}

func (i1 *I1Device) process() {
	for {
		select {
		case request, open := <-i1.sendCh:
			if !open {
				close(i1.upstreamSendCh)
				return
			}

			i1.queue = append(i1.queue, request)
			if len(i1.queue) == 1 {
				i1.send()
			}
		case msg, open := <-i1.recvCh:
			if !open {
				close(i1.upstreamSendCh)
				return
			}
			i1.receive(msg)
		case request := <-i1.doneCh:
			i1.queue[0].Ack = request.Ack
			i1.queue[0].Err = request.Err
			i1.queue[0].DoneCh <- i1.queue[0]
			if i1.queue[0].RecvCh != nil {
				if i1.queue[0].Err == nil {
					i1.waitRequest = i1.queue[0]
					i1.waitRequest.timeout = time.Now().Add(i1.timeout)
				} else {
					close(i1.queue[0].RecvCh)
				}
			}
			i1.queue = i1.queue[1:]
			i1.send()
		case <-time.After(i1.timeout):
			// prevent head of line blocking for a request that hasn't signaled it is done
			if i1.waitRequest != nil && i1.waitRequest.timeout.Before(time.Now()) {
				close(i1.waitRequest.RecvCh)
				i1.waitRequest = nil
				i1.send()
			}
		}
	}
}

func (i1 *I1Device) send() {
	if i1.waitRequest == nil && len(i1.queue) > 0 {
		request := i1.queue[0]
		flags := StandardDirectMessage
		if len(request.Payload) > 0 {
			flags = ExtendedDirectMessage
		}

		i1.upstreamSendCh <- &MessageRequest{
			Message: &Message{
				Flags:   flags,
				Command: request.Command,
				Payload: request.Payload,
			},
			DoneCh: i1.doneCh,
		}
	}
}

func (i1 *I1Device) receive(msg *Message) {
	if i1.waitRequest != nil {
		doneCh := make(chan bool, 1)
		i1.waitRequest.RecvCh <- &CommandResponse{Message: msg, DoneCh: doneCh}
		if <-doneCh {
			close(i1.waitRequest.RecvCh)
			i1.waitRequest = nil
			i1.send()
		}
	}
}

// SendCommandAndListen performs the same function as SendCommand.  However, instead of returning
// the Ack/Nak command, it returns a channel that can be read to get messages received after
// the command was sent.  This is useful for things like retrieving the link database where the
// response information is not in the Ack but in one or more ALDB responses.  Once all information
// has been received the command response DoneCh should be sent a "false" value to indicate no
// more messages are expected.
func (i1 *I1Device) SendCommandAndListen(command Command, payload []byte) (<-chan *CommandResponse, error) {
	recvCh := make(chan *CommandResponse, 1)
	_, err := i1.sendCommand(command, payload, recvCh)
	return recvCh, err
}

func (i1 *I1Device) sendCommand(command Command, payload []byte, recvCh chan<- *CommandResponse) (response Command, err error) {
	doneCh := make(chan *CommandRequest, 1)
	request := &CommandRequest{
		Command: command,
		Payload: payload,
		DoneCh:  doneCh,
		RecvCh:  recvCh,
	}

	i1.sendCh <- request
	<-doneCh

	if request.Err == nil {
		response = request.Ack.Command
	}

	return response, request.Err
}

// SendCommand will send the given command bytes to the device including
// a payload (for extended messages). If payload length is zero then a standard
// length message is used to deliver the commands. The command bytes from the
// response ack are returned as well as any error
func (i1 *I1Device) SendCommand(command Command, payload []byte) (response Command, err error) {
	return i1.sendCommand(command, payload, nil)
}

// Address is the Insteon address of the device
func (i1 *I1Device) Address() Address {
	return i1.address
}

func extractError(v interface{}, err error) error {
	return err
}

// AssignToAllLinkGroup will inform the device what group should be used during an All-Linking
// session
func (i1 *I1Device) AssignToAllLinkGroup(group Group) error {
	return extractError(i1.SendCommand(CmdAssignToAllLinkGroup.SubCommand(int(group)), nil))
}

// DeleteFromAllLinkGroup will inform the device which group should be unlinked during an
// All-Link unlinking session
func (i1 *I1Device) DeleteFromAllLinkGroup(group Group) (err error) {
	return extractError(i1.SendCommand(CmdDeleteFromAllLinkGroup.SubCommand(int(group)), nil))
}

// ProductData will retrieve the device's product data
func (i1 *I1Device) ProductData() (data *ProductData, err error) {
	recvCh, err := i1.SendCommandAndListen(CmdProductDataReq, nil)
	for resp := range recvCh {
		if resp.Message.Command&0xff00 == CmdProductDataResp&0xff00 {
			data = &ProductData{}
			err = data.UnmarshalBinary(resp.Message.Payload)
			resp.DoneCh <- true
		} else {
			resp.DoneCh <- false
		}
	}
	return data, err
}

// Ping will send a Ping command to the device
func (i1 *I1Device) Ping() (err error) {
	return extractError(i1.SendCommand(CmdPing, nil))
}

// SetAllLinkCommandAlias will set the device's standard command to be used
// when the given alias command is sent
func (i1 *I1Device) SetAllLinkCommandAlias(match, replace *Command) error {
	// TODO implement
	return ErrNotImplemented
}

// SetAllLinkCommandAliasData will set any extended data required by the alias
// command
func (i1 *I1Device) SetAllLinkCommandAliasData(data []byte) error {
	// TODO implement
	return ErrNotImplemented
}

// BlockDataTransfer will retrieve a block of memory from the device
func (i1 *I1Device) BlockDataTransfer(start, end MemAddress, length int) ([]byte, error) {
	// TODO implement
	return nil, ErrNotImplemented
}

// String returns the string "I1 Device (<address>)" where <address> is the destination
// address of the device
func (i1 *I1Device) String() string {
	return sprintf("I1 Device (%s)", i1.Address())
}
