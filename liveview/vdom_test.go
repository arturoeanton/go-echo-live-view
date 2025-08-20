package liveview

import (
	"fmt"
	"strings"
	"testing"
)

func TestCreateElement(t *testing.T) {
	props := map[string]interface{}{
		"class": "button",
		"id":    "submit-btn",
	}
	
	child1 := CreateText("Click me")
	element := CreateElement("button", props, child1)
	
	if element.Type != ElementNode {
		t.Errorf("Expected ElementNode, got %v", element.Type)
	}
	
	if element.Tag != "button" {
		t.Errorf("Expected tag 'button', got %s", element.Tag)
	}
	
	if len(element.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(element.Children))
	}
	
	if element.Props["class"] != "button" {
		t.Errorf("Expected class 'button', got %v", element.Props["class"])
	}
	
	if element.Hash == "" {
		t.Error("Expected hash to be calculated")
	}
}

func TestCreateElementWithKey(t *testing.T) {
	props := map[string]interface{}{
		"key":   "unique-key",
		"class": "item",
	}
	
	element := CreateElement("div", props, nil)
	
	if element.Key != "unique-key" {
		t.Errorf("Expected key 'unique-key', got %s", element.Key)
	}
	
	// Key should be removed from props
	if _, exists := props["key"]; exists {
		t.Error("Key should be removed from props")
	}
}

func TestCreateText(t *testing.T) {
	text := CreateText("Hello World")
	
	if text.Type != TextNode {
		t.Errorf("Expected TextNode, got %v", text.Type)
	}
	
	if text.Text != "Hello World" {
		t.Errorf("Expected text 'Hello World', got %s", text.Text)
	}
	
	if text.Hash == "" {
		t.Error("Expected hash to be calculated")
	}
}

func TestCreateFragment(t *testing.T) {
	child1 := CreateText("First")
	child2 := CreateText("Second")
	
	fragment := CreateFragment(child1, child2)
	
	if fragment.Type != FragmentNode {
		t.Errorf("Expected FragmentNode, got %v", fragment.Type)
	}
	
	if len(fragment.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(fragment.Children))
	}
}

func TestVDomInitialRender(t *testing.T) {
	vdom := NewVDom(nil)
	
	// Create a simple tree
	root := CreateElement("div", nil,
		CreateElement("h1", nil, CreateText("Title")),
		CreateElement("p", nil, CreateText("Content")),
	)
	
	patches := vdom.Render(root)
	
	// Should have one create patch for initial render
	if len(patches) != 1 {
		t.Errorf("Expected 1 patch for initial render, got %d", len(patches))
	}
	
	if patches[0].Type != PatchCreate {
		t.Errorf("Expected PatchCreate, got %v", patches[0].Type)
	}
}

func TestVDomTextUpdate(t *testing.T) {
	vdom := NewVDom(nil)
	
	// Initial render
	old := CreateElement("div", nil, CreateText("Old text"))
	vdom.Render(old)
	
	// Update render
	new := CreateElement("div", nil, CreateText("New text"))
	patches := vdom.Render(new)
	
	// Should have text update patch
	foundTextUpdate := false
	for _, patch := range patches {
		if patch.Type == PatchUpdateText {
			foundTextUpdate = true
			if patch.Value != "New text" {
				t.Errorf("Expected new text 'New text', got %v", patch.Value)
			}
			if patch.OldValue != "Old text" {
				t.Errorf("Expected old text 'Old text', got %v", patch.OldValue)
			}
		}
	}
	
	if !foundTextUpdate {
		t.Error("Expected text update patch")
	}
}

func TestVDomPropUpdate(t *testing.T) {
	vdom := NewVDom(nil)
	
	// Initial render
	oldProps := map[string]interface{}{"class": "old"}
	old := CreateElement("div", oldProps, nil)
	vdom.Render(old)
	
	// Update render
	newProps := map[string]interface{}{"class": "new", "id": "test"}
	new := CreateElement("div", newProps, nil)
	patches := vdom.Render(new)
	
	// Should have prop update patch
	foundPropUpdate := false
	for _, patch := range patches {
		if patch.Type == PatchUpdateProps {
			foundPropUpdate = true
			if patch.Props["class"] != "new" {
				t.Errorf("Expected class 'new', got %v", patch.Props["class"])
			}
			if patch.Props["id"] != "test" {
				t.Errorf("Expected id 'test', got %v", patch.Props["id"])
			}
		}
	}
	
	if !foundPropUpdate {
		t.Error("Expected prop update patch")
	}
}

