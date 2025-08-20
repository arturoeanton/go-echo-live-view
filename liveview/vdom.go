package liveview

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"html"
	"regexp"
	"strings"
	"sync"
)

// VNode represents a virtual DOM node
type VNode struct {
	Type       NodeType               `json:"type"`
	Tag        string                 `json:"tag,omitempty"`
	Props      map[string]interface{} `json:"props,omitempty"`
	Children   []*VNode               `json:"children,omitempty"`
	Text       string                 `json:"text,omitempty"`
	Key        string                 `json:"key,omitempty"`
	Component  Component              `json:"-"`
	Hash       string                 `json:"hash,omitempty"`
	EventHandlers map[string]EventHandler `json:"-"`
}

// NodeType represents the type of virtual node
type NodeType int

const (
	// ElementNode is an HTML element
	ElementNode NodeType = iota
	// TextNode is a text node
	TextNode
	// ComponentNode is a component node
	ComponentNode
	// FragmentNode is a fragment node
	FragmentNode
)

// Patch represents a DOM update operation
type Patch struct {
	Type      PatchType              `json:"type"`
	Path      string                 `json:"path"`
	NodeID    string                 `json:"nodeId,omitempty"`
	Value     interface{}            `json:"value,omitempty"`
	OldValue  interface{}            `json:"oldValue,omitempty"`
	Props     map[string]interface{} `json:"props,omitempty"`
	Children  []*VNode               `json:"children,omitempty"`
	Index     int                    `json:"index,omitempty"`
	MoveIndex int                    `json:"moveIndex,omitempty"`
}

// PatchType represents the type of patch operation
type PatchType int

const (
	// PatchCreate creates a new node
	PatchCreate PatchType = iota
	// PatchRemove removes a node
	PatchRemove
	// PatchReplace replaces a node
	PatchReplace
	// PatchUpdateText updates text content
	PatchUpdateText
	// PatchUpdateProps updates properties
	PatchUpdateProps
	// PatchReorder reorders children
	PatchReorder
	// PatchMove moves a node
	PatchMove
)

// VDom manages virtual DOM operations
type VDom struct {
	mu            sync.RWMutex
	root          *VNode
	oldRoot       *VNode
	patches       []Patch
	nodeMap       map[string]*VNode
	keyedElements map[string]*VNode
	config        *VDomConfig
}

// VDomConfig configures the virtual DOM
type VDomConfig struct {
	EnableKeys        bool
	EnableComponents  bool
	MaxDepth          int
	BatchUpdates      bool
	MinimizePatches   bool
	TrackChanges      bool
}

// DefaultVDomConfig returns default configuration
func DefaultVDomConfig() *VDomConfig {
	return &VDomConfig{
		EnableKeys:       true,
		EnableComponents: true,
		MaxDepth:         100,
		BatchUpdates:     true,
		MinimizePatches:  true,
		TrackChanges:     true,
	}
}

// NewVDom creates a new virtual DOM instance
func NewVDom(config *VDomConfig) *VDom {
	if config == nil {
		config = DefaultVDomConfig()
	}

	return &VDom{
		patches:       make([]Patch, 0),
		nodeMap:       make(map[string]*VNode),
		keyedElements: make(map[string]*VNode),
		config:        config,
	}
}

// CreateElement creates a virtual element node
func CreateElement(tag string, props map[string]interface{}, children ...*VNode) *VNode {
	node := &VNode{
		Type:     ElementNode,
		Tag:      tag,
		Props:    props,
		Children: children,
	}
	
	// Extract key if present
	if props != nil {
		if key, ok := props["key"].(string); ok {
			node.Key = key
			delete(props, "key") // Remove key from props
		}
	}
	
	// Calculate hash
	node.Hash = calculateNodeHash(node)
	
	return node
}

// CreateText creates a virtual text node
func CreateText(text string) *VNode {
	node := &VNode{
		Type: TextNode,
		Text: text,
	}
	node.Hash = calculateNodeHash(node)
	return node
}

