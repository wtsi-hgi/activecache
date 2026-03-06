package activecache_test

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/wtsi-hgi/activecache"
)

func Example() {
	var counter atomic.Uint64

	cache := activecache.New(10*time.Millisecond, func(k string) (string, error) {
		return k + "_" + strconv.FormatUint(counter.Add(1), 10), nil
	})

	valA, _ := cache.Get("keyA")
	valB, _ := cache.Get("keyB")

	fmt.Printf("valA = %s\nvalB = %s\n", valA, valB)

	valA, _ = cache.Get("keyA")
	valB, _ = cache.Get("keyB")

	fmt.Printf("valA = %s\nvalB = %s\n", valA, valB)

	time.Sleep(15 * time.Millisecond)

	newValA, _ := cache.Get("keyA")
	newValB, _ := cache.Get("keyB")

	newAPrefix := newValA[:5]
	newBPrefix := newValB[:5]

	fmt.Printf(
		"newAPrefix = %s\nnewValA == valA = %v\nnewBPrefix = %s\nnewValB == valB = %v",
		newAPrefix, valA == newAPrefix,
		newBPrefix, valB == newBPrefix,
	)

	// Output:
	// valA = keyA_1
	// valB = keyB_2
	// valA = keyA_1
	// valB = keyB_2
	// newAPrefix = keyA_
	// newValA == valA = false
	// newBPrefix = keyB_
	// newValB == valB = false
}