func TestVDomNodeReplace(t *testing.T) {
	vdom := NewVDom(nil)
	
	// Initial render
	old := CreateElement("div", nil, CreateText("Content"))
	vdom.Render(old)
	
	// Replace with different element
	new := CreateElement("span", nil, CreateText("Content"))
	patches := vdom.Render(new)
	
	// Should have replace patch
	foundReplace := false
	for _, patch := range patches {
		if patch.Type == PatchReplace {
			foundReplace = true
			oldNode := patch.OldValue.(*VNode)
			newNode := patch.Value.(*VNode)
			
			if oldNode.Tag != "div" {
				t.Errorf("Expected old tag 'div', got %s", oldNode.Tag)
			}
			if newNode.Tag != "span" {
				t.Errorf("Expected new tag 'span', got %s", newNode.Tag)
			}
		}
	}
	
	if !foundReplace {
		t.Error("Expected replace patch")
	}
}

func TestVDomChildrenAddRemove(t *testing.T) {
	vdom := NewVDom(&VDomConfig{
		MinimizePatches: false,
	})
	
	// Initial render with 2 children
	old := CreateElement("div", nil,
		CreateText("Child 1"),
		CreateText("Child 2"),
	)
	vdom.Render(old)
	
	// Update with 3 children
	new := CreateElement("div", nil,
		CreateText("Child 1"),
		CreateText("Child 2"),
		CreateText("Child 3"),
	)
	patches := vdom.Render(new)
	
	// Should have create patch for new child
	foundCreate := false
	for _, patch := range patches {
		if patch.Type == PatchCreate {
			foundCreate = true
			node := patch.Value.(*VNode)
			if node.Text != "Child 3" {
				t.Errorf("Expected 'Child 3', got %s", node.Text)
			}
		}
	}
	
	if !foundCreate {
		t.Error("Expected create patch for new child")
	}
	
	// Now remove children
	new2 := CreateElement("div", nil,
		CreateText("Child 1"),
	)
	patches2 := vdom.Render(new2)
	
	// Should have remove patches
	removeCount := 0
	for _, patch := range patches2 {
		if patch.Type == PatchRemove {
			removeCount++
		}
		// Debug: print all patches
		t.Logf("Patch: Type=%v, Path=%s", patch.Type, patch.Path)
	}
	
	// We expect 2 children to be removed (Child 2 and Child 3)
	if removeCount < 2 {
		t.Errorf("Expected at least 2 remove patches, got %d", removeCount)
	}
}

func TestVDomKeyedChildren(t *testing.T) {
	vdom := NewVDom(&VDomConfig{EnableKeys: true})
	
	// Initial render with keyed children
	old := CreateElement("ul", nil,
		CreateElement("li", map[string]interface{}{"key": "a"}, CreateText("Item A")),
		CreateElement("li", map[string]interface{}{"key": "b"}, CreateText("Item B")),
		CreateElement("li", map[string]interface{}{"key": "c"}, CreateText("Item C")),
	)
	vdom.Render(old)
	
	// Reorder children
	new := CreateElement("ul", nil,
		CreateElement("li", map[string]interface{}{"key": "c"}, CreateText("Item C")),
		CreateElement("li", map[string]interface{}{"key": "a"}, CreateText("Item A")),
		CreateElement("li", map[string]interface{}{"key": "b"}, CreateText("Item B")),
	)
	patches := vdom.Render(new)
	
	// Should have move patches
	moveCount := 0
	for _, patch := range patches {
		if patch.Type == PatchMove {
			moveCount++
		}
	}
	
	if moveCount == 0 {
		t.Error("Expected move patches for reordered keyed children")
	}
}

func TestVDomKeyedChildrenAddRemove(t *testing.T) {
	vdom := NewVDom(&VDomConfig{EnableKeys: true})
	
	// Initial render
	old := CreateElement("ul", nil,
		CreateElement("li", map[string]interface{}{"key": "a"}, CreateText("Item A")),
		CreateElement("li", map[string]interface{}{"key": "b"}, CreateText("Item B")),
	)
	vdom.Render(old)
	
	// Add and remove children
	new := CreateElement("ul", nil,
		CreateElement("li", map[string]interface{}{"key": "b"}, CreateText("Item B")),
		CreateElement("li", map[string]interface{}{"key": "c"}, CreateText("Item C")),
	)
	patches := vdom.Render(new)
	
	// Should have remove patch for "a" and create patch for "c"
	foundRemove := false
	foundCreate := false
	
	for _, patch := range patches {
		if patch.Type == PatchRemove {
			foundRemove = true
		}
		if patch.Type == PatchCreate {
			foundCreate = true
		}
	}
	
	if !foundRemove {
		t.Error("Expected remove patch for removed keyed child")
	}
	
	if !foundCreate {
		t.Error("Expected create patch for new keyed child")
	}
}

