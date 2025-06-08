#!/bin/bash

# Build script for gym-tracker-api Lambda function

set -e  # Exit on any error

echo "🏗️  Building gym-tracker-api..."

# Clean up previous builds
echo "🧹 Cleaning up previous builds..."
rm -f main lambda.zip

# Build the Go binary for Linux (Lambda runtime)
echo "🔨 Compiling Go binary for Linux..."
GOOS=linux GOARCH=amd64 go build -o main ./cmd/api/main.go

# Verify the binary was created
if [ ! -f "main" ]; then
    echo "❌ Build failed: main binary not found"
    exit 1
fi

echo "📦 Creating lambda.zip..."
zip lambda.zip main

# Verify the zip was created
if [ ! -f "lambda.zip" ]; then
    echo "❌ Zip creation failed: lambda.zip not found"
    exit 1
fi

echo "✅ Build complete!"
echo "📄 Binary size: $(ls -lh main | awk '{print $5}')"
echo "📦 Package size: $(ls -lh lambda.zip | awk '{print $5}')"
echo ""
echo "Ready for deployment! 🚀"