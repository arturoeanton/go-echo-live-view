package liveview

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// MaxMessageSize límite máximo de tamaño de mensaje WebSocket (1MB)
	MaxMessageSize = 1024 * 1024
	// MaxTemplateSize límite máximo de tamaño de template (500KB)
	MaxTemplateSize = 500 * 1024
	// MaxPathDepth profundidad máxima de rutas
	MaxPathDepth = 10
	// MaxEventNameLength longitud máxima del nombre de evento
	MaxEventNameLength = 100
	// MaxFieldCount número máximo de campos en un mensaje
	MaxFieldCount = 100
)

var (
	// ErrMessageTooLarge error cuando el mensaje excede el tamaño máximo
	ErrMessageTooLarge = errors.New("message size exceeds maximum allowed")
	// ErrTemplateTooLarge error cuando el template excede el tamaño máximo
	ErrTemplateTooLarge = errors.New("template size exceeds maximum allowed")
	// ErrInvalidPath error cuando la ruta contiene path traversal
	ErrInvalidPath = errors.New("invalid path: potential path traversal detected")
	// ErrInvalidEventName error cuando el nombre del evento es inválido
	ErrInvalidEventName = errors.New("invalid event name")
	// ErrTooManyFields error cuando hay demasiados campos
	ErrTooManyFields = errors.New("message contains too many fields")
	// ErrInvalidJSON error cuando el JSON es inválido
	ErrInvalidJSON = errors.New("invalid JSON structure")
)

// SecurityConfig configuración de seguridad
type SecurityConfig struct {
	MaxMessageSize     int
	MaxTemplateSize    int
	EnableSanitization bool
	AllowedEvents      map[string]bool
	RestrictedPaths    []string
}

var defaultSecurityConfig = &SecurityConfig{
	MaxMessageSize:     MaxMessageSize,
	MaxTemplateSize:    MaxTemplateSize,
	EnableSanitization: true,
	AllowedEvents:      make(map[string]bool),
	RestrictedPaths:    []string{"/etc", "/proc", "/sys", "/dev"},
}

// WebSocketMessage estructura validada de mensaje WebSocket
type WebSocketMessage struct {
	Type    string                 `json:"type"`
	ID      string                 `json:"id"`
	Event   string                 `json:"event"`
	Data    interface{}            `json:"data"`
	IdRet   string                 `json:"id_ret,omitempty"`
	SubType string                 `json:"sub_type,omitempty"`
}

// ValidateWebSocketMessage valida un mensaje WebSocket entrante
func ValidateWebSocketMessage(msgBytes []byte) (*WebSocketMessage, error) {
	// Verificar tamaño del mensaje
	if len(msgBytes) > defaultSecurityConfig.MaxMessageSize {
		return nil, ErrMessageTooLarge
	}
	
	// Verificar estructura JSON válida
	var msg WebSocketMessage
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	
	// Validar tipo de mensaje
	validTypes := map[string]bool{
		"data":   true,
		"get":    true,
		"fill":   true,
		"remove": true,
		"set":    true,
		"script": true,
		"style":  true,
		"text":   true,
		"propertie": true,
		"addNode": true,
	}
	
	if !validTypes[msg.Type] {
		return nil, fmt.Errorf("invalid message type: %s", msg.Type)
	}
	
	// Validar ID (prevenir inyección)
	if !isValidID(msg.ID) {
		return nil, fmt.Errorf("invalid ID format: %s", msg.ID)
	}
	
	// Validar nombre de evento
	if msg.Event != "" && !isValidEventName(msg.Event) {
		return nil, ErrInvalidEventName
	}
	
	// Validar estructura de datos
	if msg.Data != nil {
		if err := validateDataStructure(msg.Data); err != nil {
			return nil, err
		}
	}
	
	return &msg, nil
}

// isValidID valida que un ID solo contenga caracteres seguros
func isValidID(id string) bool {
	if id == "" || len(id) > 100 {
		return false
	}
	// Solo permitir alfanuméricos, guiones y underscores
	validID := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return validID.MatchString(id)
}

