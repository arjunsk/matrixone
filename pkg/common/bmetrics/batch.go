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
	metricCounters map[string]Counter
	batchInterval  time.Duration
}

var (
	metrics   *BatchedMetrics
	setupOnce sync.Once
)

func Get() *BatchedMetrics {
	setupOnce.Do(func() {
		batchDuration := time.Second
		metrics = createBatchedMetrics(batchDuration)
	})

	return metrics
}

func createBatchedMetrics(batchDuration time.Duration) *BatchedMetrics {
	o := &BatchedMetrics{
		metricCounters: make(map[string]Counter),
		batchInterval:  batchDuration,
	}
	registerMetrics(o)
	go o.startBackgroundPublisher()
	return o
}

func registerMetrics(o *BatchedMetrics) {
	o.metricCounters["MemCacheRead"] = Counter{localCounter: atomic.NewFloat64(0), globalCounter: metric.MemCacheReadCounter()}
	o.metricCounters["S3FsRead"] = Counter{localCounter: atomic.NewFloat64(0), globalCounter: metric.S3ReadCounter()}
}

func (b *BatchedMetrics) Incr(metricName string) {
	b.metricCounters[metricName].localCounter.Add(1)
}

func (b *BatchedMetrics) MergeAndReset() {
	b.Lock()
	defer b.Unlock()
	for _, v := range b.metricCounters {
		v.globalCounter.Add(v.localCounter.Load())
		v.localCounter.Store(0)
	}
}

func (b *BatchedMetrics) startBackgroundPublisher() {
	ticker := time.NewTicker(b.batchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.MergeAndReset()
		}
	}
}
