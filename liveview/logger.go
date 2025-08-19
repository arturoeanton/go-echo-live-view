package liveview

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel define los niveles de log
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// Logger estructura principal del logger
type Logger struct {
	mu       sync.RWMutex
	level    LogLevel
	verbose  bool
	prefix   string
	output   *log.Logger
	colored  bool
}

var (
	// DefaultLogger instancia global del logger
	DefaultLogger *Logger
	once          sync.Once
)

// Colores ANSI para terminal
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[37m"
	colorWhite  = "\033[97m"
)

// InitLogger inicializa el logger global
func InitLogger(verbose bool) {
	once.Do(func() {
		DefaultLogger = &Logger{
			level:   LogLevelInfo,
			verbose: verbose,
			prefix:  "[LiveView]",
			output:  log.New(os.Stdout, "", log.LstdFlags),
			colored: true,
		}
		
		if verbose {
			DefaultLogger.level = LogLevelDebug
		}
	})
}

// GetLogger obtiene el logger global
func GetLogger() *Logger {
	if DefaultLogger == nil {
		InitLogger(false)
	}
	return DefaultLogger
}

// SetVerbose activa/desactiva el modo verbose
func SetVerbose(verbose bool) {
	logger := GetLogger()
	logger.mu.Lock()
	defer logger.mu.Unlock()
	
	logger.verbose = verbose
	if verbose {
		logger.level = LogLevelDebug
	} else {
		logger.level = LogLevelInfo
	}
}

// IsVerbose retorna si el modo verbose está activo
func IsVerbose() bool {
	logger := GetLogger()
	logger.mu.RLock()
	defer logger.mu.RUnlock()
	return logger.verbose
}

// formatMessage formatea el mensaje con información adicional
func (l *Logger) formatMessage(level LogLevel, format string, args ...interface{}) string {
	levelStr := l.getLevelString(level)
	levelColor := l.getLevelColor(level)
	
	// Obtener información del caller
	_, file, line, _ := runtime.Caller(2)
	parts := strings.Split(file, "/")
	fileName := parts[len(parts)-1]
	
	// Construir el mensaje
	msg := fmt.Sprintf(format, args...)
	
	if l.colored {
		return fmt.Sprintf("%s%s%s [%s:%d] %s%s", 
			levelColor, levelStr, colorReset,
			fileName, line,
			msg, colorReset)
	}
	
	return fmt.Sprintf("%s [%s:%d] %s", levelStr, fileName, line, msg)
}

// getLevelString retorna el string del nivel
func (l *Logger) getLevelString(level LogLevel) string {
	switch level {
	case LogLevelDebug:
		return "[DEBUG]"
	case LogLevelInfo:
		return "[INFO] "
	case LogLevelWarn:
		return "[WARN] "
	case LogLevelError:
		return "[ERROR]"
	case LogLevelFatal:
		return "[FATAL]"
	default:
		return "[UNKNOWN]"
	}
}

// getLevelColor retorna el color para el nivel
func (l *Logger) getLevelColor(level LogLevel) string {
	switch level {
	case LogLevelDebug:
		return colorCyan
	case LogLevelInfo:
		return colorGreen
	case LogLevelWarn:
		return colorYellow
	case LogLevelError:
		return colorRed
	case LogLevelFatal:
		return colorPurple
	default:
		return colorWhite
	}
}

// Debug log de nivel debug
func Debug(format string, args ...interface{}) {
	logger := GetLogger()
	if logger.level <= LogLevelDebug {
		msg := logger.formatMessage(LogLevelDebug, format, args...)
		logger.output.Println(msg)
	}
}

// Info log de nivel info
func Info(format string, args ...interface{}) {
	logger := GetLogger()
	if logger.level <= LogLevelInfo {
		msg := logger.formatMessage(LogLevelInfo, format, args...)
		logger.output.Println(msg)
	}
}

// Warn log de nivel warning
func Warn(format string, args ...interface{}) {
	logger := GetLogger()
	if logger.level <= LogLevelWarn {
		msg := logger.formatMessage(LogLevelWarn, format, args...)
		logger.output.Println(msg)
	}
}

// Error log de nivel error
func Error(format string, args ...interface{}) {
	logger := GetLogger()
	if logger.level <= LogLevelError {
		msg := logger.formatMessage(LogLevelError, format, args...)
		logger.output.Println(msg)
	}
}

// Fatal log de nivel fatal y termina el programa
func Fatal(format string, args ...interface{}) {
	logger := GetLogger()
	msg := logger.formatMessage(LogLevelFatal, format, args...)
	logger.output.Fatal(msg)
}

// LogComponent logs específicos para componentes
func LogComponent(componentID string, action string, details ...interface{}) {
	if IsVerbose() {
		Debug("[Component:%s] %s %v", componentID, action, details)
	}
}

// LogWebSocket logs específicos para WebSocket
func LogWebSocket(action string, data interface{}) {
	if IsVerbose() {
		Debug("[WebSocket] %s: %+v", action, data)
	}
}

// LogEvent logs específicos para eventos
func LogEvent(componentID string, eventName string, data interface{}) {
	if IsVerbose() {
		Debug("[Event] Component:%s Event:%s Data:%+v", componentID, eventName, data)
	}
}

// LogTemplate logs para templates
func LogTemplate(componentID string, action string, details string) {
	if IsVerbose() {
		Debug("[Template:%s] %s: %s", componentID, action, details)
	}
}

// LogPerformance logs de performance
func LogPerformance(operation string, duration time.Duration) {
	if IsVerbose() {
		if duration > time.Millisecond*100 {
			Warn("[Performance] %s took %v (slow)", operation, duration)
		} else {
			Debug("[Performance] %s took %v", operation, duration)
		}
	}
}

// LogMemory logs de memoria
func LogMemory(context string) {
	if IsVerbose() {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		Debug("[Memory:%s] Alloc=%vMB Sys=%vMB NumGC=%v", 
			context,
			m.Alloc/1024/1024,
			m.Sys/1024/1024,
			m.NumGC)
	}
}

// Benchmark mide el tiempo de una operación
func Benchmark(name string, fn func()) {
	if IsVerbose() {
		start := time.Now()
		fn()
		LogPerformance(name, time.Since(start))
	} else {
		fn()
	}
}