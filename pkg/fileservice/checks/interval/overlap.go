// Copyright 2023 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package interval

import (
	"fmt"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
)

type OverlapChecker struct {
	tag          string
	keyIntervals map[string]IntervalTree
}

func NewOverlapChecker(tag string) *OverlapChecker {
	return &OverlapChecker{
		tag:          tag,
		keyIntervals: make(map[string]IntervalTree),
	}
}

func (i *OverlapChecker) Insert(key string, low, high int64) error {
	interval := NewInt64Interval(low, high)

	if _, ok := i.keyIntervals[key]; !ok {
		// If key is not present, create a new tree.
		i.keyIntervals[key] = NewIntervalTree()
	} else if i.keyIntervals[key].Intersects(interval) {
		// check if we have an overlap with existing ranges.
		overlaps := i.keyIntervals[key].Stab(interval)
		overlapsMsg := ""
		for _, v := range overlaps {
			overlapsMsg += fmt.Sprintf("[%d %d), ", v.Ivl.Begin, v.Ivl.End)
		}
		return moerr.NewInternalErrorNoCtx("Duplicate key range found in %s when inserting [%d %d). The key %s contains overlapping intervals %s", i.tag, low, high, key, overlapsMsg)
	}

	// true is just a placeholder.
	i.keyIntervals[key].Insert(interval, true)
	return nil
}

func (i *OverlapChecker) Remove(key string, low, high int64) error {
	interval := NewInt64Interval(low, high)
	if _, ok := i.keyIntervals[key]; !ok {
		return moerr.NewInternalErrorNoCtx("Key Range not found for removal in %s", i.tag)
	}

	i.keyIntervals[key].Delete(interval)

	if i.keyIntervals[key].Len() == 0 {
		delete(i.keyIntervals, key)
	}

	return nil
}