func TestVDomToHTML(t *testing.T) {
	vdom := NewVDom(nil)
	
	// Create a tree
	node := CreateElement("div", map[string]interface{}{"class": "container"},
		CreateElement("h1", nil, CreateText("Title")),
		CreateElement("p", map[string]interface{}{"id": "content"},
			CreateText("This is "),
			CreateElement("strong", nil, CreateText("important")),
			CreateText(" text."),
		),
	)
	
	html := vdom.ToHTML(node)
	
	// Check generated HTML
	expected := `<div class="container"><h1>Title</h1><p id="content">This is <strong>important</strong> text.</p></div>`
	
	if html != expected {
		t.Errorf("Expected HTML:\n%s\nGot:\n%s", expected, html)
	}
}

func TestVDomSelfClosingTags(t *testing.T) {
	vdom := NewVDom(nil)
	
	// Create self-closing elements
	node := CreateElement("div", nil,
		CreateElement("img", map[string]interface{}{"src": "image.jpg", "alt": "Test"}, nil),
		CreateElement("br", nil, nil),
		CreateElement("input", map[string]interface{}{"type": "text", "value": "test"}, nil),
	)
	
	html := vdom.ToHTML(node)
	
	// Check self-closing tags
	if !strings.Contains(html, `<img src="image.jpg" alt="Test" />`) {
		t.Error("img tag should be self-closing")
	}
	
	if !strings.Contains(html, `<br />`) {
		t.Error("br tag should be self-closing")
	}
	
	if !strings.Contains(html, `<input type="text" value="test" />`) {
		t.Error("input tag should be self-closing")
	}
}

func TestVDomHTMLEscaping(t *testing.T) {
	vdom := NewVDom(nil)
	
	// Create node with special characters
	node := CreateElement("div", map[string]interface{}{"title": `"Special" & <chars>`},
		CreateText(`Text with <html> & "quotes"`),
	)
	
	html := vdom.ToHTML(node)
	
	// Check escaping
	if !strings.Contains(html, `title="&#34;Special&#34; &amp; &lt;chars&gt;"`) {
		t.Error("Attributes should be escaped")
	}
	
	if !strings.Contains(html, `Text with &lt;html&gt; &amp; &#34;quotes&#34;`) {
		t.Error("Text content should be escaped")
	}
}

func TestVDomOptimizePatches(t *testing.T) {
	vdom := NewVDom(&VDomConfig{MinimizePatches: true})
	
	// Create initial tree
	old := CreateElement("div", map[string]interface{}{"class": "old"},
		CreateText("Old text"),
	)
	vdom.Render(old)
	
	// Multiple updates
	new := CreateElement("div", map[string]interface{}{"class": "new", "id": "test"},
		CreateText("New text"),
	)
	patches := vdom.Render(new)
	
	// Patches should be optimized/combined
	if len(patches) > 2 {
		t.Errorf("Expected optimized patches, got %d patches", len(patches))
	}
}

func TestParseHTML(t *testing.T) {
	tests := []struct {
		html     string
		expected string
	}{
		{
			html:     "Plain text",
			expected: "Plain text",
		},
		{
			html:     `<div class="test">Content</div>`,
			expected: "div",
		},
		{
			html:     `<img src="test.jpg" alt="Test" />`,
			expected: "img",
		},
	}
	
	for _, test := range tests {
		node, err := ParseHTML(test.html)
		if err != nil {
			t.Errorf("ParseHTML error for %s: %v", test.html, err)
			continue
		}
		
		if node == nil {
			t.Errorf("ParseHTML returned nil for %s", test.html)
			continue
		}
		
		if node.Type == TextNode {
			if node.Text != test.expected {
				t.Errorf("Expected text %s, got %s", test.expected, node.Text)
			}
		} else if node.Type == ElementNode {
			if node.Tag != test.expected {
				t.Errorf("Expected tag %s, got %s", test.expected, node.Tag)
			}
		}
	}
}

func TestVDomRenderer(t *testing.T) {
	// Create a test component
	component := &TestComponent{}
	renderer := NewVDomRenderer(component)
	
	// Initial render
	patches, err := renderer.Render()
	if err != nil {
		t.Errorf("Render error: %v", err)
	}
	
	if len(patches) == 0 {
		t.Error("Expected patches from initial render")
	}
	
	// Get HTML
	html := renderer.GetHTML()
	if html == "" {
		t.Error("Expected HTML output")
	}
}

