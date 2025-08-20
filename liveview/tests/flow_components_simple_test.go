package tests_test

import (
	"testing"

	"github.com/arturoeanton/go-echo-live-view/components"
)

func TestFlowBoxSimple(t *testing.T) {
	t.Run("Create FlowBox", func(t *testing.T) {
		box := components.NewFlowBox("test-box", "Test Node", components.BoxTypeProcess, 100, 200)
		
		if box.ID != "test-box" {
			t.Errorf("Expected ID 'test-box', got '%s'", box.ID)
		}
		
		if box.Label != "Test Node" {
			t.Errorf("Expected label 'Test Node', got '%s'", box.Label)
		}
		
		if box.Type != components.BoxTypeProcess {
			t.Errorf("Expected type Process, got %v", box.Type)
		}
		
		if box.X != 100 || box.Y != 200 {
			t.Errorf("Expected position (100, 200), got (%d, %d)", box.X, box.Y)
		}
		
		if box.Width != 150 || box.Height != 80 {
			t.Errorf("Expected size 150x80, got %dx%d", box.Width, box.Height)
		}
	})
	
	t.Run("Decision box has different height", func(t *testing.T) {
		box := components.NewFlowBox("decision", "Decision", components.BoxTypeDecision, 0, 0)
		
		if box.Height != 100 {
			t.Errorf("Expected decision box height 100, got %d", box.Height)
		}
	})
	
	t.Run("Box colors by type", func(t *testing.T) {
		tests := []struct {
			boxType components.BoxType
			color   string
		}{
			{components.BoxTypeStart, "#dcfce7"},
			{components.BoxTypeEnd, "#fee2e2"},
			{components.BoxTypeProcess, "#dbeafe"},
			{components.BoxTypeDecision, "#fef3c7"},
			{components.BoxTypeData, "#e9d5ff"},
			{components.BoxTypeCustom, "#f3f4f6"},
		}
		
		for _, tt := range tests {
			box := components.NewFlowBox("test", "Test", tt.boxType, 0, 0)
			if box.Color != tt.color {
				t.Errorf("Expected color %s for type %v, got %s", tt.color, tt.boxType, box.Color)
			}
		}
	})
	
	t.Run("Box ports generation", func(t *testing.T) {
		tests := []struct {
			boxType   components.BoxType
			portCount int
		}{
			{components.BoxTypeStart, 1},    // Only output
			{components.BoxTypeEnd, 1},      // Only input
			{components.BoxTypeProcess, 2},  // Input and output
			{components.BoxTypeDecision, 3}, // 1 input, 2 outputs
			{components.BoxTypeData, 2},     // Top input, bottom output
		}
		
		for _, tt := range tests {
			box := components.NewFlowBox("test", "Test", tt.boxType, 0, 0)
			if len(box.Ports) != tt.portCount {
				t.Errorf("Expected %d ports for type %v, got %d", tt.portCount, tt.boxType, len(box.Ports))
			}
		}
	})
	
	t.Run("Box position and size", func(t *testing.T) {
		box := components.NewFlowBox("test", "Test", components.BoxTypeProcess, 100, 100)
		
		// Test position
		box.X = 200
		box.Y = 300
		
		if box.X != 200 || box.Y != 300 {
			t.Errorf("Expected position (200, 300), got (%d, %d)", box.X, box.Y)
		}
		
		// Test size
		box.Width = 250
		box.Height = 150
		
		if box.Width != 250 || box.Height != 150 {
			t.Errorf("Expected size 250x150, got %dx%d", box.Width, box.Height)
		}
	})
	
	t.Run("Box selection state", func(t *testing.T) {
		box := components.NewFlowBox("test", "Test", components.BoxTypeProcess, 0, 0)
		
		if box.Selected {
			t.Error("Expected box not to be selected initially")
		}
		
		box.Selected = true
		if !box.Selected {
			t.Error("Expected box to be selected")
		}
	})
}

