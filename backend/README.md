# Renderowl 2.0 API

A Go-based REST API for the Renderowl video timeline editor, built with Gin and GORM.

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- (Optional) Docker for database

### Setup

1. **Clone and navigate to the project:**
   ```bash
   cd /projects/renderowl2.0/backend
   ```

2. **Copy environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials
   ```

3. **Set up PostgreSQL:**
   ```bash
   # Using Docker
   docker run -d \
     --name renderowl2-db \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=postgres \
     -e POSTGRES_DB=renderowl2 \
     -p 5432:5432 \
     postgres:15-alpine
   ```

4. **Install dependencies:**
   ```bash
   go mod download
   ```

5. **Run the server:**
   ```bash
   go run cmd/api/main.go
   ```

The API will be available at `http://localhost:8080`

## 📚 API Endpoints

### Health Check
- `GET /health` - Health check endpoint

### Timelines
- `POST /api/v1/timeline` - Create a new timeline
- `GET /api/v1/timeline/:id` - Get timeline by ID
- `PUT /api/v1/timeline/:id` - Update timeline
- `DELETE /api/v1/timeline/:id` - Delete timeline
- `GET /api/v1/timelines` - List all timelines (with pagination)
- `GET /api/v1/timelines/me` - Get current user's timelines

### Example Requests

**Create a timeline:**
```bash
curl -X POST http://localhost:8080/api/v1/timeline \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My Project",
    "description": "A cool video project"
  }'
```

**Get a timeline:**
```bash
curl http://localhost:8080/api/v1/timeline/1
```

## 📁 Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go           # Application entry point
├── internal/
│   ├── domain/
│   │   └── timeline.go       # Domain models (Timeline, Track, Clip)
│   ├── handlers/
│   │   └── timeline.go       # HTTP handlers
│   ├── repository/
│   │   └── timeline.go       # Database repository
│   └── service/
│       └── timeline.go       # Business logic
├── pkg/                      # Shared packages
├── migrations/               # Database migrations
├── scripts/                  # Utility scripts
├── .env.example              # Environment template
├── go.mod                    # Go module definition
├── go.sum                    # Go dependencies
└── README.md                 # This file
```

## 🏗️ Architecture

The project follows **Clean Architecture** principles:

- **Domain**: Core business logic and models
- **Repository**: Data access layer (GORM)
- **Service**: Business logic layer
- **Handlers**: HTTP transport layer (Gin)

## 🧪 Development

### Run Tests
```bash
go test ./...
```

### Run with Hot Reload
```bash
go install github.com/air-verse/air@latest
air
```

### Format Code
```bash
go fmt ./...
```

### Lint
```bash
go vet ./...
```

## 🐳 Docker

### Build Image
```bash
docker build -t renderowl2-api .
```

### Run Container
```bash
docker run -p 8080:8080 --env-file .env renderowl2-api
```

## 📝 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | 8080 |
| `GIN_MODE` | Gin mode (release/debug) | release |
| `DATABASE_URL` | PostgreSQL connection string | - |
| `DB_HOST` | Database host | localhost |
| `DB_PORT` | Database port | 5432 |
| `DB_USER` | Database user | postgres |
| `DB_PASSWORD` | Database password | postgres |
| `DB_NAME` | Database name | renderowl2 |
| `DB_DEBUG` | Enable SQL logging | false |
| `JWT_SECRET` | JWT signing secret | - |
| `LOG_LEVEL` | Logging level | info |

## 🔮 Future Features

- [ ] JWT Authentication
- [ ] File upload (video/audio assets)
- [ ] WebSocket for real-time collaboration
- [ ] Export API (video rendering)
- [ ] Swagger/OpenAPI docs
- [ ] Rate limiting
- [ ] Request validation middleware

## 📄 License

Private - Renderowl Project
