package liveview

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewTemplateCache(t *testing.T) {
	tc := NewTemplateCache(nil)
	
	if tc == nil {
		t.Fatal("NewTemplateCache returned nil")
	}
	
	if tc.config == nil {
		t.Error("Expected default config to be set")
	}
	
	if len(tc.cache) != 0 {
		t.Error("Expected empty cache")
	}
	
	if tc.compiler == nil {
		t.Error("Expected compiler to be set")
	}
}

func TestTemplateCacheCompile(t *testing.T) {
	tc := NewTemplateCache(nil)
	
	source := `<div>Hello {{.Name}}</div>`
	
	cached, err := tc.Compile("test", source, nil)
	if err != nil {
		t.Errorf("Compile() error = %v", err)
	}
	
	if cached == nil {
		t.Fatal("Expected cached template")
	}
	
	if cached.ID != "test" {
		t.Errorf("Expected ID 'test', got %s", cached.ID)
	}
	
	if cached.Hash == "" {
		t.Error("Expected hash to be calculated")
	}
	
	if cached.Size != int64(len(source)) {
		t.Errorf("Expected size %d, got %d", len(source), cached.Size)
	}
	
	// Execute template
	var buf bytes.Buffer
	data := map[string]string{"Name": "World"}
	err = cached.Compiled.Execute(&buf, data)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
	
	expected := "<div>Hello World</div>"
	if buf.String() != expected {
		t.Errorf("Expected output %s, got %s", expected, buf.String())
	}
}

func TestTemplateCacheGet(t *testing.T) {
	tc := NewTemplateCache(nil)
	
	source := `<div>Test</div>`
	
	// Compile and cache
	_, err := tc.Compile("test", source, nil)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	
	// Get from cache
	cached, exists := tc.Get("test")
	if !exists {
		t.Error("Expected template to exist in cache")
	}
	
	if cached == nil {
		t.Error("Expected cached template")
	}
	
	if cached.UseCount != 2 { // 1 from compile, 1 from get
		t.Errorf("Expected use count 2, got %d", cached.UseCount)
	}
	
	// Get non-existent
	_, exists = tc.Get("nonexistent")
	if exists {
		t.Error("Expected template to not exist")
	}
}

func TestTemplateCacheWithFuncs(t *testing.T) {
	tc := NewTemplateCache(nil)
	
	source := `<div>{{upper .Text}}</div>`
	
	funcs := template.FuncMap{
		"upper": strings.ToUpper,
	}
	
	cached, err := tc.Compile("test", source, funcs)
	if err != nil {
		t.Errorf("Compile() error = %v", err)
	}
	
	// Execute with custom function
	var buf bytes.Buffer
	data := map[string]string{"Text": "hello"}
	err = cached.Compiled.Execute(&buf, data)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
	
	expected := "<div>HELLO</div>"
	if buf.String() != expected {
		t.Errorf("Expected output %s, got %s", expected, buf.String())
	}
}

func TestTemplateCacheTTL(t *testing.T) {
	config := &TemplateCacheConfig{
		TTL:          100 * time.Millisecond,
		EnableStats:  true,
	}
	tc := NewTemplateCache(config)
	
	source := `<div>Test</div>`
	
	// Compile and cache
	_, err := tc.Compile("test", source, nil)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	
	// Should exist immediately
	_, exists := tc.Get("test")
	if !exists {
		t.Error("Expected template to exist")
	}
	
	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)
	
	// Should not exist after TTL
	_, exists = tc.Get("test")
	if exists {
		t.Error("Expected template to expire after TTL")
	}
}

func TestTemplateCacheInvalidate(t *testing.T) {
	tc := NewTemplateCache(nil)
	
	source := `<div>Test</div>`
	
	// Compile and cache
	_, err := tc.Compile("test", source, nil)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	
	// Should exist
	_, exists := tc.Get("test")
	if !exists {
		t.Error("Expected template to exist")
	}
	
	// Invalidate
	tc.Invalidate("test")
	
	// Should not exist after invalidation
	_, exists = tc.Get("test")
	if exists {
		t.Error("Expected template to be invalidated")
	}
}

func TestTemplateCacheInvalidateAll(t *testing.T) {
	tc := NewTemplateCache(nil)
	
	// Compile multiple templates
	for i := 0; i < 5; i++ {
		source := fmt.Sprintf(`<div>Test %d</div>`, i)
		_, err := tc.Compile(fmt.Sprintf("test%d", i), source, nil)
		if err != nil {
			t.Fatalf("Compile() error = %v", err)
		}
	}
	
	// All should exist
	for i := 0; i < 5; i++ {
		_, exists := tc.Get(fmt.Sprintf("test%d", i))
		if !exists {
			t.Errorf("Expected template test%d to exist", i)
		}
	}
	
	// Invalidate all
	tc.InvalidateAll()
	
	// None should exist
	for i := 0; i < 5; i++ {
		_, exists := tc.Get(fmt.Sprintf("test%d", i))
		if exists {
			t.Errorf("Expected template test%d to be invalidated", i)
		}
	}
}

