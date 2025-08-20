# Tests del Framework Go Echo LiveView

Este directorio contiene los tests unitarios para el framework Go Echo LiveView.

## Estructura de Tests

- `model_test.go` - Tests para componentes y drivers
- `page_content_test.go` - Tests para el control de páginas y generación HTML
- `utils_test.go` - Tests para funciones utilitarias
- `bimap_test.go` - Tests para la estructura de datos BiMap
- `fxtemplate_test.go` - Tests para funciones de template

## Ejecutar Tests

### Todos los tests
```bash
go test -v ./liveview/tests/...
```

### Con cobertura
```bash
go test -v -coverprofile=coverage.out ./liveview/tests/...
go tool cover -html=coverage.out -o coverage.html
```

### Script automatizado
```bash
./liveview/tests/run_tests.sh
```

### Con benchmarks
```bash
./liveview/tests/run_tests.sh --bench
```

## Cobertura Actual

Los tests cubren:
- ✅ Creación e inicialización de componentes
- ✅ Sistema de eventos
- ✅ Montaje de componentes
- ✅ Generación de HTML
- ✅ Operaciones de archivos
- ✅ Estructura BiMap
- ✅ Funciones de template
- ✅ Métodos del driver

## Añadir Nuevos Tests

Al añadir nuevas funcionalidades, asegúrate de:
1. Crear tests que cubran casos normales y edge cases
2. Incluir tests de error/panic donde sea apropiado
3. Mantener la cobertura por encima del 80%
4. Documentar comportamientos esperados

## Tests de Integración

Para tests más complejos que requieren WebSocket, ver los ejemplos en el directorio `example/` que sirven como tests de integración manuales.