# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PassDB is a password dump database API server that provides REST endpoints for querying password breach data stored in Google BigQuery. It integrates with the Have I Been Pwned (HIBP) API to provide breach information for email addresses.

## Development Commands

### Running the Application
```bash
# Set environment variables (see Environment Variables section)
source .env
go run main.go [port]
```

### Building and Testing
```bash
# Build the application
go build -o passdb main.go

# Run with Go modules
go mod tidy
go mod download

# Install dependencies
go get
```

### Docker
```bash
# Build Docker image
docker build -t passdb-server .

# Run with Docker
docker run --env-file .env -p 3000:3000 passdb-server
```

## Environment Variables

Required environment variables (create a `.env` file):
```bash
# Google Cloud Project Name
GOOGLE_CLOUD_PROJECT=your-project-name

# Format: $project.$dataset.$tablename
GOOGLE_BIGQUERY_TABLE=your-project.your-dataset.your-table

# Path to Google Cloud credentials JSON file
GOOGLE_APPLICATION_CREDENTIALS=./credentials.json

# Have I Been Pwned API key
HIBP_API_KEY=your-hibp-api-key

# Cache Configuration (optional)
CACHE_ENABLED=true                  # Enable/disable caching (default: true)
CACHE_DB_PATH=./cache.db           # Cache database file path (default: ./cache.db)
CACHE_DEFAULT_TTL=720h             # Default cache TTL - 30 days (default: 720h)
CACHE_TTL_BREACHES=168h            # Breach data cache TTL - 7 days (default: 168h)
CACHE_TTL_USERNAMES=720h           # Username cache TTL - 30 days (default: 720h)  
CACHE_TTL_PASSWORDS=720h           # Password cache TTL - 30 days (default: 720h)
CACHE_TTL_DOMAINS=720h             # Domain cache TTL - 30 days (default: 720h)
CACHE_TTL_EMAILS=720h              # Email cache TTL - 30 days (default: 720h)
```

## Architecture

### Core Components

- **main.go**: Main HTTP server using Chi router with BigQuery integration
- **hibp/hibp.go**: Have I Been Pwned API client for breach data
- **cache.go**: BBolt-based caching middleware with per-endpoint TTL configuration

### Key Patterns

- **Parameterized queries**: All BigQuery queries use parameterized queries (see `parameterize` function in main.go:195) for SQL injection prevention
- **Error handling**: JSON error responses using `JSONError` function (main.go:251)
- **Record structure**: All password dump data follows the `record` struct format (main.go:72)
- **Caching middleware**: Chi middleware caches responses with configurable TTL per endpoint type (cache.go:89)

### API Endpoints

The server provides REST endpoints for querying breach data:
- `GET /usernames/{username}` - Find records by username
- `GET /passwords/{password}` - Find records by password  
- `GET /domains/{domain}` - Find records by domain
- `GET /emails/{email}` - Find records by email (splits into username@domain)
- `GET /breaches/{email}` - Get HIBP breach data for email

Cache management endpoints:
- `GET /cache/stats` - View cache statistics and configuration
- `DELETE /cache` - Clear entire cache
- `DELETE /cache/{pattern}` - Clear cache entries matching pattern

### Data Flow

1. HTTP requests are routed through Chi middleware (logging, caching, CORS, recovery)
2. Cache middleware checks for existing responses and returns cached data if valid
3. On cache miss, URL parameters are extracted and used to build parameterized BigQuery queries
4. Results are fetched from BigQuery and marshaled to JSON
5. Successful responses are cached with appropriate TTL based on endpoint type
6. HIBP endpoints make external API calls to haveibeenpwned.com (also cached)

## Security Notes

- Never commit `credentials.json` or `.env` files
- All BigQuery queries use parameterization to prevent SQL injection
- HIBP API key is required and should be kept secure
- Cache database (`cache.db`) contains response data - ensure proper file permissions

## Caching

The application implements a persistent file-based cache using BBolt to reduce BigQuery costs:

- **Cache storage**: Single `cache.db` file with concurrent read access
- **Cache keys**: HTTP method + URL path (e.g., "GET:/usernames/john")
- **TTL configuration**: Different cache durations per endpoint type
- **Cache headers**: Responses include `X-Cache: HIT/MISS` header
- **Cost optimization**: Password dump data cached for 30 days, breach data for 7 days