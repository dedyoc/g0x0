# g0x0 - File Hosting Service

g0x0 is a lightweight, self-hosted file sharing service built with Go and PostgreSQL. It provides temporary file hosting with configurable expiration times, secret links, and a simple web interface.

## Features

- **Fast file uploads** - Efficient file handling with SHA256 deduplication
- **Secret links** - Optional secret tokens for private file sharing
- **Automatic expiration** - Files expire based on size and custom settings
- **Deduplication** - Identical files share storage space
- **RESTful API** - Simple HTTP API for programmatic access
- **Web interface** - User-friendly upload form with HTMX
- **Docker support** - Easy deployment with Docker Compose
- **PostgreSQL backend** - Reliable metadata storage

## Quick Start

### Using Docker Compose (Recommended)

1. Clone the repository:
```bash
git clone https://github.com/dedyoc/g0x0.git
cd g0x0
```

2. Start the services:
```bash
docker compose up -d
```

3. Access the web interface at `http://localhost:3000`

### Manual Setup

#### Prerequisites
- Go 1.23+
- PostgreSQL 15+

#### Installation

1. Clone and build:
```bash
git clone https://github.com/dedyoc/g0x0.git
cd g0x0
make build
```

2. Set up PostgreSQL and run the initialization script:
```bash
psql -U postgres -f sql/init.sql
```

3. Configure environment variables (optional):
```bash
export DATABASE_URL="postgres://g0x0:g0x0@localhost:5432/g0x0?sslmode=disable"
export STORAGE_PATH="./uploads"
export MAX_FILE_SIZE="268435456"  # 256MB
```

4. Run the server:
```bash
make run
```

## API Reference

### Upload File

**POST /** 

Upload a file with optional parameters.

#### Parameters (multipart/form-data)
- `file` (required) - The file to upload
- `expires` (optional) - Expiration time in hours
- `secret` (optional) - Set to "true" to create a secret link

#### Response
```json
{
  "url": "/2843fd09-80c2-4994-bef2-4a04b8bb7ab6",
  "expires": 1782016144,
  "token": "VpGawgWwWlAp0lT7uaEijPhDsC3oVJZz"
}
```

#### Example
```bash
# Basic upload
curl -F "file=@example.txt" http://localhost:8080/

# Upload with custom expiration (2 hours)
curl -F "file=@example.txt" -F "expires=2" http://localhost:8080/

# Upload with secret link
curl -F "file=@example.txt" -F "secret=true" http://localhost:8080/
```

### Download File

**GET /:id**

Download a file by its ID.

#### Example
```bash
curl http://localhost:8080/2843fd09-80c2-4994-bef2-4a04b8bb7ab6
```

### Download Secret File

**GET /s/:secret/:id**

Download a file using its secret token.

#### Example
```bash
curl http://localhost:8080/s/Cy-moJpCwDzXJyTa/4af952b1-07c1-4580-a278-929162c054f7
```

### File Management

**POST /:id**

Manage a file using its management token.

#### Parameters (multipart/form-data)
- `token` (required) - Management token returned during upload
- `delete` (optional) - Set to "true" to delete the file

#### Example
```bash
# Delete a file
curl -X POST -F "token=VpGawgWwWlAp0lT7uaEijPhDsC3oVJZz" -F "delete=true" \
  http://localhost:8080/2843fd09-80c2-4994-bef2-4a04b8bb7ab6
```

### Health Check

**GET /health**

Check service health.

#### Response
```json
{
  "status": "ok"
}
```

## Configuration

Configure the service using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://g0x0:g0x0@localhost:5432/g0x0?sslmode=disable` | PostgreSQL connection string |
| `STORAGE_PATH` | `./uploads` | Directory to store uploaded files |
| `MAX_FILE_SIZE` | `268435456` | Maximum file size in bytes (256MB) |
| `MAX_EXPIRATION` | `8760h` | Maximum file lifetime (1 year) |
| `MIN_EXPIRATION` | `720h` | Minimum file lifetime (30 days) |
| `SECRET_BYTES` | `16` | Length of secret tokens |
| `PORT` | `8080` | Server port |

## File Expiration

Files have automatic expiration based on their size:

- **Large files** (near max size): Expire closer to the minimum expiration time
- **Small files**: Can be kept longer, up to the maximum expiration time
- **Custom expiration**: Users can set custom expiration (limited by file size rules)

## Architecture

### Technology Stack
- **Backend**: Go 1.23 with Echo framework
- **Database**: PostgreSQL 15 with UUID extension
- **Frontend**: HTML with HTMX for dynamic interactions
- **ORM**: Jet for type-safe SQL queries
- **Containerization**: Docker and Docker Compose

### Project Structure
```
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection
│   ├── features/
│   │   └── files/       # File handling logic
│   ├── server/          # HTTP server setup
│   └── utils/           # Utility functions
├── sql/                 # Database schemas
├── web/templates/       # HTML templates
├── uploads/             # File storage (created at runtime)
└── docker-compose.yml   # Docker setup
```

### Key Features Implementation

#### File Deduplication
Files are identified by SHA256 hash. When uploading a duplicate file:
- Storage is shared between instances
- Expiration time is updated to the latest upload
- New management tokens are not issued for existing files

#### Security
- Files can be protected with secret tokens
- Management tokens allow file deletion
- IP addresses and user agents are logged
- File access is validated against expiration and removal status

#### Storage Management
- Files are stored using their SHA256 hash as filename
- Database tracks metadata separately from file content
- Removed files are marked in database but deleted from filesystem

## Development

### Build Commands
```bash
# Build binary
make build

# Run in development mode
make dev

# Run tests
make test

# Clean build artifacts
make clean

# Start Docker services
make docker-up

# Stop Docker services
make docker-down
```

### Database Management

The service uses PostgreSQL with the following tables:

- **files**: Stores file metadata, tokens, and expiration info
- **urls** (WIP): Reserved for future URL shortening features

Database migrations are handled via the `sql/init.sql` file.

## API Testing Examples

Here are some practical examples for testing the API:

```bash
# Health check
curl http://localhost:8080/health

# Upload a text file
echo "Hello World" > test.txt
curl -F "file=@test.txt" http://localhost:8080/

# Upload with 1-hour expiration
curl -F "file=@test.txt" -F "expires=1" http://localhost:8080/

# Upload a secret file
curl -F "file=@test.txt" -F "secret=true" http://localhost:8080/

# Download file (replace with actual ID)
curl http://localhost:8080/FILE_ID_HERE

# Download secret file (replace with actual secret and ID)
curl http://localhost:8080/s/SECRET_HERE/FILE_ID_HERE

# Delete file (replace with actual ID and token)
curl -X POST -F "token=TOKEN_HERE" -F "delete=true" \
  http://localhost:8080/FILE_ID_HERE

# Clean up
rm test.txt
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/dedyoc/g0x0).
