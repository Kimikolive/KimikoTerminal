package main

import (
    "container/heap"
    "context"
    "encoding/json"
    "sync"
    "time"
)

type MemoryBuffer struct {
    shortTerm      *MemoryHeap
    longTerm       *MemoryStore
    workingMemory  []Memory
    associations   map[string][]string
    mu             sync.RWMutex

    // Memory management parameters
    maxShortTerm   int
    maxLongTerm    int
    maxWorking     int
    decayRate      float64
}

type Memory struct {
    Content       string
    Type          string
    Timestamp     time.Time
    Importance    float64
    EmotionalTag  string
    Associations  []string
    AccessCount   int
    LastAccessed  time.Time
    Metadata      map[string]interface{}
}

type MemoryStore struct {
    memories     map[string]Memory
    indices      map[string][]string
    totalSize    int
    maxSize      int
}

type MemoryHeap []Memory

// Implement heap.Interface for MemoryHeap
func (h MemoryHeap) Len() int { return len(h) }
func (h MemoryHeap) Less(i, j int) bool {
    return h[i].Importance > h[j].Importance
}
func (h MemoryHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *MemoryHeap) Push(x interface{}) {
    *h = append(*h, x.(Memory))
}
func (h *MemoryHeap) Pop() interface{} {
    old := *h
    n := len(old)
    x := old[n-1]
    *h = old[0 : n-1]
    return x
}

func NewMemoryBuffer(config MemoryConfig) *MemoryBuffer {
    mb := &MemoryBuffer{
        shortTerm:     &MemoryHeap{},
        longTerm:      NewMemoryStore(config.MaxLongTerm),
        workingMemory: make([]Memory, 0, config.MaxWorking),
        associations:  make(map[string][]string),
        maxShortTerm: config.MaxShortTerm,
        maxLongTerm:  config.MaxLongTerm,
        maxWorking:   config.MaxWorking,
        decayRate:    config.DecayRate,
    }
    
    heap.Init(mb.shortTerm)
    go mb.runMemoryMaintenance()
    
    return mb
}

func (mb *MemoryBuffer) AddMemory(content string, memType string, importance float64) {
    mb.mu.Lock()
    defer mb.mu.Unlock()

    memory := Memory{
        Content:      content,
        Type:         memType,
        Timestamp:    time.Now(),
        Importance:   importance,
        EmotionalTag: mb.analyzeEmotionalContent(content),
        Associations: mb.findAssociations(content),
        AccessCount:  1,
        LastAccessed: time.Now(),
        Metadata:     make(map[string]interface{}),
    }

    // Add to short-term memory
    heap.Push(mb.shortTerm, memory)

    // Maintain short-term memory size
    if mb.shortTerm.Len() > mb.maxShortTerm {
        mb.consolidateMemory()
    }

    // Update associations
    mb.updateAssociations(memory)
}

func (mb *MemoryBuffer) consolidateMemory() {
    // Move least important memories to long-term storage
    for mb.shortTerm.Len() > mb.maxShortTerm {
        memory := heap.Pop(mb.shortTerm).(Memory)
        if memory.Importance > mb.calculateConsolidationThreshold() {
            mb.longTerm.Store(memory)
        }
    }
}

func (mb *MemoryBuffer) Recall(query string, limit int) []Memory {
    mb.mu.Lock()
    defer mb.mu.Unlock()

    // Search through all memory stores
    var results []Memory
    
    // Check working memory first
    results = append(results, mb.searchWorkingMemory(query)...)
    
    // Then short-term memory
    results = append(results, mb.searchShortTermMemory(query)...)
    
    // Finally, long-term memory
    results = append(results, mb.searchLongTermMemory(query)...)
    
    // Sort by relevance and importance
    mb.sortMemoriesByRelevance(results, query)
    
    // Update access counts and timestamps
    mb.updateMemoryAccess(results)
    
    // Return limited results
    if len(results) > limit {
        results = results[:limit]
    }
    
    return results
}

func (mb *MemoryBuffer) runMemoryMaintenance() {
    ticker := time.NewTicker(time.Hour)
    defer ticker.Stop()

    for range ticker.C {
        mb.mu.Lock()
        
        // Apply memory decay
        mb.applyMemoryDecay()
        
        // Consolidate memories
        mb.consolidateMemories()
        
        // Clean up old associations
        mb.cleanupAssociations()
        
        mb.mu.Unlock()
    }
}

func (mb *MemoryBuffer) applyMemoryDecay() {
    // Apply decay to short-term memories
    for i := range *mb.shortTerm {
        memory := &(*mb.shortTerm)[i]
        timeSinceAccess := time.Since(memory.LastAccessed)
        memory.Importance *= math.Exp(-mb.decayRate * timeSinceAccess.Hours())
    }
    
    // Reheap after modification
    heap.Init(mb.shortTerm)
}

func (mb *MemoryBuffer) updateAssociations(memory Memory) {
    // Extract keywords and create associations
    keywords := extractKeywords(memory.Content)
    
    for _, keyword := range keywords {
        mb.associations[keyword] = append(mb.associations[keyword], memory.Content)
        
        // Limit association list size
        if len(mb.associations[keyword]) > 100 {
            mb.associations[keyword] = mb.associations[keyword][1:]
        }
    }
}