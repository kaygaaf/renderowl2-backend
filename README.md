# Renderowl 2.0 Backend API

Go + Gin backend API with PostgreSQL, JWT authentication, and comprehensive health checks.

## 🔐 Authentication

**Decision: Clerk** (instead of Auth0)

### Why Clerk?

| Factor | Clerk | Auth0 |
|--------|-------|-------|
| **Pricing** | Generous free tier (10k MAU) | Limited free tier (25k MAU but limited features) |
| **Next.js Integration** | First-class SDK | Good but requires more setup |
| **UX** | Pre-built, customizable components | Requires custom UI |
| **Developer Experience** | Simple setup, excellent DX | More complex configuration |
| **Pricing at Scale** | Predictable, cheaper for our use case | Can get expensive |

### Clerk Implementation

The API uses Clerk JWT tokens for authentication. Protected endpoints require a `Bearer` token in the Authorization header.

```
Authorization: Bearer <clerk_jwt_token>
```

### Protected Endpoints

All `/api/v1/*` endpoints require authentication except:
- `GET /health` - Health check
- `GET /health/ready` - Readiness probe
- `GET /health/live` - Liveness probe

## 🚀 API Endpoints

### Health Checks
```
GET  /health              → Basic health check
GET  /health/ready        → Readiness probe (checks DB, Redis, Remotion)
GET  /health/live         → Liveness probe
```

### Timelines
```
GET    /api/v1/timelines       → List all timelines
POST   /api/v1/timelines       → Create timeline
GET    /api/v1/timelines/:id   → Get timeline
PUT    /api/v1/timelines/:id   → Update timeline
DELETE /api/v1/timelines/:id   → Delete timeline
```

### Clips
```
GET    /api/v1/timelines/:id/clips  → List clips for timeline
POST   /api/v1/timelines/:id/clips  → Create clip
GET    /api/v1/clips/:clipId        → Get clip
PUT    /api/v1/clips/:clipId        → Update clip
DELETE /api/v1/clips/:clipId        → Delete clip
```

### Tracks
```
GET    /api/v1/timelines/:id/tracks  → List tracks for timeline
POST   /api/v1/timelines/:id/tracks  → Create track
PUT    /api/v1/tracks/:trackId       → Update track
DELETE /api/v1/tracks/:trackId       → Delete track
PATCH  /api/v1/tracks/:trackId/reorder   → Reorder tracks
PATCH  /api/v1/tracks/:trackId/mute      → Toggle mute
PATCH  /api/v1/tracks/:trackId/solo      → Toggle solo
```

## 🌐 CORS Configuration

CORS is configured to allow:
- Production frontend: `https://renderowl.app`
- Staging frontend: `https://staging.renderowl.app`
- Local development: `http://localhost:3000`, `http://localhost:3001`

Credentials are enabled for authenticated requests.

## 📊 Health Check Details

The `/health/ready` endpoint checks:

1. **Database**: Connection pool status, including open/in-use/idle connections
2. **Redis**: Queue connectivity (placeholder)
3. **Remotion**: Video rendering service (placeholder)

Returns `503 Service Unavailable` if any critical dependency is down.

## 🔧 Development

### Setup

```bash
cd backend
go mod download
```

### Environment Variables

```bash
ENVIRONMENT=development
PORT=8080
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/renderowl
REDIS_URL=redis://localhost:6379
CLERK_SECRET_KEY=sk_test_...
FRONTEND_URL=http://localhost:3000
```

### Run

```bash
go run ./cmd/api
```

### Build

```bash
go build -o bin/api ./cmd/api
```

## 📁 Project Structure

```
backend/
├── cmd/api/
│   └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration
│   ├── domain/
│   │   └── models.go        # Domain entities
│   ├── handlers/
│   │   ├── health.go        # Health handlers
│   │   ├── timeline.go      # Timeline handlers
│   │   ├── clip.go          # Clip handlers
│   │   └── track.go         # Track handlers
│   ├── middleware/
│   │   ├── auth.go          # JWT authentication
│   │   ├── cors.go          # CORS setup
│   │   └── errors.go        # Error handling
│   ├── repository/
│   │   ├── models.go        # DB models
│   │   ├── timeline.go      # Timeline repository
│   │   ├── clip.go          # Clip repository
│   │   └── track.go         # Track repository
│   └── service/
│       ├── timeline.go      # Timeline service
│       ├── clip.go          # Clip service
│       └── track.go         # Track service
├── go.mod
└── README.md
```

## 🧪 Testing with curl

### Health Check
```bash
curl http://localhost:8080/health
```

### Create Timeline (Authenticated)
```bash
curl -X POST http://localhost:8080/api/v1/timelines \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "My Timeline", "description": "Test"}'
```

### CORS Test
```bash
curl -X OPTIONS http://localhost:8080/api/v1/timelines \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  -v
```
