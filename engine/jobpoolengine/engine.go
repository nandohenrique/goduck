package jobpoolengine

import (
	"context"
	"io"

	"github.com/arquivei/goduck"
	"github.com/arquivei/goduck/gokithelper"
	"github.com/go-kit/kit/endpoint"

	"github.com/arquivei/foundationkit/errors"
)

// JobPoolEngine processes the messages from a JobPool in parallel, without any
// ordering guarantees
type JobPoolEngine struct {
	queue       goduck.MessagePool
	nextMessage chan goduck.RawMessage
	nWorkers    int
	processor   goduck.Processor

	cancelFn       func()
	processorError error
}

// NewFromEndpoint creates a JobPoolEngine from a go-kit endpoint
func NewFromEndpoint(
	e endpoint.Endpoint,
	decoder goduck.EndpointDecoder,
	queue goduck.MessagePool,
	nWorkers int,
) *JobPoolEngine {
	return New(
		queue,
		gokithelper.MustNewEndpointProcessor(e, decoder),
		nWorkers,
	)
}

// New creates a new JobPoolEngine
func New(queue goduck.MessagePool, processor goduck.Processor, nWorkers int) *JobPoolEngine {
	engine := &JobPoolEngine{
		queue:          queue,
		nextMessage:    make(chan goduck.RawMessage),
		nWorkers:       nWorkers,
		processor:      processor,
		cancelFn:       nil,
		processorError: nil,
	}
	return engine
}

// Run starts processing the messages, until @ctx is closed
func (e *JobPoolEngine) Run(ctx context.Context) error {
	ctx, e.cancelFn = context.WithCancel(ctx)
	for i := 0; i < e.nWorkers; i++ {
		go e.handleMessages(context.Background())
	}
	e.pollMessages(ctx)
	return e.processorError
}

func (e *JobPoolEngine) pollMessages(ctx context.Context) {
	defer close(e.nextMessage)
	for {
		msg, err := e.queue.Next(ctx)
		if err == io.EOF {
			return
		}
		if err != nil {
			continue
		}
		select {
		case e.nextMessage <- msg:
			continue
		case <-ctx.Done():
			return
		}
	}

}

func (e *JobPoolEngine) handleMessages(ctx context.Context) {
	for {
		msg, ok := <-e.nextMessage
		if !ok {
			return
		}
		e.handleMessage(ctx, msg)
	}
}

func (e *JobPoolEngine) handleMessage(ctx context.Context, msg goduck.RawMessage) {
	err := e.processor.Process(ctx, msg.Bytes())
	if err == nil {
		e.queue.Done(ctx, msg)
	} else {
		e.queue.Failed(ctx, msg)
		if errors.GetSeverity(err) == errors.SeverityFatal {
			e.selfClose(err)
		}
	}
	// Ack/Nack errors are ignored
}

func (e *JobPoolEngine) selfClose(err error) {
	e.cancelFn()
	e.processorError = err
}
