package insteon

import (
	"time"
)

// PacketRequest is used to request that a packetized (marshaled) insteon
// message be sent to the network. Once the upstream device (PLM usually)
// has attempted to send the packet, the Err field will be assigned and
// DoneCh will be written to and closed
type PacketRequest struct {
	Payload []byte
	Err     error
	DoneCh  chan<- *PacketRequest
}

// MessageRequest is used to request a message be sent to a specific device.
// Once the connection has sent the message and either received an ack or
// encountered an error, the Ack and Err fields will be filled and DoneCh
// will be written to and closed
type MessageRequest struct {
	Message *Message
	timeout time.Time
	Ack     *Message
	Err     error
	DoneCh  chan<- *MessageRequest
}

// Network is the main means to communicate with
// devices on the Insteon network
type Network struct {
	timeout     time.Duration
	DB          ProductDatabase
	connections []chan<- *Message

	sendCh       chan<- *PacketRequest
	recvCh       <-chan []byte
	connectCh    chan chan<- *Message
	disconnectCh chan chan<- *Message
	closeCh      chan chan error
}

// New creates a new Insteon network instance for the send and receive channels.  The timeout
// indicates how long the network (and subsuquent devices) should wait when expecting incoming
// messages/responses
func New(sendCh chan<- *PacketRequest, recvCh <-chan []byte, timeout time.Duration) *Network {
	network := &Network{
		timeout: timeout,
		DB:      NewProductDB(),

		sendCh:       sendCh,
		recvCh:       recvCh,
		connectCh:    make(chan chan<- *Message),
		disconnectCh: make(chan chan<- *Message),
		closeCh:      make(chan chan error),
	}

	go network.process()
	return network
}

func (network *Network) process() {
	for {
		select {
		case pkt, open := <-network.recvCh:
			if !open {
				network.close()
				return
			}
			network.receive(pkt)
		case connection := <-network.connectCh:
			network.connections = append(network.connections, connection)
		case connection := <-network.disconnectCh:
			network.disconnect(connection)
		case ch := <-network.closeCh:
			ch <- network.close()
			return
		}
	}
}

func (network *Network) receive(buf []byte) {
	msg := &Message{}
	if err := msg.UnmarshalBinary(buf); err == nil {
		Log.Tracef("Received Insteon Message %v", msg)
		if msg.Broadcast() {
			// Set Button Pressed Controller/Responder
			if msg.Command&0xff00 == 0x0100 || msg.Command&0xff00 == 0x0200 {
				network.DB.UpdateFirmwareVersion(msg.Src, FirmwareVersion(msg.Dst[2]))
				network.DB.UpdateDevCat(msg.Src, DevCat{msg.Dst[0], msg.Dst[1]})
			}
		} else if msg.Ack() && msg.Command&0xff00 == 0x0d00 {
			// Engine Version Request ACK
			network.DB.UpdateEngineVersion(msg.Src, EngineVersion(0x00ff&msg.Command))
		}

		for _, connection := range network.connections {
			connection <- msg
		}
	} else {
		Log.Tracef("Failed unmarshalling message received from network: %v", err)
	}
}

func (network *Network) disconnect(connection chan<- *Message) {
	for i, conn := range network.connections {
		if conn == connection {
			close(conn)
			network.connections = append(network.connections[0:i], network.connections[i+1:]...)
			break
		}
	}
}

func (network *Network) sendMessage(msg *Message) error {
	buf, err := msg.MarshalBinary()

	if err == nil {
		Log.Tracef("Sending %v to network", msg)
		if info, found := network.DB.Find(msg.Dst); found {
			if msg.Flags.Extended() && info.EngineVersion == VerI2Cs {
				buf[len(buf)-1] = checksum(buf[7:22])
			}
		}

		doneCh := make(chan *PacketRequest, 1)
		request := &PacketRequest{buf, nil, doneCh}
		network.sendCh <- request
		<-doneCh
		err = request.Err
	}
	return err
}

// EngineVersion will query the dst device to determine its Insteon engine
// version
func (network *Network) EngineVersion(dst Address) (engineVersion EngineVersion, err error) {
	conn := network.connect(dst, 1, CmdGetEngineVersion)
	defer func() { close(conn.sendCh) }()

	doneCh := make(chan *MessageRequest, 1)
	request := &MessageRequest{Message: &Message{Command: CmdGetEngineVersion, Flags: StandardDirectMessage}, DoneCh: doneCh}
	conn.sendCh <- request
	<-doneCh

	if request.Err == nil {
		engineVersion = EngineVersion(request.Ack.Command & 0x00ff)
	}
	return engineVersion, request.Err
}

