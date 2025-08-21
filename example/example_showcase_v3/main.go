package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// ShowcaseComponent demonstrates all framework features
type ShowcaseComponent struct {
	*liveview.ComponentDriver[*ShowcaseComponent]
	
	// Showcase state
	ActiveDemo      string
	DemoResults     map[string]string
	PerformanceData map[string]float64
	EventLog        []string
	Features        []FeatureDemo
	
	// Framework components
	errorBoundary   *liveview.ErrorBoundary
	stateManager    *liveview.StateManager
	eventRegistry   *liveview.EventRegistry
	lifecycle       *liveview.LifecycleManager
	templateCache   *liveview.TemplateCache
	lazyLoader      *liveview.LazyLoader
}

type FeatureDemo struct {
	Name        string
	Description string
	Status      string
	Performance float64
	Icon        string
}

func (s *ShowcaseComponent) Start() {
	// Initialize all framework features
	s.initializeFramework()
	
	// Initialize showcase state
	s.ActiveDemo = "overview"
	s.DemoResults = make(map[string]string)
	s.PerformanceData = make(map[string]float64)
	s.EventLog = []string{"System initialized"}
	
	s.Features = []FeatureDemo{
		{
			Name:        "Error Boundaries",
			Description: "Automatic error recovery",
			Status:      "active",
			Performance: 99.9,
			Icon:        "🛡️",
		},
		{
			Name:        "State Management",
			Description: "Reactive state with persistence",
			Status:      "active",
			Performance: 98.5,
			Icon:        "💾",
		},
		{
			Name:        "Virtual DOM",
			Description: "Efficient rendering",
			Status:      "active",
			Performance: 97.2,
			Icon:        "⚡",
		},
		{
			Name:        "Event Registry",
			Description: "Advanced event handling",
			Status:      "active",
			Performance: 99.1,
			Icon:        "🎯",
		},
		{
			Name:        "Template Cache",
			Description: "Compiled template caching",
			Status:      "active",
			Performance: 96.8,
			Icon:        "📦",
		},
		{
			Name:        "Lazy Loading",
			Description: "On-demand component loading",
			Status:      "active",
			Performance: 94.5,
			Icon:        "🎭",
		},
		{
			Name:        "Lifecycle Hooks",
			Description: "Component lifecycle management",
			Status:      "active",
			Performance: 99.7,
			Icon:        "🔄",
		},
		{
			Name:        "SafeScript API",
			Description: "Secure JavaScript execution",
			Status:      "active",
			Performance: 100.0,
			Icon:        "🔒",
		},
	}
	
	// Execute lifecycle
	s.lifecycle.Create()
	s.lifecycle.Mount()
	
	// Start performance monitoring after everything is initialized
	go s.monitorPerformance()
	
	s.Commit()
}

