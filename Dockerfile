# Build frontend
FROM node:18-alpine AS frontend-builder
RUN npm install -g pnpm
WORKDIR /app/web
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm run build

# Build backend
FROM golang:1.25-alpine AS backend-builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/web/out ./static/out
RUN go build -tags netgo -ldflags '-s -w' -o app

# Runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=backend-builder /app/app .
EXPOSE 8080
CMD ["./app", "start"]