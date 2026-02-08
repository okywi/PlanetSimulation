#!/bin/bash
if [ ! -d "build" ]; then
    mkdir build
fi
echo "Building for linux..."
GOOS=linux go build -o build/PlanetSimulation cmd/planetsimulation/main.go 
echo "Building for windows..."
GOOS=windows go build -o build/PlanetSimulation.exe cmd/planetsimulation/main.go 