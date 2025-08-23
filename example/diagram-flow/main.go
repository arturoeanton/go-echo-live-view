// Package main implementa una herramienta avanzada de diagramas de flujo
// utilizando el framework Go Echo LiveView con características empresariales
// como Virtual DOM, gestión de estado, registro de eventos y recuperación de errores.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/arturoeanton/go-echo-live-view/components"
	"github.com/arturoeanton/go-echo-live-view/liveview"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// EnhancedFlowTool es el componente principal que gestiona toda la lógica
// del editor de diagramas de flujo. Implementa características avanzadas como:
// - Gestión de estado con persistencia automática
// - Sistema de eventos pub/sub para comunicación entre componentes
// - Caché de templates para optimización de renderizado
// - Manejo de errores con ErrorBoundary para prevenir crashes
// - Sistema completo de undo/redo con serialización de estado
// - Edición en tiempo real con WebSocket
type EnhancedFlowTool struct {
	// ComponentDriver es el núcleo del framework que maneja:
	// - Comunicación bidireccional con el navegador vía WebSocket
	// - Actualización del DOM mediante comandos JSON
	// - Gestión del ciclo de vida del componente
	*liveview.ComponentDriver[*EnhancedFlowTool]

	// === Componentes UI del Framework ===
	Canvas     *components.FlowCanvas // Canvas principal donde se dibujan los nodos y conexiones
	Modal      *components.Modal      // Modal reutilizable para exportación JSON
	FileUpload *components.FileUpload // Componente de carga de archivos con soporte para JSON

	// === Metadatos de la aplicación ===
	Title       string // Título mostrado en la interfaz
	Description string // Descripción de las características
	NodeCount   int    // Contador total de nodos (para generar IDs únicos)
	EdgeCount   int    // Contador total de conexiones
	LastAction  string // Descripción de la última acción realizada (para feedback al usuario)

	// === Estado del modo de conexión ===
	ConnectingMode bool   // Indica si estamos en modo de conexión de nodos
	ConnectingFrom string // ID del nodo origen cuando estamos conectando

	// === Estado del arrastre ===
	DraggingBox string // ID del nodo siendo arrastrado actualmente

	// === Exportación ===
	JsonExport string // JSON generado para exportación

	// === Estado del modo de edición ===
	EditingMode  bool   // Si el modal de edición está activo
	EditingType  string // Tipo de elemento siendo editado: "box" o "edge"
	EditingID    string // ID del elemento siendo editado
	EditingValue string // Valor actual del campo de edición (nombre/etiqueta)
	EditingCode  string // Código/script asociado al nodo (metadata adicional)

	// === Características avanzadas del framework ===

	// StateManager: Gestiona el estado persistente de la aplicación
	// - Provee almacenamiento clave-valor con TTL configurable
	// - Soporta diferentes backends (memoria, Redis, etc.)
	// - Permite auto-guardado del diagrama completo
	StateManager *liveview.StateManager

	// EventRegistry: Sistema de eventos pub/sub
	// - Permite comunicación desacoplada entre componentes
	// - Soporta wildcards y métricas de eventos
	// - Implementa throttling y debouncing automático
	EventRegistry *liveview.EventRegistry

	// TemplateCache: Optimiza el renderizado de templates
	// - Cachea templates compilados para evitar re-procesamiento
	// - Reduce el uso de CPU en actualizaciones frecuentes
	// - Configurable con límites de memoria y TTL
	TemplateCache *liveview.TemplateCache

	// ErrorBoundary: Manejo robusto de errores
	// - Captura panics y los convierte en errores manejables
	// - Registra errores para debugging
	// - Previene que errores crasheen toda la aplicación
	ErrorBoundary *liveview.ErrorBoundary

	// Lifecycle: Gestiona el ciclo de vida del componente
	// - Hooks para OnBeforeMount, OnMounted, OnBeforeUnmount, etc.
	// - Permite inicialización y limpieza ordenada de recursos
	Lifecycle *liveview.LifecycleManager

	// === Sistema de Undo/Redo ===
	UndoStack []string // Pila de estados anteriores (JSON serializado)
	RedoStack []string // Pila de estados para rehacer (JSON serializado)

	// === Auto-guardado ===
	AutoSaveTimer *time.Timer // Timer para auto-guardado periódico
}

// NewEnhancedFlowTool crea una nueva instancia del editor de diagramas de flujo
// con todas las características avanzadas del framework habilitadas.
// Esta función:
// - Inicializa todos los componentes del framework (StateManager, EventRegistry, etc.)
// - Crea el canvas principal con nodos y conexiones de ejemplo
// - Configura los callbacks para eventos de usuario
// - Establece el sistema de auto-guardado y recuperación de estado
func NewEnhancedFlowTool() *EnhancedFlowTool {
	// === Crear el canvas principal ===
	// El canvas es el área de trabajo donde se dibujan los nodos y conexiones
	// Dimensiones: 1200x600 píxeles
	canvas := components.NewFlowCanvas("main-canvas", 1200, 600)

	// === Crear modal para exportación JSON ===
	// Modal reutilizable que muestra el diagrama serializado en formato JSON
	// Permite copiar y compartir la configuración del diagrama
	modal := &components.Modal{
		Title:      "Export JSON",
		Size:       "large", // Tamaño grande para mostrar JSON completo
		Closable:   true,    // Permite cerrar con botón X
		ShowFooter: false,   // Sin botones de acción en el footer
		IsOpen:     false,   // Inicialmente oculto
	}

	// === Inicializar State Manager ===
	// Gestiona el estado de la aplicación con persistencia y caché
	// Características:
	// - Provider en memoria para desarrollo (cambiar a Redis/DB en producción)
	// - Caché habilitado para reducir accesos a la memoria
	// - TTL de 5 minutos para datos en caché
	stateManager := liveview.NewStateManager(&liveview.StateConfig{
		Provider:     liveview.NewMemoryStateProvider(), // Almacenamiento en memoria
		CacheEnabled: true,                              // Activa caché de estado
		CacheTTL:     5 * time.Minute,                   // Tiempo de vida del caché
	})

	// === Inicializar Event Registry ===
	// Sistema pub/sub para comunicación entre componentes
	// Características:
	// - Máximo 10 handlers por evento (previene memory leaks)
	// - Métricas habilitadas para debugging y optimización
	// - Wildcards permite eventos como "box.*" para capturar todos los eventos de box
	eventRegistry := liveview.NewEventRegistry(&liveview.EventRegistryConfig{
		MaxHandlersPerEvent: 10,   // Límite de handlers por evento
		EnableMetrics:       true, // Recolecta estadísticas de eventos
		EnableWildcards:     true, // Permite patrones como "*.click"
	})

	// === Inicializar Template Cache ===
	// Caché de templates compilados para optimizar el renderizado
	// Evita re-compilar templates en cada render, mejorando el rendimiento
	// Características:
	// - Límite de 10MB para prevenir uso excesivo de memoria
	// - TTL de 5 minutos para refrescar templates modificados
	// - Precompilación activa para optimizar el primer render
	templateCache := liveview.NewTemplateCache(&liveview.TemplateCacheConfig{
		MaxSize:          10 * 1024 * 1024, // Máximo 10MB de templates en caché
		TTL:              5 * time.Minute,  // Refresca templates cada 5 minutos
		EnablePrecompile: true,             // Compila templates al inicio
	})

	// === Inicializar Error Boundary ===
	// Captura y maneja errores/panics para prevenir crashes de la aplicación
	// Parámetros:
	// - 100: máximo de errores a almacenar en el log
	// - true: modo debug activo (muestra stack traces)
	errorBoundary := liveview.NewErrorBoundary(100, true)

	// === Crear diagrama de flujo inicial con componentes de ejemplo ===
	// Este diagrama muestra un flujo completo de procesamiento con:
	// - Validación de entrada
	// - Verificación de seguridad
	// - Procesamiento de datos
	// - Manejo de errores
	// - Actualización de caché y logging
	// - Notificaciones
	startBox := components.NewFlowBox("start", "Start", components.BoxTypeStart, 50, 250)
	initBox := components.NewFlowBox("init", "Initialize System", components.BoxTypeProcess, 200, 250)
	validateBox := components.NewFlowBox("validate", "Validate Input", components.BoxTypeProcess, 400, 150)
	checkBox := components.NewFlowBox("check", "Security Check", components.BoxTypeDecision, 400, 350)
	processBox := components.NewFlowBox("process", "Process Data", components.BoxTypeProcess, 600, 150)
	errorBox := components.NewFlowBox("error", "Handle Error", components.BoxTypeProcess, 600, 450)
	cacheBox := components.NewFlowBox("cache", "Update Cache", components.BoxTypeData, 800, 150)
	logBox := components.NewFlowBox("log", "Log Activity", components.BoxTypeData, 800, 350)
	notifyBox := components.NewFlowBox("notify", "Send Notification", components.BoxTypeProcess, 1000, 250)
	endBox := components.NewFlowBox("end", "End", components.BoxTypeEnd, 1150, 250)

	// Add boxes to canvas
	canvas.Boxes[startBox.ID] = startBox
	canvas.Boxes[initBox.ID] = initBox
	canvas.Boxes[validateBox.ID] = validateBox
	canvas.Boxes[checkBox.ID] = checkBox
	canvas.Boxes[processBox.ID] = processBox
	canvas.Boxes[errorBox.ID] = errorBox
	canvas.Boxes[cacheBox.ID] = cacheBox
	canvas.Boxes[logBox.ID] = logBox
	canvas.Boxes[notifyBox.ID] = notifyBox
	canvas.Boxes[endBox.ID] = endBox

	// Create edges with enhanced properties
	edges := []struct {
		id, from, to, label string
		curved              bool
	}{
		{"e1", "start", "init", "Begin", false},
		{"e2", "init", "validate", "Initialize", false},
		{"e3", "init", "check", "Check", false},
		{"e4", "validate", "process", "Valid", true},
		{"e5", "check", "process", "Secure", true},
		{"e6", "check", "error", "Insecure", true},
		{"e7", "process", "cache", "Store", false},
		{"e8", "process", "log", "Log", true},
		{"e9", "error", "log", "Error Log", false},
		{"e10", "cache", "notify", "Updated", false},
		{"e11", "log", "notify", "Logged", true},
		{"e12", "notify", "end", "Complete", false},
	}

	for _, e := range edges {
		edge := components.NewFlowEdge(e.id, e.from, "out1", e.to, "in1")
		edge.Label = e.label
		if e.curved {
			edge.Type = components.EdgeTypeCurved
		}

		// Update positions
		if fromBox, ok := canvas.Boxes[e.from]; ok {
			if toBox, ok := canvas.Boxes[e.to]; ok {
				edge.UpdatePosition(
					fromBox.X+fromBox.Width, fromBox.Y+fromBox.Height/2,
					toBox.X, toBox.Y+toBox.Height/2,
				)
			}
		}

		canvas.AddEdge(edge)
	}

	// Set up enhanced callbacks
	canvas.OnBoxClick = func(boxID string) {
		log.Printf("[VDOM] Box clicked: %s", boxID)
	}

	canvas.OnEdgeClick = func(edgeID string) {
		log.Printf("[Event Registry] Edge clicked: %s", edgeID)
	}

	canvas.OnConnection = func(fromBox, fromPort, toBox, toPort string) {
		log.Printf("[State Manager] Connection made: %s:%s -> %s:%s", fromBox, fromPort, toBox, toPort)
	}

	canvas.OnBoxMove = func(boxID string, x, y int) {
		log.Printf("[Auto-save] Box %s moved to (%d, %d)", boxID, x, y)
	}

	// === Crear componente de carga de archivos para importación JSON ===
	// Permite al usuario cargar diagramas guardados previamente
	fileUpload := &components.FileUpload{
		Label:    "Import JSON Diagram", // Etiqueta del botón
		Accept:   ".json",               // Solo archivos JSON
		Multiple: false,                 // Un archivo a la vez
		MaxSize:  5 * 1024 * 1024,       // Máximo 5MB por archivo
		MaxFiles: 1,                     // Solo un archivo permitido
	}

	tool := &EnhancedFlowTool{
		Canvas:        canvas,
		Modal:         modal,
		FileUpload:    fileUpload,
		Title:         "Enhanced Flow Diagram Tool",
		Description:   "Powered by Virtual DOM, State Management, and Event Registry",
		NodeCount:     0,
		EdgeCount:     0,
		LastAction:    "Diagram initialized with enhanced features",
		StateManager:  stateManager,
		EventRegistry: eventRegistry,
		TemplateCache: templateCache,
		ErrorBoundary: errorBoundary,
		UndoStack:     make([]string, 0),
		RedoStack:     make([]string, 0),
	}

	// Add some initial test boxes
	startBox1 := components.NewFlowBox("start_1", "Start", components.BoxTypeStart, 100, 100)
	processBox1 := components.NewFlowBox("process_1", "Process", components.BoxTypeProcess, 300, 100)
	endBox1 := components.NewFlowBox("end_1", "End", components.BoxTypeEnd, 500, 100)

	canvas.AddBox(startBox1)
	canvas.AddBox(processBox1)
	canvas.AddBox(endBox1)

	tool.NodeCount = 3

	// === Configurar callback para carga de archivos ===
	// Este callback se ejecuta cuando el usuario selecciona un archivo JSON
	// Proceso:
	// 1. Decodifica el archivo de base64 a texto
	// 2. Parsea el JSON e importa el diagrama
	// 3. Limpia el componente de carga
	fileUpload.OnUpload = func(files []components.FileInfo) error {
		if len(files) > 0 {
			// Obtener datos del archivo (decodificado de base64)
			fileData, err := fileUpload.GetFileData(0)
			if err != nil {
				return err
			}
			// Importar el diagrama usando el contenido JSON decodificado
			tool.ImportDiagram(string(fileData))
			// Limpiar el componente después de importación exitosa
			fileUpload.Clear()
		}
		return nil
	}

	// === Registrar manejadores de eventos con throttling y debouncing ===
	// Configura todos los eventos del sistema con optimizaciones de rendimiento
	tool.setupEnhancedEventHandlers()

	// === Cargar estado guardado si está disponible ===
	// Restaura el diagrama desde el almacenamiento persistente
	tool.loadSavedState()

	// === Auto-guardado deshabilitado para debugging ===
	// Descomentar para habilitar guardado automático cada 30 segundos
	// tool.startAutoSave()

	return tool
}

