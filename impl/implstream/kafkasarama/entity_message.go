package kafkasarama

type goduckMsg struct {
	key      []byte
	value    []byte
	metadata map[string][]byte
}

func (r goduckMsg) Key() []byte {
	return r.key
}

func (r goduckMsg) Value() []byte {
	return r.value
}

func (r goduckMsg) Metadata() map[string][]byte {
	return r.metadata
}
