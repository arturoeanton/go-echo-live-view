#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Go Echo LiveView - Collaborative Examples${NC}"
echo ""

# Check if WASM is built
if [ ! -f "assets/json.wasm" ]; then
    echo -e "${YELLOW}⚠️  WASM module not found. Building...${NC}"
    cd cmd/wasm/
    GOOS=js GOARCH=wasm go build -o ../../assets/json.wasm
    cd ../..
    echo -e "${GREEN}✅ WASM module built${NC}"
fi

echo ""
echo -e "${GREEN}Available Examples:${NC}"
echo "1) Kanban Simple - Basic kanban board"
echo "2) Kanban App - Full project management"
echo "3) Collaborative Working - All demos"
echo "4) Build WASM only"
echo "5) Exit"
echo ""
read -p "Select an example to run (1-5): " choice

case $choice in
    1)
        echo -e "${BLUE}Starting Kanban Simple...${NC}"
        echo -e "${YELLOW}Open http://localhost:8080 in your browser${NC}"
        cd example/kanban_simple
        go run main.go
        ;;
    2)
        echo -e "${BLUE}Starting Kanban App...${NC}"
        echo -e "${YELLOW}Open http://localhost:8080 in your browser${NC}"
        cd example/kanban_app
        go run main.go
        ;;
    3)
        echo -e "${BLUE}Starting Collaborative Working...${NC}"
        echo -e "${YELLOW}Open http://localhost:8080 in your browser${NC}"
        cd example/collaborative_working
        go run main.go
        ;;
    4)
        echo -e "${BLUE}Building WASM module...${NC}"
        cd cmd/wasm/
        GOOS=js GOARCH=wasm go build -o ../../assets/json.wasm
        echo -e "${GREEN}✅ WASM module built successfully${NC}"
        ;;
    5)
        echo -e "${GREEN}Goodbye!${NC}"
        exit 0
        ;;
    *)
        echo -e "${RED}Invalid option${NC}"
        exit 1
        ;;
esac