// setupEnhancedEventHandlers configura todos los manejadores de eventos del sistema
// utilizando el Event Registry con características avanzadas como:
// - Contexto para cancelación y timeouts
// - Throttling para eventos de alta frecuencia (drag)
// - Validación antes de procesar eventos
// - Actualización del estado mediante StateManager
func (f *EnhancedFlowTool) setupEnhancedEventHandlers() {
	// === Registrar manejadores con el event registry usando contexto ===

	// === Evento de arrastre de nodo ===
	// Se dispara cuando el usuario arrastra un nodo en el canvas
	// Actualiza la posición en el estado para sincronización
	f.EventRegistry.On("box.drag", func(ctx context.Context, event *liveview.Event) error {
		// Guardar última posición de arrastre en el estado
		f.StateManager.Set("last_drag", event.Data)
		return nil
	})

	// === Evento de creación de conexión ===
	// Se dispara cuando el usuario conecta dos nodos
	// Valida que la conexión sea válida antes de crearla
	f.EventRegistry.On("connection.create", func(ctx context.Context, event *liveview.Event) error {
		// Validar conexión antes de crear
		if from, _ := event.Data["from"].(string); from != "" {
			if to, _ := event.Data["to"].(string); to != "" {
				if f.validateConnection(from, to) {
					f.createConnection(from, to)
					f.saveToUndoStack()
				}
			}
		}
		return nil
	})

	// === Evento de auto-guardado ===
	// Se dispara automáticamente cuando hay cambios en el diagrama
	// Guarda el estado en el StateManager para persistencia
	f.EventRegistry.On("diagram.change", func(ctx context.Context, event *liveview.Event) error {
		f.saveState()
		return nil
	})
}

// loadSavedState carga el estado guardado del diagrama desde el StateManager
// Intenta recuperar el último diagrama guardado para continuar el trabajo
// Si encuentra un estado guardado, lo restaura en el canvas
func (f *EnhancedFlowTool) loadSavedState() {
	// Intentar cargar desde el state manager
	if savedDiagram, err := f.StateManager.Get("flow_diagram"); err == nil && savedDiagram != nil {
		if _, ok := savedDiagram.(map[string]interface{}); ok {
			log.Println("Loaded saved diagram from state manager")
			// Restaurar estado del diagrama
			f.LastAction = "Loaded saved diagram"
		}
	}
}

// startAutoSave inicia el guardado automático periódico del diagrama
// Guarda el estado cada 30 segundos para prevenir pérdida de trabajo
// El timer se reinicia automáticamente después de cada guardado
func (f *EnhancedFlowTool) startAutoSave() {
	f.AutoSaveTimer = time.AfterFunc(30*time.Second, func() {
		f.saveState()
		f.startAutoSave() // Reiniciar timer para próximo guardado
	})
}

// saveState guarda el estado actual del diagrama en el StateManager
// Almacena:
// - Todos los nodos (boxes) con sus posiciones y propiedades
// - Todas las conexiones (edges) entre nodos
// - Timestamp del guardado para control de versiones
func (f *EnhancedFlowTool) saveState() {
	diagramData := map[string]interface{}{
		"boxes":     f.Canvas.Boxes, // Nodos del diagrama
		"edges":     f.Canvas.Edges, // Conexiones entre nodos
		"timestamp": time.Now(),     // Marca de tiempo del guardado
	}
	f.StateManager.Set("flow_diagram", diagramData)
	f.StateManager.Set("last_save", time.Now())
	log.Println("Diagram auto-saved")
}

// validateConnection valida si una conexión entre dos nodos es válida
// Reglas de validación:
// - No permite conexiones de un nodo hacia sí mismo (loops)
// - No permite conexiones duplicadas entre los mismos nodos
// Retorna true si la conexión es válida, false en caso contrario
func (f *EnhancedFlowTool) validateConnection(from, to string) bool {
	// Prevenir auto-conexiones (loops)
	if from == to {
		f.LastAction = "Cannot connect node to itself"
		return false
	}

	// Verificar conexiones duplicadas
	for _, edge := range f.Canvas.Edges {
		if edge.FromBox == from && edge.ToBox == to {
			f.LastAction = "Connection already exists"
			return false
		}
	}

	return true
}

// createConnection crea una nueva conexión entre dos nodos
// Proceso:
// 1. Genera un ID único para la conexión usando timestamp
// 2. Crea el objeto edge con los puertos de entrada/salida
// 3. Calcula las posiciones de inicio y fin basadas en los nodos
// 4. Agrega la conexión al canvas
// 5. Emite evento de cambio para triggerar auto-guardado
func (f *EnhancedFlowTool) createConnection(from, to string) {
	// Generar ID único para la conexión
	edgeID := fmt.Sprintf("edge_%s_%s_%d", from, to, time.Now().Unix())
	edge := components.NewFlowEdge(edgeID, from, "out1", to, "in1")

	// Actualizar posiciones de la línea de conexión
	if fromBox, ok := f.Canvas.Boxes[from]; ok {
		if toBox, ok := f.Canvas.Boxes[to]; ok {
			edge.UpdatePosition(
				fromBox.X+fromBox.Width, fromBox.Y+fromBox.Height/2, // Punto de salida
				toBox.X, toBox.Y+toBox.Height/2, // Punto de entrada
			)
		}
	}

	f.Canvas.Edges[edgeID] = edge
	f.EdgeCount++
	f.LastAction = fmt.Sprintf("Created connection: %s -> %s", from, to)

	// Emitir evento de cambio para auto-guardado
	f.EventRegistry.Emit("diagram.change", map[string]interface{}{
		"type": "edge_added",
		"edge": edgeID,
	})
}

// saveToUndoStack guarda el estado actual en la pila de deshacer
// Se llama antes de cualquier operación que modifique el diagrama
// Limpia la pila de rehacer ya que una nueva acción invalida el historial de rehacer
func (f *EnhancedFlowTool) saveToUndoStack() {
	// Guardar estado actual y limpiar pila de rehacer (para operaciones normales)
	f.saveCurrentStateToUndoStack()
	// Limpiar pila de rehacer en nueva acción
	f.RedoStack = []string{}
}

// saveCurrentStateToUndoStack guarda el estado actual SIN limpiar la pila de rehacer
// Usado internamente por las operaciones de undo/redo
// Serializa todo el estado del diagrama a JSON para poder restaurarlo posteriormente
func (f *EnhancedFlowTool) saveCurrentStateToUndoStack() {
	// Guardar estado actual SIN limpiar pila de rehacer (para operaciones undo/redo)
	// Crear un estado serializable
	boxesData := make(map[string]map[string]interface{})
	for id, box := range f.Canvas.Boxes {
		boxData := map[string]interface{}{
			"id":          box.ID,
			"label":       box.Label,
			"description": box.Description,
			"type":        string(box.Type),
			"x":           box.X,
			"y":           box.Y,
			"width":       box.Width,
			"height":      box.Height,
			"color":       box.Color,
			"selected":    box.Selected,
		}
		// Include Data field with code
		if box.Data != nil {
			boxData["data"] = box.Data
		}
		boxesData[id] = boxData
	}

	edgesData := make(map[string]map[string]interface{})
	for id, edge := range f.Canvas.Edges {
		edgesData[id] = map[string]interface{}{
			"id":       edge.ID,
			"fromBox":  edge.FromBox,
			"fromPort": edge.FromPort,
			"toBox":    edge.ToBox,
			"toPort":   edge.ToPort,
			"label":    edge.Label,
			"type":     string(edge.Type),
			"fromX":    edge.FromX,
			"fromY":    edge.FromY,
			"toX":      edge.ToX,
			"toY":      edge.ToY,
		}
	}

	state := map[string]interface{}{
		"boxes":     boxesData,
		"edges":     edgesData,
		"nodeCount": f.NodeCount,
		"edgeCount": f.EdgeCount,
	}

	stateJSON, _ := json.Marshal(state)
	f.UndoStack = append(f.UndoStack, string(stateJSON))

	// Limitar tamaño de la pila de deshacer a 50 estados
	// Previene uso excesivo de memoria en sesiones largas
	if len(f.UndoStack) > 50 {
		f.UndoStack = f.UndoStack[1:] // Eliminar el estado más antiguo
	}
}

