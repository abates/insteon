package insteon

type testConnection int

func (testConnection) Write(*Message) (ack *Message, err error)  { return nil, nil }
func (testConnection) Subscribe(match ...*Command) chan *Message { return nil }
func (testConnection) Unsubscribe(chan *Message)                 {}

/*
func TestCleanup(t *testing.T) {
	var flags RecordControlFlags
	flags.setController()

	link1 := func(available bool) *Link {
		if available {
			flags.setAvailable()
		}
		return &Link{Flags: flags, Group: Group(0x01), Address: Address{0x01, 0x02, 0x03}, Data: [3]byte{0x04, 0x05, 0x06}}
	}

	link2 := func(available bool) *Link {
		if available {
			flags.setAvailable()
		}
		return &Link{Flags: flags, Group: Group(0x01), Address: Address{0x07, 0x08, 0x09}, Data: [3]byte{0x0a, 0x0b, 0x0c}}
	}

	tests := []struct {
		input    []*Link
		expected []*Link
	}{
		{
			input:    []*Link{link1(false), link1(false), link1(false), link2(false), link2(false), link2(false)},
			expected: []*Link{link1(false), link1(true), link1(true), link2(false), link2(true), link2(true)},
		},
	}

	for i, test := range tests {
		linkdb := DeviceLinkDB{
			conn:  testConnection(i),
			links: test.input,
		}
		linkdb.Cleanup()
		if !reflect.DeepEqual(test.expected, linkdb.links) {
			t.Errorf("tests[%d] expected:\n%s\ngot\n%s", i, test.expected, linkdb.links)
		}
	}
}*/
