package devices

import "github.com/abates/insteon"

type FilterFunc func(next MessageWriter) MessageWriter

func (ff FilterFunc) Filter(next MessageWriter) MessageWriter {
	return ff(next)
}

type Filter interface {
	Filter(next MessageWriter) MessageWriter
}

type filter struct {
	read  func() (*insteon.Message, error)
	write func(*insteon.Message) (*insteon.Message, error)
}

func (f *filter) Read() (*insteon.Message, error) {
	return f.read()
}

func (f *filter) Write(msg *insteon.Message) (*insteon.Message, error) {
	return f.write(msg)
}

func FilterDuplicates() Filter {
	return FilterFunc(func(mw MessageWriter) MessageWriter {
		cache := NewCache(10)
		mw = cache.Filter(mw)
		read := func() (*insteon.Message, error) {
			msg, err := mw.Read()
		top:
			for ; err == nil; msg, err = mw.Read() {
				if _, found := cache.Lookup(DuplicateMatcher(msg)); found {
					LogDebug.Printf("Dropping duplicate message %v", msg)
					continue top
				}
				break
			}
			return msg, err
		}

		return &filter{
			read:  read,
			write: mw.Write,
		}
	})
}

func TTL(ttl int) Filter {
	return FilterFunc(func(mw MessageWriter) MessageWriter {
		return &filter{
			read: mw.Read,
			write: func(msg *insteon.Message) (*insteon.Message, error) {
				msg.SetMaxTTL(uint8(ttl))
				msg.SetTTL(uint8(ttl))
				return mw.Write(msg)
			},
		}
	})
}

func NewCache(size int, messages ...*insteon.Message) *CacheFilter {
	if size < len(messages) {
		size = len(messages)
	}

	c := &CacheFilter{
		Messages: make([]*insteon.Message, 0, size),
	}

	for i, msg := range messages {
		c.i = i
		c.Messages = append(c.Messages, msg)
	}
	return c
}

func (c *CacheFilter) Filter(next MessageWriter) MessageWriter {
	c.filter.read = func() (*insteon.Message, error) {
		msg, err := next.Read()
		c.push(msg)
		return msg, err
	}

	c.filter.write = func(msg *insteon.Message) (*insteon.Message, error) {
		c.push(msg)
		return next.Write(msg)
	}
	return c
}

type CacheFilter struct {
	filter
	Messages []*insteon.Message
	i        int
}

func (c *CacheFilter) push(msg *insteon.Message) {
	if msg == nil {
		return
	}

	if len(c.Messages) > 0 {
		c.i++
	}

	if len(c.Messages) < cap(c.Messages) {
		c.Messages = append(c.Messages, msg)
	} else {
		if c.i == len(c.Messages) {
			c.i = 0
		}
		c.Messages[c.i] = msg
	}
}

func (c *CacheFilter) Lookup(matcher Matcher) (*insteon.Message, bool) {
	if len(c.Messages) == 0 {
		return nil, false
	}

	j := c.i + 1
	if j == len(c.Messages) {
		j = 0
	}

	for i := c.i; ; i-- {
		if i < 0 {
			i = len(c.Messages) - 1
		}
		if matcher.Matches(c.Messages[i]) {
			return c.Messages[i], true
		}
		if i == j {
			break
		}
	}
	return nil, false
}
