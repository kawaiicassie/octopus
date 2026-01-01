#!/bin/bash

# Wrapper script for Render to use the main build script
# Only builds for current platform without Docker/archives

set -e

echo "Building Octopus for Render deployment..."

# Build frontend
cd web
npm install
NEXT_PUBLIC_APP_VERSION="$(git describe --tags --abbrev=0 2>/dev/null || echo 'dev')" npm run build
cd ..

# Move frontend to static
rm -rf static/out
mv web/out static/

# Get build metadata
GIT_VERSION="$(git describe --tags --abbrev=0 2>/dev/null || echo 'dev')"
COMMIT_ID="$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
BUILD_TIME="$(TZ='UTC' date +'%F %T %z')"

# Build Go binary with version info
go build -tags netgo \
  -ldflags "-X 'github.com/kawaiicassie/octopus/internal/conf.Version=${GIT_VERSION}' \
            -X 'github.com/kawaiicassie/octopus/internal/conf.Commit=${COMMIT_ID}' \
            -X 'github.com/kawaiicassie/octopus/internal/conf.BuildTime=${BUILD_TIME}' \
            -X 'github.com/kawaiicassie/octopus/internal/conf.Author=kawaiicassie' \
            -s -w" \
  -o app

echo "✅ Build completed successfully!"