// CreateComponent creates a virtual component node
func CreateComponent(component Component, props map[string]interface{}) *VNode {
	return &VNode{
		Type:      ComponentNode,
		Component: component,
		Props:     props,
	}
}

// CreateFragment creates a virtual fragment node
func CreateFragment(children ...*VNode) *VNode {
	return &VNode{
		Type:     FragmentNode,
		Children: children,
	}
}

// Render updates the virtual DOM tree and returns patches
func (vd *VDom) Render(newRoot *VNode) []Patch {
	vd.mu.Lock()
	defer vd.mu.Unlock()

	// Clear previous patches
	vd.patches = make([]Patch, 0)
	
	// Store old root
	vd.oldRoot = vd.root
	
	// Diff the trees
	if vd.oldRoot == nil {
		// Initial render
		vd.createPatches(newRoot, "root", 0)
	} else {
		// Update render
		vd.diffNodes(vd.oldRoot, newRoot, "root", 0)
	}
	
	// Update root
	vd.root = newRoot
	
	// Optimize patches if configured
	if vd.config.MinimizePatches {
		vd.optimizePatches()
	}
	
	return vd.patches
}

// diffNodes compares two nodes and generates patches
func (vd *VDom) diffNodes(oldNode, newNode *VNode, path string, index int) {
	// Both nil
	if oldNode == nil && newNode == nil {
		return
	}
	
	// Node removed
	if oldNode != nil && newNode == nil {
		vd.addPatch(Patch{
			Type:  PatchRemove,
			Path:  path,
			Index: index,
		})
		return
	}
	
	// Node added
	if oldNode == nil && newNode != nil {
		vd.createPatches(newNode, path, index)
		return
	}
	
	// Different types - replace
	if oldNode.Type != newNode.Type || oldNode.Tag != newNode.Tag {
		vd.addPatch(Patch{
			Type:     PatchReplace,
			Path:     path,
			Value:    newNode,
			OldValue: oldNode,
			Index:    index,
		})
		return
	}
	
	// Same node type - check for updates
	switch oldNode.Type {
	case TextNode:
		vd.diffText(oldNode, newNode, path)
	case ElementNode:
		vd.diffElement(oldNode, newNode, path)
	case ComponentNode:
		vd.diffComponent(oldNode, newNode, path)
	case FragmentNode:
		vd.diffChildren(oldNode.Children, newNode.Children, path)
	}
}

// diffText compares text nodes
func (vd *VDom) diffText(oldNode, newNode *VNode, path string) {
	if oldNode.Text != newNode.Text {
		vd.addPatch(Patch{
			Type:     PatchUpdateText,
			Path:     path,
			Value:    newNode.Text,
			OldValue: oldNode.Text,
		})
	}
}

// diffElement compares element nodes
func (vd *VDom) diffElement(oldNode, newNode *VNode, path string) {
	// Check props
	if vd.propsChanged(oldNode.Props, newNode.Props) {
		vd.addPatch(Patch{
			Type:  PatchUpdateProps,
			Path:  path,
			Props: newNode.Props,
		})
	}
	
	// Check children
	vd.diffChildren(oldNode.Children, newNode.Children, path)
}

// diffComponent compares component nodes
func (vd *VDom) diffComponent(oldNode, newNode *VNode, path string) {
	// Component-specific diffing logic
	if oldNode.Component != newNode.Component {
		vd.addPatch(Patch{
			Type:     PatchReplace,
			Path:     path,
			Value:    newNode,
			OldValue: oldNode,
		})
		return
	}
	
	// Check props
	if vd.propsChanged(oldNode.Props, newNode.Props) {
		// Update component props
		vd.addPatch(Patch{
			Type:  PatchUpdateProps,
			Path:  path,
			Props: newNode.Props,
		})
	}
}

// diffChildren compares children using keys when available
func (vd *VDom) diffChildren(oldChildren, newChildren []*VNode, parentPath string) {
	// Build key maps if keys are enabled
	if vd.config.EnableKeys {
		vd.diffKeyedChildren(oldChildren, newChildren, parentPath)
	} else {
		vd.diffNonKeyedChildren(oldChildren, newChildren, parentPath)
	}
}

