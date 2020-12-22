package tableengine

import (
	"context"
	"sync"

	"github.com/arquivei/goduck"
	"github.com/arquivei/goduck/engine/streamengine"
)

// TableEngine leverages the existing "streamengine" to handle key-value
// messages as a table CDC. This allows the worker to materialize a stream
// topic to an in-memory table for fast access.
// Things to bear in mind when using this engine (specially for kafka):
// - The input stream should never commit the offsets, or else old data may be lost
// - The input stream should never load-balance the data from the topic (aka:
// don't reuse the same groupid), as this would discard part of the data
// - The consumer should expect eventually consistent data
type TableEngine struct {
	engine   *streamengine.StreamEngine
	internal *internalProcessor
}

type internalProcessor struct {
	items map[string][]byte
	lock  *sync.RWMutex
}

func (p *internalProcessor) Process(ctx context.Context, message goduck.Message) error {
	rawKey := message.Key()
	// messages without key should be ignored
	if rawKey == nil {
		return nil
	}

	key := string(rawKey)
	value := message.Value()
	p.lock.Lock()
	defer p.lock.Unlock()
	if value == nil {
		delete(p.items, key)
	} else {
		p.items[key] = value
	}
	return nil
}

// New creates a new StreamEngine
func New(streams []goduck.Stream) *TableEngine {
	processor := &internalProcessor{
		items: make(map[string][]byte),
		lock:  &sync.RWMutex{},
	}
	engine := &TableEngine{
		engine:   streamengine.New(processor, streams),
		internal: processor,
	}

	return engine
}

// Run starts processing the messages, until @ctx is closed
func (e *TableEngine) Run(ctx context.Context) error {
	return e.engine.Run(ctx)
}

// Get returns the value associated with the @key. If @key doesn't exist, nil
// is returned.
// This method is thread safe.
func (e *TableEngine) Get(key string) []byte {
	e.internal.lock.RLock()
	defer e.internal.lock.RUnlock()
	return e.internal.items[key]
}

// Iterate through all the table data.
// This method is thread safe. Note: All writes will be stopped while
// iterating, therefore, you should avoid calling this method, and make
// @fn as fast as possible.
func (e *TableEngine) Iterate(fn func(key string, value []byte)) {
	e.internal.lock.RLock()
	defer e.internal.lock.RUnlock()
	for k, v := range e.internal.items {
		fn(k, v)
	}

}
