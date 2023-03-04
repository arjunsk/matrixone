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

type Metrics struct {
	sync.Mutex
	metrics  map[string]Counter
	duration time.Duration
}

var (
	metrics   *Metrics
	setupOnce sync.Once
)

func Get() *Metrics {
	setupOnce.Do(func() {
		batchDuration := time.Second
		metrics = createBatch(batchDuration)
	})

	return metrics
}

func createBatch(batchDuration time.Duration) *Metrics {
	o := &Metrics{
		metrics:  make(map[string]Counter),
		duration: batchDuration,
	}
	registerMetrics(o)
	go o.startAsyncPublisher()
	return o
}

func registerMetrics(o *Metrics) {
	o.metrics["MemCacheRead"] = Counter{localCounter: atomic.NewFloat64(0), globalCounter: metric.MemCacheReadCounter()}
	o.metrics["S3FsRead"] = Counter{localCounter: atomic.NewFloat64(0), globalCounter: metric.S3ReadCounter()}
}

func (b *Metrics) Incr(metricName string) {
	b.metrics[metricName].localCounter.Add(1)
}

func (b *Metrics) ResetLocalCounters() {
	b.Lock()
	defer b.Unlock()
	for _, v := range b.metrics {
		v.localCounter.Store(0)
	}
}

func (b *Metrics) MergeCounters() {
	for _, v := range b.metrics {
		v.globalCounter.Add(v.localCounter.Load())
	}
}

func (b *Metrics) startAsyncPublisher() {
	ticker := time.NewTicker(b.duration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			{
				b.MergeCounters()
				b.ResetLocalCounters()

				ticker.Reset(b.duration)
				continue
			}
		}
	}
}