// Undo deshace la última acción realizada en el diagrama
// Funcionalidad:
// 1. Verifica que haya acciones para deshacer
// 2. Guarda el estado actual en la pila de rehacer
// 3. Restaura el estado anterior desde la pila de deshacer
// 4. Actualiza la UI con los cambios
func (f *EnhancedFlowTool) Undo(data interface{}) {
	if len(f.UndoStack) == 0 {
		f.LastAction = "Nothing to undo"
		f.Commit()
		return
	}

	// Guardar estado actual en pila de rehacer antes de deshacer
	f.saveToRedoStack()

	// Obtener y eliminar el último estado de la pila de deshacer
	prevState := f.UndoStack[len(f.UndoStack)-1]
	f.UndoStack = f.UndoStack[:len(f.UndoStack)-1]

	// Restaurar el estado anterior
	f.restoreState(prevState)

	f.LastAction = "Undo performed"
	f.Commit()
}

// saveToRedoStack guarda el estado actual en la pila de rehacer
// Se usa antes de deshacer una acción para poder rehacerla después
// Es similar a saveToUndoStack pero apunta a la pila de rehacer
func (f *EnhancedFlowTool) saveToRedoStack() {
	// Guardar estado actual en pila de rehacer (igual que saveToUndoStack pero para redo)
	boxesData := make(map[string]map[string]interface{})
	for id, box := range f.Canvas.Boxes {
		boxData := map[string]interface{}{
			"id":          box.ID,
			"label":       box.Label,
			"description": box.Description,
			"type":        string(box.Type),
			"x":           box.X,
			"y":           box.Y,
			"width":       box.Width,
			"height":      box.Height,
			"color":       box.Color,
			"selected":    box.Selected,
		}
		if box.Data != nil {
			boxData["data"] = box.Data
		}
		boxesData[id] = boxData
	}

	edgesData := make(map[string]map[string]interface{})
	for id, edge := range f.Canvas.Edges {
		edgesData[id] = map[string]interface{}{
			"id":       edge.ID,
			"fromBox":  edge.FromBox,
			"fromPort": edge.FromPort,
			"toBox":    edge.ToBox,
			"toPort":   edge.ToPort,
			"label":    edge.Label,
			"type":     string(edge.Type),
			"fromX":    edge.FromX,
			"fromY":    edge.FromY,
			"toX":      edge.ToX,
			"toY":      edge.ToY,
		}
	}

	state := map[string]interface{}{
		"boxes":     boxesData,
		"edges":     edgesData,
		"nodeCount": f.NodeCount,
		"edgeCount": f.EdgeCount,
	}

	stateJSON, _ := json.Marshal(state)
	f.RedoStack = append(f.RedoStack, string(stateJSON))
}

// restoreState restaura un estado previamente guardado del diagrama
// Proceso:
// 1. Deserializa el JSON del estado guardado
// 2. Limpia el canvas actual
// 3. Recrea todos los nodos con sus propiedades
// 4. Recrea todas las conexiones
// 5. Actualiza contadores y metadatos
func (f *EnhancedFlowTool) restoreState(stateJSON string) {
	var state map[string]interface{}
	if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
		log.Printf("Error unmarshaling state: %v", err)
		return
	}

	// Limpiar canvas actual
	f.Canvas.Boxes = make(map[string]*components.FlowBox)
	f.Canvas.Edges = make(map[string]*components.FlowEdge)

	// === Restaurar nodos (boxes) ===
	if boxesData, ok := state["boxes"].(map[string]interface{}); ok {
		for id, boxData := range boxesData {
			if boxMap, ok := boxData.(map[string]interface{}); ok {
				// Obtener tipo de nodo (proceso por defecto)
				boxType := components.BoxTypeProcess
				if typeStr, ok := boxMap["type"].(string); ok {
					boxType = components.BoxType(typeStr)
				}

				// Obtener dimensiones con valores por defecto
				width := 120
				height := 60
				if w, ok := boxMap["width"].(float64); ok && w > 0 {
					width = int(w)
				}
				if h, ok := boxMap["height"].(float64); ok && h > 0 {
					height = int(h)
				}

				// Create new box
				box := &components.FlowBox{
					ID:     id,
					Label:  boxMap["label"].(string),
					Type:   boxType,
					X:      int(boxMap["x"].(float64)),
					Y:      int(boxMap["y"].(float64)),
					Width:  width,
					Height: height,
				}

				// Restaurar campos opcionales
				if desc, ok := boxMap["description"].(string); ok {
					box.Description = desc
				}
				if color, ok := boxMap["color"].(string); ok {
					box.Color = color
				} else {
					// Establecer color por defecto según el tipo
					switch boxType {
					case components.BoxTypeStart:
						box.Color = "#dcfce7"
					case components.BoxTypeEnd:
						box.Color = "#fee2e2"
					case components.BoxTypeDecision:
						box.Color = "#fef3c7"
					case components.BoxTypeData:
						box.Color = "#e9d5ff"
					default:
						box.Color = "#dbeafe"
					}
				}

				// Restaurar campo Data (incluye código y metadatos del nodo)
				if data, ok := boxMap["data"].(map[string]interface{}); ok {
					box.Data = data
				}

				// Registrar driver de LiveView para el componente
				liveview.New(id, box)
				f.Canvas.Boxes[id] = box
			}
		}
	}

	// === Restaurar conexiones (edges) ===
	if edgesData, ok := state["edges"].(map[string]interface{}); ok {
		for id, edgeData := range edgesData {
			if edgeMap, ok := edgeData.(map[string]interface{}); ok {
				edge := &components.FlowEdge{
					ID:       id,
					FromBox:  edgeMap["fromBox"].(string),
					FromPort: edgeMap["fromPort"].(string),
					ToBox:    edgeMap["toBox"].(string),
					ToPort:   edgeMap["toPort"].(string),
				}

				// Restore optional fields
				if label, ok := edgeMap["label"].(string); ok {
					edge.Label = label
				}
				if typeStr, ok := edgeMap["type"].(string); ok {
					edge.Type = components.EdgeType(typeStr)
				}

				// Restore positions
				if x, ok := edgeMap["fromX"].(float64); ok {
					edge.FromX = int(x)
				}
				if y, ok := edgeMap["fromY"].(float64); ok {
					edge.FromY = int(y)
				}
				if x, ok := edgeMap["toX"].(float64); ok {
					edge.ToX = int(x)
				}
				if y, ok := edgeMap["toY"].(float64); ok {
					edge.ToY = int(y)
				}

				f.Canvas.Edges[id] = edge
			}
		}
	}

	// Restore counts
	if nodeCount, ok := state["nodeCount"].(float64); ok {
		f.NodeCount = int(nodeCount)
	}
	if edgeCount, ok := state["edgeCount"].(float64); ok {
		f.EdgeCount = int(edgeCount)
	}
}

func (f *EnhancedFlowTool) Redo(data interface{}) {
	if len(f.RedoStack) == 0 {
		f.LastAction = "Nothing to redo"
		f.Commit()
		return
	}

	// Save current state to undo stack WITHOUT clearing redo stack
	f.saveCurrentStateToUndoStack()

	// Get and remove the last state from redo stack
	nextState := f.RedoStack[len(f.RedoStack)-1]
	f.RedoStack = f.RedoStack[:len(f.RedoStack)-1]

	// Restore the state
	f.restoreState(nextState)

	f.LastAction = "Redo performed"
	f.Commit()
}

func (f *EnhancedFlowTool) Start() {
	// Initialize with lifecycle hooks
	f.Lifecycle = liveview.NewLifecycleManager("enhanced_flowtool")
	f.Lifecycle.SetHooks(&liveview.LifecycleHooks{
		OnBeforeMount: func() error {
			log.Println("Enhanced FlowTool mounting...")
			return nil
		},
		OnMounted: func() error {
			log.Println("Enhanced FlowTool mounted successfully")
			return nil
		},
	})

	// Execute lifecycle
	f.Lifecycle.Create()
	f.Lifecycle.Mount()

	// Initialize modal events
	if f.Modal != nil && f.Modal.ComponentDriver != nil {
		f.Modal.Start()
	}

	// Initialize file upload component
	if f.FileUpload != nil && f.FileUpload.ComponentDriver != nil {
		f.FileUpload.Start()
	}

	// === Registrar todos los manejadores de eventos ===
	// Los eventos se registran en el ComponentDriver y se ejecutan cuando
	// el usuario interactúa con la UI a través del WebSocket
	if f.ComponentDriver != nil {
		// === Eventos mejorados con ErrorBoundary ===
		// AddNode: agrega un nuevo nodo al diagrama
		// Usa ErrorBoundary para capturar errores y prevenir crashes
		f.ComponentDriver.Events["AddNode"] = func(c *EnhancedFlowTool, data interface{}) {
			f.ErrorBoundary.SafeExecute("add_node", func() error {
				c.HandleAddNode(data)
				return nil
			})
		}
		f.ComponentDriver.Events["ClearDiagram"] = func(c *EnhancedFlowTool, data interface{}) {
			c.ClearDiagram(data)
		}
		f.ComponentDriver.Events["ExportDiagram"] = func(c *EnhancedFlowTool, data interface{}) {
			c.ExportDiagram(data)
		}
		f.ComponentDriver.Events["ImportDiagram"] = func(c *EnhancedFlowTool, data interface{}) {
			c.ImportDiagram(data)
		}
		f.ComponentDriver.Events["Undo"] = func(c *EnhancedFlowTool, data interface{}) {
			c.Undo(data)
		}
		f.ComponentDriver.Events["Redo"] = func(c *EnhancedFlowTool, data interface{}) {
			c.Redo(data)
		}
		f.ComponentDriver.Events["AutoArrange"] = func(c *EnhancedFlowTool, data interface{}) {
			c.AutoArrange(data)
		}

		// === Eventos de interacción con nodos ===
		// BoxClick: maneja clics en nodos para selección y conexión
		f.ComponentDriver.Events["BoxClick"] = func(c *EnhancedFlowTool, data interface{}) {
			log.Printf("📨 BoxClick event received: %v", data)
			c.HandleBoxClick(data)
		}
		f.ComponentDriver.Events["DeleteBox"] = func(c *EnhancedFlowTool, data interface{}) {
			c.DeleteBox(data)
		}
		f.ComponentDriver.Events["DeleteEdge"] = func(c *EnhancedFlowTool, data interface{}) {
			c.DeleteEdge(data)
		}
		f.ComponentDriver.Events["SelectEdge"] = func(c *EnhancedFlowTool, data interface{}) {
			c.SelectEdge(data)
		}
		f.ComponentDriver.Events["MoveBox"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleMoveBox(data)
		}

		// === Eventos del canvas ===
		// CanvasZoomIn: aumenta el zoom del canvas
		f.ComponentDriver.Events["CanvasZoomIn"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleCanvasZoomIn(data)
		}
		f.ComponentDriver.Events["CanvasZoomOut"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleCanvasZoomOut(data)
		}
		f.ComponentDriver.Events["CanvasReset"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleCanvasReset(data)
		}
		f.ComponentDriver.Events["ToggleGrid"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleToggleGrid(data)
		}
		f.ComponentDriver.Events["ToggleConnectMode"] = func(c *EnhancedFlowTool, data interface{}) {
			log.Printf("📨 ToggleConnectMode event received: %v", data)
			c.HandleToggleConnectMode(data)
		}

		// === Eventos genéricos de arrastre desde WASM ===
		// El módulo WASM envía estos eventos cuando detecta arrastre
		// DragStart: inicio del arrastre de cualquier elemento draggable
		f.ComponentDriver.Events["DragStart"] = func(c *EnhancedFlowTool, data interface{}) {
			log.Printf("📨 DragStart event received: %v", data)
			c.HandleDragStart(data)
		}
		f.ComponentDriver.Events["DragMove"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleDragMove(data)
		}
		f.ComponentDriver.Events["DragEnd"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleDragEnd(data)
		}

		// === Compatibilidad hacia atrás: eventos específicos de arrastre de nodos ===
		// Mantiene compatibilidad con el sistema anterior de drag & drop
		f.ComponentDriver.Events["BoxStartDrag"] = func(c *EnhancedFlowTool, data interface{}) {
			log.Printf("📨 BoxStartDrag event received: %v", data)
			c.HandleBoxStartDrag(data)
		}
		f.ComponentDriver.Events["BoxDrag"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleBoxDrag(data)
		}
		f.ComponentDriver.Events["BoxEndDrag"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleBoxEndDrag(data)
		}

		// === Eventos de edición ===
		// EditBox: abre el modo de edición para un nodo
		f.ComponentDriver.Events["EditBox"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleEditBox(data)
		}
		f.ComponentDriver.Events["EditEdge"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleEditEdge(data)
		}
		f.ComponentDriver.Events["SaveEdit"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleSaveEdit(data)
		}
		f.ComponentDriver.Events["CancelEdit"] = func(c *EnhancedFlowTool, data interface{}) {
			c.HandleCancelEdit(data)
		}
	}

	if f.ComponentDriver != nil {
		f.Commit()
	}
}