// diffKeyedChildren handles children with keys
func (vd *VDom) diffKeyedChildren(oldChildren, newChildren []*VNode, parentPath string) {
	oldKeyMap := make(map[string]*VNode)
	oldIndexMap := make(map[string]int)
	
	// Build maps for old children
	for i, child := range oldChildren {
		if child != nil && child.Key != "" {
			oldKeyMap[child.Key] = child
			oldIndexMap[child.Key] = i
		}
	}
	
	// Track processed keys
	processedKeys := make(map[string]bool)
	
	// Process new children
	for i, newChild := range newChildren {
		if newChild == nil {
			continue
		}
		
		childPath := fmt.Sprintf("%s.children[%d]", parentPath, i)
		
		if newChild.Key != "" {
			processedKeys[newChild.Key] = true
			
			if oldChild, exists := oldKeyMap[newChild.Key]; exists {
				// Same key exists - check if moved
				oldIndex := oldIndexMap[newChild.Key]
				
				if oldIndex != i {
					// Child moved
					vd.addPatch(Patch{
						Type:      PatchMove,
						Path:      childPath,
						Index:     i,
						MoveIndex: oldIndex,
					})
				}
				
				// Diff the nodes
				vd.diffNodes(oldChild, newChild, childPath, i)
			} else {
				// New keyed child
				vd.createPatches(newChild, childPath, i)
			}
		} else {
			// Non-keyed child in keyed list
			if i < len(oldChildren) {
				vd.diffNodes(oldChildren[i], newChild, childPath, i)
			} else {
				vd.createPatches(newChild, childPath, i)
			}
		}
	}
	
	// Remove old children that weren't processed
	for key, oldChild := range oldKeyMap {
		if !processedKeys[key] {
			oldIndex := oldIndexMap[key]
			childPath := fmt.Sprintf("%s.children[%d]", parentPath, oldIndex)
			vd.addPatch(Patch{
				Type:  PatchRemove,
				Path:  childPath,
				Index: oldIndex,
				Value: oldChild,
			})
		}
	}
}

// diffNonKeyedChildren handles children without keys
func (vd *VDom) diffNonKeyedChildren(oldChildren, newChildren []*VNode, parentPath string) {
	// Process all new children
	for i := 0; i < len(newChildren); i++ {
		childPath := fmt.Sprintf("%s.children[%d]", parentPath, i)
		
		if i < len(oldChildren) {
			// Diff existing child
			vd.diffNodes(oldChildren[i], newChildren[i], childPath, i)
		} else {
			// Add new child
			vd.diffNodes(nil, newChildren[i], childPath, i)
		}
	}
	
	// Remove extra old children
	for i := len(newChildren); i < len(oldChildren); i++ {
		childPath := fmt.Sprintf("%s.children[%d]", parentPath, i)
		vd.diffNodes(oldChildren[i], nil, childPath, i)
	}
}

// propsChanged checks if properties have changed
func (vd *VDom) propsChanged(oldProps, newProps map[string]interface{}) bool {
	if len(oldProps) != len(newProps) {
		return true
	}
	
	for key, oldValue := range oldProps {
		newValue, exists := newProps[key]
		if !exists || !equalValues(oldValue, newValue) {
			return true
		}
	}
	
	return false
}

// createPatches generates patches for creating new nodes
func (vd *VDom) createPatches(node *VNode, path string, index int) {
	vd.addPatch(Patch{
		Type:  PatchCreate,
		Path:  path,
		Value: node,
		Index: index,
	})
}

// addPatch adds a patch to the list
func (vd *VDom) addPatch(patch Patch) {
	vd.patches = append(vd.patches, patch)
}

