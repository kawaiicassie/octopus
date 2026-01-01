# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Octopus is a full-stack LLM API aggregation and load balancing service that enables connecting multiple LLM providers through a unified management panel with intelligent load balancing, protocol conversion, and comprehensive analytics.

**Tech Stack:**
- Backend: Go with Gin framework, GORM ORM
- Frontend: Next.js 16 + React 19 with TypeScript, Tailwind CSS v4
- Database: SQLite (default), MySQL, or PostgreSQL

## Common Development Commands

### Running the Application

```bash
# Backend (runs on http://localhost:8080)
go run . start

# Frontend development (runs on http://localhost:3000)
cd web && npm run dev

# Build for production (multi-platform)
bash scripts/build.sh release

# Run tests
go test ./...
```

### Building from Source

```bash
# 1. Build frontend
cd web
npm install
npm run build
cd ..

# 2. Move frontend to static directory
mv web/out static/

# 3. Run backend (embeds frontend)
go run . start
```

## High-Level Architecture

### Request Flow Architecture

The system uses a transformer pattern for protocol conversion:

```
Client Request → Inbound Adapter → Internal Format → Load Balancer
→ Channel Selection → Outbound Adapter → Provider API → Response Transform → Client
```

**Key Components:**
- **Inbound Adapters**: Convert OpenAI Chat/Responses, Anthropic, Gemini formats to internal representation
- **Load Balancers**: Four strategies implemented in `internal/loadbalance/`:
  - RoundRobin (atomic counter-based)
  - Random
  - Failover (priority-based)
  - Weighted distribution
- **Outbound Adapters**: Convert internal format to provider-specific formats

### Core Business Logic Organization

The codebase follows a clean layered architecture:

1. **Entry Points**:
   - `main.go` → `cmd/start.go` → `internal/server/server.go`
   - HTTP handlers in `internal/server/handlers/`

2. **Business Operations** (`internal/op/`):
   - Each operation module handles specific domain logic
   - Key modules: channel, group, statistics, cache, apikey, user
   - Operations coordinate between database, cache, and external services

3. **Relay System** (`internal/relay/`):
   - Core request forwarding logic
   - Manages streaming (SSE) and non-streaming responses
   - Handles token counting and billing

4. **Database Layer** (`internal/model/`):
   - GORM models with auto-migration
   - Key entities: Channel, Group, LLMInfo, Statistics, User, APIKey

### Frontend Architecture

The frontend uses a modern React stack with:
- **State Management**: Zustand stores in `web/src/stores/`
- **API Integration**: TanStack Query with centralized API client
- **Components**: Modular components in `web/src/components/`
- **Routing**: Next.js App Router with internationalization

### Statistics System

The statistics system uses an in-memory aggregation pattern:
- Statistics are collected in memory using atomic operations
- Periodically batch-written to database (configurable interval)
- Graceful shutdown ensures all stats are persisted
- Multiple granularities: total, daily, hourly, per-model, per-channel

## Key Design Patterns

### Load Balancing Implementation

The load balancers in `internal/loadbalance/` use atomic operations for thread safety:
- **RoundRobin**: Atomic counter with modulo operation
- **Weighted**: Cumulative weight calculation with random selection
- **Failover**: Priority-based with automatic fallback

### Protocol Transformation

Transformers in `internal/transformer/` handle bidirectional conversion:
- Each transformer implements both request and response conversion
- Streaming responses handled via Server-Sent Events (SSE)
- Token counting integrated for accurate billing

### Cache Architecture

The cache system in `internal/op/cache/` uses a sharded design:
- Multiple shards to reduce lock contention
- Atomic operations for thread safety
- TTL-based expiration with background cleanup

## Database Schema

Key relationships:
- **Channels** belong to **Groups** (many-to-many via group_channels)
- **Statistics** track usage per channel, model, and API key
- **LLMInfo** stores model pricing (input, output, cache read/write tokens)
- **Users** own **APIKeys** for authentication

## Configuration

Configuration priority (highest to lowest):
1. Environment variables (`OCTOPUS_*` prefix)
2. Configuration file (`data/config.json`)
3. Default values

Key environment variables:
- `OCTOPUS_DATABASE_PATH`: Database connection string
- `OCTOPUS_DATABASE_TYPE`: sqlite/mysql/postgres
- `OCTOPUS_SERVER_PORT`: Server port (default 8080)

## Testing Approach

- Backend: Standard Go testing with examples in `internal/price/price_test.go`
- Run all tests: `go test ./...`
- Frontend: ESLint configured, manual testing recommended

## Important Implementation Details

### API Authentication
- Admin panel: JWT-based authentication
- API requests: Bearer token or x-api-key header
- API key format: `sk-octopus-*`

### Channel Configuration
Base URLs should exclude specific endpoints - the system automatically appends:
- OpenAI Chat: Base URL + `/chat/completions`
- OpenAI Responses: Base URL + `/responses`
- Anthropic: Base URL + `/messages`
- Gemini: Base URL + `/models/:model:generateContent`

### Group Management
- Group names become the exposed model names
- Load balancing modes determine request distribution
- Failed channels are automatically retried with next available

### Graceful Shutdown
Always use proper shutdown (Ctrl+C or SIGTERM) to ensure:
- In-memory statistics are persisted
- Active connections are closed properly
- Background tasks complete cleanly