package insteon

type Outlet struct {
	*Switch
}

func NewOutlet(device Device, bus Bus, info DeviceInfo) *Outlet {
	//outlet := &Outlet{
	sd := &Switch{Device: device, bus: bus, info: info}

	sd.On(And(AllLinkMatcher(), CmdMatcher(CmdLightOn)), sd.onTurnOn)
	sd.On(And(AllLinkMatcher(), CmdMatcher(CmdLightOff)), sd.onTurnOff)
	return &Outlet{sd}
}


