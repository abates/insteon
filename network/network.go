package network

import "github.com/abates/insteon"

type Handler interface {
	Handle(msg *insteon.Message)
}

type HandlerFunc func(msg *insteon.Message)

func (hf HandlerFunc) Handle(msg *insteon.Message) {
	hf(msg)
}

type Middleware func(next Handler) Handler

func ChainMiddleware(middleware ...Middleware) Handler {
	handler := Handler(nil)

	for i := len(middleware) - 1; i >= 0; i-- {
		m := middleware[i]
		handler = m(handler)
	}
	return handler
}

func DatabaseMiddleware(db Database) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(msg *insteon.Message) {
			info, _ := db.Get(msg.Src)

			if !msg.Flags.Ack() && (msg.Command.Matches(insteon.CmdSetButtonPressedController) || msg.Command.Matches(insteon.CmdSetButtonPressedResponder)) {
				info.Address = msg.Src
				info.DevCat = insteon.DevCat{msg.Dst[0], msg.Dst[1]}
				info.FirmwareVersion = insteon.FirmwareVersion(msg.Dst[2])
				db.Put(info)
			} else if msg.Flags.Ack() && msg.Command.Matches(insteon.CmdGetEngineVersion) {
				info.Address = msg.Src
				info.EngineVersion = insteon.EngineVersion(msg.Command.Command2())
				db.Put(info)
			}
			next.Handle(msg)
		})
	}
}