func TestTemplateCacheMaxEntries(t *testing.T) {
	config := &TemplateCacheConfig{
		MaxEntries:  3,
		EnableStats: true,
	}
	tc := NewTemplateCache(config)
	
	// Add templates up to limit
	for i := 0; i < 4; i++ {
		source := fmt.Sprintf(`<div>Test %d</div>`, i)
		_, err := tc.Compile(fmt.Sprintf("test%d", i), source, nil)
		if err != nil {
			t.Fatalf("Compile() error = %v", err)
		}
	}
	
	// Only last 3 should exist (LRU eviction)
	tc.mu.RLock()
	count := len(tc.cache)
	tc.mu.RUnlock()
	
	if count != 3 {
		t.Errorf("Expected 3 cached entries, got %d", count)
	}
	
	// Check stats
	stats := tc.GetStats()
	if stats.Evictions < 1 {
		t.Error("Expected at least 1 eviction")
	}
}

func TestTemplateCacheMaxSize(t *testing.T) {
	config := &TemplateCacheConfig{
		MaxSize:     100, // 100 bytes
		EnableStats: true,
	}
	tc := NewTemplateCache(config)
	
	// Add templates that exceed size limit
	source1 := strings.Repeat("a", 40)
	source2 := strings.Repeat("b", 40)
	source3 := strings.Repeat("c", 40) // Total would be 120 bytes
	
	tc.Compile("test1", source1, nil)
	tc.Compile("test2", source2, nil)
	tc.Compile("test3", source3, nil)
	
	// Check total size doesn't exceed limit
	stats := tc.GetStats()
	if stats.TotalSize > 100 {
		t.Errorf("Expected total size <= 100, got %d", stats.TotalSize)
	}
	
	// Should have evictions
	if stats.Evictions < 1 {
		t.Error("Expected at least 1 eviction due to size limit")
	}
}

func TestTemplateCachePrecompile(t *testing.T) {
	tc := NewTemplateCache(nil)
	
	templates := map[string]string{
		"header": `<header>{{.Title}}</header>`,
		"footer": `<footer>{{.Copyright}}</footer>`,
		"body":   `<main>{{.Content}}</main>`,
	}
	
	err := tc.Precompile(templates, nil)
	if err != nil {
		t.Errorf("Precompile() error = %v", err)
	}
	
	// All templates should be cached
	for key := range templates {
		_, exists := tc.Get(key)
		if !exists {
			t.Errorf("Expected template %s to be precompiled", key)
		}
	}
}

func TestTemplateCacheValidators(t *testing.T) {
	tc := NewTemplateCache(nil)
	
	// Add validator that rejects templates with script tags
	tc.AddValidator(func(source string) error {
		if strings.Contains(source, "<script>") {
			return fmt.Errorf("script tags not allowed")
		}
		return nil
	})
	
	// Valid template
	validSource := `<div>Safe content</div>`
	_, err := tc.Compile("valid", validSource, nil)
	if err != nil {
		t.Errorf("Expected valid template to compile: %v", err)
	}
	
	// Invalid template
	invalidSource := `<div><script>alert('xss')</script></div>`
	_, err = tc.Compile("invalid", invalidSource, nil)
	if err == nil {
		t.Error("Expected invalid template to fail validation")
	}
}

func TestTemplateCacheExtractDependencies(t *testing.T) {
	tc := NewTemplateCache(nil)
	
	// Need to provide mount function
	funcs := template.FuncMap{
		"mount": func(name string) string {
			return fmt.Sprintf("<!-- mount: %s -->", name)
		},
	}
	
	source := `
		<div>
			{{template "header" .}}
			{{mount "navbar"}}
			{{template "footer" .}}
		</div>
	`
	
	cached, err := tc.Compile("test", source, funcs)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	
	// Check dependencies
	expectedDeps := []string{"header", "mount:navbar", "footer"}
	
	if len(cached.Dependencies) != len(expectedDeps) {
		t.Errorf("Expected %d dependencies, got %d", len(expectedDeps), len(cached.Dependencies))
	}
	
	// Check each dependency
	depMap := make(map[string]bool)
	for _, dep := range cached.Dependencies {
		depMap[dep] = true
	}
	
	for _, expected := range expectedDeps {
		if !depMap[expected] {
			t.Errorf("Missing expected dependency: %s", expected)
		}
	}
}

