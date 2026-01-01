#!/bin/bash

# Get version info
GIT_VERSION="$(git describe --tags --abbrev=0 2>/dev/null || echo 'dev')"
COMMIT_ID="$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
BUILD_TIME="$(date -u +'%Y-%m-%d %H:%M:%S UTC')"

# Install Node.js dependencies và build frontend với version
cd web
npm install
NEXT_PUBLIC_APP_VERSION="$GIT_VERSION" npm run build
cd ..

# Di chuyển frontend build vào static
mv web/out static/

# Build Go binary với version info
go build -tags netgo -ldflags "-X 'github.com/bestruirui/octopus/internal/conf.Version=$GIT_VERSION' \
  -X 'github.com/bestruirui/octopus/internal/conf.Commit=$COMMIT_ID' \
  -X 'github.com/bestruirui/octopus/internal/conf.BuildTime=$BUILD_TIME' \
  -s -w" -o app