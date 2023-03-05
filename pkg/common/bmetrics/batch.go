// Copyright 2023 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bmetrics

import (
	"github.com/matrixorigin/matrixone/pkg/util/metric"
	"sync"
	"time"
)

type Counter struct {
	localCounter  *MetricCounter
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

func GetInstance() *BatchedMetrics {
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
	o.metricCounters["MemCacheRead"] = Counter{localCounter: NewCounter(), globalCounter: metric.MemCacheReadCounter()}
	o.metricCounters["S3FsRead"] = Counter{localCounter: NewCounter(), globalCounter: metric.S3ReadCounter()}
}

func (b *BatchedMetrics) GetCounter(metricName string) *MetricCounter {
	return b.metricCounters[metricName].localCounter
}

func (b *BatchedMetrics) MergeAndReset() {
	for _, v := range b.metricCounters {
		b.Lock()
		v.globalCounter.Add(v.localCounter.Load())
		v.localCounter.Reset()
		b.Unlock()
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