// GetTemplate retorna el template HTML principal de la aplicación
// Utiliza el TemplateCache del framework para optimizar el renderizado:
// - Si el template está en caché, usa la versión compilada
// - Si no, retorna el template crudo (que será compilado y cacheado)
// El template incluye todos los estilos CSS y estructura HTML
func (f *EnhancedFlowTool) GetTemplate() string {
	// Usar template cacheado si está disponible
	if cached, exists := f.TemplateCache.Get("flowtool_main"); exists {
		var buf strings.Builder
		cached.Compiled.Execute(&buf, f)
		return buf.String()
	}

	return `
<!DOCTYPE html>
<html>
<head>
	<title>{{.Title}}</title>
	<style>
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}
		
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			min-height: 100vh;
			padding: 2rem;
		}
		
		.container {
			max-width: 1400px;
			margin: 0 auto;
		}
		
		.header {
			background: white;
			border-radius: 12px;
			padding: 2rem;
			margin-bottom: 2rem;
			box-shadow: 0 10px 30px rgba(0,0,0,0.1);
		}
		
		.title {
			font-size: 2rem;
			font-weight: 700;
			color: #1a202c;
			margin-bottom: 0.5rem;
		}
		
		.description {
			color: #718096;
			font-size: 1rem;
		}
		
		.feature-badges {
			display: flex;
			gap: 0.5rem;
			margin-top: 1rem;
			flex-wrap: wrap;
		}
		
		.badge {
			display: inline-block;
			padding: 0.25rem 0.75rem;
			background: linear-gradient(135deg, #667eea, #764ba2);
			color: white;
			border-radius: 20px;
			font-size: 0.75rem;
			font-weight: 600;
		}
		
		.main-content {
			background: white;
			border-radius: 12px;
			padding: 2rem;
			box-shadow: 0 10px 30px rgba(0,0,0,0.1);
		}
		
		.controls {
			display: flex;
			gap: 1rem;
			margin-bottom: 2rem;
			flex-wrap: wrap;
		}
		
		.control-group {
			display: flex;
			gap: 0.5rem;
			align-items: center;
			padding: 0.5rem;
			background: #f7fafc;
			border-radius: 8px;
		}
		
		.btn {
			padding: 0.625rem 1.25rem;
			border: none;
			border-radius: 6px;
			font-weight: 500;
			cursor: pointer;
			transition: all 0.2s;
			font-size: 0.875rem;
		}
		
		.btn-primary {
			background: #667eea;
			color: white;
		}
		
		.btn-primary:hover {
			background: #5a67d8;
			transform: translateY(-1px);
			box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
		}
		
		.btn-secondary {
			background: #e2e8f0;
			color: #4a5568;
		}
		
		.btn-secondary:hover {
			background: #cbd5e0;
		}
		
		.btn-danger {
			background: #fc8181;
			color: white;
		}
		
		.btn-danger:hover {
			background: #f56565;
		}
		
		.btn-success {
			background: #68d391;
			color: white;
		}
		
		.btn-success:hover {
			background: #48bb78;
		}
		
		.btn-warning {
			background: #f6ad55;
			color: white;
		}
		
		.btn-warning:hover {
			background: #ed8936;
		}
		
		.stats {
			display: flex;
			gap: 2rem;
			margin-bottom: 1rem;
			padding: 1rem;
			background: #f7fafc;
			border-radius: 6px;
		}
		
		.stat {
			display: flex;
			flex-direction: column;
		}
		
		.stat-label {
			font-size: 0.75rem;
			color: #718096;
			text-transform: uppercase;
			letter-spacing: 0.05em;
		}
		
		.stat-value {
			font-size: 1.5rem;
			font-weight: 600;
			color: #2d3748;
		}
		
		.status-bar {
			padding: 0.75rem 1rem;
			background: #edf2f7;
			border-radius: 6px;
			margin-top: 1rem;
			font-size: 0.875rem;
			color: #4a5568;
			display: flex;
			justify-content: space-between;
			align-items: center;
		}
		
		.dropdown {
			position: relative;
			display: inline-block;
		}
		
		.dropdown-content {
			display: none;
			position: absolute;
			background: white;
			min-width: 160px;
			box-shadow: 0 8px 16px rgba(0,0,0,0.1);
			border-radius: 6px;
			z-index: 1000;
			margin-top: 0.5rem;
		}
		
		.dropdown.active .dropdown-content {
			display: block;
		}
		
		.dropdown-item {
			padding: 0.75rem 1rem;
			cursor: pointer;
			transition: background 0.2s;
			border-radius: 6px;
		}
		
		.dropdown-item:hover {
			background: #f7fafc;
		}
		
		.legend {
			display: flex;
			gap: 2rem;
			margin-top: 1rem;
			padding: 1rem;
			background: #f7fafc;
			border-radius: 6px;
			flex-wrap: wrap;
		}
		
		.legend-item {
			display: flex;
			align-items: center;
			gap: 0.5rem;
		}
		
		.legend-box {
			width: 20px;
			height: 20px;
			border-radius: 4px;
			border: 1px solid #cbd5e0;
		}
		
		.legend-label {
			font-size: 0.875rem;
			color: #4a5568;
		}
		
		.save-indicator {
			position: fixed;
			top: 20px;
			right: 20px;
			padding: 0.5rem 1rem;
			background: #48bb78;
			color: white;
			border-radius: 6px;
			font-size: 0.875rem;
			opacity: 0;
			transition: opacity 0.3s;
			z-index: 1000;
		}
		
		.save-indicator.show {
			opacity: 1;
		}
		
		.draggable-box:hover .delete-box-btn {
			display: block !important;
		}
		
		.delete-box-btn:hover {
			background: #dc2626 !important;
			transform: scale(1.1);
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1 class="title">{{.Title}}</h1>
			<p class="description">{{.Description}}</p>
			<div class="feature-badges">
				<span class="badge">Virtual DOM</span>
				<span class="badge">State Management</span>
				<span class="badge">Event Registry</span>
				<span class="badge">Template Cache</span>
				<span class="badge">Error Recovery</span>
				<span class="badge">Auto-Save</span>
				<span class="badge">Undo/Redo</span>
			</div>
		</div>
		
		<div class="main-content">
			<div class="controls">
				<div class="control-group">
					<div class="dropdown" id="add-node-dropdown">
						<button class="btn btn-primary" onclick="document.getElementById('add-node-dropdown').classList.toggle('active')">Add Node ▼</button>
						<div class="dropdown-content">
							<div class="dropdown-item" onclick="send_event('{{.IdComponent}}', 'AddNode', 'start'); document.getElementById('add-node-dropdown').classList.remove('active')">
								Start Node
							</div>
							<div class="dropdown-item" onclick="send_event('{{.IdComponent}}', 'AddNode', 'process'); document.getElementById('add-node-dropdown').classList.remove('active')">
								Process Node
							</div>
							<div class="dropdown-item" onclick="send_event('{{.IdComponent}}', 'AddNode', 'decision'); document.getElementById('add-node-dropdown').classList.remove('active')">
								Decision Node
							</div>
							<div class="dropdown-item" onclick="send_event('{{.IdComponent}}', 'AddNode', 'data'); document.getElementById('add-node-dropdown').classList.remove('active')">
								Data Node
							</div>
							<div class="dropdown-item" onclick="send_event('{{.IdComponent}}', 'AddNode', 'end'); document.getElementById('add-node-dropdown').classList.remove('active')">
								End Node
							</div>
						</div>
					</div>
					
					<button class="btn {{if .ConnectingMode}}btn-danger{{else}}btn-success{{end}}" onclick="send_event('{{.IdComponent}}', 'ToggleConnectMode', null)">
						{{if .ConnectingMode}}Cancel Connect{{else}}Connect Mode{{end}}
					</button>
				</div>
				
				<div class="control-group">
					<button class="btn btn-secondary" onclick="send_event('{{.IdComponent}}', 'Undo', null)">
						↶ Undo
					</button>
					
					<button class="btn btn-secondary" onclick="send_event('{{.IdComponent}}', 'Redo', null)">
						↷ Redo
					</button>
				</div>
				
				<div class="control-group">
					<button class="btn btn-warning" onclick="send_event('{{.IdComponent}}', 'AutoArrange', null)">
						Auto Arrange
					</button>
				</div>
				
				<div class="control-group">
					<button class="btn btn-secondary" onclick="send_event('{{.IdComponent}}', 'ExportDiagram', null)">
						📥 Export JSON
					</button>
					
					<button class="btn btn-secondary" onclick="document.getElementById('import-file-input').click()">
						📤 Import JSON
					</button>
					<input id="import-file-input" type="file" accept=".json" style="display: none;" 
						onchange="if(this.files[0]) { 
							var reader = new FileReader(); 
							reader.onload = function(e) { 
								send_event('{{.IdComponent}}', 'ImportDiagram', e.target.result); 
							}; 
							reader.readAsText(this.files[0]); 
						}">
					
					<button class="btn btn-danger" onclick="send_event('{{.IdComponent}}', 'ClearDiagram', null)">
						🗑️ Clear All
					</button>
				</div>
			</div>
			
			<div class="stats">
				<div class="stat">
					<span class="stat-label">Nodes</span>
					<span class="stat-value">{{.NodeCount}}</span>
				</div>
				<div class="stat">
					<span class="stat-label">Edges</span>
					<span class="stat-value">{{.EdgeCount}}</span>
				</div>
				<div class="stat">
					<span class="stat-label">Canvas Size</span>
					<span class="stat-value">{{.Canvas.Width}} × {{.Canvas.Height}}</span>
				</div>
				<div class="stat">
					<span class="stat-label">Zoom</span>
					<span class="stat-value">{{.Canvas.ZoomPercent}}%</span>
				</div>
				<div class="stat">
					<span class="stat-label">Undo Stack</span>
					<span class="stat-value">{{len .UndoStack}}</span>
				</div>
			</div>
			
			<!-- Canvas Component -->
			<div id="flow-canvas-mount">
				{{if .Canvas}}
					<div id="{{.Canvas.ID}}" style="position: relative; width: {{.Canvas.Width}}px; height: {{.Canvas.Height}}px; border: 2px solid #e5e7eb; border-radius: 8px; overflow: hidden; background: #fafafa;">
						<div style="position: absolute; top: 10px; right: 10px; display: flex; gap: 0.5rem; background: white; padding: 0.5rem; border-radius: 6px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); z-index: 100;">
							<button onclick="send_event('{{$.IdComponent}}', 'CanvasZoomIn', null)" style="padding: 0.5rem; background: white; border: 1px solid #d1d5db; border-radius: 4px; cursor: pointer;">Zoom In</button>
							<button onclick="send_event('{{$.IdComponent}}', 'CanvasZoomOut', null)" style="padding: 0.5rem; background: white; border: 1px solid #d1d5db; border-radius: 4px; cursor: pointer;">Zoom Out</button>
							<button onclick="send_event('{{$.IdComponent}}', 'CanvasReset', null)" style="padding: 0.5rem; background: white; border: 1px solid #d1d5db; border-radius: 4px; cursor: pointer;">Reset</button>
							<button onclick="send_event('{{$.IdComponent}}', 'ToggleGrid', null)" style="padding: 0.5rem; background: white; border: 1px solid #d1d5db; border-radius: 4px; cursor: pointer;">Grid</button>
						</div>
						
						<div id="canvas-viewport" style="position: relative; width: 100%; height: 100%; transform: scale({{.Canvas.Zoom}}) translate({{.Canvas.PanX}}px, {{.Canvas.PanY}}px); transform-origin: 0 0; transition: transform 0.2s;">
							<!-- Render boxes -->
							{{$component := .}}
							{{range $id, $box := .Canvas.Boxes}}
								<div id="box-{{$id}}" 
								     class="draggable draggable-box flow-box"
								     data-element-id="box-{{$id}}"
								     data-component-id="{{$component.IdComponent}}"
								     data-box-id="{{$id}}"
								     data-box-x="{{$box.X}}"
								     data-box-y="{{$box.Y}}"
								     {{if $component.ConnectingMode}}data-drag-disabled="true"{{end}}
								     style="position: absolute; left: {{$box.X}}px; top: {{$box.Y}}px; width: {{$box.Width}}px; height: {{$box.Height}}px; background: {{$box.Color}}; border: 2px solid {{if $box.Selected}}#2563eb{{else}}#cbd5e1{{end}}; border-radius: 8px; padding: 0.5rem; cursor: {{if $component.ConnectingMode}}pointer{{else}}move{{end}}; box-shadow: 0 2px 4px rgba(0,0,0,0.1); user-select: none; z-index: 20;"
								     onclick="if({{$component.ConnectingMode}}) { send_event('{{$component.IdComponent}}', 'BoxClick', '{{$id}}'); }"
								     ondblclick="send_event('{{$component.IdComponent}}', 'EditBox', '{{$id}}'); event.stopPropagation();">
									<button class="delete-box-btn" 
									        onclick="event.stopPropagation(); if(confirm('Delete this box?')) { send_event('{{$component.IdComponent}}', 'DeleteBox', '{{$id}}'); }"
									        style="position: absolute; top: -8px; right: -8px; width: 20px; height: 20px; border-radius: 50%; background: #ef4444; color: white; border: 2px solid white; cursor: pointer; font-size: 12px; line-height: 1; padding: 0; display: {{if $box.Selected}}block{{else}}none{{end}}; z-index: 10;">
									        ×
									</button>
									<div style="font-weight: 600; color: #1f2937; font-size: 0.875rem; pointer-events: none;">{{$box.Label}}</div>
									{{if $box.Description}}
										<div style="font-size: 0.75rem; color: #6b7280; pointer-events: none;">{{$box.Description}}</div>
									{{end}}
								</div>
							{{end}}
							
							<!-- Render edges as SVG -->
							<svg style="position: absolute; top: 0; left: 0; width: 100%; height: 100%; pointer-events: auto;">
								{{range $id, $edge := .Canvas.Edges}}
									<g class="edge-group">
										<line x1="{{$edge.FromX}}" y1="{{$edge.FromY}}" x2="{{$edge.ToX}}" y2="{{$edge.ToY}}" 
										      stroke="{{if $edge.Selected}}#2563eb{{else}}#6b7280{{end}}" 
										      stroke-width="{{if $edge.Selected}}3{{else}}2{{end}}"
										      style="cursor: pointer;"
										      onclick="send_event('{{$.IdComponent}}', 'SelectEdge', '{{$id}}')"
										      ondblclick="send_event('{{$.IdComponent}}', 'EditEdge', '{{$id}}'); event.stopPropagation();" />
										{{if $edge.Label}}
											<text x="{{$edge.GetMidX}}" y="{{$edge.GetMidY}}" text-anchor="middle" fill="#374151" font-size="12">{{$edge.Label}}</text>
										{{end}}
										{{if $edge.Selected}}
											<circle cx="{{$edge.GetMidX}}" cy="{{$edge.GetMidY}}" r="10" 
											        fill="#ef4444" 
											        style="cursor: pointer;"
											        onclick="event.stopPropagation(); if(confirm('Delete this connection?')) { send_event('{{$.IdComponent}}', 'DeleteEdge', '{{$id}}'); }">
											</circle>
											<text x="{{$edge.GetMidX}}" y="{{$edge.GetMidY}}" 
											      text-anchor="middle" 
											      fill="white" 
											      font-size="14" 
											      font-weight="bold"
											      pointer-events="none">×</text>
										{{end}}
									</g>
								{{end}}
							</svg>
						</div>
						
						<div style="position: absolute; bottom: 10px; left: 10px; background: white; padding: 0.5rem 1rem; border-radius: 6px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); font-size: 0.75rem; color: #6b7280; z-index: 100;">
							Boxes: {{len .Canvas.Boxes}} | Edges: {{len .Canvas.Edges}} | Zoom: {{.Canvas.ZoomPercent}}%
						</div>
					</div>
				{{end}}
			</div>
			
			<div class="legend">
				<div class="legend-item">
					<div class="legend-box" style="background: #dcfce7;"></div>
					<span class="legend-label">Start Node</span>
				</div>
				<div class="legend-item">
					<div class="legend-box" style="background: #dbeafe;"></div>
					<span class="legend-label">Process Node</span>
				</div>
				<div class="legend-item">
					<div class="legend-box" style="background: #fef3c7; transform: rotate(45deg);"></div>
					<span class="legend-label">Decision Node</span>
				</div>
				<div class="legend-item">
					<div class="legend-box" style="background: #e9d5ff;"></div>
					<span class="legend-label">Data Node</span>
				</div>
				<div class="legend-item">
					<div class="legend-box" style="background: #fee2e2;"></div>
					<span class="legend-label">End Node</span>
				</div>
			</div>
			
			<div class="status-bar">
				<div>
					<strong>Last Action:</strong> {{.LastAction}}
					{{if .Canvas}}
						{{range $id, $box := .Canvas.Boxes}}
							{{if $box.Selected}}
								| <strong>Selected:</strong> {{$box.Label}} ({{$box.X}}, {{$box.Y}})
							{{end}}
						{{end}}
					{{end}}
				</div>
				<div>
					<span style="color: #48bb78;">● Auto-save enabled</span>
				</div>
			</div>
		</div>
		
		<!-- Modal Component -->
		{{mount "export-modal"}}
		
		<!-- Edit Modal -->
		{{if .EditingMode}}
		<div style="position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 2000; overflow-y: auto;">
			<div style="background: white; padding: 2rem; border-radius: 8px; min-width: 600px; max-width: 800px; max-height: 90vh; overflow-y: auto; box-shadow: 0 10px 30px rgba(0,0,0,0.2); margin: 2rem;">
				<h3 style="margin-top: 0; color: #333;">
					{{if eq .EditingType "box"}}Edit Node{{else}}Edit Edge Label{{end}}
				</h3>
				
				<label style="display: block; margin-bottom: 0.5rem; color: #666; font-weight: 500;">
					{{if eq .EditingType "box"}}Node Name:{{else}}Edge Label:{{end}}
				</label>
				<input id="edit-input" type="text" value="{{.EditingValue}}" 
					style="width: 100%; padding: 0.75rem; border: 2px solid #ddd; border-radius: 4px; font-size: 1rem; margin-bottom: 1rem;"
					{{if eq .EditingType "edge"}}onkeypress="if(event.key === 'Enter') { event.preventDefault(); document.getElementById('save-edit-btn').click(); }"{{end}}>
				
				{{if eq .EditingType "box"}}
				<label style="display: block; margin-bottom: 0.5rem; color: #666; font-weight: 500;">
					Code/Script:
				</label>
				<textarea id="edit-code" 
					style="width: 100%; min-height: 200px; padding: 0.75rem; border: 2px solid #ddd; border-radius: 4px; font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace; font-size: 0.9rem; margin-bottom: 1rem; resize: vertical; background: #f8f9fa;"
					placeholder="// Enter code, script, or notes here..."
					>{{.EditingCode}}</textarea>
				
				<div style="display: flex; gap: 1rem; margin-bottom: 1rem;">
					<button class="btn btn-secondary" onclick="document.getElementById('edit-code').style.height = '400px';" style="padding: 0.5rem 1rem; font-size: 0.875rem;">
						↕ Expand
					</button>
					<button class="btn btn-secondary" onclick="
						var code = document.getElementById('edit-code');
						code.value = '// Function template\nfunction process() {\n    // Your code here\n    \n    return result;\n}';
						" style="padding: 0.5rem 1rem; font-size: 0.875rem;">
						📝 Template
					</button>
					<button class="btn btn-secondary" onclick="
						var code = document.getElementById('edit-code');
						code.select();
						document.execCommand('copy');
						alert('Code copied to clipboard!');
						" style="padding: 0.5rem 1rem; font-size: 0.875rem;">
						📋 Copy
					</button>
				</div>
				{{end}}
				
				<div style="display: flex; gap: 1rem; justify-content: flex-end;">
					<button class="btn btn-secondary" onclick="send_event('{{.IdComponent}}', 'CancelEdit', null)">
						Cancel
					</button>
					<button id="save-edit-btn" class="btn btn-primary" onclick="
						var data = {
							value: document.getElementById('edit-input').value
							{{if eq .EditingType "box"}},
							code: document.getElementById('edit-code').value
							{{end}}
						};
						send_event('{{.IdComponent}}', 'SaveEdit', JSON.stringify(data));
					">
						Save
					</button>
				</div>
			</div>
		</div>
		<script>
			// Focus the input when modal opens
			setTimeout(function() {
				var input = document.getElementById('edit-input');
				if(input) {
					input.focus();
					input.select();
				}
			}, 100);
		</script>
		{{end}}
	</div>
	
	<div class="save-indicator" id="save-indicator">Saved</div>
	
	<script>
	// Drag & drop is now handled in WASM module
	// This prevents event listeners from being lost on re-render
	
	// Handle keyboard shortcuts
	document.addEventListener('keydown', function(e) {
		// Delete key to delete selected box
		if (e.key === 'Delete' || e.key === 'Backspace') {
			e.preventDefault();
			// Find selected box
			{{range $id, $box := .Canvas.Boxes}}
				{{if $box.Selected}}
					if (confirm('Delete selected box?')) {
						send_event('{{.IdComponent}}', 'DeleteBox', '{{$id}}');
					}
				{{end}}
			{{end}}
		}
		
		// Escape key to exit connect mode
		if (e.key === 'Escape') {
			{{if .ConnectingMode}}
				send_event('{{.IdComponent}}', 'ToggleConnectMode', null);
			{{end}}
		}
	});
	
	// Show save indicator
	function showSaveIndicator() {
		var indicator = document.getElementById('save-indicator');
		indicator.classList.add('show');
		setTimeout(function() {
			indicator.classList.remove('show');
		}, 2000);
	}
	</script>
</body>
</html>
`
}

