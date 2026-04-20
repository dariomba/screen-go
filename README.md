[![CI](https://github.com/dariomba/screen-go/actions/workflows/ci.yaml/badge.svg)](https://github.com/dariomba/screen-go/actions/workflows/ci.yaml)
# Screen-Go

A self-hosted screenshot-as-a-service API built with Go. It uses headless Chrome to capture website screenshots and processes jobs asynchronously.

## Quick Start

### Using Docker Compose

1. Clone the repository:
   ```bash
   git clone https://github.com/dariomba/screen-go.git
   cd screen-go
   ```

2. Start the project:
   ```bash
   docker-compose up -d
   ```

3. Run database migrations:
   ```bash
   make migrate
   ```

The API will be available at `http://localhost:8080`.

### Local Development

1. Install dependencies:
   - Go 1.26+
   - Docker
   - Chrome/Chromium

2. Start the development server (includes Docker DB):
   ```bash
   make run
   ```

The API will be available at `http://localhost:8080`.

## Features

- Asynchronous job processing for screenshot requests
- Supports PNG and PDF output formats
- Full page screenshots (captures entire scrollable pages)
- Flexible storage backends (local filesystem, S3 compatible object storage)
- RESTful API with OpenAPI 3.0 specification

## Tech Stack

- **Language**: Go 1.26+
- **Database**: PostgreSQL and sqlc to generate SQL queries
- **Browser Driver**: Chromedp
- **Storage**: Local filesystem or MinIO
- **API**: OpenAPI 3.0 with oapi-codegen to generate server code
- **Migrations**: golang-migrate
- **Configuration**: Viper with environment variables
- **Logging**: Zerolog
- **CLI**: Cobra
- **Testing**: Testify, Testcontainers

## API Usage

### Submit a Screenshot Job

```bash
curl -X POST http://localhost:8080/v1/job \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://github.com",
    "format": "png",
    "width": 1280,
    "height": 800,
    "full_page": false
  }'
```

Response:
```json
{
  "job_id": "01J4KZQX8G3N2P7M5R9T0VWYE6",
  "status": "pending",
  "status_url": "/v1/job/01J4KZQX8G3N2P7M5R9T0VWYE6"
}
```

### Check Job Status

```bash
curl http://localhost:8080/v1/job/01J4KZQX8G3N2P7M5R9T0VWYE6
```

When complete, the response includes a `screenshot_url` to download the image.

### Download Screenshot

```bash
curl http://localhost:8080/v1/screenshot/01J4KZQX8G3N2P7M5R9T0VWYE6 > screenshot.png
```

## Development

### Code Generation

Generate OpenAPI server code and SQL queries:

```bash
make generate
```

### Database Migrations

Create a new migration:

```bash
migrate create -ext sql -dir tools/migrate -seq add_new_table
```

Run migrations:

```bash
make migrate
```

### Testing

Run unit tests:

```bash
make test
```

Run integration tests (requires Docker):

```bash
make test-integration
```

Run all tests:

```bash
make test-all
```

### Building

Build the binary:

```bash
go build -o bin/screen-go main.go
```

Build with Docker:

```bash
docker build -t screen-go .
```

## API Reference

The complete API documentation is available in the [OpenAPI specification](api/openapi.yaml).

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `make test-all`
6. Submit a pull request

## License

MIT License: see [LICENSE](LICENSE) file for details.