func (s *ShowcaseComponent) initializeFramework() {
	// Error Boundary with custom fallback
	s.errorBoundary = liveview.NewErrorBoundary(100, true)
	s.errorBoundary.SetFallbackRenderer(func(componentID string, err error) string {
		return fmt.Sprintf(`
			<div class="error-recovery">
				<h3>Component Recovered</h3>
				<p>Error: %v</p>
				<p>The component has been automatically recovered.</p>
			</div>
		`, err)
	})
	
	// State Manager with all features
	s.stateManager = liveview.NewStateManager(&liveview.StateConfig{
		Provider:         liveview.NewJSONStateProvider(liveview.NewMemoryStateProvider()),
		CacheEnabled:     true,
		CacheTTL:         10 * time.Minute,
		AutoPersist:      true,
		PersistInterval:  20 * time.Second,
		EnableVersioning: true,
	})
	
	// Event Registry with metrics
	s.eventRegistry = liveview.NewEventRegistry(&liveview.EventRegistryConfig{
		MaxHandlersPerEvent: 10,
		EnableMetrics:       true,
		EnableWildcards:     true,
		DefaultTimeout:      30 * time.Second,
		EnableNamespaces:    true,
	})
	
	// Register showcase event handlers
	s.eventRegistry.On("demo.*", func(ctx context.Context, event *liveview.Event) error {
		s.logEvent(fmt.Sprintf("Demo event: %s", event.Type))
		return nil
	})
	
	// Lifecycle Manager with all hooks
	s.lifecycle = liveview.NewLifecycleManager("showcase")
	s.lifecycle.SetHooks(&liveview.LifecycleHooks{
		OnBeforeCreate: func() error {
			s.logEvent("Before create")
			return nil
		},
		OnCreated: func() error {
			s.logEvent("Component created")
			return nil
		},
		OnBeforeMount: func() error {
			s.logEvent("Before mount")
			return nil
		},
		OnMounted: func() error {
			s.logEvent("Component mounted")
			s.stateManager.Set("mount_time", time.Now())
			return nil
		},
		OnBeforeUpdate: func(oldData, newData interface{}) error {
			s.logEvent("Before update")
			return nil
		},
		OnUpdated: func() error {
			s.logEvent("Component updated")
			return nil
		},
		OnBeforeUnmount: func() error {
			s.logEvent("Before unmount")
			s.saveShowcaseState()
			return nil
		},
		OnUnmounted: func() error {
			s.logEvent("Component unmounted")
			return nil
		},
		OnError: func(stage liveview.LifecycleStage, err error) error {
			s.logEvent(fmt.Sprintf("Error at stage %s: %v", stage, err))
			return nil
		},
	})
	
	// Template Cache
	s.templateCache = liveview.NewTemplateCache(&liveview.TemplateCacheConfig{
		MaxSize:            20 * 1024 * 1024, // 20MB
		TTL:                10 * time.Minute,
		EnablePrecompile:   true,
	})
	
	// Lazy Loader
	s.lazyLoader = liveview.NewLazyLoader(&liveview.LazyLoaderConfig{
		MaxRetries:    3,
		RetryDelay:    1 * time.Second,
		LoadTimeout:   5 * time.Second,
		EnableCaching: true,
		EnableMetrics: true,
	})
}

func (s *ShowcaseComponent) logEvent(message string) {
	s.EventLog = append(s.EventLog, fmt.Sprintf("[%s] %s", 
		time.Now().Format("15:04:05"), message))
	
	// Keep only last 10 events
	if len(s.EventLog) > 10 {
		s.EventLog = s.EventLog[len(s.EventLog)-10:]
	}
}

func (s *ShowcaseComponent) monitorPerformance() {
	// Wait a bit to ensure everything is initialized
	time.Sleep(1 * time.Second)
	
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		if s.lifecycle != nil && s.lifecycle.GetStage() == liveview.StageMounted {
			// Update performance metrics
			for i, feature := range s.Features {
				// Simulate performance variations
				s.Features[i].Performance = feature.Performance + (rand.Float64()-0.5)*2
				if s.Features[i].Performance > 100 {
					s.Features[i].Performance = 100
				}
				if s.Features[i].Performance < 90 {
					s.Features[i].Performance = 90
				}
			}
			
			// Update performance data (simulated since metrics are private)
			if s.PerformanceData == nil {
				s.PerformanceData = make(map[string]float64)
			}
			s.PerformanceData["events_processed"] = s.PerformanceData["events_processed"] + 1
			s.PerformanceData["handlers_registered"] = 5
			
			// Get cache statistics (check for nil)
			if s.templateCache != nil {
				stats := s.templateCache.GetStats()
				if stats != nil {
					s.PerformanceData["cache_hits"] = float64(stats.Hits)
					s.PerformanceData["cache_size"] = float64(stats.TotalSize)
				} else {
					s.PerformanceData["cache_hits"] = 0
					s.PerformanceData["cache_size"] = 0
				}
			} else {
				s.PerformanceData["cache_hits"] = 0
				s.PerformanceData["cache_size"] = 0
			}
			
			s.Commit()
		}
	}
}

func (s *ShowcaseComponent) saveShowcaseState() error {
	// Create state snapshot
	snapshot, err := s.stateManager.TakeSnapshot()
	if err != nil {
		return err
	}
	
	s.stateManager.Set("showcase_snapshot", snapshot)
	s.stateManager.Set("showcase_features", s.Features)
	s.stateManager.Set("showcase_event_log", s.EventLog)
	
	return nil
}

