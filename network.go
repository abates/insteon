package insteon

import (
	"sync"
	"time"
)

var (
	// Timeout is a time.Duration that indicates how long
	// various operations should wait on a device to respond
	// defaults to 5 seconds
	Timeout = 5 * time.Second
)

// DeviceInfo is a record of information about active
// devices on the network
type DeviceInfo struct {
	Address         Address
	DevCat          DevCat
	FirmwareVersion FirmwareVersion
	EngineVersion   EngineVersion
	mutex           sync.Mutex
}

// ProductDatabase is a registry of all the devices on
// the local Insteon network including the device category
// and firmware. ProductDatabase implementations must be
// thread safe as the methods can be called from multiple
// go routines
type ProductDatabase interface {
	UpdateDevCat(address Address, devCat DevCat)
	UpdateEngineVersion(address Address, engineVersion EngineVersion)
	UpdateFirmwareVersion(address Address, firmwareVersion FirmwareVersion)
	Find(address Address) (deviceInfo DeviceInfo, found bool)
}

type productDatabase struct {
	devices map[Address]*DeviceInfo
	mutex   sync.Mutex
}

// NewProductDB will initialize a product database for
// use in the network object
func NewProductDB() ProductDatabase {
	return &productDatabase{
		devices: make(map[Address]*DeviceInfo),
	}
}

func (pdb *productDatabase) Find(address Address) (deviceInfo DeviceInfo, found bool) {
	pdb.mutex.Lock()
	di, found := pdb.devices[address]
	pdb.mutex.Unlock()
	if found {
		deviceInfo = *di
	}
	return deviceInfo, found
}

func (pdb *productDatabase) Update(address Address, callback func(*DeviceInfo)) {
	pdb.mutex.Lock()
	deviceInfo, found := pdb.devices[address]
	if !found {
		deviceInfo = &DeviceInfo{}
		pdb.devices[address] = deviceInfo
	}
	pdb.mutex.Unlock()

	deviceInfo.mutex.Lock()
	callback(deviceInfo)
	deviceInfo.mutex.Unlock()
}

func (pdb *productDatabase) UpdateFirmwareVersion(address Address, firmwareVersion FirmwareVersion) {
	pdb.Update(address, func(deviceInfo *DeviceInfo) { deviceInfo.FirmwareVersion = firmwareVersion })
}

func (pdb *productDatabase) UpdateEngineVersion(address Address, engineVersion EngineVersion) {
	pdb.Update(address, func(deviceInfo *DeviceInfo) { deviceInfo.EngineVersion = engineVersion })
}

func (pdb *productDatabase) UpdateDevCat(address Address, devCat DevCat) {
	pdb.Update(address, func(deviceInfo *DeviceInfo) { deviceInfo.DevCat = devCat })
}

type Bridge interface {
	SendMessage(*Message) error
}

// Network is any implementation that aggregates control of
// all the devices on a physical Insteon network
type Network interface {
	// Notify is called by the upstream component (such as the PLM) to
	// deliver data that has been received from the physical network
	Notify([]byte) error

	// SendMessage will marshall the message and deliver the packet
	// to the upstream component (such as a PLM) for physical delivery
	// to the network
	SendMessage(message *Message) error

	// Dial will return a basic device object that can appropriately communicate
	// with the physical device out on the insteon network. Dial will determine
	// the engine version (1, 2, or 2CS) that the device is running and return
	// either an I1Device, I2Device or I2CSDevice. For a fully initialized
	// device (dimmer, switch, thermostat, etc) use Connect
	Dial(dst Address) (device Device, err error)

	// Connect will Dial the destination device and then determine the device category
	// in order to return a category specific device (dimmer, switch, etc). If, for
	// some reason, the devcat cannot be determined, then the device returned
	// by Dial is returned
	Connect(dst Address) (device Device, err error)
}

// NetworkImpl is the main means to communicate with
// devices on the Insteon network
type NetworkImpl struct {
	ProductDatabase
	mutex      sync.Mutex
	devices    map[Address]Device
	retries    int
	ackTimeout time.Duration

	bridge          Bridge
	engineVersionCh chan *Message
	idRequestCh     chan *Message
}

func New(bridge Bridge) Network {
	network := &NetworkImpl{
		ProductDatabase: NewProductDB(),
		devices:         make(map[Address]Device),
		retries:         3,
		ackTimeout:      time.Second,

		bridge:          bridge,
		engineVersionCh: make(chan *Message),
		idRequestCh:     make(chan *Message),
	}

	return network
}