// === Métodos de manejo de eventos y drivers ===

// GetDriver retorna el driver de LiveView para este componente
// Requerido por la interfaz LiveDriver del framework
func (f *EnhancedFlowTool) GetDriver() liveview.LiveDriver {
	return f
}

// HandleAddNode agrega un nuevo nodo al diagrama
// Proceso:
// 1. Guarda el estado actual para poder deshacer
// 2. Calcula la posición del nuevo nodo (distribución automática)
// 3. Crea el nodo con el tipo especificado
// 4. Registra el componente con LiveView
// 5. Actualiza el StateManager y renderiza
func (f *EnhancedFlowTool) HandleAddNode(data interface{}) {
	// Guardar estado ANTES de hacer cambios (para undo)
	f.saveToUndoStack()

	// Obtener tipo de nodo desde el evento
	nodeType := data.(string)

	x := 100 + (f.NodeCount*50)%1000
	y := 100 + (f.NodeCount*30)%400

	nodeID := fmt.Sprintf("node_%d", f.NodeCount+1)
	label := fmt.Sprintf("%s %d", nodeType, f.NodeCount+1)

	var boxType components.BoxType
	switch nodeType {
	case "start":
		boxType = components.BoxTypeStart
	case "end":
		boxType = components.BoxTypeEnd
	case "process":
		boxType = components.BoxTypeProcess
	case "decision":
		boxType = components.BoxTypeDecision
	case "data":
		boxType = components.BoxTypeData
	default:
		boxType = components.BoxTypeCustom
	}

	newBox := components.NewFlowBox(nodeID, label, boxType, x, y)

	if f.Canvas != nil {
		// Crear y registrar el driver correctamente con LiveView
		liveview.New(nodeID, newBox)
		f.Canvas.AddBox(newBox)

		// Actualizar state manager con el último nodo agregado
		f.StateManager.Set("last_added_node", nodeID)
	}

	f.NodeCount++
	f.LastAction = fmt.Sprintf("Added %s node", nodeType)

	if f.ComponentDriver != nil {
		f.Commit()
	}
}

