package insteon

type FilterFunc func(next MessageWriter) MessageWriter

func (ff FilterFunc) Filter(next MessageWriter) MessageWriter {
	return ff(next)
}

type Filter interface {
	Filter(next MessageWriter) MessageWriter
}

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
	return FilterFunc(func(mw MessageWriter) MessageWriter {
		cache := NewCache(10)
		mw = cache.Filter(mw)
		read := func() (*Message, error) {
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
			write: func(msg *Message) (*Message, error) {
				msg.SetMaxTTL(uint8(ttl))
				msg.SetTTL(uint8(ttl))
				return mw.Write(msg)
			},
		}
	})
}

type CacheFilter interface {
	Filter
	Lookup(Matcher) (match *Message, found bool)
}

func NewCache(size int, messages ...*Message) CacheFilter {
	return newCache(size, messages...)
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

func (c *cache) Filter(next MessageWriter) MessageWriter {
	c.filter.read = func() (*Message, error) {
		msg, err := next.Read()
		if msg != nil {
			c.push(msg)
		}
		return msg, err
	}

	c.filter.write = func(msg *Message) (*Message, error) {
		if msg != nil {
			c.push(msg)
		}
		return next.Write(msg)
	}
	return c
}

type cache struct {
	filter
	messages []*Message
	i        int
	length   int
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

func (c *cache) Lookup(matcher Matcher) (*Message, bool) {
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
