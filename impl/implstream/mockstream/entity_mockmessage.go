package mockstream

type goduckMessage struct {
	data []byte
	idx  int
}

func (m goduckMessage) Key() []byte {
	return nil
}

func (m goduckMessage) Value() []byte {
	return m.data
}

func (m goduckMessage) Metadata() map[string][]byte {
	return nil
}