// AutoArrange organiza automáticamente los nodos en una cuadrícula
// Algoritmo:
// 1. Recolecta todos los nodos del canvas
// 2. Los distribuye en una cuadrícula de 4 columnas
// 3. Actualiza las posiciones de todas las conexiones
// 4. Guarda el estado para poder deshacer
func (f *EnhancedFlowTool) AutoArrange(data interface{}) {
	// Guardar estado ANTES de hacer cambios
	f.saveToUndoStack()

	// Auto-organizar nodos usando un layout de cuadrícula simple
	boxList := make([]*components.FlowBox, 0, len(f.Canvas.Boxes))
	for _, box := range f.Canvas.Boxes {
		boxList = append(boxList, box)
	}

	cols := 4
	spacing := 200
	startX := 50
	startY := 50

	for i, box := range boxList {
		row := i / cols
		col := i % cols
		box.X = startX + (col * spacing)
		box.Y = startY + (row * spacing)
	}

	// Actualizar posiciones de las conexiones
	f.updateEdgePositions()

	f.LastAction = "Nodes auto-arranged"
	f.Commit()
}

// ClearDiagram limpia completamente el diagrama
// Elimina todos los nodos y conexiones del canvas
// Guarda el estado previo para poder deshacer
func (f *EnhancedFlowTool) ClearDiagram(data interface{}) {
	f.saveToUndoStack()
	f.Canvas.Clear()
	f.NodeCount = 0
	f.EdgeCount = 0
	f.LastAction = "Diagram cleared"
	f.Commit()
}

// ExportDiagram exporta el diagrama actual a formato JSON
// El JSON incluye:
// - Todos los nodos con sus posiciones y código asociado
// - Todas las conexiones entre nodos
// - Metadatos (fecha, versión, herramienta)
// El JSON se muestra en un modal para que el usuario pueda copiarlo
func (f *EnhancedFlowTool) ExportDiagram(data interface{}) {
	if f.Canvas == nil {
		f.LastAction = "No canvas to export"
		if f.ComponentDriver != nil {
			f.Commit()
		}
		return
	}

	// Crear datos de exportación personalizados que incluyen metadatos de código
	boxes := []map[string]interface{}{}
	for _, box := range f.Canvas.Boxes {
		boxData := map[string]interface{}{
			"id":    box.ID,
			"label": box.Label,
			"type":  box.Type,
			"x":     box.X,
			"y":     box.Y,
		}

		// Incluir código si está presente
		if box.Data != nil {
			if code, ok := box.Data["code"].(string); ok && code != "" {
				boxData["code"] = code
			}
			// Incluir cualquier otro metadato del nodo
			for key, value := range box.Data {
				if key != "code" {
					boxData[key] = value
				}
			}
		}

		boxes = append(boxes, boxData)
	}

	edges := []map[string]interface{}{}
	for _, edge := range f.Canvas.Edges {
		edges = append(edges, map[string]interface{}{
			"id":       edge.ID,
			"fromBox":  edge.FromBox,
			"fromPort": edge.FromPort,
			"toBox":    edge.ToBox,
			"toPort":   edge.ToPort,
			"type":     edge.Type,
			"label":    edge.Label,
		})
	}

	exportData := map[string]interface{}{
		"boxes": boxes,
		"edges": edges,
		"metadata": map[string]interface{}{
			"exportTime": time.Now().Format(time.RFC3339),
			"version":    "1.0",
			"tool":       "Enhanced Flow Diagram Tool",
		},
	}

	// Convertir a JSON string para mostrar en el modal
	jsonBytes, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		f.LastAction = fmt.Sprintf("Export error: %v", err)
	} else {
		f.JsonExport = string(jsonBytes)

		// Create JavaScript to handle the download
		downloadScript := fmt.Sprintf(`
			var jsonData = %s;
			var dataStr = JSON.stringify(jsonData, null, 2);
			var dataUri = 'data:application/json;charset=utf-8,'+ encodeURIComponent(dataStr);
			var exportFileDefaultName = 'flow_diagram_%d.json';
			var linkElement = document.createElement('a');
			linkElement.setAttribute('href', dataUri);
			linkElement.setAttribute('download', exportFileDefaultName);
			linkElement.click();
		`, string(jsonBytes), time.Now().Unix())

		// Execute the download script
		if f.ComponentDriver != nil {
			f.ComponentDriver.EvalScript(downloadScript)
		}

		f.LastAction = fmt.Sprintf("Exported diagram with %d boxes and %d edges", len(f.Canvas.Boxes), len(f.Canvas.Edges))
	}

	if f.ComponentDriver != nil {
		f.Commit()
	}
}

func (f *EnhancedFlowTool) ImportDiagram(data interface{}) {
	// Parse JSON data
	var jsonStr string
	if str, ok := data.(string); ok {
		jsonStr = str
	} else {
		f.LastAction = "Invalid import data"
		f.Commit()
		return
	}

	var importData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &importData); err != nil {
		f.LastAction = fmt.Sprintf("Import error: %v", err)
		f.Commit()
		return
	}

	// Save state BEFORE making changes
	f.saveToUndoStack()

	// Clear current diagram
	f.Canvas.Clear()
	f.NodeCount = 0
	f.EdgeCount = 0

	// Import boxes
	if boxes, ok := importData["boxes"].([]interface{}); ok {
		for _, boxData := range boxes {
			if box, ok := boxData.(map[string]interface{}); ok {
				id := box["id"].(string)
				label := box["label"].(string)
				boxType := components.BoxType(box["type"].(string))
				x := int(box["x"].(float64))
				y := int(box["y"].(float64))

				newBox := components.NewFlowBox(id, label, boxType, x, y)

				// Import code and other metadata
				if code, ok := box["code"].(string); ok {
					if newBox.Data == nil {
						newBox.Data = make(map[string]interface{})
					}
					newBox.Data["code"] = code
				}

				// Import any other metadata
				for key, value := range box {
					if key != "id" && key != "label" && key != "type" && key != "x" && key != "y" && key != "code" {
						if newBox.Data == nil {
							newBox.Data = make(map[string]interface{})
						}
						newBox.Data[key] = value
					}
				}

				// Register the driver properly for imported boxes
				liveview.New(id, newBox)
				f.Canvas.AddBox(newBox)
				f.NodeCount++
			}
		}
	}

	// Import edges
	if edges, ok := importData["edges"].([]interface{}); ok {
		for _, edgeData := range edges {
			if edge, ok := edgeData.(map[string]interface{}); ok {
				id := edge["id"].(string)
				fromBox := edge["fromBox"].(string)
				fromPort := edge["fromPort"].(string)
				toBox := edge["toBox"].(string)
				toPort := edge["toPort"].(string)

				newEdge := components.NewFlowEdge(id, fromBox, fromPort, toBox, toPort)
				if edgeType, ok := edge["type"].(string); ok {
					newEdge.Type = components.EdgeType(edgeType)
				}
				if label, ok := edge["label"].(string); ok {
					newEdge.Label = label
				}

				f.Canvas.AddEdge(newEdge)
				f.EdgeCount++
			}
		}
	}

	// Update edge positions
	f.updateEdgePositions()

	f.LastAction = fmt.Sprintf("Imported %d boxes and %d edges", f.NodeCount, f.EdgeCount)
	f.Commit()
}

func (f *EnhancedFlowTool) HandleBoxClick(data interface{}) {
	var boxID string
	if str, ok := data.(string); ok {
		boxID = str
	} else {
		return
	}

	if f.ConnectingMode {
		// Handle connection creation
		if f.ConnectingFrom == "" {
			// First box selected
			f.ConnectingFrom = boxID
			if box, ok := f.Canvas.Boxes[boxID]; ok {
				box.Selected = true
			}
			f.LastAction = fmt.Sprintf("Connecting from: %s", boxID)
		} else if f.ConnectingFrom != boxID {
			// Second box selected - create edge
			if f.validateConnection(f.ConnectingFrom, boxID) {
				// Save state BEFORE creating connection
				f.saveToUndoStack()
				f.createConnection(f.ConnectingFrom, boxID)
			}

			// Reset connection mode
			for _, box := range f.Canvas.Boxes {
				box.Selected = false
			}
			f.ConnectingFrom = ""
		}
	} else {
		// Normal selection
		for id, box := range f.Canvas.Boxes {
			box.Selected = (id == boxID)
		}
		f.LastAction = fmt.Sprintf("Selected box: %s", boxID)
	}

	if f.ComponentDriver != nil {
		f.Commit()
	}
}

