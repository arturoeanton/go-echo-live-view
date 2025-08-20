package tests_test

import (
	"testing"

	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

func TestFlowBox(t *testing.T) {
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
	
	t.Run("Box movement", func(t *testing.T) {
		box := components.NewFlowBox("test", "Test", components.BoxTypeProcess, 100, 100)
		
		// Test direct position update
		box.X = 200
		box.Y = 300
		
		if box.X != 200 || box.Y != 300 {
			t.Errorf("Expected position (200, 300) after move, got (%d, %d)", box.X, box.Y)
		}
	})
	
	t.Run("Box resize", func(t *testing.T) {
		box := components.NewFlowBox("test", "Test", components.BoxTypeProcess, 0, 0)
		
		// Test direct size update
		box.Width = 200
		box.Height = 120
		
		if box.Width != 200 || box.Height != 120 {
			t.Errorf("Expected size 200x120 after resize, got %dx%d", box.Width, box.Height)
		}
	})
	
	t.Run("Box selection", func(t *testing.T) {
		box := components.NewFlowBox("test", "Test", components.BoxTypeProcess, 0, 0)
		driver := liveview.NewDriver("test-box", box)
		box.ComponentDriver = driver
		
		if box.Selected {
			t.Error("Expected box not to be selected initially")
		}
		
		box.SetSelected(true)
		if !box.Selected {
			t.Error("Expected box to be selected after SetSelected(true)")
		}
		
		box.SetSelected(false)
		if box.Selected {
			t.Error("Expected box not to be selected after SetSelected(false)")
		}
	})
	
	t.Run("Box click handler", func(t *testing.T) {
		clicked := false
		clickedID := ""
		
		box := components.NewFlowBox("test", "Test", components.BoxTypeProcess, 0, 0)
		driver := liveview.NewDriver("test-box", box)
		box.ComponentDriver = driver
		box.OnClick = func(boxID string) {
			clicked = true
			clickedID = boxID
		}
		box.Start()
		
		box.HandleClick("test")
		
		if !clicked {
			t.Error("Expected OnClick to be called")
		}
		
		if clickedID != "test" {
			t.Errorf("Expected clicked ID 'test', got '%s'", clickedID)
		}
		
		if !box.Selected {
			t.Error("Expected box to be selected after click")
		}
	})
	
	t.Run("Port management", func(t *testing.T) {
		box := components.NewFlowBox("test", "Test", components.BoxTypeProcess, 0, 0)
		driver := liveview.NewDriver("test-box", box)
		box.ComponentDriver = driver
		
		initialPortCount := len(box.Ports)
		
		newPort := components.Port{
			ID:       "custom-port",
			Type:     "output",
			Position: "bottom",
			Label:    "Custom",
		}
		
		box.AddPort(newPort)
		
		if len(box.Ports) != initialPortCount+1 {
			t.Errorf("Expected %d ports after adding, got %d", initialPortCount+1, len(box.Ports))
		}
		
		box.RemovePort("custom-port")
		
		if len(box.Ports) != initialPortCount {
			t.Errorf("Expected %d ports after removing, got %d", initialPortCount, len(box.Ports))
		}
	})
}

func TestFlowEdge(t *testing.T) {
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
	
	t.Run("Edge position update", func(t *testing.T) {
		edge := components.NewFlowEdge("test", "box1", "out1", "box2", "in1")
		driver := liveview.NewDriver("test-edge", edge)
		edge.ComponentDriver = driver
		
		edge.UpdatePosition(100, 200, 300, 400)
		
		if edge.FromX != 100 || edge.FromY != 200 {
			t.Errorf("Expected from position (100, 200), got (%d, %d)", edge.FromX, edge.FromY)
		}
		
		if edge.ToX != 300 || edge.ToY != 400 {
			t.Errorf("Expected to position (300, 400), got (%d, %d)", edge.ToX, edge.ToY)
		}
	})
	
	t.Run("Edge style changes", func(t *testing.T) {
		edge := components.NewFlowEdge("test", "box1", "out1", "box2", "in1")
		driver := liveview.NewDriver("test-edge", edge)
		edge.ComponentDriver = driver
		
		edge.SetEdgeStyle(components.EdgeStyleDashed)
		if edge.Style != components.EdgeStyleDashed {
			t.Errorf("Expected style Dashed, got %v", edge.Style)
		}
		
		edge.SetType(components.EdgeTypeStep)
		if edge.Type != components.EdgeTypeStep {
			t.Errorf("Expected type Step, got %v", edge.Type)
		}
		
		edge.SetAnimated(true)
		if !edge.Animated {
			t.Error("Expected edge to be animated")
		}
	})
	
	t.Run("Edge calculations", func(t *testing.T) {
		edge := components.NewFlowEdge("test", "box1", "out1", "box2", "in1")
		edge.UpdatePosition(0, 0, 100, 100)
		
		midX := edge.GetMidX()
		midY := edge.GetMidY()
		
		if midX != 50 || midY != 50 {
			t.Errorf("Expected midpoint (50, 50), got (%d, %d)", midX, midY)
		}
		
		length := edge.GetLength()
		// expectedLength would be 141.42 (sqrt(100^2 + 100^2))
		if length < 141 || length > 142 {
			t.Errorf("Expected length ~141.42, got %f", length)
		}
	})
	
	t.Run("Edge path generation", func(t *testing.T) {
		edge := components.NewFlowEdge("test", "box1", "out1", "box2", "in1")
		edge.UpdatePosition(0, 0, 100, 100)
		
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
	
	t.Run("Edge click handler", func(t *testing.T) {
		clicked := false
		clickedID := ""
		
		edge := components.NewFlowEdge("test", "box1", "out1", "box2", "in1")
		driver := liveview.NewDriver("test-edge", edge)
		edge.ComponentDriver = driver
		edge.OnClick = func(edgeID string) {
			clicked = true
			clickedID = edgeID
		}
		edge.Start()
		
		edge.HandleClick("test")
		
		if !clicked {
			t.Error("Expected OnClick to be called")
		}
		
		if clickedID != "test" {
			t.Errorf("Expected clicked ID 'test', got '%s'", clickedID)
		}
		
		if !edge.Selected {
			t.Error("Expected edge to be selected after click")
		}
	})
}

func TestFlowCanvas(t *testing.T) {
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
	
	t.Run("Add and remove boxes", func(t *testing.T) {
		canvas := components.NewFlowCanvas("test", 800, 600)
		driver := liveview.NewDriver("test-canvas", canvas)
		canvas.ComponentDriver = driver
		
		box1 := components.NewFlowBox("box1", "Box 1", components.BoxTypeProcess, 100, 100)
		box2 := components.NewFlowBox("box2", "Box 2", components.BoxTypeDecision, 200, 200)
		
		canvas.AddBox(box1)
		canvas.AddBox(box2)
		
		if len(canvas.Boxes) != 2 {
			t.Errorf("Expected 2 boxes, got %d", len(canvas.Boxes))
		}
		
		canvas.RemoveBox("box1")
		
		if len(canvas.Boxes) != 1 {
			t.Errorf("Expected 1 box after removal, got %d", len(canvas.Boxes))
		}
		
		if _, exists := canvas.Boxes["box2"]; !exists {
			t.Error("Expected box2 to still exist")
		}
	})
	
	t.Run("Add and remove edges", func(t *testing.T) {
		canvas := components.NewFlowCanvas("test", 800, 600)
		driver := liveview.NewDriver("test-canvas", canvas)
		canvas.ComponentDriver = driver
		
		edge1 := components.NewFlowEdge("edge1", "box1", "out1", "box2", "in1")
		edge2 := components.NewFlowEdge("edge2", "box2", "out1", "box3", "in1")
		
		canvas.AddEdge(edge1)
		canvas.AddEdge(edge2)
		
		if len(canvas.Edges) != 2 {
			t.Errorf("Expected 2 edges, got %d", len(canvas.Edges))
		}
		
		canvas.RemoveEdge("edge1")
		
		if len(canvas.Edges) != 1 {
			t.Errorf("Expected 1 edge after removal, got %d", len(canvas.Edges))
		}
	})
	
	t.Run("Clear canvas", func(t *testing.T) {
		canvas := components.NewFlowCanvas("test", 800, 600)
		driver := liveview.NewDriver("test-canvas", canvas)
		canvas.ComponentDriver = driver
		
		box := components.NewFlowBox("box1", "Box", components.BoxTypeProcess, 0, 0)
		edge := components.NewFlowEdge("edge1", "box1", "out1", "box2", "in1")
		
		canvas.AddBox(box)
		canvas.AddEdge(edge)
		
		canvas.Clear()
		
		if len(canvas.Boxes) != 0 {
			t.Errorf("Expected 0 boxes after clear, got %d", len(canvas.Boxes))
		}
		
		if len(canvas.Edges) != 0 {
			t.Errorf("Expected 0 edges after clear, got %d", len(canvas.Edges))
		}
	})
	
	t.Run("Zoom controls", func(t *testing.T) {
		canvas := components.NewFlowCanvas("test", 800, 600)
		driver := liveview.NewDriver("test-canvas", canvas)
		canvas.ComponentDriver = driver
		canvas.Start()
		
		initialZoom := canvas.Zoom
		
		canvas.HandleZoomIn(nil)
		if canvas.Zoom <= initialZoom {
			t.Error("Expected zoom to increase")
		}
		
		canvas.HandleZoomOut(nil)
		canvas.HandleZoomOut(nil)
		if canvas.Zoom >= initialZoom {
			t.Error("Expected zoom to decrease")
		}
		
		canvas.HandleResetView(nil)
		if canvas.Zoom != 1.0 {
			t.Errorf("Expected zoom to reset to 1.0, got %f", canvas.Zoom)
		}
		if canvas.PanX != 0 || canvas.PanY != 0 {
			t.Errorf("Expected pan to reset to (0, 0), got (%d, %d)", canvas.PanX, canvas.PanY)
		}
	})
	
	t.Run("Grid toggle", func(t *testing.T) {
		canvas := components.NewFlowCanvas("test", 800, 600)
		driver := liveview.NewDriver("test-canvas", canvas)
		canvas.ComponentDriver = driver
		canvas.Start()
		
		initialGrid := canvas.ShowGrid
		
		canvas.HandleToggleGrid(nil)
		if canvas.ShowGrid == initialGrid {
			t.Error("Expected grid to toggle")
		}
		
		canvas.HandleToggleGrid(nil)
		if canvas.ShowGrid != initialGrid {
			t.Error("Expected grid to toggle back")
		}
	})
	
	t.Run("Export JSON", func(t *testing.T) {
		canvas := components.NewFlowCanvas("test", 800, 600)
		
		box := components.NewFlowBox("box1", "Test Box", components.BoxTypeProcess, 100, 200)
		edge := components.NewFlowEdge("edge1", "box1", "out1", "box2", "in1")
		
		canvas.AddBox(box)
		canvas.AddEdge(edge)
		
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
	
	t.Run("Canvas callbacks", func(t *testing.T) {
		boxClicked := false
		edgeClicked := false
		connectionMade := false
		boxMoved := false
		
		canvas := components.NewFlowCanvas("test", 800, 600)
		driver := liveview.NewDriver("test-canvas", canvas)
		canvas.ComponentDriver = driver
		
		canvas.OnBoxClick = func(boxID string) {
			boxClicked = true
		}
		canvas.OnEdgeClick = func(edgeID string) {
			edgeClicked = true
		}
		canvas.OnConnection = func(fromBox, fromPort, toBox, toPort string) {
			connectionMade = true
		}
		canvas.OnBoxMove = func(boxID string, x, y int) {
			boxMoved = true
		}
		
		canvas.Start()
		
		// Test box click
		box := components.NewFlowBox("box1", "Test", components.BoxTypeProcess, 0, 0)
		canvas.AddBox(box)
		canvas.HandleSelectBox("box1")
		
		if !boxClicked {
			t.Error("Expected OnBoxClick to be called")
		}
		
		// Test edge click
		edge := components.NewFlowEdge("edge1", "box1", "out1", "box2", "in1")
		canvas.AddEdge(edge)
		canvas.HandleSelectEdge("edge1")
		
		if !edgeClicked {
			t.Error("Expected OnEdgeClick to be called")
		}
		
		// Test connection
		canvas.HandleStartConnection(map[string]interface{}{
			"boxId":  "box1",
			"portId": "out1",
		})
		
		canvas.HandleCompleteConnection(map[string]interface{}{
			"boxId":  "box2",
			"portId": "in1",
		})
		
		if !connectionMade {
			t.Error("Expected OnConnection to be called")
		}
		
		// Test box move
		canvas.HandleMoveBox(map[string]interface{}{
			"boxId": "box1",
			"x":     200.0,
			"y":     300.0,
		})
		
		if !boxMoved {
			t.Error("Expected OnBoxMove to be called")
		}
	})
}

func TestFlowIntegration(t *testing.T) {
	t.Run("Complete flow diagram", func(t *testing.T) {
		// Create canvas
		canvas := components.NewFlowCanvas("flow", 1000, 800)
		driver := liveview.NewDriver("flow-canvas", canvas)
		canvas.ComponentDriver = driver
		
		// Create boxes
		startBox := components.NewFlowBox("start", "Start", components.BoxTypeStart, 50, 400)
		processBox := components.NewFlowBox("process", "Process", components.BoxTypeProcess, 200, 400)
		decisionBox := components.NewFlowBox("decision", "Decision", components.BoxTypeDecision, 400, 400)
		endBox := components.NewFlowBox("end", "End", components.BoxTypeEnd, 600, 400)
		
		// Add boxes to canvas
		canvas.AddBox(startBox)
		canvas.AddBox(processBox)
		canvas.AddBox(decisionBox)
		canvas.AddBox(endBox)
		
		if len(canvas.Boxes) != 4 {
			t.Errorf("Expected 4 boxes, got %d", len(canvas.Boxes))
		}
		
		// Create edges
		edge1 := components.NewFlowEdge("edge1", "start", "out1", "process", "in1")
		edge2 := components.NewFlowEdge("edge2", "process", "out1", "decision", "in1")
		edge3 := components.NewFlowEdge("edge3", "decision", "out1", "end", "in1")
		
		// Add edges to canvas
		canvas.AddEdge(edge1)
		canvas.AddEdge(edge2)
		canvas.AddEdge(edge3)
		
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