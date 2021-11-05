package insteon

type Outlet struct {
	*Switch
}

func NewOutlet(d *BasicDevice) *Outlet {
	sd := &Switch{BasicDevice: d}

	return &Outlet{sd}
}
