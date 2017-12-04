package insteon

type I2Device struct {
	*I1Device
}

func NewI2Device(address Address, bridge Bridge) *I2Device {
	return &I2Device{NewI1Device(address, bridge)}
}
