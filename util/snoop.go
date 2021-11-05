package util

import (
	"encoding"
	"fmt"
	"io"
	"strings"

	"github.com/abates/insteon"
)

type snoop struct {
	cache insteon.CacheFilter
	db    Database
	mw    insteon.MessageWriter
	out   io.Writer
}

func (s *snoop) Filter(next insteon.MessageWriter) insteon.MessageWriter {
	s.mw = s.cache.Filter(next)
	return s
}

func (s *snoop) Read() (*insteon.Message, error) {
	msg, err := s.mw.Read()
	if msg != nil {
		s.print(msg)
	}
	return msg, err
}

func (s *snoop) Write(msg *insteon.Message) (*insteon.Message, error) {
	s.print(msg)
	msg, err := s.mw.Write(msg)
	// print the ACK too
	if msg != nil {
		s.print(msg)
	}
	return msg, err
}

func (s *snoop) print(msg *insteon.Message) {
	if msg.Type() == insteon.MsgTypeAllLinkBroadcast {
		fmt.Fprintf(s.out, "All-Link Broadcast %s -> ff.ff.ff", msg.Src)
	} else if msg.Type() == insteon.MsgTypeBroadcast {
		devCat := insteon.DevCat{msg.Dst[0], msg.Dst[1]}
		firmware := insteon.FirmwareVersion(msg.Dst[2])
		fmt.Fprintf(s.out, "         Broadcast %s -> ff.ff.ff DevCat %v Firmware %s", msg.Src, devCat, firmware)
	} else if msg.Type() == insteon.MsgTypeAllLinkCleanup {
		fmt.Fprintf(s.out, "  All-Link Cleanup %s -> %s", msg.Src, msg.Dst)
	} else {
		fmt.Fprintf(s.out, "            Direct %s -> %s", msg.Src, msg.Dst)
	}
	fmt.Fprintf(s.out, " %d:%d", msg.MaxTTL(), msg.TTL())

	if msg.Ack() {
		prev, found := s.cache.Lookup(insteon.MatchAck(msg))
		if found {
			fmt.Fprintf(s.out, " %v ACK", prev.Command)
		} else {
			fmt.Fprintf(s.out, " %d.%d (unknown ACK)", msg.Command.Command1(), msg.Command.Command2())
		}
	} else if msg.Type() == insteon.MsgTypeAllLinkBroadcast {
		if insteon.CmdMatcher(insteon.CmdAllLinkSuccessReport).Matches(msg) {
			// this is ugly
			fmt.Fprintf(s.out, " %v: %v Group %d (cleanup %d, failed %d)", msg.Command&0xffff00, insteon.Command(0x0c0000)|insteon.Command(msg.Dst[0])<<8, msg.Dst[2], msg.Dst[1], msg.Command.Command2())
		} else {
			fmt.Fprintf(s.out, " %v Group %d", msg.Command&0xffff00, msg.Dst[2])
		}
	} else if msg.Type() == insteon.MsgTypeAllLinkCleanup {
		fmt.Fprintf(s.out, " %v Group %d", msg.Command&0xffff00, msg.Command.Command2())
	} else {
		fmt.Fprintf(s.out, " %v", msg.Command)
	}

	if msg.Extended() {
		var data encoding.BinaryUnmarshaler
		switch {
		case msg.Command.Matches(insteon.CmdProductDataResp):
			data = &insteon.ProductData{}
		case msg.Command.Matches(insteon.CmdReadWriteALDB):
			data = &insteon.LinkRequest{}
		case msg.Command.Matches(insteon.CmdExtendedGetSet):
			if s.db != nil {
				if info, found := s.db.Get(msg.Src); found {
					switch info.DevCat.Domain() {
					case insteon.DimmerDomain:
						data = &insteon.DimmerConfig{}
					case insteon.SwitchDomain:
						data = &insteon.SwitchConfig{}
					}
				}
			}
		}
		payload := ""
		if data != nil {
			err := data.UnmarshalBinary(msg.Payload)
			if err == nil {
				payload = fmt.Sprintf("%v", data)
			} else {
				payload = fmt.Sprintf("payload error [%v] %v", s.payloadStr(msg.Payload), err)
			}
		} else {
			payload = fmt.Sprintf("unknown payload [%v]", s.payloadStr(msg.Payload))
		}
		fmt.Fprint(s.out, payload)
	}
	fmt.Fprintln(s.out, "")
}

func (s *snoop) payloadStr(payload []byte) string {
	builder := &strings.Builder{}
	for i, value := range payload {
		fmt.Fprintf(builder, "%02x", value)
		if i < len(payload)-1 {
			fmt.Fprintf(builder, ", ")
		}
	}
	return builder.String()
}

func Snoop(out io.Writer, db Database) insteon.Filter {
	return &snoop{
		db:    db,
		cache: insteon.NewCache(10),
		out:   out,
	}
}