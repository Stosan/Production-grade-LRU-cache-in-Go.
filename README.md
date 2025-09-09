# HyperLRU 🚀

A **high-throughput, concurrent-safe LRU (Least Recently Used) cache** for Go — engineered to handle **millions of keys** with minimal lock contention.  

Built for **high-impact systems** where performance, scalability, and reliability are non-negotiable.  

---

## ⚡ Why HyperLRU?
- **Concurrency-first** → optimized for multi-goroutine workloads.  
- **Scales to millions** → configurable max capacity for large datasets.  
- **Minimal lock contention** → efficient synchronization for speed.  
- **True LRU eviction** → least recently used items are automatically removed.  
- **Production-r**

```go
package main

import (
    "fmt"
    "github.com/<your-username>/hyperlru"
)

func main() {
    // Create a new cache with a capacity of 1 million keys
    cache := hyperlru.NewCache(1_000_000)

    // Put values
    cache.Put("foo", "bar")
    cache.Put("number", 42)

    // Get values
    if val, ok := cache.Get("foo"); ok {
        fmt.Println("Found:", val) // Output: Found: bar
    }

    if _, ok := cache.Get("missing"); !ok {
        fmt.Println("Key not found")
    }
}

```


## 🏗️ Use Cases

High-throughput caching in distributed systems

Real-time analytics pipelines

Web services with massive request volume

Systems where predictable eviction (LRU) is critical
