package insteon

type EngineVersion int

const (
	VerI1 EngineVersion = iota
	VerI2
	VerI2Cs
)

func (ver EngineVersion) String() string {
	switch ver {
	case VerI1:
		return "I1"
	case VerI2:
		return "I2"
	case VerI2Cs:
		return "I2CS"
	}
	return "Unknown"
}

type DeviceInitializer func(Device) Device

var Devices DeviceRegistry

type DeviceRegistry struct {
	// devices key is the first byte of the
	// Category.  Documentation simply calls this
	// the category and the second byte the sub
	// category, but we've combined both bytes
	// into a single type
	devices map[byte]DeviceInitializer
}

func (dr *DeviceRegistry) Register(category byte, initializer DeviceInitializer) {
	if dr.devices == nil {
		dr.devices = make(map[byte]DeviceInitializer)
	}
	dr.devices[category] = initializer
}

func (dr *DeviceRegistry) Find(category Category) DeviceInitializer {
	return dr.devices[category[0]]
}

func (dr *DeviceRegistry) Initialize(device Device) (Device, error) {
	// query the device
	pd, err := device.ProductData()

	// construct device for device type
	if err == nil {
		initializer := dr.Find(pd.Category)
		if initializer != nil {
			device = initializer(device)
		}
	} else if err == ErrReadTimeout {
		Log.Infof("Timed out waiting for product data response. Returning standard device")
		err = nil
	}

	return device, err
}

type Device interface {
	Linkable
	ProductData() (*ProductData, error)
	FXUsername() (string, error)
	TextString() (string, error)
	EngineVersion() (EngineVersion, error)
	Ping() error
	Close() error
}