// optimizePatches optimizes the patch list
func (vd *VDom) optimizePatches() {
	if len(vd.patches) == 0 {
		return
	}
	
	// Group patches by path
	patchMap := make(map[string][]Patch)
	for _, patch := range vd.patches {
		patchMap[patch.Path] = append(patchMap[patch.Path], patch)
	}
	
	// Optimize grouped patches
	optimized := make([]Patch, 0, len(vd.patches))
	
	for path, patches := range patchMap {
		// If node is being replaced, ignore other patches
		for _, patch := range patches {
			if patch.Type == PatchReplace || patch.Type == PatchRemove {
				optimized = append(optimized, patch)
				continue
			}
		}
		
		// Combine property updates
		var combinedProps map[string]interface{}
		for _, patch := range patches {
			if patch.Type == PatchUpdateProps {
				if combinedProps == nil {
					combinedProps = make(map[string]interface{})
				}
				for k, v := range patch.Props {
					combinedProps[k] = v
				}
			} else {
				optimized = append(optimized, patch)
			}
		}
		
		if combinedProps != nil {
			optimized = append(optimized, Patch{
				Type:  PatchUpdateProps,
				Path:  path,
				Props: combinedProps,
			})
		}
	}
	
	vd.patches = optimized
}

// ApplyPatches applies patches to the actual DOM
func (vd *VDom) ApplyPatches(patches []Patch) error {
	for _, patch := range patches {
		if err := vd.applyPatch(patch); err != nil {
			return fmt.Errorf("failed to apply patch: %w", err)
		}
	}
	return nil
}

// applyPatch applies a single patch
func (vd *VDom) applyPatch(patch Patch) error {
	switch patch.Type {
	case PatchCreate:
		return vd.applyCreate(patch)
	case PatchRemove:
		return vd.applyRemove(patch)
	case PatchReplace:
		return vd.applyReplace(patch)
	case PatchUpdateText:
		return vd.applyUpdateText(patch)
	case PatchUpdateProps:
		return vd.applyUpdateProps(patch)
	case PatchMove:
		return vd.applyMove(patch)
	case PatchReorder:
		return vd.applyReorder(patch)
	}
	return nil
}

// applyCreate handles node creation
func (vd *VDom) applyCreate(patch Patch) error {
	// Implementation depends on actual DOM manipulation
	// This is where you'd create actual DOM elements
	Debug("Creating node at %s", patch.Path)
	return nil
}

// applyRemove handles node removal
func (vd *VDom) applyRemove(patch Patch) error {
	Debug("Removing node at %s", patch.Path)
	return nil
}

// applyReplace handles node replacement
func (vd *VDom) applyReplace(patch Patch) error {
	Debug("Replacing node at %s", patch.Path)
	return nil
}

// applyUpdateText handles text updates
func (vd *VDom) applyUpdateText(patch Patch) error {
	Debug("Updating text at %s: %v -> %v", patch.Path, patch.OldValue, patch.Value)
	return nil
}

// applyUpdateProps handles property updates
func (vd *VDom) applyUpdateProps(patch Patch) error {
	Debug("Updating props at %s", patch.Path)
	return nil
}

// applyMove handles node movement
func (vd *VDom) applyMove(patch Patch) error {
	Debug("Moving node from %d to %d at %s", patch.MoveIndex, patch.Index, patch.Path)
	return nil
}

// applyReorder handles children reordering
func (vd *VDom) applyReorder(patch Patch) error {
	Debug("Reordering children at %s", patch.Path)
	return nil
}

// ToHTML converts a virtual node to HTML
func (vd *VDom) ToHTML(node *VNode) string {
	if node == nil {
		return ""
	}
	
	var buf bytes.Buffer
	vd.nodeToHTML(node, &buf)
	return buf.String()
}