// IDRequest will send an ID Request message to the destination device and wait for
// either a "Set-button Pressed Controller" or "Set-button Pressed Responder" broadcast
// message. This message includes the device category and firmaware information which
// is then returned in the DeviceInfo object.  It should be noted that the returned
// DeviceInfo object will not have the engine version field populated as this information
// is not included in the broadcast response.
func (network *Network) IDRequest(dst Address) (info DeviceInfo, err error) {
	info = DeviceInfo{
		Address: dst,
	}
	conn := network.connect(dst, 1, CmdSetButtonPressedResponder, CmdSetButtonPressedController)
	defer func() { close(conn.sendCh) }()
	doneCh := make(chan *MessageRequest, 1)
	request := &MessageRequest{Message: &Message{Command: CmdIDRequest, Flags: StandardDirectMessage}, DoneCh: doneCh}
	conn.sendCh <- request
	<-doneCh
	err = request.Err
	if err == nil {
		for {
			select {
			case msg := <-conn.recvCh:
				if msg.Broadcast() {
					info, _ = network.DB.Find(dst)
					return
				}
			case <-time.After(network.timeout):
				err = ErrReadTimeout
				return
			}
		}
	}
	return
}

func (network *Network) connect(dst Address, version EngineVersion, match ...Command) *connection {
	sendCh := make(chan *MessageRequest, 1)
	recvCh := make(chan *Message, 1)
	go func() {
		for request := range sendCh {
			request.Err = network.sendMessage(request.Message)
			request.DoneCh <- request
		}
		network.disconnectCh <- recvCh
	}()
	connection := newConnection(sendCh, recvCh, dst, version, network.timeout, match...)
	network.connectCh <- recvCh
	return connection
}

// Dial will return a basic device object that can appropriately communicate
// with the physical device out on the insteon network. Dial will determine
// the engine version (1, 2, or 2CS) that the device is running and return
// either an I1Device, I2Device or I2CSDevice. For a fully initialized
// device (dimmer, switch, thermostat, etc) use Connect
func (network *Network) Dial(dst Address) (device Device, err error) {
	var info DeviceInfo
	var found bool
	if info, found = network.DB.Find(dst); !found {
		info.EngineVersion, err = network.EngineVersion(dst)
		// ErrNotLinked here is only returned by i2cs devices
		if err == ErrNotLinked {
			network.DB.UpdateEngineVersion(dst, VerI2Cs)
			info.EngineVersion = VerI2Cs
		}
	}

	if err == nil {
		var constructor DeviceConstructor
		connection := network.connect(dst, info.EngineVersion)
		constructor, err = BaseConstructor(info.EngineVersion)
		if err == nil {
			device, err = constructor(info, dst, connection.sendCh, connection.recvCh, network.timeout)
		}
	}
	return device, err
}

// Connect will Dial the destination device and then determine the device category
// in order to return a category specific device (dimmer, switch, etc). If, for
// some reason, the devcat cannot be determined, then the device returned
// by Dial is returned
func (network *Network) Connect(dst Address) (device Device, err error) {
	var info DeviceInfo
	var found bool
	if info, found = network.DB.Find(dst); !found {
		info.EngineVersion, err = network.EngineVersion(dst)
		if err == nil {
			info, err = network.IDRequest(dst)
		}
	}

	if err == nil {
		var constructor DeviceConstructor
		var found bool
		if constructor, found = Devices.Find(info.DevCat.Category()); !found {
			constructor, err = BaseConstructor(info.EngineVersion)
		}

		if err == nil {
			connection := network.connect(dst, info.EngineVersion)
			device, err = constructor(info, dst, connection.sendCh, connection.recvCh, network.timeout)
		}
	}
	return
}

func (network *Network) close() error {
	for _, connection := range network.connections {
		close(connection)
	}
	return nil
}

// Close will cleanup/close open connections and disconnect gracefully
func (network *Network) Close() error {
	ch := make(chan error)
	network.closeCh <- ch
	close(network.closeCh)
	err := <-ch
	if err == nil {
		close(network.sendCh)
	}
	return err
}
