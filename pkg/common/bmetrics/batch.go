package bmetrics

import (
	"github.com/matrixorigin/matrixone/pkg/util/metric"
	"go.uber.org/atomic"
	"sync"
	"time"
)

type Counter struct {
	localCounter  *atomic.Float64
	globalCounter metric.Counter
}

type BatchedMetrics struct {
	sync.Mutex
	metrics       map[string]Counter
	batchDuration time.Duration
}

var (
	metrics   *BatchedMetrics
	setupOnce sync.Once
)

func Get() *BatchedMetrics {
	setupOnce.Do(func() {
		batchDuration := time.Second
		metrics = createBatch(batchDuration)
	})

	return metrics
}

func createBatch(batchDuration time.Duration) *BatchedMetrics {
	o := &BatchedMetrics{
		metrics:       make(map[string]Counter),
		batchDuration: batchDuration,
	}
	registerMetrics(o)
	go o.startAsyncPublisher()
	return o
}

func registerMetrics(o *BatchedMetrics) {
	o.metrics["MemCacheRead"] = Counter{localCounter: atomic.NewFloat64(0), globalCounter: metric.MemCacheReadCounter()}
	o.metrics["S3FsRead"] = Counter{localCounter: atomic.NewFloat64(0), globalCounter: metric.S3ReadCounter()}
}

func (b *BatchedMetrics) Incr(metricName string) {
	b.metrics[metricName].localCounter.Add(1)
}

func (b *BatchedMetrics) ResetLocalCounters() {
	b.Lock()
	defer b.Unlock()
	for _, v := range b.metrics {
		v.localCounter.Store(0)
	}
}

func (b *BatchedMetrics) MergeCounters() {
	for _, v := range b.metrics {
		v.globalCounter.Add(v.localCounter.Load())
	}
}

func (b *BatchedMetrics) startAsyncPublisher() {
	ticker := time.NewTicker(b.batchDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			{
				b.MergeCounters()
				b.ResetLocalCounters()

				ticker.Reset(b.batchDuration)
				continue
			}
		}
	}
}
