package containers

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {

	opts := withAllocator(Options{})
	vec := NewStlVector[int64](opts)
	now := time.Now()

	for i := 0; i < 500; i++ {
		vec.Append(int64(i))
	}
	t.Log(time.Since(now))
	t.Log(vec.String())

	vec.Close()
}
