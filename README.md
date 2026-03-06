# activecache

[![CI](https://github.com/wtsi-hgi/activecache/actions/workflows/go-checks.yml/badge.svg)](https://github.com/wtsi-hgi/activecache/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/wtsi-hgi/activecache.svg)](https://pkg.go.dev/github.com/wtsi-hgi/activecache)
[![Go Report Card](https://goreportcard.com/badge/github.com/wtsi-hgi/activecache)](https://goreportcard.com/report/github.com/wtsi-hgi/activecache)

--
    import "github.com/wtsi-hgi/activecache"

Package activecache implements a generic, self-populating cache, that automatically refreshes stored data.

## Highlights

 - Lazy loads data upon request.
 - Actively refreshes stored data using the supplied function in a thread-safe way.
 - Fail-safe background retrieval that does not overwrite successful data retrieval.

## Usage

```go
package main

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/wtsi-hgi/activecache"
)

func main() {
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
```

## Documentation

Full API docs can be found at:

https://pkg.go.dev/github.com/wtsi-hgi/activecache
