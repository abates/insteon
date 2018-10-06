package insteon

import "time"

// I2CsDevice can communicate with Version 2 (checksum) Insteon Engines
type I2CsDevice struct {
	*I2Device
}

// NewI2CsDevice will initialize a new I2CsDevice object and make
// it ready for use
func NewI2CsDevice(address Address, sendCh chan<- *MessageRequest, recvCh <-chan *Message, timeout time.Duration) Device {
	return &I2CsDevice{NewI2Device(address, sendCh, recvCh, timeout).(*I2Device)}
}

// EnterLinkingMode will put the device into linking mode. This is
// equivalent to holding down the set button until the device
// beeps and the indicator light starts flashing
func (i2cs *I2CsDevice) EnterLinkingMode(group Group) (err error) {
	return extractError(i2cs.SendCommand(CmdEnterLinkingModeExt.SubCommand(int(group)), make([]byte, 14)))
}

// String returns the string "I2CS Device (<address>)" where <address> is the destination
// address of the device
func (i2cs *I2CsDevice) String() string {
	return sprintf("I2CS Device (%s)", i2cs.Address())
}