// isValidEventName valida el nombre del evento
func isValidEventName(event string) bool {
	if event == "" || len(event) > MaxEventNameLength {
		return false
	}
	// Solo permitir nombres de eventos alfanuméricos con CamelCase
	validEvent := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`)
	return validEvent.MatchString(event)
}

// validateDataStructure valida la estructura de datos
func validateDataStructure(data interface{}) error {
	switch v := data.(type) {
	case map[string]interface{}:
		if len(v) > MaxFieldCount {
			return ErrTooManyFields
		}
		for key, value := range v {
			if !isValidFieldName(key) {
				return fmt.Errorf("invalid field name: %s", key)
			}
			if err := validateDataStructure(value); err != nil {
				return err
			}
		}
	case []interface{}:
		if len(v) > MaxFieldCount {
			return ErrTooManyFields
		}
		for _, item := range v {
			if err := validateDataStructure(item); err != nil {
				return err
			}
		}
	case string:
		// Strings are validated when used
	case float64, int, bool, nil:
		// Primitive types are safe
	default:
		return fmt.Errorf("unsupported data type: %T", v)
	}
	return nil
}

// isValidFieldName valida nombres de campos
func isValidFieldName(name string) bool {
	if name == "" || len(name) > 100 {
		return false
	}
	validField := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return validField.MatchString(name)
}

// SanitizeTemplate sanitiza un template HTML
func SanitizeTemplate(template string) (string, error) {
	// Verificar tamaño
	if len(template) > defaultSecurityConfig.MaxTemplateSize {
		return "", ErrTemplateTooLarge
	}
	
	if !defaultSecurityConfig.EnableSanitization {
		return template, nil
	}
	
	// Escapar HTML peligroso pero preservar Go templates
	sanitized := template
	
	// Proteger Go template tags temporalmente
	goTemplateRegex := regexp.MustCompile(`{{[^}]*}}`)
	placeholders := make(map[string]string)
	index := 0
	
	sanitized = goTemplateRegex.ReplaceAllStringFunc(sanitized, func(match string) string {
		placeholder := fmt.Sprintf("__GO_TEMPLATE_%d__", index)
		placeholders[placeholder] = match
		index++
		return placeholder
	})
	
	// Remover scripts peligrosos
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	sanitized = scriptRegex.ReplaceAllString(sanitized, "")
	
	// Remover event handlers peligrosos (excepto los necesarios para LiveView)
	eventRegex := regexp.MustCompile(`(?i)\s*on\w+\s*=\s*["'][^"']*["']`)
	sanitized = eventRegex.ReplaceAllStringFunc(sanitized, func(match string) string {
		// Preservar handlers necesarios para LiveView
		safeHandlers := []string{
			"send_event",
			"event.preventDefault",
			"event.stopPropagation", 
			"event.dataTransfer",
			"this.value",
			"this.checked",
			"this.classList",
			"this.files",
			"this.id",
			"JSON.stringify",
		}
		
		// Verificar si contiene algún handler seguro
		for _, safe := range safeHandlers {
			if strings.Contains(match, safe) {
				return match
			}
		}
		return ""
	})
	
	// Remover iframes
	iframeRegex := regexp.MustCompile(`(?i)<iframe[^>]*>.*?</iframe>`)
	sanitized = iframeRegex.ReplaceAllString(sanitized, "")
	
	// Remover object y embed
	objectRegex := regexp.MustCompile(`(?i)<(object|embed)[^>]*>.*?</(object|embed)>`)
	sanitized = objectRegex.ReplaceAllString(sanitized, "")
	
	// Restaurar Go templates
	for placeholder, original := range placeholders {
		sanitized = strings.ReplaceAll(sanitized, placeholder, original)
	}
	
	return sanitized, nil
}

// ValidatePath valida una ruta para prevenir path traversal
func ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}
	
	// Limpiar la ruta
	cleanPath := filepath.Clean(path)
	
	// Verificar path traversal
	if strings.Contains(path, "..") {
		return ErrInvalidPath
	}
	
	// Verificar profundidad máxima
	parts := strings.Split(cleanPath, string(filepath.Separator))
	if len(parts) > MaxPathDepth {
		return fmt.Errorf("path too deep: %d levels (max %d)", len(parts), MaxPathDepth)
	}
	
	// Verificar rutas restringidas
	for _, restricted := range defaultSecurityConfig.RestrictedPaths {
		if strings.HasPrefix(cleanPath, restricted) {
			return fmt.Errorf("access to path %s is restricted", restricted)
		}
	}
	
	// Verificar caracteres peligrosos
	dangerousChars := []string{";", "|", "&", "$", "`", "\n", "\r", "\x00"}
	for _, char := range dangerousChars {
		if strings.Contains(path, char) {
			return fmt.Errorf("path contains dangerous character: %s", char)
		}
	}
	
	return nil
}

