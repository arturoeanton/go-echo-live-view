package liveview

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"
)

// TemplateCache manages compiled template caching
type TemplateCache struct {
	mu         sync.RWMutex
	cache      map[string]*CachedTemplate
	config     *TemplateCacheConfig
	stats      *CacheStats
	compiler   TemplateCompiler
	validators []TemplateValidator
}

// CachedTemplate represents a cached compiled template
type CachedTemplate struct {
	ID          string
	Template    *template.Template
	Compiled    CompiledTemplate
	Hash        string
	Size        int64
	CreatedAt   time.Time
	LastUsed    time.Time
	UseCount    int64
	Metadata    map[string]interface{}
	Dependencies []string
}

// CompiledTemplate represents a compiled template ready for execution
type CompiledTemplate interface {
	Execute(w io.Writer, data interface{}) error
	ExecuteTemplate(w io.Writer, name string, data interface{}) error
	Clone() (CompiledTemplate, error)
}

// TemplateCompiler compiles raw templates
type TemplateCompiler interface {
	Compile(source string) (CompiledTemplate, error)
	CompileWithFuncs(source string, funcs template.FuncMap) (CompiledTemplate, error)
}

// TemplateValidator validates templates before caching
type TemplateValidator func(source string) error

// TemplateCacheConfig configures the template cache
type TemplateCacheConfig struct {
	MaxSize          int64         // Max cache size in bytes
	MaxEntries       int           // Max number of cached templates
	TTL              time.Duration // Time to live for cached entries
	CleanupInterval  time.Duration // Cleanup interval
	EnableStats      bool          // Enable statistics
	EnablePrecompile bool          // Enable precompilation
	WarmupOnStart    bool          // Warmup cache on start
}

// CacheStats tracks cache statistics
type CacheStats struct {
	mu          sync.RWMutex
	Hits        int64
	Misses      int64
	Evictions   int64
	TotalSize   int64
	CompileTime time.Duration
	LastCleanup time.Time
}

// DefaultTemplateCacheConfig returns default configuration
func DefaultTemplateCacheConfig() *TemplateCacheConfig {
	return &TemplateCacheConfig{
		MaxSize:          100 * 1024 * 1024, // 100MB
		MaxEntries:       1000,
		TTL:              1 * time.Hour,
		CleanupInterval:  5 * time.Minute,
		EnableStats:      true,
		EnablePrecompile: true,
		WarmupOnStart:    false,
	}
}

// NewTemplateCache creates a new template cache
func NewTemplateCache(config *TemplateCacheConfig) *TemplateCache {
	if config == nil {
		config = DefaultTemplateCacheConfig()
	}

	tc := &TemplateCache{
		cache:      make(map[string]*CachedTemplate),
		config:     config,
		compiler:   NewDefaultCompiler(),
		validators: make([]TemplateValidator, 0),
	}

	if config.EnableStats {
		tc.stats = &CacheStats{}
	}

	// Start cleanup goroutine
	if config.CleanupInterval > 0 {
		go tc.cleanupLoop()
	}

	return tc
}

// Get retrieves a cached template
func (tc *TemplateCache) Get(key string) (*CachedTemplate, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	entry, exists := tc.cache[key]
	if !exists {
		if tc.stats != nil {
			tc.stats.recordMiss()
		}
		return nil, false
	}

	// Check TTL
	if tc.config.TTL > 0 && time.Since(entry.CreatedAt) > tc.config.TTL {
		// Entry expired
		if tc.stats != nil {
			tc.stats.recordMiss()
		}
		return nil, false
	}

	// Update last used
	entry.LastUsed = time.Now()
	entry.UseCount++

	if tc.stats != nil {
		tc.stats.recordHit()
	}

	return entry, true
}

