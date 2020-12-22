package pubsubqueue

import "cloud.google.com/go/pubsub"

type goduckMessage struct {
	msg      *pubsub.Message
	metadata map[string][]byte
}

func (r goduckMessage) Key() []byte {
	return []byte(r.msg.OrderingKey)
}

func (r goduckMessage) Value() []byte {
	return r.msg.Data
}

func (r goduckMessage) Metadata() map[string][]byte {
	return r.metadata
}
