package insteon

import (
	"encoding"
	"fmt"
	"io"
	"strings"
)

type cache struct {
	messages []*Message
	i        int
	length   int
}

func newCache(size int, messages ...*Message) *cache {
	c := &cache{
		messages: make([]*Message, size),
		length:   0,
	}

	for i, msg := range messages {
		c.i = i
		c.length++
		c.messages[i] = msg
	}
	return c
}

func (c *cache) push(msg *Message) {
	if c.length > 0 {
		c.i++
		if c.i == len(c.messages) {
			c.i = 0
		}
	}

	if c.length < len(c.messages) {
		c.length++
	}

	c.messages[c.i] = msg
}

func (c *cache) lookup(matcher Matcher) (*Message, bool) {
	if c.length == 0 {
		return nil, false
	}

	j := c.i + 1
	if j == c.length {
		j = 0
	}

	for i := c.i; ; i-- {
		if i < 0 {
			i = c.length - 1
		}
		if matcher.Matches(c.messages[i]) {
			return c.messages[i], true
		}
		if i == j {
			break
		}
	}
	return nil, false
}

type Filter func(next MessageWriter) MessageWriter

type filter struct {
	read  func() (*Message, error)
	write func(*Message) (*Message, error)
}

func (f *filter) Read() (*Message, error) {
	return f.read()
}

func (f *filter) Write(msg *Message) (*Message, error) {
	return f.write(msg)
}

func FilterDuplicates() Filter {
	return func(mw MessageWriter) MessageWriter {
		msgs := newCache(10)
		read := func() (*Message, error) {
			msg, err := mw.Read()
		top:
			for ; err == nil; msg, err = mw.Read() {
				if _, found := msgs.lookup(DuplicateMatcher(msg)); found {
					LogDebug.Printf("Dropping duplicate message %v", msg)
					continue top
				}
				msgs.push(msg)
				break
			}
			return msg, err
		}

		return &filter{
			read:  read,
			write: mw.Write,
		}
	}
}

func TTL(ttl int) Filter {
	return func(mw MessageWriter) MessageWriter {
		return &filter{
			read: mw.Read,
			write: func(msg *Message) (*Message, error) {
				msg.SetMaxTTL(uint8(ttl))
				msg.SetTTL(uint8(ttl))
				return mw.Write(msg)
			},
		}
	}
}

type snoop struct {
	cache *cache
	db    Database
	mw    MessageWriter
	out   io.Writer
}

func (s *snoop) Read() (*Message, error) {
	msg, err := s.mw.Read()
	if msg != nil {
		s.print(msg)
	}
	return msg, err
}

func (s *snoop) Write(msg *Message) (*Message, error) {
	s.print(msg)
	msg, err := s.mw.Write(msg)
	return msg, err
}

func (s *snoop) print(msg *Message) {
	if msg.Type() == MsgTypeAllLinkBroadcast {
		fmt.Fprintf(s.out, "All-Link Broadcast %s -> ff.ff.ff", msg.Src)
	} else if msg.Type() == MsgTypeBroadcast {
		devCat := DevCat{msg.Dst[0], msg.Dst[1]}
		firmware := FirmwareVersion(msg.Dst[2])
		fmt.Fprintf(s.out, "         Broadcast %s -> ff.ff.ff DevCat %v Firmware %s", msg.Src, devCat, firmware)
	} else if msg.Type() == MsgTypeAllLinkCleanup {
		fmt.Fprintf(s.out, "  All-Link Cleanup %s -> %s", msg.Src, msg.Dst)
	} else {
		fmt.Fprintf(s.out, "            Direct %s -> %s", msg.Src, msg.Dst)
	}
	fmt.Fprintf(s.out, " %d:%d", msg.MaxTTL(), msg.TTL())

	if msg.Ack() {
		prev, found := s.cache.lookup(MatchAck(msg))
		if found {
			fmt.Fprintf(s.out, " %v ACK", prev.Command)
		} else {
			fmt.Fprintf(s.out, " %d.%d (unknown ACK)", msg.Command.Command1(), msg.Command.Command2())
		}
	} else if msg.Type() == MsgTypeAllLinkBroadcast {
		if CmdMatcher(CmdAllLinkSuccessReport).Matches(msg) {
			fmt.Fprintf(s.out, " %v: %v Group %d (cleanup %d, failed %d)", msg.Command&0xffff00, Command(0x0c0000)|Command(msg.Dst[0])<<8, msg.Dst[2], msg.Dst[1], msg.Command.Command2())
		} else {
			fmt.Fprintf(s.out, " %v Group %d", msg.Command&0xffff00, msg.Dst[2])
		}
	} else if msg.Type() == MsgTypeAllLinkCleanup {
		fmt.Fprintf(s.out, " %v Group %d", msg.Command&0xffff00, msg.Command.Command2())
	} else {
		fmt.Fprintf(s.out, " %v", msg.Command)
	}

	if msg.Extended() {
		var data encoding.BinaryUnmarshaler
		switch {
		case msg.Command.Matches(CmdProductDataResp):
			data = &ProductData{}
		case msg.Command.Matches(CmdReadWriteALDB):
			data = &linkRequest{}
		case msg.Command.Matches(CmdExtendedGetSet):
			if s.db != nil {
				if info, found := s.db.Get(msg.Src); found {
					switch info.DevCat.Domain() {
					case DimmerDomain:
						data = &DimmerConfig{}
					case SwitchDomain:
						data = &SwitchConfig{}
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

func Snoop(out io.Writer, db Database) Filter {
	return func(mw MessageWriter) MessageWriter {
		return &snoop{
			mw:    mw,
			db:    db,
			cache: newCache(10),
			out:   out,
		}
	}
}