func TestVDomComplexDiff(t *testing.T) {
	vdom := NewVDom(nil)
	
	// Complex initial tree
	old := CreateElement("div", map[string]interface{}{"class": "app"},
		CreateElement("header", nil,
			CreateElement("h1", nil, CreateText("App Title")),
			CreateElement("nav", nil,
				CreateElement("a", map[string]interface{}{"href": "/home"}, CreateText("Home")),
				CreateElement("a", map[string]interface{}{"href": "/about"}, CreateText("About")),
			),
		),
		CreateElement("main", map[string]interface{}{"id": "content"},
			CreateElement("article", nil,
				CreateElement("h2", nil, CreateText("Article Title")),
				CreateElement("p", nil, CreateText("Article content...")),
			),
		),
		CreateElement("footer", nil,
			CreateText("© 2025"),
		),
	)
	
	vdom.Render(old)
	
	// Complex update
	new := CreateElement("div", map[string]interface{}{"class": "app dark-mode"},
		CreateElement("header", nil,
			CreateElement("h1", nil, CreateText("Updated App Title")),
			CreateElement("nav", nil,
				CreateElement("a", map[string]interface{}{"href": "/home", "class": "active"}, CreateText("Home")),
				CreateElement("a", map[string]interface{}{"href": "/about"}, CreateText("About")),
				CreateElement("a", map[string]interface{}{"href": "/contact"}, CreateText("Contact")),
			),
		),
		CreateElement("main", map[string]interface{}{"id": "content", "role": "main"},
			CreateElement("article", nil,
				CreateElement("h2", nil, CreateText("New Article Title")),
				CreateElement("p", nil, CreateText("Updated content...")),
				CreateElement("aside", nil, CreateText("Related links")),
			),
		),
		CreateElement("footer", nil,
			CreateText("© 2025 - Updated"),
		),
	)
	
	patches := vdom.Render(new)
	
	// Should have various types of patches
	patchTypes := make(map[PatchType]int)
	for _, patch := range patches {
		patchTypes[patch.Type]++
	}
	
	// Verify we have different patch types
	if patchTypes[PatchUpdateProps] == 0 {
		t.Error("Expected property update patches")
	}
	
	if patchTypes[PatchUpdateText] == 0 {
		t.Error("Expected text update patches")
	}
	
	if patchTypes[PatchCreate] == 0 {
		t.Error("Expected create patches for new elements")
	}
}

func BenchmarkVDomRender(b *testing.B) {
	vdom := NewVDom(nil)
	
	// Create a moderately complex tree
	createTree := func(text string) *VNode {
		return CreateElement("div", map[string]interface{}{"class": "container"},
			CreateElement("header", nil,
				CreateElement("h1", nil, CreateText(text)),
			),
			CreateElement("main", nil,
				CreateElement("ul", nil,
					CreateElement("li", nil, CreateText("Item 1")),
					CreateElement("li", nil, CreateText("Item 2")),
					CreateElement("li", nil, CreateText("Item 3")),
				),
			),
		)
	}
	
	old := createTree("Old Title")
	vdom.Render(old)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		new := createTree(fmt.Sprintf("Title %d", i))
		vdom.Render(new)
	}
}

func BenchmarkVDomKeyedChildren(b *testing.B) {
	vdom := NewVDom(&VDomConfig{EnableKeys: true})
	
	// Create list with keyed children
	createList := func(count int) *VNode {
		children := make([]*VNode, count)
		for i := 0; i < count; i++ {
			children[i] = CreateElement("li", 
				map[string]interface{}{"key": fmt.Sprintf("item-%d", i)},
				CreateText(fmt.Sprintf("Item %d", i)),
			)
		}
		return CreateElement("ul", nil, children...)
	}
	
	old := createList(100)
	vdom.Render(old)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Shuffle some items
		new := createList(100)
		vdom.Render(new)
	}
}

func BenchmarkVDomToHTML(b *testing.B) {
	vdom := NewVDom(nil)
	
	// Create a complex tree
	node := CreateElement("div", map[string]interface{}{"class": "container"},
		CreateElement("header", nil,
			CreateElement("h1", nil, CreateText("Title")),
			CreateElement("nav", nil,
				CreateElement("ul", nil,
					CreateElement("li", nil, CreateElement("a", nil, CreateText("Link 1"))),
					CreateElement("li", nil, CreateElement("a", nil, CreateText("Link 2"))),
					CreateElement("li", nil, CreateElement("a", nil, CreateText("Link 3"))),
				),
			),
		),
		CreateElement("main", nil,
			CreateElement("article", nil,
				CreateElement("h2", nil, CreateText("Article")),
				CreateElement("p", nil, CreateText("Content...")),
			),
		),
	)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = vdom.ToHTML(node)
	}
}