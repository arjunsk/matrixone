package trace

import (
	"context"
	"github.com/matrixorigin/matrixone/pkg/util/export/observability/trace"
	"sync"
	"time"
)

type AsyncTrace struct {
	sync.Mutex
	// TODO: make this map concurrent?
	// TODO: replace int64 with the spanMetrics object.
	spans    map[string]int64
	duration time.Duration
}

func NewAsyncTrace(duration time.Duration) *AsyncTrace {
	o := &AsyncTrace{
		spans:    make(map[string]int64),
		duration: duration,
	}
	//TODO: see how to use taskservice here.
	go o.startAsyncPublisher()
	return o
}

// Start Accumulates to the batched span information.
func (t *AsyncTrace) Start(ctx context.Context, spanName string, opts ...SpanOption) {
	//TODO: need to do the accumulation part here.
	//TODO: should I create an interface for sync-trace and async-trace? Or this approach is fine.
	t.spans[spanName]++
}

func (t *AsyncTrace) End(spanName string) {
}

func (t *AsyncTrace) ResetMetrics() {
	// TODO: should I lock this before reset. Need to check the best practise.
	for k, _ := range t.spans {
		t.spans[k] = 0
	}
}

func (t *AsyncTrace) startAsyncPublisher() {
	//TODO: should async publisher push the data before closing?

	ticker := time.NewTicker(t.duration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			{
				for k, v := range t.spans {
					// TODO: Trace.Add() is not available. Need to figure out this part
					trace.Add(ctx, k, v)
				}
				t.ResetMetrics()
				ticker.Reset(t.duration)
				continue
			}
		}
	}
}
