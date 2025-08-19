package liveview

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// BenchmarkSuite provides comprehensive benchmarking utilities for LiveView components
type BenchmarkSuite struct {
	Name       string
	Components map[string]Component
	Metrics    *BenchmarkMetrics
	ctx        context.Context
	cancel     context.CancelFunc
}

// BenchmarkMetrics stores benchmark results
type BenchmarkMetrics struct {
	RenderTime       time.Duration
	EventProcessTime time.Duration
	MemoryUsed       uint64
	GoroutinesCount  int
	Operations       int64
	Errors           int64
	StartTime        time.Time
	EndTime          time.Time
	mu               sync.RWMutex
}

// NewBenchmarkSuite creates a new benchmark suite
func NewBenchmarkSuite(name string) *BenchmarkSuite {
	ctx, cancel := context.WithCancel(context.Background())
	return &BenchmarkSuite{
		Name:       name,
		Components: make(map[string]Component),
		Metrics:    &BenchmarkMetrics{StartTime: time.Now()},
		ctx:        ctx,
		cancel:     cancel,
	}
}

// RegisterComponent registers a component for benchmarking
func (bs *BenchmarkSuite) RegisterComponent(id string, component Component) {
	bs.Components[id] = component
}

// BenchmarkComponentRender benchmarks component rendering performance
func BenchmarkComponentRender(b *testing.B, component Component) {
	// Get template once
	tmpl := component.GetTemplate()
	t, err := template.New("bench").Parse(tmpl)
	if err != nil {
		b.Fatal(err)
	}
	
	// Initialize component
	component.Start()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		err := t.Execute(&buf, component)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEventProcessing benchmarks event processing performance
func BenchmarkEventProcessing(b *testing.B, component Component, eventName string, data interface{}) {
	driver := NewDriver("bench", component)
	component.Start()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		driver.ExecuteEvent(eventName, data)
	}
}

// BenchmarkConcurrentEvents benchmarks concurrent event processing
func BenchmarkConcurrentEvents(b *testing.B, component Component, eventName string, concurrency int) {
	driver := NewDriver("bench", component)
	component.Start()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			driver.ExecuteEvent(eventName, nil)
		}
	})
}

// BenchmarkMemoryUsage measures memory usage of components
func BenchmarkMemoryUsage(b *testing.B, componentFactory func() Component) {
	b.ReportAllocs()
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	allocBefore := m.Alloc
	
	components := make([]Component, b.N)
	for i := 0; i < b.N; i++ {
		components[i] = componentFactory()
		components[i].Start()
	}
	
	runtime.ReadMemStats(&m)
	allocAfter := m.Alloc
	
	avgMemoryPerComponent := (allocAfter - allocBefore) / uint64(b.N)
	b.ReportMetric(float64(avgMemoryPerComponent), "bytes/component")
}