func TestTemplateCacheStats(t *testing.T) {
	config := &TemplateCacheConfig{
		EnableStats: true,
	}
	tc := NewTemplateCache(config)
	
	source := `<div>Test</div>`
	
	// Compile (initial compile doesn't count as miss since it's not a Get)
	tc.Compile("test", source, nil)
	
	// Get (hit)
	tc.Get("test")
	
	// Get non-existent (miss)
	tc.Get("nonexistent")
	
	// Check stats
	stats := tc.GetStats()
	
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
	
	// During Compile, we call Get internally which causes a miss, plus the explicit miss
	if stats.Misses < 1 {
		t.Errorf("Expected at least 1 miss, got %d", stats.Misses)
	}
	
	if stats.TotalSize != int64(len(source)) {
		t.Errorf("Expected total size %d, got %d", len(source), stats.TotalSize)
	}
}

func TestTemplateCacheConcurrent(t *testing.T) {
	tc := NewTemplateCache(nil)
	
	var wg sync.WaitGroup
	numGoroutines := 100
	
	// Concurrent compile and get
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			key := fmt.Sprintf("test%d", id%10)
			source := fmt.Sprintf(`<div>Test %d</div>`, id%10)
			
			// Compile
			tc.Compile(key, source, nil)
			
			// Get
			tc.Get(key)
		}(i)
	}
	
	wg.Wait()
	
	// Should have at most 10 cached templates
	tc.mu.RLock()
	count := len(tc.cache)
	tc.mu.RUnlock()
	
	if count > 10 {
		t.Errorf("Expected at most 10 cached templates, got %d", count)
	}
}

func TestTemplateRegistry(t *testing.T) {
	registry := NewTemplateRegistry()
	
	// Get or create cache
	cache1 := registry.GetCache("components")
	cache2 := registry.GetCache("layouts")
	
	if cache1 == nil || cache2 == nil {
		t.Error("Expected caches to be created")
	}
	
	// Same category should return same cache
	cache1Again := registry.GetCache("components")
	if cache1 != cache1Again {
		t.Error("Expected same cache instance for same category")
	}
	
	// Add templates to different caches
	cache1.Compile("header", `<header>Test</header>`, nil)
	cache2.Compile("main", `<main>Test</main>`, nil)
	
	// Invalidate category
	registry.InvalidateCategory("components")
	
	// Check only components cache was invalidated
	_, exists1 := cache1.Get("header")
	_, exists2 := cache2.Get("main")
	
	if exists1 {
		t.Error("Expected components cache to be invalidated")
	}
	
	if !exists2 {
		t.Error("Expected layouts cache to remain")
	}
}

func TestFastTemplateEngine(t *testing.T) {
	engine := NewFastTemplateEngine("/templates")
	
	// Register custom function
	engine.RegisterFunc("double", func(n int) int {
		return n * 2
	})
	
	// Since we can't load from disk in test, compile directly
	source := `<div>{{double .Number}}</div>`
	engine.cache.Compile("test", source, engine.funcMap)
	
	// Render
	result, err := engine.Render("test", map[string]int{"Number": 5})
	if err != nil {
		t.Errorf("Render() error = %v", err)
	}
	
	expected := "<div>10</div>"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestTemplateHelpers(t *testing.T) {
	tests := []struct {
		name     string
		helper   string
		input    interface{}
		expected interface{}
	}{
		{"upper", "upper", "hello", "HELLO"},
		{"lower", "lower", "HELLO", "hello"},
		{"trim", "trim", "  hello  ", "hello"},
		{"contains", "contains", []interface{}{"hello world", "world"}, true},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fn, exists := TemplateHelpers[test.helper]
			if !exists {
				t.Errorf("Helper %s not found", test.helper)
				return
			}
			
			// Test based on function signature
			switch test.helper {
			case "upper", "lower", "trim":
				result := fn.(func(string) string)(test.input.(string))
				if result != test.expected {
					t.Errorf("Expected %v, got %v", test.expected, result)
				}
			case "contains":
				args := test.input.([]interface{})
				result := fn.(func(string, string) bool)(args[0].(string), args[1].(string))
				if result != test.expected {
					t.Errorf("Expected %v, got %v", test.expected, result)
				}
			}
		})
	}
}

func BenchmarkTemplateCacheCompile(b *testing.B) {
	tc := NewTemplateCache(nil)
	source := `<div>{{.Title}}</div>`
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("test%d", i)
		tc.Compile(key, source, nil)
	}
}

func BenchmarkTemplateCacheGet(b *testing.B) {
	tc := NewTemplateCache(nil)
	source := `<div>{{.Title}}</div>`
	
	// Pre-compile
	tc.Compile("test", source, nil)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		tc.Get("test")
	}
}

func BenchmarkTemplateExecute(b *testing.B) {
	tc := NewTemplateCache(nil)
	source := `<div>{{.Title}} - {{.Content}}</div>`
	
	cached, _ := tc.Compile("test", source, nil)
	data := map[string]string{
		"Title":   "Test Title",
		"Content": "Test Content",
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		cached.Compiled.Execute(&buf, data)
	}
}