func TestFlowEdgeSimple(t *testing.T) {
	t.Run("Create FlowEdge", func(t *testing.T) {
		edge := components.NewFlowEdge("test-edge", "box1", "out1", "box2", "in1")
		
		if edge.ID != "test-edge" {
			t.Errorf("Expected ID 'test-edge', got '%s'", edge.ID)
		}
		
		if edge.FromBox != "box1" || edge.ToBox != "box2" {
			t.Errorf("Expected connection box1->box2, got %s->%s", edge.FromBox, edge.ToBox)
		}
		
		if edge.Type != components.EdgeTypeCurved {
			t.Errorf("Expected default type Curved, got %v", edge.Type)
		}
		
		if edge.Style != components.EdgeStyleSolid {
			t.Errorf("Expected default style Solid, got %v", edge.Style)
		}
		
		if !edge.ArrowHead {
			t.Error("Expected ArrowHead to be true by default")
		}
	})
	
	t.Run("Edge position", func(t *testing.T) {
		edge := components.NewFlowEdge("test", "box1", "out1", "box2", "in1")
		
		edge.FromX = 100
		edge.FromY = 200
		edge.ToX = 300
		edge.ToY = 400
		
		if edge.FromX != 100 || edge.FromY != 200 {
			t.Errorf("Expected from position (100, 200), got (%d, %d)", edge.FromX, edge.FromY)
		}
		
		if edge.ToX != 300 || edge.ToY != 400 {
			t.Errorf("Expected to position (300, 400), got (%d, %d)", edge.ToX, edge.ToY)
		}
	})
	
	t.Run("Edge styles", func(t *testing.T) {
		edge := components.NewFlowEdge("test", "box1", "out1", "box2", "in1")
		
		edge.Style = components.EdgeStyleDashed
		if edge.Style != components.EdgeStyleDashed {
			t.Errorf("Expected style Dashed, got %v", edge.Style)
		}
		
		edge.Type = components.EdgeTypeStep
		if edge.Type != components.EdgeTypeStep {
			t.Errorf("Expected type Step, got %v", edge.Type)
		}
		
		edge.Animated = true
		if !edge.Animated {
			t.Error("Expected edge to be animated")
		}
	})
	
	t.Run("Edge calculations", func(t *testing.T) {
		edge := components.NewFlowEdge("test", "box1", "out1", "box2", "in1")
		edge.FromX = 0
		edge.FromY = 0
		edge.ToX = 100
		edge.ToY = 100
		
		midX := edge.GetMidX()
		midY := edge.GetMidY()
		
		if midX != 50 || midY != 50 {
			t.Errorf("Expected midpoint (50, 50), got (%d, %d)", midX, midY)
		}
		
		length := edge.GetLength()
		// Expected length is approximately 141.42 (sqrt(100^2 + 100^2))
		if length < 141 || length > 142 {
			t.Errorf("Expected length ~141.42, got %f", length)
		}
	})
	
	t.Run("Edge path generation", func(t *testing.T) {
		edge := components.NewFlowEdge("test", "box1", "out1", "box2", "in1")
		edge.FromX = 0
		edge.FromY = 0
		edge.ToX = 100
		edge.ToY = 100
		
		// Test curved path
		curvedPath := edge.CalculateCurvedPath()
		if curvedPath == "" {
			t.Error("Expected non-empty curved path")
		}
		
		// Test step path
		stepPath := edge.CalculateStepPath()
		if stepPath == "" {
			t.Error("Expected non-empty step path")
		}
		
		// Test bezier path
		bezierPath := edge.CalculateBezierPath()
		if bezierPath == "" {
			t.Error("Expected non-empty bezier path")
		}
	})
}