// BenchmarkWebSocketThroughput benchmarks WebSocket message throughput
func BenchmarkWebSocketThroughput(b *testing.B, messageSize int) {
	// Create test message
	msg := make(map[string]interface{})
	msg["type"] = "data"
	msg["id"] = "test"
	msg["data"] = make([]byte, messageSize)
	
	data, err := json.Marshal(msg)
	if err != nil {
		b.Fatal(err)
	}
	
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		var decoded map[string]interface{}
		err := json.Unmarshal(data, &decoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkComponentLifecycle benchmarks full component lifecycle
func BenchmarkComponentLifecycle(b *testing.B, componentFactory func() Component) {
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		component := componentFactory()
		driver := NewDriver(fmt.Sprintf("bench-%d", i), component)
		
		// Start
		component.Start()
		
		// Render
		tmpl := component.GetTemplate()
		t, _ := template.New("bench").Parse(tmpl)
		var buf bytes.Buffer
		t.Execute(&buf, component)
		
		// Execute some events
		driver.ExecuteEvent("Click", nil)
		driver.ExecuteEvent("Change", "test")
		
		// Cleanup
		if closer, ok := component.(interface{ Close() }); ok {
			closer.Close()
		}
	}
}

// LoadTestComponent performs a load test on a component
func LoadTestComponent(component Component, duration time.Duration, rps int) *LoadTestResult {
	result := &LoadTestResult{
		StartTime: time.Now(),
		Duration:  duration,
		RPS:       rps,
	}
	
	driver := NewDriver("loadtest", component)
	component.Start()
	
	ticker := time.NewTicker(time.Second / time.Duration(rps))
	defer ticker.Stop()
	
	timeout := time.After(duration)
	var operations int64
	var errors int64
	
	for {
		select {
		case <-ticker.C:
			go func() {
				err := simulateUserAction(driver)
				if err != nil {
					atomic.AddInt64(&errors, 1)
				} else {
					atomic.AddInt64(&operations, 1)
				}
			}()
		case <-timeout:
			result.EndTime = time.Now()
			result.TotalOperations = atomic.LoadInt64(&operations)
			result.TotalErrors = atomic.LoadInt64(&errors)
			result.ActualDuration = result.EndTime.Sub(result.StartTime)
			result.ActualRPS = float64(result.TotalOperations) / result.ActualDuration.Seconds()
			return result
		}
	}
}

// LoadTestResult contains load test results
type LoadTestResult struct {
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
	ActualDuration  time.Duration
	RPS             int
	ActualRPS       float64
	TotalOperations int64
	TotalErrors     int64
}

// simulateUserAction simulates a user action for load testing
func simulateUserAction(driver LiveDriver) error {
	events := []string{"Click", "Change", "Submit", "Hover", "Focus"}
	event := events[time.Now().UnixNano()%int64(len(events))]
	driver.ExecuteEvent(event, "test-data")
	return nil
}

// ProfileComponent profiles a component's performance
func ProfileComponent(component Component, duration time.Duration) *ComponentProfile {
	profile := &ComponentProfile{
		ComponentName: fmt.Sprintf("%T", component),
		StartTime:     time.Now(),
	}
	
	driver := NewDriver("profile", component)
	component.Start()
	
	// Measure initial memory
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	profile.InitialMemory = m.Alloc
	profile.InitialGoroutines = runtime.NumGoroutine()
	
	// Run component for duration
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	timeout := time.After(duration)
	
	var renderCount int64
	var eventCount int64
	
	for {
		select {
		case <-ticker.C:
			// Render
			tmpl := component.GetTemplate()
			t, _ := template.New("profile").Parse(tmpl)
			var buf bytes.Buffer
			t.Execute(&buf, component)
			atomic.AddInt64(&renderCount, 1)
			
			// Process event
			driver.ExecuteEvent("ProfileEvent", nil)
			atomic.AddInt64(&eventCount, 1)
			
		case <-timeout:
			// Final measurements
			runtime.ReadMemStats(&m)
			profile.FinalMemory = m.Alloc
			profile.FinalGoroutines = runtime.NumGoroutine()
			profile.EndTime = time.Now()
			profile.RenderCount = atomic.LoadInt64(&renderCount)
			profile.EventCount = atomic.LoadInt64(&eventCount)
			profile.Duration = profile.EndTime.Sub(profile.StartTime)
			
			// Calculate metrics
			profile.MemoryGrowth = int64(profile.FinalMemory) - int64(profile.InitialMemory)
			profile.GoroutineLeaks = profile.FinalGoroutines - profile.InitialGoroutines
			profile.AvgRenderTime = profile.Duration / time.Duration(profile.RenderCount)
			profile.AvgEventTime = profile.Duration / time.Duration(profile.EventCount)
			
			return profile
		}
	}
}

// ComponentProfile contains profiling results for a component
type ComponentProfile struct {
	ComponentName     string
	StartTime         time.Time
	EndTime           time.Time
	Duration          time.Duration
	InitialMemory     uint64
	FinalMemory       uint64
	MemoryGrowth      int64
	InitialGoroutines int
	FinalGoroutines   int
	GoroutineLeaks    int
	RenderCount       int64
	EventCount        int64
	AvgRenderTime     time.Duration
	AvgEventTime      time.Duration
}

// String returns a formatted string representation of the profile
func (cp *ComponentProfile) String() string {
	return fmt.Sprintf(`
Component Profile: %s
==================
Duration: %v
Memory Growth: %d bytes
Goroutine Leaks: %d
Total Renders: %d (avg: %v)
Total Events: %d (avg: %v)
`, cp.ComponentName, cp.Duration, cp.MemoryGrowth, cp.GoroutineLeaks,
		cp.RenderCount, cp.AvgRenderTime, cp.EventCount, cp.AvgEventTime)
}

// BenchmarkScenario represents a complete benchmark scenario
type BenchmarkScenario struct {
	Name        string
	Description string
	Setup       func() Component
	Events      []BenchmarkEvent
	Duration    time.Duration
	Concurrent  bool
	Workers     int
}

// BenchmarkEvent represents an event in a benchmark scenario
type BenchmarkEvent struct {
	Name  string
	Data  interface{}
	Delay time.Duration
}

// RunBenchmarkScenario executes a complete benchmark scenario
func RunBenchmarkScenario(b *testing.B, scenario BenchmarkScenario) {
	b.Run(scenario.Name, func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			component := scenario.Setup()
			driver := NewDriver(fmt.Sprintf("scenario-%d", i), component)
			component.Start()
			
			if scenario.Concurrent {
				runConcurrentEvents(driver, scenario.Events, scenario.Workers)
			} else {
				runSequentialEvents(driver, scenario.Events)
			}
		}
	})
}