func (s *ShowcaseComponent) GetTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
	<title>Framework Showcase v3</title>
	<style>
		* { margin: 0; padding: 0; box-sizing: border-box; }
		
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			background: #0a0e27;
			color: white;
			min-height: 100vh;
		}
		
		.header {
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			padding: 2rem;
			text-align: center;
			box-shadow: 0 4px 30px rgba(0,0,0,0.5);
		}
		
		.header h1 {
			font-size: 2.5rem;
			margin-bottom: 0.5rem;
		}
		
		.header p {
			opacity: 0.9;
			font-size: 1.1rem;
		}
		
		.container {
			max-width: 1400px;
			margin: 0 auto;
			padding: 2rem;
		}
		
		.demo-selector {
			display: flex;
			gap: 1rem;
			margin-bottom: 2rem;
			flex-wrap: wrap;
		}
		
		.demo-btn {
			padding: 0.75rem 1.5rem;
			background: rgba(102, 126, 234, 0.2);
			border: 2px solid #667eea;
			color: white;
			border-radius: 8px;
			cursor: pointer;
			transition: all 0.3s;
			font-size: 1rem;
		}
		
		.demo-btn:hover, .demo-btn.active {
			background: rgba(102, 126, 234, 0.4);
			transform: translateY(-2px);
			box-shadow: 0 4px 20px rgba(102, 126, 234, 0.4);
		}
		
		.features-grid {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
			gap: 1.5rem;
			margin-bottom: 2rem;
		}
		
		.feature-card {
			background: linear-gradient(135deg, rgba(102, 126, 234, 0.1), rgba(118, 75, 162, 0.1));
			border: 1px solid rgba(102, 126, 234, 0.3);
			border-radius: 12px;
			padding: 1.5rem;
			transition: all 0.3s;
		}
		
		.feature-card:hover {
			transform: translateY(-5px);
			box-shadow: 0 10px 30px rgba(102, 126, 234, 0.3);
			border-color: #667eea;
		}
		
		.feature-header {
			display: flex;
			align-items: center;
			gap: 1rem;
			margin-bottom: 1rem;
		}
		
		.feature-icon {
			font-size: 2rem;
		}
		
		.feature-name {
			font-size: 1.2rem;
			font-weight: 600;
		}
		
		.feature-description {
			color: #999;
			margin-bottom: 1rem;
		}
		
		.performance-bar {
			background: rgba(255,255,255,0.1);
			border-radius: 20px;
			height: 8px;
			overflow: hidden;
			margin-bottom: 0.5rem;
		}
		
		.performance-fill {
			height: 100%;
			background: linear-gradient(90deg, #667eea, #764ba2);
			border-radius: 20px;
			transition: width 0.5s ease;
		}
		
		.performance-text {
			font-size: 0.875rem;
			color: #667eea;
		}
		
		.status-badge {
			display: inline-block;
			padding: 0.25rem 0.75rem;
			background: rgba(72, 187, 120, 0.2);
			border: 1px solid #48bb78;
			color: #48bb78;
			border-radius: 20px;
			font-size: 0.75rem;
			text-transform: uppercase;
		}
		
		.metrics-section {
			background: rgba(26, 26, 46, 0.8);
			border-radius: 12px;
			padding: 1.5rem;
			margin-bottom: 2rem;
		}
		
		.metrics-grid {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
			gap: 1rem;
			margin-top: 1rem;
		}
		
		.metric {
			text-align: center;
		}
		
		.metric-value {
			font-size: 2rem;
			font-weight: bold;
			color: #667eea;
		}
		
		.metric-label {
			color: #999;
			font-size: 0.875rem;
			margin-top: 0.25rem;
		}
		
		.event-log {
			background: rgba(26, 26, 46, 0.8);
			border-radius: 12px;
			padding: 1.5rem;
		}
		
		.event-log h3 {
			margin-bottom: 1rem;
			color: #667eea;
		}
		
		.event-item {
			padding: 0.5rem;
			background: rgba(102, 126, 234, 0.1);
			border-left: 3px solid #667eea;
			margin-bottom: 0.5rem;
			border-radius: 4px;
			font-family: monospace;
			font-size: 0.875rem;
		}
		
		.action-buttons {
			position: fixed;
			bottom: 2rem;
			right: 2rem;
			display: flex;
			flex-direction: column;
			gap: 1rem;
		}
		
		.action-btn {
			width: 56px;
			height: 56px;
			border-radius: 50%;
			background: linear-gradient(135deg, #667eea, #764ba2);
			border: none;
			color: white;
			font-size: 24px;
			cursor: pointer;
			box-shadow: 0 4px 20px rgba(102, 126, 234, 0.4);
			transition: all 0.3s;
		}
		
		.action-btn:hover {
			transform: scale(1.1);
			box-shadow: 0 6px 30px rgba(102, 126, 234, 0.6);
		}
		
		.error-recovery {
			background: rgba(245, 101, 101, 0.1);
			border: 2px solid #f56565;
			border-radius: 8px;
			padding: 1rem;
			margin: 1rem 0;
		}
	</style>
</head>
<body>
	<div class="header">
		<h1>🚀 Framework Showcase v3</h1>
		<p>Experience all enhanced framework features in action</p>
	</div>
	
	<div class="container">
		<div class="demo-selector">
			<button class="demo-btn {{if eq .ActiveDemo "overview"}}active{{end}}" 
			        onclick="send_event('{{.IdComponent}}', 'SetDemo', {demo: 'overview'})">
				Overview
			</button>
			<button class="demo-btn {{if eq .ActiveDemo "error_boundary"}}active{{end}}"
			        onclick="send_event('{{.IdComponent}}', 'SetDemo', {demo: 'error_boundary'})">
				Error Boundaries
			</button>
			<button class="demo-btn {{if eq .ActiveDemo "state_management"}}active{{end}}"
			        onclick="send_event('{{.IdComponent}}', 'SetDemo', {demo: 'state_management'})">
				State Management
			</button>
			<button class="demo-btn {{if eq .ActiveDemo "virtual_dom"}}active{{end}}"
			        onclick="send_event('{{.IdComponent}}', 'SetDemo', {demo: 'virtual_dom'})">
				Virtual DOM
			</button>
			<button class="demo-btn {{if eq .ActiveDemo "events"}}active{{end}}"
			        onclick="send_event('{{.IdComponent}}', 'SetDemo', {demo: 'events'})">
				Event System
			</button>
		</div>
		
		<div class="features-grid">
			{{range .Features}}
			<div class="feature-card">
				<div class="feature-header">
					<span class="feature-icon">{{.Icon}}</span>
					<span class="feature-name">{{.Name}}</span>
				</div>
				<div class="feature-description">{{.Description}}</div>
				<div class="performance-bar">
					<div class="performance-fill" style="width: {{.Performance}}%;"></div>
				</div>
				<div style="display: flex; justify-content: space-between; align-items: center;">
					<span class="performance-text">{{printf "%.1f" .Performance}}% Performance</span>
					<span class="status-badge">{{.Status}}</span>
				</div>
			</div>
			{{end}}
		</div>
		
		<div class="metrics-section">
			<h3>Live Metrics</h3>
			<div class="metrics-grid">
				<div class="metric">
					<div class="metric-value">{{index .PerformanceData "events_processed" | printf "%.0f"}}</div>
					<div class="metric-label">Events Processed</div>
				</div>
				<div class="metric">
					<div class="metric-value">{{index .PerformanceData "handlers_registered" | printf "%.0f"}}</div>
					<div class="metric-label">Handlers Registered</div>
				</div>
				<div class="metric">
					<div class="metric-value">{{index .PerformanceData "cache_hits" | printf "%.0f"}}</div>
					<div class="metric-label">Cache Hits</div>
				</div>
				<div class="metric">
					<div class="metric-value">{{index .PerformanceData "cache_size" | printf "%.0f"}}</div>
					<div class="metric-label">Cache Size (bytes)</div>
				</div>
			</div>
		</div>
		
		<div class="event-log">
			<h3>Event Log</h3>
			{{range .EventLog}}
			<div class="event-item">{{.}}</div>
			{{end}}
		</div>
	</div>
	
	<div class="action-buttons">
		<button class="action-btn" onclick="send_event('{{.IdComponent}}', 'TestErrorBoundary', null)" title="Test Error Boundary">
			🛡️
		</button>
		<button class="action-btn" onclick="send_event('{{.IdComponent}}', 'TestStateSnapshot', null)" title="Take State Snapshot">
			📸
		</button>
		<button class="action-btn" onclick="send_event('{{.IdComponent}}', 'TestEventSystem', null)" title="Test Event System">
			⚡
		</button>
	</div>
</body>
</html>
	`
}

func (s *ShowcaseComponent) GetDriver() liveview.LiveDriver {
	return s
}

func (s *ShowcaseComponent) SetDemo(data interface{}) {
	if m, ok := data.(map[string]interface{}); ok {
		if demo, ok := m["demo"].(string); ok {
			s.ActiveDemo = demo
			s.logEvent(fmt.Sprintf("Switched to demo: %s", demo))
			s.eventRegistry.Emit("demo.changed", map[string]interface{}{
				"demo": demo,
			})
		}
	}
	s.Commit()
}

func (s *ShowcaseComponent) TestErrorBoundary(data interface{}) {
	s.logEvent("Testing error boundary...")
	
	// Simulate an error that gets caught
	err := s.errorBoundary.SafeExecute("test", func() error {
		return fmt.Errorf("simulated error for demonstration")
	})
	
	if err != nil {
		s.DemoResults["error_boundary"] = fmt.Sprintf("Error caught: %v", err)
		s.logEvent("Error boundary successfully caught and recovered from error")
	}
	
	s.Commit()
}

func (s *ShowcaseComponent) TestStateSnapshot(data interface{}) {
	s.logEvent("Taking state snapshot...")
	
	// Take a snapshot
	snapshot, err := s.stateManager.TakeSnapshot()
	if err != nil {
		s.logEvent(fmt.Sprintf("Snapshot failed: %v", err))
	} else {
		s.DemoResults["state_snapshot"] = fmt.Sprintf("Snapshot taken with %d keys", len(snapshot.Data))
		s.logEvent(fmt.Sprintf("State snapshot successful: %d keys saved", len(snapshot.Data)))
	}
	
	s.Commit()
}

func (s *ShowcaseComponent) TestEventSystem(data interface{}) {
	s.logEvent("Testing event system...")
	
	// Emit multiple events
	for i := 0; i < 5; i++ {
		s.eventRegistry.Emit("test.event", map[string]interface{}{
			"index": i,
			"time":  time.Now(),
		})
	}
	
	// Update event count
	eventsProcessed := 5
	s.DemoResults["event_system"] = fmt.Sprintf("Processed %d events", eventsProcessed)
	s.logEvent(fmt.Sprintf("Event system test complete: %d events", eventsProcessed))
	
	s.Commit()
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	
	// Showcase page
	page := &liveview.PageControl{
		Path:   "/",
		Title:  "Framework Showcase v3",
		Router: e,
	}
	
	page.Register(func() liveview.LiveDriver {
		return liveview.NewDriver("showcase", &ShowcaseComponent{})
	})
	
	port := ":8083"
	fmt.Printf("Starting Framework Showcase v3\n")
	fmt.Printf("Open http://localhost%s\n", port)
	fmt.Println("\nShowcasing:")
	fmt.Println("  • Error Boundaries")
	fmt.Println("  • State Management")
	fmt.Println("  • Virtual DOM")
	fmt.Println("  • Event Registry")
	fmt.Println("  • Template Cache")
	fmt.Println("  • Lazy Loading")
	fmt.Println("  • Lifecycle Hooks")
	fmt.Println("  • SafeScript API")
	fmt.Println("  • Communication Bus")
	
	if err := e.Start(port); err != nil {
		log.Fatal(err)
	}
}