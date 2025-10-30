# Localization Manager Backend (Go)

A high-performance backend service for managing localized React components with templates, implemented in Go.

## 🚀 Quick Start

```bash
# Install dependencies
make install

# Start Redis (in terminal 1)
make docker-up

# Run server (in terminal 2)
make run

# Test the server (in terminal 3)
make test
```

Visit `http://localhost:8000/health` to verify the server is running!

## Features

- 🚀 **High Performance**: Built with Go and Gin framework
- 🌍 **Multi-language Support**: English, Spanish, French, and German
- 📦 **Dual-layer Caching**: In-memory TTL cache + Redis for distributed caching
- ⚡ **Concurrency Control**: Built-in request limiting to prevent overload
- 🎯 **Component Templates**: Pre-defined React component templates with localization
- 🔄 **Cache Management**: Automatic TTL-based cache expiration

## Architecture

### Caching Strategy

The service implements a two-tier caching system:

1. **In-Memory TTL Cache** (10 minutes)
   - LRU eviction policy
   - Maximum 50 entries
   - Fast local access

2. **Redis Cache** (30 minutes)
   - Distributed caching
   - Longer TTL for persistence
   - Automatic fallback if unavailable

### Cache Flow

```
Request → TTL Cache → Redis → Generate Component → Store in both caches
```

When the TTL cache expires but Redis still has the data, it's fetched from Redis and stored back in the TTL cache. The Redis TTL is also refreshed on each read to keep frequently accessed data available.

## Prerequisites

- Go 1.23.1 or higher
- Redis (optional, but recommended)
- Docker & Docker Compose (for easy Redis setup)
- Make (for running Makefile commands)

## Available Make Commands

```bash
make help         # Show all available commands
make install      # Install Go dependencies
make build        # Build the server binary
make run          # Run the server in development mode
make test         # Run the test script
make docker-up    # Start Redis using Docker Compose
make docker-down  # Stop Redis
make clean        # Remove binary and cache files
```

## Installation

### Step 1: Install Dependencies

```bash
cd go-localization-manager-backend
make install
```

### Step 2: Start Redis (Recommended)

**Option A: Using Docker Compose (Easiest)**
```bash
make docker-up
```

**Option B: Local Redis Installation**
```bash
# macOS
brew install redis && redis-server

# Linux
sudo apt-get install redis-server && redis-server
```

### Step 3: Run the Server

```bash
make run
```

That's it! The server will be available at `http://localhost:8000`.

## Configuration

Set environment variables for Redis configuration:

```bash
export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=  # Leave empty if no password
```

Or create a `.env` file:
```
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
```

## Running the Server

### Development Mode

```bash
make run
```

### Production Mode

```bash
# Build the binary
make build

# Run the binary
./localization-server
```

The server will start on `http://localhost:8000`.

### Quick Start (All in One)

```bash
# Terminal 1: Start Redis
make docker-up

# Terminal 2: Run the server
make run

# Terminal 3: Run tests
make test

# When done, stop Redis
make docker-down
```

## API Endpoints

### Health Check

```bash
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "localization-manager-backend",
  "version": "0.1.0",
  "cache_size": 5,
  "concurrency_limit": 2,
  "redis_status": "connected"
}
```

### Get Localized Component

```bash
GET /api/component/:component_type?lang=:language_code
```

**Parameters:**
- `component_type` (path): Component type (`welcome`, `navigation`, `user_profile`, `footer`)
- `lang` (query): Language code (`en`, `es`, `fr`, `de`) - defaults to `en`

**Example:**
```bash
curl "http://localhost:8000/api/component/welcome?lang=es"
```

**Response:**
```json
{
  "component_name": "WelcomeComponent",
  "component_type": "functional",
  "language": "es",
  "template": "...",
  "localized_data": {
    "welcome_title": "Bienvenido a Nuestra App",
    "welcome_subtitle": "Tu viaje comienza aquí",
    "login_button": "Iniciar Sesión",
    "signup_button": "Registrarse"
  },
  "metadata": {
    "component_id": "welcome_es_1234",
    "last_updated": "2024-01-15T10:30:00Z",
    "required_keys": ["welcome_title", "welcome_subtitle", "login_button", "signup_button"]
  },
  "cached": false
}
```

## Available Components

- `welcome` - Welcome page with title, subtitle, and action buttons
- `navigation` - Navigation menu with home, about, and contact links
- `user_profile` - User profile section with title and edit button
- `footer` - Footer with copyright information

## Supported Languages

- `en` - English
- `es` - Spanish
- `fr` - French
- `de` - German

## Testing

### Automated Testing

Run the comprehensive test script:
```bash
make test
```

This will test all endpoints, caching behavior, and error handling.

### Manual Testing

#### Test All Components

```bash
# English
curl "http://localhost:8000/api/component/welcome?lang=en"
curl "http://localhost:8000/api/component/navigation?lang=en"
curl "http://localhost:8000/api/component/user_profile?lang=en"
curl "http://localhost:8000/api/component/footer?lang=en"

# Spanish
curl "http://localhost:8000/api/component/welcome?lang=es"

# French
curl "http://localhost:8000/api/component/welcome?lang=fr"

# German
curl "http://localhost:8000/api/component/welcome?lang=de"
```

### Test Caching

```bash
# First request (not cached)
curl "http://localhost:8000/api/component/welcome?lang=en"

# Second request (cached)
curl "http://localhost:8000/api/component/welcome?lang=en"
# Look for "cached": true in response
```

### Test Invalid Component

```bash
curl "http://localhost:8000/api/component/invalid?lang=en"
# Returns error with list of available components
```

## Performance Characteristics

- **Concurrency Limit**: 2 concurrent requests (configurable)
- **TTL Cache**: 10 minutes, 50 entries max
- **Redis Cache**: 30 minutes
- **Cache Hit Performance**: < 1ms (TTL cache), < 5ms (Redis)
- **Cache Miss Performance**: < 10ms (template interpolation)

## Project Structure

```
.
├── main.go              # Main application with all logic
├── go.mod               # Go module dependencies
├── go.sum               # Dependency checksums
├── Makefile             # Build and run commands
├── docker-compose.yml   # Redis setup
├── test.sh              # Automated test script
├── .gitignore           # Git ignore rules
└── README.md            # This file
```

## Common Commands

```bash
# Development workflow
make install          # First time setup
make docker-up        # Start Redis
make run              # Run the server
make test             # Test all endpoints

# Build for production
make build            # Creates ./localization-server binary
./localization-server # Run the binary

# Cleanup
make docker-down      # Stop Redis
make clean            # Remove build artifacts
```

## Dependencies

- `github.com/gin-gonic/gin` - HTTP web framework
- `github.com/redis/go-redis/v9` - Redis client

## License

MIT