// runSequentialEvents runs events sequentially
func runSequentialEvents(driver LiveDriver, events []BenchmarkEvent) {
	for _, event := range events {
		if event.Delay > 0 {
			time.Sleep(event.Delay)
		}
		driver.ExecuteEvent(event.Name, event.Data)
	}
}

// runConcurrentEvents runs events concurrently
func runConcurrentEvents(driver LiveDriver, events []BenchmarkEvent, workers int) {
	eventChan := make(chan BenchmarkEvent, len(events))
	var wg sync.WaitGroup
	
	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for event := range eventChan {
				if event.Delay > 0 {
					time.Sleep(event.Delay)
				}
				driver.ExecuteEvent(event.Name, event.Data)
			}
		}()
	}
	
	// Send events
	for _, event := range events {
		eventChan <- event
	}
	close(eventChan)
	
	wg.Wait()
}

// CompareBenchmarks compares two benchmark results
func CompareBenchmarks(old, new *BenchmarkMetrics) *BenchmarkComparison {
	comp := &BenchmarkComparison{
		Old: old,
		New: new,
	}
	
	// Calculate differences
	comp.RenderTimeDiff = float64(new.RenderTime-old.RenderTime) / float64(old.RenderTime) * 100
	comp.EventTimeDiff = float64(new.EventProcessTime-old.EventProcessTime) / float64(old.EventProcessTime) * 100
	comp.MemoryDiff = float64(new.MemoryUsed-old.MemoryUsed) / float64(old.MemoryUsed) * 100
	comp.OperationsDiff = float64(new.Operations-old.Operations) / float64(old.Operations) * 100
	
	// Determine if improved
	comp.Improved = comp.RenderTimeDiff < 0 && comp.EventTimeDiff < 0 && comp.MemoryDiff < 0
	
	return comp
}

// BenchmarkComparison contains comparison results
type BenchmarkComparison struct {
	Old            *BenchmarkMetrics
	New            *BenchmarkMetrics
	RenderTimeDiff float64 // Percentage difference
	EventTimeDiff  float64
	MemoryDiff     float64
	OperationsDiff float64
	Improved       bool
}

// String returns a formatted comparison report
func (bc *BenchmarkComparison) String() string {
	status := "REGRESSION"
	if bc.Improved {
		status = "IMPROVEMENT"
	}
	
	return fmt.Sprintf(`
Benchmark Comparison - %s
================================
Render Time: %.2f%% 
Event Processing: %.2f%%
Memory Usage: %.2f%%
Operations: %.2f%%
`, status, bc.RenderTimeDiff, bc.EventTimeDiff, bc.MemoryDiff, bc.OperationsDiff)
}

// GenerateBenchmarkReport generates a comprehensive benchmark report
func GenerateBenchmarkReport(results map[string]*BenchmarkMetrics) string {
	report := "LiveView Benchmark Report\n"
	report += "=========================\n\n"
	
	for name, metrics := range results {
		report += fmt.Sprintf("Component: %s\n", name)
		report += fmt.Sprintf("  Render Time: %v\n", metrics.RenderTime)
		report += fmt.Sprintf("  Event Time: %v\n", metrics.EventProcessTime)
		report += fmt.Sprintf("  Memory: %d bytes\n", metrics.MemoryUsed)
		report += fmt.Sprintf("  Operations: %d\n", metrics.Operations)
		report += fmt.Sprintf("  Errors: %d\n", metrics.Errors)
		report += fmt.Sprintf("  Duration: %v\n\n", metrics.EndTime.Sub(metrics.StartTime))
	}
	
	return report
}