func (network *NetworkImpl) Notify(pkt []byte) error {
	msg := &Message{}
	err := msg.UnmarshalBinary(pkt)
	if err == nil {
		err = network.process(msg)
	}
	return err
}

func (network *NetworkImpl) process(msg *Message) (err error) {
	if msg.Broadcast() {
		// Set Button Pressed Controller/Responder
		if msg.Command[0] == 0x01 || msg.Command[0] == 0x02 {
			network.UpdateFirmwareVersion(msg.Src, FirmwareVersion(msg.Dst[2]))
			network.UpdateDevCat(msg.Src, DevCat{msg.Dst[0], msg.Dst[1]})

			writeToCh(network.idRequestCh, msg)
		}
	} else if msg.Ack() && msg.Command[0] == 0x0d {
		// try to deliver the version if this was requested locally
		writeToCh(network.engineVersionCh, msg)

		// Engine Version Request ACK
		network.UpdateEngineVersion(msg.Src, EngineVersion(msg.Command[1]))
	}

	network.mutex.Lock()
	if device, found := network.devices[msg.Src]; found {
		err = device.Notify(msg)
	}
	network.mutex.Unlock()

	return err
}

func (network *NetworkImpl) SendMessage(message *Message) error {
	if info, found := network.Find(message.Dst); found {
		message.version = info.EngineVersion
	}
	return network.bridge.SendMessage(message)
}

// EngineVersion will query the dst device to determine its Insteon engine
// version
func (network *NetworkImpl) EngineVersion(dst Address) (engineVersion EngineVersion, err error) {
	err = network.SendMessage(&Message{
		Dst:     dst,
		Flags:   StandardDirectMessage,
		Command: CmdGetEngineVersion,
	})

	if err == nil {
		var msg *Message
		msg, err = readFromCh(network.engineVersionCh)
		if err == nil {
			engineVersion = EngineVersion(msg.Command[1])
		}
	}
	return
}

func (network *NetworkImpl) IDRequest(dst Address) (firmwareVersion FirmwareVersion, devCat DevCat, err error) {
	err = network.SendMessage(&Message{
		Dst:     dst,
		Flags:   StandardDirectMessage,
		Command: CmdIDRequest,
	})

	if err == nil {
		var msg *Message
		msg, err = readFromCh(network.idRequestCh)
		if err == nil {
			firmwareVersion = FirmwareVersion(msg.Dst[2])
			devCat = DevCat{msg.Dst[0], msg.Dst[1]}
			network.UpdateFirmwareVersion(dst, firmwareVersion)
			network.UpdateDevCat(dst, devCat)
		}
	}
	return
}

// Dial will return a basic device object that can appropriately communicate
// with the physical device out on the insteon network. Dial will determine
// the engine version (1, 2, or 2CS) that the device is running and return
// either an I1Device, I2Device or I2CSDevice. For a fully initialized
// device (dimmer, switch, thermostat, etc) use Connect
func (network *NetworkImpl) Dial(dst Address) (device Device, err error) {
	var version EngineVersion
	if info, found := network.Find(dst); found {
		version = info.EngineVersion
	} else {
		version, err = network.EngineVersion(dst)
		// ErrNotLinked here is only returned by i2cs devices
		if err == ErrNotLinked {
			network.UpdateEngineVersion(dst, VerI2Cs)
			Log.Debugf("Got ErrNotLinked, creating I2CS device")
			err = nil
			version = VerI2Cs
		}
	}

	switch version {
	case VerI1:
		Log.Debugf("Version 1 device detected")
		device = NewI1Device(dst, network)
	case VerI2:
		Log.Debugf("Version 2 device detected")
		device = NewI2Device(dst, network)
	case VerI2Cs:
		Log.Debugf("Version 2 CS device detected")
		device = NewI2CsDevice(dst, network)
	}
	network.mutex.Lock()
	network.devices[dst] = device
	network.mutex.Unlock()
	return device, err
}

// Connect will Dial the destination device and then determine the device category
// in order to return a category specific device (dimmer, switch, etc). If, for
// some reason, the devcat cannot be determined, then the device returned
// by Dial is returned
func (network *NetworkImpl) Connect(dst Address) (device Device, err error) {
	var devCat DevCat
	device, err = network.Dial(dst)
	if err == nil {
		if info, found := network.Find(dst); found {
			devCat = info.DevCat
		} else {
			_, devCat, err = network.IDRequest(dst)
		}

		if err == nil {
			if initializer, found := Devices.Find(devCat.Category()); found {
				info, _ := network.Find(dst)
				device = initializer(device, info)
			}
		}
	}
	return
}
