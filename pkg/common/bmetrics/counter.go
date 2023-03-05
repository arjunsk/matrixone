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

import "go.uber.org/atomic"

type MetricCounter struct {
	counter *atomic.Float64
}

func NewCounter() *MetricCounter {
	return &MetricCounter{
		counter: atomic.NewFloat64(0),
	}
}

func (l *MetricCounter) Incr() {
	l.counter.Add(1)
}

func (l *MetricCounter) Load() float64 {
	return l.counter.Load()
}

func (l *MetricCounter) Reset() {
	l.counter.Store(0)
}
