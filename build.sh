#!/bin/bash

# Exit on error
set -e

# Get version info
GIT_VERSION="$(git describe --tags --abbrev=0 2>/dev/null || echo 'dev')"
COMMIT_ID="$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
BUILD_TIME="$(date -u +'%Y-%m-%d %H:%M:%S UTC')"

echo "🚀 Building Octopus version: $GIT_VERSION (commit: $COMMIT_ID)"

# Build frontend
echo "📦 Building frontend..."
cd web
npm ci --production=false  # Dùng ci thay vì install cho production
NEXT_PUBLIC_APP_VERSION="$GIT_VERSION" npm run build
cd ..

# Clean and prepare static directory
echo "📁 Preparing static directory..."
rm -rf static/out
mv web/out static/

# Build Go binary
echo "🔨 Building Go binary..."
CGO_ENABLED=0 go build -tags netgo \
  -ldflags "-X 'github.com/kawaiicassie/octopus/internal/conf.Version=$GIT_VERSION' \
            -X 'github.com/kawaiicassie/octopus/internal/conf.Commit=$COMMIT_ID' \
            -X 'github.com/kawaiicassie/octopus/internal/conf.BuildTime=$BUILD_TIME' \
            -extldflags '-static' -s -w" \
  -o app

echo "✅ Build completed successfully!"
echo "   Version: $GIT_VERSION"
echo "   Commit: $COMMIT_ID"
echo "   Binary: ./app"