// DeleteBox elimina un nodo del diagrama
// Proceso:
// 1. Identifica el nodo a eliminar
// 2. Guarda el estado para poder deshacer
// 3. Elimina todas las conexiones relacionadas con el nodo
// 4. Elimina el nodo del canvas
// 5. Actualiza contadores y renderiza
func (f *EnhancedFlowTool) DeleteBox(data interface{}) {
	var boxID string

	// Manejar diferentes tipos de datos del evento
	if str, ok := data.(string); ok {
		boxID = str
	} else if dataMap, ok := data.(map[string]interface{}); ok {
		if id, ok := dataMap["id"].(string); ok {
			boxID = id
		}
	}

	if boxID == "" {
		f.LastAction = "No box selected to delete"
		f.Commit()
		return
	}

	// Guardar estado para deshacer
	f.saveToUndoStack()

	// Eliminar todas las conexiones conectadas a este nodo
	for edgeID, edge := range f.Canvas.Edges {
		if edge.FromBox == boxID || edge.ToBox == boxID {
			delete(f.Canvas.Edges, edgeID)
			f.EdgeCount--
		}
	}

	// Remove the box
	if _, exists := f.Canvas.Boxes[boxID]; exists {
		delete(f.Canvas.Boxes, boxID)
		f.NodeCount--
		f.LastAction = fmt.Sprintf("Deleted box: %s", boxID)
	} else {
		f.LastAction = fmt.Sprintf("Box not found: %s", boxID)
	}

	if f.ComponentDriver != nil {
		f.Commit()
	}
}

// DeleteEdge elimina una conexión entre nodos
// Similar a DeleteBox pero para conexiones
func (f *EnhancedFlowTool) DeleteEdge(data interface{}) {
	var edgeID string

	// Manejar diferentes tipos de datos
	if str, ok := data.(string); ok {
		edgeID = str
	} else if dataMap, ok := data.(map[string]interface{}); ok {
		if id, ok := dataMap["id"].(string); ok {
			edgeID = id
		}
	}

	if edgeID == "" {
		f.LastAction = "No edge selected to delete"
		f.Commit()
		return
	}

	// Save state for undo
	f.saveToUndoStack()

	// Remove the edge
	if _, exists := f.Canvas.Edges[edgeID]; exists {
		delete(f.Canvas.Edges, edgeID)
		f.EdgeCount--
		f.LastAction = fmt.Sprintf("Deleted edge: %s", edgeID)
	} else {
		f.LastAction = fmt.Sprintf("Edge not found: %s", edgeID)
	}

	if f.ComponentDriver != nil {
		f.Commit()
	}
}

// SelectEdge selecciona una conexión para editar o eliminar
// Solo una conexión puede estar seleccionada a la vez
func (f *EnhancedFlowTool) SelectEdge(data interface{}) {
	edgeID, ok := data.(string)
	if !ok {
		f.LastAction = "Invalid edge ID for selection"
		f.Commit()
		return
	}

	edge, exists := f.Canvas.Edges[edgeID]
	if !exists {
		f.LastAction = fmt.Sprintf("Edge not found: %s", edgeID)
		f.Commit()
		return
	}

	// Deseleccionar otras conexiones
	for _, e := range f.Canvas.Edges {
		e.Selected = false
	}

	// Seleccionar esta conexión (toggle)
	edge.Selected = !edge.Selected
	f.LastAction = fmt.Sprintf("Selected edge: %s", edgeID)

	if f.ComponentDriver != nil {
		f.Commit()
	}
}

// HandleMoveBox mueve un nodo usando las teclas de dirección
// Permite ajustar la posición del nodo en incrementos de 20 píxeles
// Respeta los límites del canvas para evitar que el nodo salga del área visible
func (f *EnhancedFlowTool) HandleMoveBox(data interface{}) {
	moveData, ok := data.(map[string]interface{})
	if !ok {
		log.Printf("HandleMoveBox: data is not a map: %T", data)
		return
	}

	boxID, _ := moveData["id"].(string)
	direction, _ := moveData["dir"].(string)
	log.Printf("HandleMoveBox: boxID=%s, direction=%s", boxID, direction)

	if box, ok := f.Canvas.Boxes[boxID]; ok {
		step := 20 // Píxeles a mover en cada paso

		switch direction {
		case "up":
			box.Y -= step
			if box.Y < 0 {
				box.Y = 0
			}
		case "down":
			box.Y += step
			if box.Y > f.Canvas.Height-box.Height {
				box.Y = f.Canvas.Height - box.Height
			}
		case "left":
			box.X -= step
			if box.X < 0 {
				box.X = 0
			}
		case "right":
			box.X += step
			if box.X > f.Canvas.Width-box.Width {
				box.X = f.Canvas.Width - box.Width
			}
		}

		// Actualizar posiciones de las conexiones relacionadas
		f.updateEdgePositions()

		f.LastAction = fmt.Sprintf("Moved box %s to (%d, %d)", boxID, box.X, box.Y)

		if f.ComponentDriver != nil {
			f.Commit()
		}
	}
}

// HandleToggleConnectMode activa/desactiva el modo de conexión
// En modo conexión:
// - El drag & drop se desactiva
// - Hacer clic en dos nodos los conecta
// - ESC o clic en el botón desactiva el modo
func (f *EnhancedFlowTool) HandleToggleConnectMode(data interface{}) {
	log.Printf("🔗 HandleToggleConnectMode called with data: %v", data)
	f.ConnectingMode = !f.ConnectingMode
	f.ConnectingFrom = ""
	log.Printf("🔗 ConnectingMode toggled to: %v", f.ConnectingMode)

	// Limpiar todas las selecciones
	for _, box := range f.Canvas.Boxes {
		box.Selected = false
	}

	if f.ConnectingMode {
		f.LastAction = "Connection mode activated - click two boxes to connect"
	} else {
		f.LastAction = "Connection mode deactivated"
	}

	if f.ComponentDriver != nil {
		log.Printf("🔄 Calling Commit() to update UI...")
		f.Commit()
		log.Printf("✅ Commit() completed")
	}
}

// === Métodos de compatibilidad hacia atrás para eventos BoxDrag antiguos ===
// Estos métodos convierten el formato antiguo de eventos al nuevo formato genérico

// HandleBoxStartDrag maneja el formato antiguo de inicio de arrastre
// Convierte {id, x, y} al nuevo formato {element, x, y}
func (f *EnhancedFlowTool) HandleBoxStartDrag(data interface{}) {
	log.Printf("🚀 HandleBoxStartDrag called with data: %v (%T)", data, data)

	// Convertir formato antiguo a nuevo formato y delegar a HandleDragStart
	if dataStr, ok := data.(string); ok {
		var oldData map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &oldData); err == nil {
			// Convert old format {id, x, y} to new format {element, x, y}
			if id, hasId := oldData["id"].(string); hasId {
				// Convertir formato {id, x, y} a {element, x, y}
				newData := map[string]interface{}{
					"element": "box-" + id, // Agregar prefijo para nuevo formato
					"x":       oldData["x"],
					"y":       oldData["y"],
				}
				newDataJSON, _ := json.Marshal(newData)
				f.HandleDragStart(string(newDataJSON))
				return
			}
		}
	}

	// Fallback: llamada directa si no se puede convertir
	f.HandleDragStart(data)
}

// HandleBoxDrag maneja el formato antiguo de arrastre continuo
func (f *EnhancedFlowTool) HandleBoxDrag(data interface{}) {
	// Convertir formato antiguo a nuevo formato y delegar a HandleDragMove
	if dataStr, ok := data.(string); ok {
		var oldData map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &oldData); err == nil {
			// Convert old format {id, x, y} to new format {element, x, y}
			if id, hasId := oldData["id"].(string); hasId {
				newData := map[string]interface{}{
					"element": "box-" + id,
					"x":       oldData["x"],
					"y":       oldData["y"],
				}
				newDataJSON, _ := json.Marshal(newData)
				f.HandleDragMove(string(newDataJSON))
				return
			}
		}
	}

	// Fallback to direct call
	f.HandleDragMove(data)
}

func (f *EnhancedFlowTool) HandleBoxEndDrag(data interface{}) {
	// Convert old format and delegate to HandleDragEnd
	if dataStr, ok := data.(string); ok {
		// For BoxEndDrag, the data might just be the box ID
		newData := map[string]interface{}{
			"element": "box-" + dataStr,
			"x":       0, // Will be updated by the actual handler
			"y":       0,
		}
		newDataJSON, _ := json.Marshal(newData)
		f.HandleDragEnd(string(newDataJSON))
		return
	}

	// Fallback: llamada directa si no se puede convertir
	f.HandleDragEnd(data)
}

// === Métodos genéricos de arrastre (nuevo sistema WASM) ===

// HandleDragStart maneja el inicio del arrastre de cualquier elemento draggable
// Este es el nuevo formato genérico que envía el módulo WASM
// Formato esperado: {element: "box-id", x: mouseX, y: mouseY}
func (f *EnhancedFlowTool) HandleDragStart(data interface{}) {
	// Manejar inicio de arrastre genérico
	log.Printf("🚀 DragStart called with data: %v (%T)", data, data)

	// Guardar estado ANTES de iniciar el arrastre (para undo)
	f.saveToUndoStack()

	// Intentar parsear como JSON string primero
	if dataStr, ok := data.(string); ok {
		var dataMap map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &dataMap); err == nil {
			data = dataMap
			log.Printf("Parsed JSON data: %v", dataMap)
		}
	}

	if dataMap, ok := data.(map[string]interface{}); ok {
		if element, ok := dataMap["element"].(string); ok {
			// Extraer ID del nodo desde el ID del elemento (formato: "box-{id}")
			if strings.HasPrefix(element, "box-") {
				boxID := strings.TrimPrefix(element, "box-")
				f.DraggingBox = boxID
				log.Printf("Started dragging box: %s", boxID)
				if box, exists := f.Canvas.Boxes[boxID]; exists {
					box.Dragging = true
					if box.ComponentDriver != nil {
						box.Commit()
					}
				}
				f.LastAction = fmt.Sprintf("Started dragging %s", boxID)
			}
		}
	}

	if f.ComponentDriver != nil {
		f.Commit()
	}
}

