package main

import (
	"fmt"
	"hash/fnv"
	"runtime"
	"sync"
	"time"
)

type Node struct {
	key   string
	value interface{}
	prev  *Node
	next  *Node
}

type Shard struct {
	mu       sync.RWMutex
	capacity int
	size     int
	cache    map[string]*Node
	head     *Node
	tail     *Node
}

type LRUCache struct {
	shards    []*Shard
	shardMask uint32
}

func nextPowerOf2(n uint32) uint32 {
	if n == 0 {
		return 1
	}
	return n
}

func newShard(capacity int) *Shard {
	shard := &Shard{
		capacity: capacity,
		size:     0,
		cache:    make(map[string]*Node),
		head:     &Node{},
		tail:     &Node{},
	}
		// Initialize doubly-linked list with sentinel nodes
	shard.head.next = shard.tail
	shard.tail.prev = shard.head

	return shard
}

func (c LRUCache) hash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

// getShard returns the appropriate shard for a key
func (c LRUCache) getShard(key string) *Shard {
	return c.shards[c.hash(key)&c.shardMask]
}

func NewLRUCache(capacity int) *LRUCache {
	numShards := nextPowerOf2(uint32(runtime.NumCPU() * 4))
	if numShards < 16 {
		numShards = 16
	}
	if numShards < 1024 {
		numShards = 1024
	}
	shardCapacity := capacity / int(numShards)
	if shardCapacity < 1 {
		shardCapacity = 1
	}

	cache := &LRUCache{
		shards:    make([]*Shard, numShards),
		shardMask: numShards - 1,
	}
	for i := uint32(0); i < numShards; i++ {
		cache.shards[i] = newShard(shardCapacity)
	}
	return cache
}

func (s *Shard) moveToFront(node *Node) {
	s.removeNode(node)
	s.addToFront(node)
}
func (s *Shard) addToFront(node *Node) {
	node.prev = s.head
	node.next = s.head.next
	s.head.next.prev = node
	s.head.next = node
}
func (s *Shard) removeNode(node *Node) {
	node.prev.next = node.next
	node.next.prev = node.prev
}
func (c *LRUCache) Get(key string) (value interface{}, ok bool) {
	shard := c.getShard(key)
	shard.mu.Lock() // Minimize lock contention
	defer shard.mu.Unlock()

	if node, exists := shard.cache[key]; exists {
		shard.moveToFront(node)
		return node.value, true
	}
	return nil, false
}
func (c *LRUCache) Put(key string, value interface{}) {
	shard := c.getShard(key)
	shard.mu.Lock() // Minimize lock contention
	defer shard.mu.Unlock()

	if node, exists := shard.cache[key]; exists {
		node.value = value
		shard.moveToFront(node)
		return
	}
	newNode := &Node{
		key:   key,
		value: value,
	}
	shard.cache[key] = newNode
	shard.addToFront(newNode)
	shard.size++

	if shard.size > shard.capacity {
		return // need a better logic
	}
}

// Clear removes all items from the cache
func (c *LRUCache) Clear() {
	for _, shard := range c.shards {
		shard.mu.Lock()
		shard.cache = make(map[string]*Node)
		shard.head.next = shard.tail
		shard.tail.prev = shard.head
		shard.size = 0
		shard.mu.Unlock()
	}
}

// evictLRU removes the least recently used item
func (s *Shard) evictLRU() {
	lru := s.tail.prev
	if lru != s.head {
		s.removeNode(lru)
		delete(s.cache, lru.key)
		s.size--
	}
}

// Size returns the total number of items in the cache
func (c *LRUCache) Size() int {
	total := 0
	for _, shard := range c.shards {
		shard.mu.RLock()
		total += shard.size
		shard.mu.RUnlock()
	}
	return total
}

func (c *LRUCache) Stats() map[string]interface{} {
	totalSize := 0
	shardSizes := make([]int, len(c.shards))

	for i, shard := range c.shards {
		shard.mu.RLock()
		shardSizes[i] = shard.size
		totalSize += shard.size
		shard.mu.RUnlock()
	}

	return map[string]interface{}{
		"total_size":  totalSize,
		"num_shards":  len(c.shards),
		"shard_sizes": shardSizes,
	}
}

func main() {
	cache := NewLRUCache(1000000)
	fmt.Println("LRU Cache has been created with 1M capacity")
	fmt.Printf("Number of shards: %d", len(cache.shards))
	cache.Put("user:1", "Tony")
	cache.Put("user:2", "Ayo")
	
	cache.Put("session:abc098", map[string]string{"token": "xyz123"})
	if value, ok := cache.Get("user:1"); ok {
		fmt.Printf("Found user:1= %v\n", value)
	}

	if _, ok := cache.Get("noneexistent"); !ok {
		fmt.Printf("Found user:1= %v\n", ok)
	}
	// Demonstrate concurrent access
	fmt.Println("\nTesting concurrent access...")

	var wg sync.WaitGroup
	numGoroutines := 100
	operationsPerGoroutine := 10000
	start := time.Now()

	// Launch concurrent workers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("worker:%d:key:%d", workerID, j)
				value := fmt.Sprintf("value_%d_%d", workerID, j)

				// Mix of puts and gets
				if j%3 == 0 {
					cache.Get(key) // May not exist yet
				} else {
					cache.Put(key, value)
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	totalOps := numGoroutines * operationsPerGoroutine
	fmt.Printf("Completed %d operations in %v\n", totalOps, duration)
	fmt.Printf("Throughput: %.0f ops/second\n", float64(totalOps)/duration.Seconds())

	// Print final stats
	stats := cache.Stats()
	fmt.Printf("\nFinal cache stats: %+v\n", stats)
}