// SanitizeHTML sanitiza contenido HTML para prevenir XSS
func SanitizeHTML(content string) string {
	// Escapar caracteres HTML peligrosos
	sanitized := html.EscapeString(content)
	return sanitized
}

// SanitizeJSON sanitiza strings en estructuras JSON
func SanitizeJSON(data interface{}) interface{} {
	switch v := data.(type) {
	case string:
		return SanitizeHTML(v)
	case map[string]interface{}:
		sanitized := make(map[string]interface{})
		for key, value := range v {
			sanitized[key] = SanitizeJSON(value)
		}
		return sanitized
	case []interface{}:
		sanitized := make([]interface{}, len(v))
		for i, value := range v {
			sanitized[i] = SanitizeJSON(value)
		}
		return sanitized
	default:
		return v
	}
}

// ValidateFileUpload valida un archivo subido
func ValidateFileUpload(filename string, size int64, mimeType string) error {
	// Validar nombre de archivo
	if err := ValidatePath(filename); err != nil {
		return fmt.Errorf("invalid filename: %w", err)
	}
	
	// Validar tamaño
	maxSize := int64(10 * 1024 * 1024) // 10MB default
	if size > maxSize {
		return fmt.Errorf("file too large: %d bytes (max %d)", size, maxSize)
	}
	
	// Validar tipo MIME
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/gif":       true,
		"image/webp":      true,
		"application/pdf": true,
		"text/plain":      true,
		"text/csv":        true,
	}
	
	if !allowedTypes[mimeType] {
		return fmt.Errorf("file type not allowed: %s", mimeType)
	}
	
	return nil
}

// RateLimiter estructura para limitar tasa de mensajes
type RateLimiter struct {
	requests map[string][]int64
	limit    int
	window   int64 // en segundos
}

// NewRateLimiter crea un nuevo limitador de tasa
func NewRateLimiter(limit int, windowSeconds int64) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]int64),
		limit:    limit,
		window:   windowSeconds,
	}
}

// Allow verifica si se permite una petición
func (r *RateLimiter) Allow(clientID string, timestamp int64) bool {
	// Limpiar peticiones antiguas
	cutoff := timestamp - r.window
	
	if timestamps, exists := r.requests[clientID]; exists {
		filtered := []int64{}
		for _, t := range timestamps {
			if t > cutoff {
				filtered = append(filtered, t)
			}
		}
		r.requests[clientID] = filtered
		
		// Verificar límite
		if len(filtered) >= r.limit {
			return false
		}
	} else {
		r.requests[clientID] = []int64{}
	}
	
	// Agregar nueva petición
	r.requests[clientID] = append(r.requests[clientID], timestamp)
	return true
}

// SetSecurityConfig actualiza la configuración de seguridad
func SetSecurityConfig(config *SecurityConfig) {
	if config.MaxMessageSize > 0 {
		defaultSecurityConfig.MaxMessageSize = config.MaxMessageSize
	}
	if config.MaxTemplateSize > 0 {
		defaultSecurityConfig.MaxTemplateSize = config.MaxTemplateSize
	}
	defaultSecurityConfig.EnableSanitization = config.EnableSanitization
	if config.AllowedEvents != nil {
		defaultSecurityConfig.AllowedEvents = config.AllowedEvents
	}
	if config.RestrictedPaths != nil {
		defaultSecurityConfig.RestrictedPaths = config.RestrictedPaths
	}
}

// GetSecurityConfig obtiene la configuración actual
func GetSecurityConfig() *SecurityConfig {
	return defaultSecurityConfig
}