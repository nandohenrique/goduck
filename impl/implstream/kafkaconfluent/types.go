package kafkaconfluent

type goduckMsg struct {
	key      []byte
	value    []byte
	metadata map[string][]byte
}

func (msg goduckMsg) Key() []byte {
	return msg.key
}

func (msg goduckMsg) Value() []byte {
	return msg.value
}

func (msg goduckMsg) Metadata() map[string][]byte {
	return msg.metadata
}

type topicPartition struct {
	topic     *string
	partition int32
}
