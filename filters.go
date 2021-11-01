package insteon

type ring struct {
	messages []*Message
	i        int
	j        int
}

func (r *ring) push(msg *Message) {
	if r.j > 0 {
		r.i++
		if r.i == len(r.messages) {
			r.i = 0
		}
	}
	if r.j < len(r.messages) {
		r.j++
	}

	r.messages[r.i] = msg
}

func (r *ring) matches(matcher Matcher) (*Message, bool) {
	if r.j == 0 {
		return nil, false
	}

	for _, msg := range r.messages[0:r.j] {
		if matcher.Matches(msg) {
			return msg, true
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
		msgs := &ring{messages: make([]*Message, 10)}
		read := func() (*Message, error) {
			msg, err := mw.Read()
		top:
			for ; err == nil; msg, err = mw.Read() {
				if _, found := msgs.matches(DuplicateMatcher(msg)); found {
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