func TestFlowCanvasSimple(t *testing.T) {
	t.Run("Create FlowCanvas", func(t *testing.T) {
		canvas := components.NewFlowCanvas("test-canvas", 800, 600)
		
		if canvas.ID != "test-canvas" {
			t.Errorf("Expected ID 'test-canvas', got '%s'", canvas.ID)
		}
		
		if canvas.Width != 800 || canvas.Height != 600 {
			t.Errorf("Expected size 800x600, got %dx%d", canvas.Width, canvas.Height)
		}
		
		if canvas.Zoom != 1.0 {
			t.Errorf("Expected default zoom 1.0, got %f", canvas.Zoom)
		}
		
		if !canvas.ShowGrid {
			t.Error("Expected grid to be shown by default")
		}
		
		if canvas.GridSize != components.GridMedium {
			t.Errorf("Expected default grid size Medium, got %v", canvas.GridSize)
		}
	})
	
	t.Run("Canvas box management", func(t *testing.T) {
		canvas := components.NewFlowCanvas("test", 800, 600)
		
		box1 := components.NewFlowBox("box1", "Box 1", components.BoxTypeProcess, 100, 100)
		box2 := components.NewFlowBox("box2", "Box 2", components.BoxTypeDecision, 200, 200)
		
		canvas.Boxes[box1.ID] = box1
		canvas.Boxes[box2.ID] = box2
		
		if len(canvas.Boxes) != 2 {
			t.Errorf("Expected 2 boxes, got %d", len(canvas.Boxes))
		}
		
		delete(canvas.Boxes, "box1")
		
		if len(canvas.Boxes) != 1 {
			t.Errorf("Expected 1 box after removal, got %d", len(canvas.Boxes))
		}
		
		if _, exists := canvas.Boxes["box2"]; !exists {
			t.Error("Expected box2 to still exist")
		}
	})
	
	t.Run("Canvas edge management", func(t *testing.T) {
		canvas := components.NewFlowCanvas("test", 800, 600)
		
		edge1 := components.NewFlowEdge("edge1", "box1", "out1", "box2", "in1")
		edge2 := components.NewFlowEdge("edge2", "box2", "out1", "box3", "in1")
		
		canvas.Edges[edge1.ID] = edge1
		canvas.Edges[edge2.ID] = edge2
		
		if len(canvas.Edges) != 2 {
			t.Errorf("Expected 2 edges, got %d", len(canvas.Edges))
		}
		
		delete(canvas.Edges, "edge1")
		
		if len(canvas.Edges) != 1 {
			t.Errorf("Expected 1 edge after removal, got %d", len(canvas.Edges))
		}
	})
	
	t.Run("Canvas zoom", func(t *testing.T) {
		canvas := components.NewFlowCanvas("test", 800, 600)
		
		canvas.Zoom = 1.5
		if canvas.Zoom != 1.5 {
			t.Errorf("Expected zoom 1.5, got %f", canvas.Zoom)
		}
		
		percent := canvas.ZoomPercent()
		if percent != 150 {
			t.Errorf("Expected zoom percent 150, got %d", percent)
		}
	})
	
	t.Run("Canvas export", func(t *testing.T) {
		canvas := components.NewFlowCanvas("test", 800, 600)
		
		box := components.NewFlowBox("box1", "Test Box", components.BoxTypeProcess, 100, 200)
		edge := components.NewFlowEdge("edge1", "box1", "out1", "box2", "in1")
		
		canvas.Boxes[box.ID] = box
		canvas.Edges[edge.ID] = edge
		
		exported := canvas.ExportJSON()
		
		boxes, ok := exported["boxes"].([]map[string]interface{})
		if !ok || len(boxes) != 1 {
			t.Error("Expected 1 box in export")
		}
		
		edges, ok := exported["edges"].([]map[string]interface{})
		if !ok || len(edges) != 1 {
			t.Error("Expected 1 edge in export")
		}
	})
}

func TestFlowIntegrationSimple(t *testing.T) {
	t.Run("Complete flow diagram", func(t *testing.T) {
		// Create canvas
		canvas := components.NewFlowCanvas("flow", 1000, 800)
		
		// Create boxes
		startBox := components.NewFlowBox("start", "Start", components.BoxTypeStart, 50, 400)
		processBox := components.NewFlowBox("process", "Process", components.BoxTypeProcess, 200, 400)
		decisionBox := components.NewFlowBox("decision", "Decision", components.BoxTypeDecision, 400, 400)
		endBox := components.NewFlowBox("end", "End", components.BoxTypeEnd, 600, 400)
		
		// Add boxes to canvas
		canvas.Boxes[startBox.ID] = startBox
		canvas.Boxes[processBox.ID] = processBox
		canvas.Boxes[decisionBox.ID] = decisionBox
		canvas.Boxes[endBox.ID] = endBox
		
		if len(canvas.Boxes) != 4 {
			t.Errorf("Expected 4 boxes, got %d", len(canvas.Boxes))
		}
		
		// Create edges
		edge1 := components.NewFlowEdge("edge1", "start", "out1", "process", "in1")
		edge2 := components.NewFlowEdge("edge2", "process", "out1", "decision", "in1")
		edge3 := components.NewFlowEdge("edge3", "decision", "out1", "end", "in1")
		
		// Add edges to canvas
		canvas.Edges[edge1.ID] = edge1
		canvas.Edges[edge2.ID] = edge2
		canvas.Edges[edge3.ID] = edge3
		
		if len(canvas.Edges) != 3 {
			t.Errorf("Expected 3 edges, got %d", len(canvas.Edges))
		}
		
		// Test export
		exported := canvas.ExportJSON()
		
		boxes := exported["boxes"].([]map[string]interface{})
		edges := exported["edges"].([]map[string]interface{})
		
		if len(boxes) != 4 {
			t.Errorf("Expected 4 boxes in export, got %d", len(boxes))
		}
		
		if len(edges) != 3 {
			t.Errorf("Expected 3 edges in export, got %d", len(edges))
		}
	})
}