// Compile compiles and caches a template
func (tc *TemplateCache) Compile(key, source string, funcs template.FuncMap) (*CachedTemplate, error) {
	// Check if already cached
	if cached, exists := tc.Get(key); exists {
		return cached, nil
	}

	// Validate template
	if err := tc.validate(source); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}

	// Compile template
	start := time.Now()
	
	var compiled CompiledTemplate
	var err error
	
	if funcs != nil {
		compiled, err = tc.compiler.CompileWithFuncs(source, funcs)
	} else {
		compiled, err = tc.compiler.Compile(source)
	}
	
	if err != nil {
		return nil, fmt.Errorf("template compilation failed: %w", err)
	}

	compileTime := time.Since(start)

	// Create cache entry
	entry := &CachedTemplate{
		ID:          key,
		Compiled:    compiled,
		Hash:        tc.hash(source),
		Size:        int64(len(source)),
		CreatedAt:   time.Now(),
		LastUsed:    time.Now(),
		UseCount:    1,
		Metadata:    make(map[string]interface{}),
		Dependencies: tc.extractDependencies(source),
	}

	// Cache the entry
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Check cache limits
	if err := tc.ensureSpace(entry.Size); err != nil {
		return nil, err
	}

	tc.cache[key] = entry

	if tc.stats != nil {
		tc.stats.mu.Lock()
		tc.stats.TotalSize += entry.Size
		tc.stats.CompileTime += compileTime
		tc.stats.mu.Unlock()
	}

	Debug("Template cached: %s (size: %d bytes)", key, entry.Size)
	return entry, nil
}

// CompileString compiles a template string without caching
func (tc *TemplateCache) CompileString(source string, funcs template.FuncMap) (CompiledTemplate, error) {
	// Generate cache key from source
	key := tc.hash(source)
	
	// Try to get from cache
	if cached, exists := tc.Get(key); exists {
		return cached.Compiled, nil
	}

	// Compile and cache
	cached, err := tc.Compile(key, source, funcs)
	if err != nil {
		return nil, err
	}

	return cached.Compiled, nil
}

// Invalidate removes a template from cache
func (tc *TemplateCache) Invalidate(key string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if entry, exists := tc.cache[key]; exists {
		delete(tc.cache, key)
		
		if tc.stats != nil {
			tc.stats.mu.Lock()
			tc.stats.TotalSize -= entry.Size
			tc.stats.Evictions++
			tc.stats.mu.Unlock()
		}
		
		Debug("Template invalidated: %s", key)
	}
}

// InvalidateAll clears the entire cache
func (tc *TemplateCache) InvalidateAll() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	count := len(tc.cache)
	tc.cache = make(map[string]*CachedTemplate)

	if tc.stats != nil {
		tc.stats.mu.Lock()
		tc.stats.TotalSize = 0
		tc.stats.Evictions += int64(count)
		tc.stats.mu.Unlock()
	}

	Debug("Template cache cleared: %d entries removed", count)
}

// Precompile precompiles a set of templates
func (tc *TemplateCache) Precompile(templates map[string]string, funcs template.FuncMap) error {
	if !tc.config.EnablePrecompile {
		return nil
	}

	for key, source := range templates {
		if _, err := tc.Compile(key, source, funcs); err != nil {
			return fmt.Errorf("failed to precompile %s: %w", key, err)
		}
	}

	Debug("Precompiled %d templates", len(templates))
	return nil
}

// AddValidator adds a template validator
func (tc *TemplateCache) AddValidator(validator TemplateValidator) {
	tc.validators = append(tc.validators, validator)
}

// validate validates a template using all validators
func (tc *TemplateCache) validate(source string) error {
	for _, validator := range tc.validators {
		if err := validator(source); err != nil {
			return err
		}
	}
	return nil
}

// hash generates a hash for template source
func (tc *TemplateCache) hash(source string) string {
	hasher := md5.New()
	hasher.Write([]byte(source))
	return hex.EncodeToString(hasher.Sum(nil))
}

// extractDependencies extracts template dependencies
func (tc *TemplateCache) extractDependencies(source string) []string {
	deps := make([]string, 0)
	
	// Extract {{template "name"}} calls
	templateRegex := regexp.MustCompile(`{{\s*template\s+"([^"]+)"`)
	matches := templateRegex.FindAllStringSubmatch(source, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			deps = append(deps, match[1])
		}
	}
	
	// Extract {{mount "component"}} calls
	mountRegex := regexp.MustCompile(`{{\s*mount\s+"([^"]+)"`)
	mountMatches := mountRegex.FindAllStringSubmatch(source, -1)
	
	for _, match := range mountMatches {
		if len(match) > 1 {
			deps = append(deps, "mount:"+match[1])
		}
	}
	
	return deps
}

