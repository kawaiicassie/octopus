#!/bin/bash

# Install Node.js dependencies và build frontend
cd web
npm install
npm run build
cd ..

# Di chuyển frontend build vào static
mv web/out static/

# Build Go binary
go build -tags netgo -ldflags '-s -w' -o app