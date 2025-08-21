#!/bin/bash

# Script to run all three enhanced examples on different ports

echo "Starting Enhanced Framework Examples..."
echo "======================================="
echo ""

# Kill any existing processes on these ports
lsof -ti:8081 | xargs kill -9 2>/dev/null
lsof -ti:8082 | xargs kill -9 2>/dev/null
lsof -ti:8083 | xargs kill -9 2>/dev/null

# Start collaborative workspace
echo "Starting Collaborative Workspace v2 on port 8081..."
(cd example/collaborative_working_v2 && go run main.go) &
PID1=$!
echo "PID: $PID1"
echo ""

# Start flow tool
echo "Starting Flow Tool v2 on port 8082..."
(cd example/example_flowtool_v2 && go run main.go) &
PID2=$!
echo "PID: $PID2"
echo ""

# Start showcase
echo "Starting Framework Showcase v3 on port 8083..."
(cd example/example_showcase_v3 && go run main.go) &
PID3=$!
echo "PID: $PID3"
echo ""

echo "======================================="
echo "All examples are running!"
echo ""
echo "Access them at:"
echo "  • Collaborative Workspace: http://localhost:8081"
echo "  • Flow Tool:              http://localhost:8082"
echo "  • Framework Showcase:      http://localhost:8083"
echo ""
echo "Press Ctrl+C to stop all examples"
echo "======================================="

# Wait for any process to exit
wait $PID1 $PID2 $PID3

# Clean up
kill $PID1 $PID2 $PID3 2>/dev/null
echo "All examples stopped."