// ensureSpace ensures there's enough space in cache
func (tc *TemplateCache) ensureSpace(size int64) error {
	if tc.config.MaxSize <= 0 && tc.config.MaxEntries <= 0 {
		return nil // No limits
	}

	// Check entry limit
	if tc.config.MaxEntries > 0 && len(tc.cache) >= tc.config.MaxEntries {
		// Evict least recently used
		tc.evictLRU()
	}

	// Check size limit
	if tc.config.MaxSize > 0 && tc.stats != nil {
		currentSize := tc.stats.TotalSize
		if currentSize+size > tc.config.MaxSize {
			// Evict until we have space
			tc.evictBySize(size)
		}
	}

	return nil
}

// evictLRU evicts the least recently used entry
func (tc *TemplateCache) evictLRU() {
	var lruKey string
	var lruTime time.Time

	for key, entry := range tc.cache {
		if lruKey == "" || entry.LastUsed.Before(lruTime) {
			lruKey = key
			lruTime = entry.LastUsed
		}
	}

	if lruKey != "" {
		delete(tc.cache, lruKey)
		if tc.stats != nil {
			tc.stats.Evictions++
		}
		Debug("Evicted LRU template: %s", lruKey)
	}
}

// evictBySize evicts entries to free up space
func (tc *TemplateCache) evictBySize(needed int64) {
	// Sort by last used time and evict oldest first
	type entry struct {
		key      string
		lastUsed time.Time
		size     int64
	}

	entries := make([]entry, 0, len(tc.cache))
	for k, v := range tc.cache {
		entries = append(entries, entry{
			key:      k,
			lastUsed: v.LastUsed,
			size:     v.Size,
		})
	}

	// Simple sort by last used
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].lastUsed.After(entries[j].lastUsed) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Evict until we have enough space
	freed := int64(0)
	for _, e := range entries {
		if freed >= needed {
			break
		}
		
		delete(tc.cache, e.key)
		freed += e.size
		
		if tc.stats != nil {
			tc.stats.Evictions++
			tc.stats.TotalSize -= e.size
		}
		
		Debug("Evicted template for space: %s", e.key)
	}
}

// cleanupLoop runs periodic cleanup
func (tc *TemplateCache) cleanupLoop() {
	ticker := time.NewTicker(tc.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		tc.cleanup()
	}
}

// cleanup removes expired entries
func (tc *TemplateCache) cleanup() {
	if tc.config.TTL <= 0 {
		return
	}

	tc.mu.Lock()
	defer tc.mu.Unlock()

	now := time.Now()
	expired := make([]string, 0)

	for key, entry := range tc.cache {
		if now.Sub(entry.CreatedAt) > tc.config.TTL {
			expired = append(expired, key)
		}
	}

	for _, key := range expired {
		if entry, exists := tc.cache[key]; exists {
			delete(tc.cache, key)
			
			if tc.stats != nil {
				tc.stats.TotalSize -= entry.Size
				tc.stats.Evictions++
			}
		}
	}

	if tc.stats != nil {
		tc.stats.LastCleanup = now
	}

	if len(expired) > 0 {
		Debug("Cleaned up %d expired templates", len(expired))
	}
}

// GetStats returns cache statistics
func (tc *TemplateCache) GetStats() *CacheStats {
	if tc.stats == nil {
		return nil
	}

	tc.stats.mu.RLock()
	defer tc.stats.mu.RUnlock()

	return &CacheStats{
		Hits:        tc.stats.Hits,
		Misses:      tc.stats.Misses,
		Evictions:   tc.stats.Evictions,
		TotalSize:   tc.stats.TotalSize,
		CompileTime: tc.stats.CompileTime,
		LastCleanup: tc.stats.LastCleanup,
	}
}

// recordHit records a cache hit
func (cs *CacheStats) recordHit() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.Hits++
}

// recordMiss records a cache miss
func (cs *CacheStats) recordMiss() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.Misses++
}

// DefaultCompiler is the default template compiler
type DefaultCompiler struct{}

// NewDefaultCompiler creates a new default compiler
func NewDefaultCompiler() *DefaultCompiler {
	return &DefaultCompiler{}
}

// Compile compiles a template
func (dc *DefaultCompiler) Compile(source string) (CompiledTemplate, error) {
	tmpl, err := template.New("template").Parse(source)
	if err != nil {
		return nil, err
	}
	return &defaultCompiledTemplate{tmpl: tmpl}, nil
}

// CompileWithFuncs compiles a template with functions
func (dc *DefaultCompiler) CompileWithFuncs(source string, funcs template.FuncMap) (CompiledTemplate, error) {
	tmpl, err := template.New("template").Funcs(funcs).Parse(source)
	if err != nil {
		return nil, err
	}
	return &defaultCompiledTemplate{tmpl: tmpl}, nil
}