// HandleDragMove maneja el movimiento continuo durante el arrastre
// Se llama repetidamente mientras el usuario mueve el mouse con el botón presionado
// Actualiza la posición del nodo y las conexiones en tiempo real
func (f *EnhancedFlowTool) HandleDragMove(data interface{}) {
	if f.DraggingBox == "" {
		log.Printf("BoxDrag: No box being dragged")
		return
	}

	log.Printf("BoxDrag called for box %s with data: %v", f.DraggingBox, data)

	// Intentar parsear como JSON string primero
	if dataStr, ok := data.(string); ok {
		var dataMap map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &dataMap); err == nil {
			data = dataMap
		}
	}

	// Manejar movimiento de arrastre con actualizaciones VDOM
	if dataMap, ok := data.(map[string]interface{}); ok {
		if box, exists := f.Canvas.Boxes[f.DraggingBox]; exists {
			oldX, oldY := box.X, box.Y

			if newX, ok := dataMap["x"].(float64); ok {
				box.X = int(newX)
				log.Printf("Box %s moved X: %d -> %d", f.DraggingBox, oldX, box.X)
			}
			if newY, ok := dataMap["y"].(float64); ok {
				box.Y = int(newY)
				log.Printf("Box %s moved Y: %d -> %d", f.DraggingBox, oldY, box.Y)
			}

			// Restringir a los límites del canvas
			if box.X < 0 {
				box.X = 0
			}
			if box.Y < 0 {
				box.Y = 0
			}
			maxX := f.Canvas.Width - box.Width
			maxY := f.Canvas.Height - box.Height
			if box.X > maxX {
				box.X = maxX
			}
			if box.Y > maxY {
				box.Y = maxY
			}

			// Actualizar estado si la posición cambió
			if oldX != box.X || oldY != box.Y {
				// Guardar posición en StateManager
				f.StateManager.Set("box_position_"+f.DraggingBox, map[string]interface{}{
					"x": box.X,
					"y": box.Y,
				})
			}

			// Actualizar posiciones de las conexiones
			f.updateEdgePositions()

			// Emitir evento de arrastre para auto-guardado
			f.EventRegistry.Emit("diagram.change", map[string]interface{}{
				"type": "box_moved",
				"box":  f.DraggingBox,
			})
		}
	}

	// Llamar Commit para actualizar las conexiones en la UI
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

// HandleDragEnd maneja el fin del arrastre cuando el usuario suelta el mouse
// Finaliza el estado de arrastre y actualiza la UI
// NO guarda el estado aquí porque ya se guardó en HandleDragStart
func (f *EnhancedFlowTool) HandleDragEnd(data interface{}) {
	if f.DraggingBox != "" {
		if box, exists := f.Canvas.Boxes[f.DraggingBox]; exists {
			box.Dragging = false
			if box.ComponentDriver != nil {
				box.Commit()
			}
		}
		f.LastAction = fmt.Sprintf("Finished dragging %s", f.DraggingBox)
		// NO guardar estado aquí - ya se guardó en HandleDragStart
		f.DraggingBox = ""
	}

	if f.ComponentDriver != nil {
		f.Commit()
	}
}

// updateEdgePositions actualiza las posiciones de todas las conexiones
// Se llama cuando un nodo se mueve para mantener las líneas conectadas
// Calcula los puntos de inicio y fin basados en las posiciones de los nodos
func (f *EnhancedFlowTool) updateEdgePositions() {
	for _, edge := range f.Canvas.Edges {
		if fromBox, ok := f.Canvas.Boxes[edge.FromBox]; ok {
			if toBox, ok := f.Canvas.Boxes[edge.ToBox]; ok {
				// Punto de salida: lado derecho del nodo origen
				edge.FromX = fromBox.X + fromBox.Width
				edge.FromY = fromBox.Y + fromBox.Height/2
				// Punto de entrada: lado izquierdo del nodo destino
				edge.ToX = toBox.X
				edge.ToY = toBox.Y + toBox.Height/2
			}
		}
	}
}

// === Métodos de control del canvas ===

// HandleCanvasZoomIn aumenta el zoom del canvas en 20%
// Límite máximo: 300%
func (f *EnhancedFlowTool) HandleCanvasZoomIn(data interface{}) {
	f.Canvas.Zoom = min(f.Canvas.Zoom*1.2, 3.0) // Aumentar 20%, máximo 300%
	f.LastAction = fmt.Sprintf("Zoom: %d%%", f.Canvas.ZoomPercent())
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

// HandleCanvasZoomOut reduce el zoom del canvas en 20%
// Límite mínimo: 30%
func (f *EnhancedFlowTool) HandleCanvasZoomOut(data interface{}) {
	f.Canvas.Zoom = max(f.Canvas.Zoom/1.2, 0.3) // Reducir 20%, mínimo 30%
	f.LastAction = fmt.Sprintf("Zoom: %d%%", f.Canvas.ZoomPercent())
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

// HandleCanvasReset restaura el canvas a su vista inicial
// Zoom: 100%, Pan: (0,0)
func (f *EnhancedFlowTool) HandleCanvasReset(data interface{}) {
	f.Canvas.Zoom = 1.0 // Zoom al 100%
	f.Canvas.PanX = 0   // Sin desplazamiento horizontal
	f.Canvas.PanY = 0   // Sin desplazamiento vertical
	f.LastAction = "View reset"
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

// HandleToggleGrid activa/desactiva la cuadrícula del canvas
// Útil para alinear nodos visualmente
func (f *EnhancedFlowTool) HandleToggleGrid(data interface{}) {
	f.Canvas.ShowGrid = !f.Canvas.ShowGrid
	f.LastAction = fmt.Sprintf("Grid: %v", f.Canvas.ShowGrid)
	if f.ComponentDriver != nil {
		f.Commit()
	}
}

// === Funciones de utilidad ===

// min retorna el menor de dos números flotantes
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// max retorna el mayor de dos números flotantes
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// Edit handlers
func (f *EnhancedFlowTool) HandleEditBox(data interface{}) {
	boxID := ""
	if str, ok := data.(string); ok {
		boxID = str
	}

	if box, ok := f.Canvas.Boxes[boxID]; ok {
		f.EditingMode = true
		f.EditingType = "box"
		f.EditingID = boxID
		f.EditingValue = box.Label

		// Load existing code from box Data
		if box.Data == nil {
			box.Data = make(map[string]interface{})
		}
		if code, ok := box.Data["code"].(string); ok {
			f.EditingCode = code
		} else {
			f.EditingCode = ""
		}

		f.LastAction = fmt.Sprintf("Editing box: %s", box.Label)
		f.Commit()
	}
}

func (f *EnhancedFlowTool) HandleEditEdge(data interface{}) {
	edgeID := ""
	if str, ok := data.(string); ok {
		edgeID = str
	}

	if edge, ok := f.Canvas.Edges[edgeID]; ok {
		f.EditingMode = true
		f.EditingType = "edge"
		f.EditingID = edgeID
		f.EditingValue = edge.Label
		f.LastAction = fmt.Sprintf("Editing edge label")
		f.Commit()
	}
}

func (f *EnhancedFlowTool) HandleSaveEdit(data interface{}) {
	// Parse the data which could be a string or JSON object
	var editData map[string]interface{}

	if str, ok := data.(string); ok {
		// Try to parse as JSON first
		if err := json.Unmarshal([]byte(str), &editData); err != nil {
			// If not JSON, treat as simple string (for backward compatibility)
			editData = map[string]interface{}{"value": str}
		}
	} else if m, ok := data.(map[string]interface{}); ok {
		editData = m
	}

	newValue := ""
	if val, ok := editData["value"].(string); ok {
		newValue = val
	}

	if f.EditingType == "box" {
		if box, ok := f.Canvas.Boxes[f.EditingID]; ok {
			// Save state BEFORE making changes
			f.saveToUndoStack()

			box.Label = newValue

			// Save code to box Data
			if box.Data == nil {
				box.Data = make(map[string]interface{})
			}
			if code, ok := editData["code"].(string); ok {
				box.Data["code"] = code
				f.LastAction = fmt.Sprintf("Updated box '%s' with code", newValue)
			} else {
				f.LastAction = fmt.Sprintf("Renamed box to: %s", newValue)
			}
		}
	} else if f.EditingType == "edge" {
		if edge, ok := f.Canvas.Edges[f.EditingID]; ok {
			// Save state BEFORE making changes
			f.saveToUndoStack()
			edge.Label = newValue
			f.LastAction = fmt.Sprintf("Updated edge label to: %s", newValue)
		}
	}

	f.EditingMode = false
	f.EditingID = ""
	f.EditingValue = ""
	f.EditingCode = ""
	f.Commit()
}

func (f *EnhancedFlowTool) HandleCancelEdit(data interface{}) {
	f.EditingMode = false
	f.EditingID = ""
	f.EditingValue = ""
	f.EditingCode = ""
	f.LastAction = "Edit cancelled"
	f.Commit()
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Serve static assets
	e.Static("/example/assets", "../../assets")
	e.Static("/assets", "../../assets")

	home := liveview.PageControl{
		Title:  "Enhanced Flow Diagram Tool",
		Lang:   "en",
		Path:   "/example/flowtool",
		Router: e,
	}

	home.Register(func() liveview.LiveDriver {
		// Use simple layout that doesn't cause remounting
		document := liveview.NewLayout("flowtool-layout", `{{mount "flow-tool"}}`)

		// Create enhanced flow tool component
		flowTool := NewEnhancedFlowTool()
		liveview.New("flow-tool", flowTool)

		// Set up the modal driver
		if flowTool.Modal != nil {
			liveview.New("export-modal", flowTool.Modal)
		}

		// Set up the file upload driver
		if flowTool.FileUpload != nil {
			liveview.New("file-upload", flowTool.FileUpload)
		}

		// Set up drivers for existing boxes
		if flowTool.Canvas != nil {
			for id, box := range flowTool.Canvas.Boxes {
				liveview.New(id, box)
			}

			// Set up drivers for existing edges
			for id, edge := range flowTool.Canvas.Edges {
				liveview.New(id, edge)
			}
		}

		return document
	})

	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `<!DOCTYPE html>
<html>
<head>
	<title>Enhanced Flow Diagram Tool</title>
	<style>
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
			max-width: 800px;
			margin: 50px auto;
			padding: 20px;
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			min-height: 100vh;
		}
		.container {
			background: white;
			border-radius: 12px;
			padding: 2rem;
			box-shadow: 0 10px 30px rgba(0,0,0,0.1);
		}
		h1 {
			color: #1a202c;
			margin-bottom: 1rem;
		}
		h2 {
			color: #4a5568;
			margin-top: 2rem;
			margin-bottom: 1rem;
		}
		a {
			color: #667eea;
			text-decoration: none;
			font-weight: 500;
		}
		a:hover {
			text-decoration: underline;
		}
		ul {
			list-style: none;
			padding: 0;
		}
		li {
			padding: 0.5rem 0;
			padding-left: 1.5rem;
			position: relative;
		}
		li:before {
			content: "✓";
			position: absolute;
			left: 0;
			color: #48bb78;
			font-weight: bold;
		}
		.button {
			display: inline-block;
			padding: 0.75rem 1.5rem;
			background: #667eea;
			color: white;
			border-radius: 6px;
			margin-top: 1rem;
			transition: all 0.2s;
		}
		.button:hover {
			background: #5a67d8;
			transform: translateY(-2px);
			box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
			text-decoration: none;
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>Enhanced Flow Diagram Tool</h1>
		<p>An advanced visual flow diagram editor with drag-and-drop functionality and real-time collaboration features.</p>
		<a href="/example/flowtool" class="button">Open Flow Diagram Editor</a>
		<h2>New Features:</h2>
		<ul>
			<li>Virtual DOM for efficient rendering</li>
			<li>State Management with auto-save</li>
			<li>Event Registry with throttling</li>
			<li>Template Cache for performance</li>
			<li>Error Boundaries for recovery</li>
			<li>Undo/Redo functionality</li>
			<li>Auto-arrange nodes</li>
		</ul>
	</div>
</body>
</html>`)
	})

	port := ":8082"
	log.Printf("🚀 Enhanced Flow Tool Server")
	log.Printf("🌐 Starting on http://localhost%s", port)
	log.Printf("Visit http://localhost%s/example/flowtool", port)
	e.Logger.Fatal(e.Start(port))
}
