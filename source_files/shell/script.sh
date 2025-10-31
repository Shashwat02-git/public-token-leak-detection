#!/bin/bash
set -e

echo "Starting Project Charlie build..."

# Step 1: Compile Java code
echo "Compiling Java..."
javac com/example/charlie/utils/S3Connector.java
echo "Java compilation complete."

# Step 2: Build Docker image
echo "Building Docker image..."
docker build -t charlie-service:latest .
echo "Docker build complete."

echo "Build successful!"