// defaultCompiledTemplate wraps a standard template
type defaultCompiledTemplate struct {
	tmpl *template.Template
}

// Execute executes the template
func (ct *defaultCompiledTemplate) Execute(w io.Writer, data interface{}) error {
	return ct.tmpl.Execute(w, data)
}

// ExecuteTemplate executes a named template
func (ct *defaultCompiledTemplate) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	return ct.tmpl.ExecuteTemplate(w, name, data)
}

// Clone clones the template
func (ct *defaultCompiledTemplate) Clone() (CompiledTemplate, error) {
	cloned, err := ct.tmpl.Clone()
	if err != nil {
		return nil, err
	}
	return &defaultCompiledTemplate{tmpl: cloned}, nil
}

// TemplateRegistry manages multiple template caches by category
type TemplateRegistry struct {
	mu     sync.RWMutex
	caches map[string]*TemplateCache
}

// NewTemplateRegistry creates a new template registry
func NewTemplateRegistry() *TemplateRegistry {
	return &TemplateRegistry{
		caches: make(map[string]*TemplateCache),
	}
}

// GetCache gets or creates a cache for a category
func (tr *TemplateRegistry) GetCache(category string) *TemplateCache {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if cache, exists := tr.caches[category]; exists {
		return cache
	}

	cache := NewTemplateCache(nil)
	tr.caches[category] = cache
	return cache
}

// InvalidateCategory invalidates all templates in a category
func (tr *TemplateRegistry) InvalidateCategory(category string) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	if cache, exists := tr.caches[category]; exists {
		cache.InvalidateAll()
	}
}

// FastTemplateEngine provides optimized template execution
type FastTemplateEngine struct {
	cache     *TemplateCache
	funcMap   template.FuncMap
	baseDir   string
	extension string
}

// NewFastTemplateEngine creates a new fast template engine
func NewFastTemplateEngine(baseDir string) *FastTemplateEngine {
	return &FastTemplateEngine{
		cache:     NewTemplateCache(nil),
		funcMap:   make(template.FuncMap),
		baseDir:   baseDir,
		extension: ".html",
	}
}

// RegisterFunc registers a template function
func (fte *FastTemplateEngine) RegisterFunc(name string, fn interface{}) {
	fte.funcMap[name] = fn
}

// Render renders a template with data
func (fte *FastTemplateEngine) Render(name string, data interface{}) (string, error) {
	// Try cache first
	cached, exists := fte.cache.Get(name)
	
	if !exists {
		// Load and compile template
		source, err := fte.loadTemplate(name)
		if err != nil {
			return "", err
		}
		
		cached, err = fte.cache.Compile(name, source, fte.funcMap)
		if err != nil {
			return "", err
		}
	}

	// Execute template
	var buf bytes.Buffer
	if err := cached.Compiled.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// loadTemplate loads a template from disk
func (fte *FastTemplateEngine) loadTemplate(name string) (string, error) {
	// This is a simplified version - in production, implement proper file loading
	return fmt.Sprintf("<!-- Template: %s -->", name), nil
}

// Preload preloads templates into cache
func (fte *FastTemplateEngine) Preload(names []string) error {
	for _, name := range names {
		source, err := fte.loadTemplate(name)
		if err != nil {
			return err
		}
		
		if _, err := fte.cache.Compile(name, source, fte.funcMap); err != nil {
			return err
		}
	}
	return nil
}

// TemplateHelpers provides template helper functions
var TemplateHelpers = template.FuncMap{
	"upper":    strings.ToUpper,
	"lower":    strings.ToLower,
	"title":    strings.Title,
	"trim":     strings.TrimSpace,
	"contains": strings.Contains,
	"replace":  strings.ReplaceAll,
	"split":    strings.Split,
	"join":     strings.Join,
	"default": func(def, val interface{}) interface{} {
		if val == nil || val == "" {
			return def
		}
		return val
	},
	"safe": func(s string) template.HTML {
		return template.HTML(s)
	},
	"attr": func(s string) template.HTMLAttr {
		return template.HTMLAttr(s)
	},
	"js": func(s string) template.JS {
		return template.JS(s)
	},
	"css": func(s string) template.CSS {
		return template.CSS(s)
	},
	"url": func(s string) template.URL {
		return template.URL(s)
	},
}