// nodeToHTML converts a node to HTML
func (vd *VDom) nodeToHTML(node *VNode, buf *bytes.Buffer) {
	switch node.Type {
	case TextNode:
		buf.WriteString(html.EscapeString(node.Text))
		
	case ElementNode:
		// Open tag
		buf.WriteString("<")
		buf.WriteString(node.Tag)
		
		// Write attributes
		if node.Props != nil {
			for key, value := range node.Props {
				buf.WriteString(" ")
				buf.WriteString(key)
				buf.WriteString(`="`)
				buf.WriteString(html.EscapeString(fmt.Sprintf("%v", value)))
				buf.WriteString(`"`)
			}
		}
		
		// Self-closing tags
		if isSelfClosing(node.Tag) {
			buf.WriteString(" />")
			return
		}
		
		buf.WriteString(">")
		
		// Write children
		for _, child := range node.Children {
			vd.nodeToHTML(child, buf)
		}
		
		// Close tag
		buf.WriteString("</")
		buf.WriteString(node.Tag)
		buf.WriteString(">")
		
	case ComponentNode:
		// Render component
		if node.Component != nil {
			html := node.Component.GetTemplate()
			buf.WriteString(html)
		}
		
	case FragmentNode:
		// Render children only
		for _, child := range node.Children {
			vd.nodeToHTML(child, buf)
		}
	}
}

// ParseHTML parses HTML into virtual nodes
func ParseHTML(html string) (*VNode, error) {
	// Simple HTML parser (for basic use cases)
	// In production, use a proper HTML parser
	html = strings.TrimSpace(html)
	
	if html == "" {
		return nil, nil
	}
	
	// Text node
	if !strings.HasPrefix(html, "<") {
		return CreateText(html), nil
	}
	
	// Parse element
	tagRegex := regexp.MustCompile(`<(\w+)([^>]*)>`)
	matches := tagRegex.FindStringSubmatch(html)
	
	if len(matches) < 2 {
		return CreateText(html), nil
	}
	
	tag := matches[1]
	propsStr := matches[2]
	
	// Parse properties
	props := parseProps(propsStr)
	
	// Create element
	return CreateElement(tag, props), nil
}

// parseProps parses HTML attributes
func parseProps(propsStr string) map[string]interface{} {
	props := make(map[string]interface{})
	
	if propsStr == "" {
		return props
	}
	
	// Simple attribute parser
	attrRegex := regexp.MustCompile(`(\w+)="([^"]*)"`)
	matches := attrRegex.FindAllStringSubmatch(propsStr, -1)
	
	for _, match := range matches {
		if len(match) >= 3 {
			props[match[1]] = match[2]
		}
	}
	
	return props
}

// calculateNodeHash calculates a hash for a node
func calculateNodeHash(node *VNode) string {
	hasher := md5.New()
	
	switch node.Type {
	case TextNode:
		hasher.Write([]byte(node.Text))
	case ElementNode:
		hasher.Write([]byte(node.Tag))
		// Include props in hash
		for key, value := range node.Props {
			hasher.Write([]byte(key))
			hasher.Write([]byte(fmt.Sprintf("%v", value)))
		}
	}
	
	return hex.EncodeToString(hasher.Sum(nil))
}

// equalValues compares two values for equality
func equalValues(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// isSelfClosing checks if a tag is self-closing
func isSelfClosing(tag string) bool {
	selfClosing := []string{
		"area", "base", "br", "col", "embed", "hr", "img", "input",
		"link", "meta", "param", "source", "track", "wbr",
	}
	
	for _, t := range selfClosing {
		if t == tag {
			return true
		}
	}
	
	return false
}

// VDomRenderer provides component rendering with virtual DOM
type VDomRenderer struct {
	vdom      *VDom
	component Component
	mu        sync.RWMutex
}

// NewVDomRenderer creates a new virtual DOM renderer
func NewVDomRenderer(component Component) *VDomRenderer {
	return &VDomRenderer{
		vdom:      NewVDom(nil),
		component: component,
	}
}

// Render renders the component and returns patches
func (vr *VDomRenderer) Render() ([]Patch, error) {
	vr.mu.Lock()
	defer vr.mu.Unlock()
	
	// Parse component template to virtual DOM
	html := vr.component.GetTemplate()
	vnode, err := ParseHTML(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}
	
	// Get patches
	patches := vr.vdom.Render(vnode)
	
	return patches, nil
}

// GetHTML returns the current HTML
func (vr *VDomRenderer) GetHTML() string {
	vr.mu.RLock()
	defer vr.mu.RUnlock()
	
	if vr.vdom.root == nil {
		return ""
	}
	
	return vr.vdom.ToHTML(vr.vdom.root)
}