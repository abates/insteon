package insteon

type Monitor struct {
	devcats  map[Address]DevCat
	firmware map[Address]FirmwareVersion
	accept   chan *Message
}

func (mon *Monitor) Accept() chan *Message {
	return mon.accept
}

func NewMonitor() *Monitor {
	return &Monitor{
		devcats:  make(map[Address]DevCat),
		firmware: make(map[Address]FirmwareVersion),
		accept:   make(chan *Message),
	}
}

func (mon *Monitor) Update(buf []byte) {
	message := &Message{}
	err := message.UnmarshalBinary(buf)
	if err == nil {
		if message.Broadcast() {
			mon.devcats[message.Src] = DevCat{message.Dst[0], message.Dst[1]}
			mon.firmware[message.Src] = FirmwareVersion(message.Dst[2])
		}

		select {
		case mon.accept <- message:
		default:
		}
	}
}

func (mon *Monitor) DumpMessage(message *Message) (str string) {
	if message.Broadcast() {
		if message.Flags.Type() == MsgTypeAllLinkBroadcast {
			str = sprintf("%s -> ff.ff.ff %v Group(%d)", message.Src, message.Flags, message.Dst[2])
		} else {
			devCat := DevCat{message.Dst[0], message.Dst[1]}
			firmware := FirmwareVersion(message.Dst[2])

			str = sprintf("%s -> ff.ff.ff %v DevCat %v Firmware %v", message.Src, message.Flags, devCat, firmware)
		}
	} else {
		str = sprintf("%s -> %s %v", message.Src, message.Dst, message.Flags)
		if message.Flags.Extended() {
			str = sprintf("%s %v", str, message.Payload)
		}
	}

	var command CommandBytes
	if message.Flags.Extended() {
		command = Commands.FindExt(mon.devcats[message.Src], message.Command)
	} else {
		command = Commands.FindStd(mon.devcats[message.Src], message.Command)
	}

	str = sprintf("%s %v", str, command)
	return str
}
