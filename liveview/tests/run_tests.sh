#!/bin/bash

# Script para ejecutar todos los tests del framework Go Echo LiveView

echo "=== Ejecutando Tests de Go Echo LiveView ==="
echo

# Colores para output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Cambiar al directorio del proyecto
cd "$(dirname "$0")/../.."

# Ejecutar tests con coverage
echo "Ejecutando tests con cobertura..."
echo

# Ejecutar tests del paquete liveview
go test -v -coverprofile=coverage.out ./liveview/tests/...

# Verificar si los tests pasaron
if [ $? -eq 0 ]; then
    echo -e "\n${GREEN}✓ Todos los tests pasaron exitosamente!${NC}\n"
    
    # Mostrar resumen de cobertura
    echo "=== Resumen de Cobertura ==="
    go tool cover -func=coverage.out | grep -E '^total:|^github.com/arturoeanton/go-echo-live-view/liveview\s'
    
    # Generar reporte HTML de cobertura
    go tool cover -html=coverage.out -o coverage.html
    echo -e "\n${YELLOW}Reporte de cobertura HTML generado: coverage.html${NC}"
else
    echo -e "\n${RED}✗ Algunos tests fallaron${NC}\n"
    exit 1
fi

# Ejecutar tests de componentes si existen
if [ -d "components/tests" ]; then
    echo -e "\n=== Ejecutando Tests de Componentes ==="
    go test -v ./components/tests/...
fi

# Ejecutar benchmarks opcionalmente
if [ "$1" == "--bench" ]; then
    echo -e "\n=== Ejecutando Benchmarks ==="
    go test -bench=. -benchmem ./liveview/tests/...
fi

# Limpiar archivos temporales
rm -f coverage.out

echo -e "\n=== Tests Completados ==="