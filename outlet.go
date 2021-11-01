package insteon

type Outlet struct {
	*Switch
}

func NewOutlet(d *device, info DeviceInfo) *Outlet {
	sd := &Switch{device: d, info: info}

	return &Outlet{sd}
}
