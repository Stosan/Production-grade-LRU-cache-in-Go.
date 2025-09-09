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

### Try it out
```go
go run main.go

```


## 🏗️ Use Cases

High-throughput caching in distributed systems

Real-time analytics pipelines

Web services with massive request volume

Systems where predictable eviction (LRU) is critical
