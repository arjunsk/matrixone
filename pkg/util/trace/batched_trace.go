package trace

import (
	"context"
	"github.com/matrixorigin/matrixone/pkg/util/export/observability/trace"
	"golang.org/x/exp/maps"
	"sync"
	"time"
)

type BatchedTrace struct {
	sync.Mutex
	// TODO: make this map concurrent?
	// TODO: replace int64 with the accumulated Span object. Please ignore int64 for now.
	spans         map[string]int64
	batchDuration time.Duration
}

func NewBatchedTrace(batchDuration time.Duration) *BatchedTrace {
	o := &BatchedTrace{
		spans:         make(map[string]int64),
		batchDuration: batchDuration,
	}
	//TODO: see how to use taskservice here.
	go o.startAsyncPublisher()
	return o
}

// Start Accumulates to the batched span information.
func (t *BatchedTrace) Start(ctx context.Context, spanName string, opts ...SpanOption) {
	//TODO: need to do the accumulation part here. May be read: https://docs.newrelic.com/docs/more-integrations/open-source-telemetry-integrations/opentelemetry/best-practices/opentelemetry-best-practices-batching/
	//TODO: should I create an interface for batched-trace and regular-trace? Or this approach is fine.
	//TODO: not sure how batching spans will impact nested traces. Do we have any corner case?
	t.spans[spanName]++
}

func (t *BatchedTrace) End(spanName string) {
}

func (t *BatchedTrace) ResetMetrics() {
	// TODO: should I do coarse locking before reset. Need to check the best practise.
	maps.Clear(t.spans)
}

func (t *BatchedTrace) startAsyncPublisher() {
	//TODO: should async publisher push the data before closing? Need to check about graceful shutdown practises.

	ticker := time.NewTicker(t.batchDuration)
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
				ticker.Reset(t.batchDuration)
				continue
			}
		}
	}
}
