# AICademy Backend

Backend API untuk aplikasi AICademy menggunakan Go dengan Clean Architecture pattern.

## Tech Stack

- **Framework**: Fiber v2
- **Database**: PostgreSQL 15
- **ORM**: GORM
- **Authentication**: JWT
- **Containerization**: Docker & Docker Compose
- **Language**: Go 1.24

## Project Structure

```
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── database.go          # Database configuration
│   ├── domain/
│   │   ├── auth/                # Authentication module
│   │   │   ├── dto.go           # Data Transfer Objects
│   │   │   ├── handler.go       # HTTP handlers
│   │   │   ├── repository.go    # Data access layer
│   │   │   └── service.go       # Business logic
│   │   └── user/
│   │       └── model.go         # User domain model
│   ├── middleware/
│   │   └── auth.go              # Authentication middleware
│   ├── migration/               # Database migrations
│   └── utils/
│       ├── jwt.go               # JWT utilities
│       ├── password.go          # Password hashing utilities
│       └── response.go          # API response utilities
├── pkg/                         # External packages
├── docker-compose.yaml          # Docker services configuration
├── Dockerfile                   # Application container
├── go.mod                       # Go module dependencies
├── go.sum                       # Go module checksums
├── .env                         # Environment variables (ignored)
└── .env.example                 # Environment template
```

## Prerequisites

- Go 1.24+
- Docker & Docker Compose
- PostgreSQL (if running locally)

## Environment Variables

Copy `.env.example` to `.env` and configure the following variables:

```bash
# Application Settings
APP_ENV=development
APP_PORT=3000
APP_NAME=AICademy

# Database Configuration
DB_HOST=localhost
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=your_db_name
DB_PORT=5432
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET=your_secure_jwt_secret_key
JWT_EXPIRE_HOURS=24
```

**Note**: Never commit your actual `.env` file to version control.

## Installation & Setup

### Using Docker (Recommended)

1. Clone the repository
```bash
git clone <repository-url>
cd AICademy-backend
```

2. Create environment file
```bash
cp .env.example .env
# Edit .env with your actual configuration values
```

3. Start services with Docker Compose
```bash
docker-compose up -d
```

4. Verify services are running
```bash
docker-compose ps
```

### Local Development

1. Install dependencies
```bash
go mod download
```

2. Setup environment variables
```bash
cp .env.example .env
# Configure your .env file with actual values
```

3. Start PostgreSQL database
```bash
docker-compose up -d postgres
```

4. Run the application
```bash
go run cmd/server/main.go
```

## Services

| Service | Port | Description |
|---------|------|-------------|
| Backend API | 3000 | Main application |
| PostgreSQL | 5432 | Database |
| pgAdmin | 8080 | Database management |

## API Endpoints

### Authentication
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `POST /api/auth/refresh` - Refresh JWT token

### Protected Routes
All protected routes require `Authorization: Bearer <token>` header.

## User Roles

The system supports the following user roles:
- **Student** - Regular students
- **Teacher** - Faculty members
- **Alumni** - Graduates
- **Admin** - System administrators

## Database

The application uses PostgreSQL with GORM as ORM. Database migrations are handled automatically on application startup.

### pgAdmin Access
- URL: http://localhost:8080
- Default credentials are configured in docker-compose.yaml

## Development

### Running Tests
```bash
go test ./...
```

### Building for Production
```bash
go build -o bin/aicademy cmd/server/main.go
```

### Database Migration
Migrations are automatically executed on application startup.

## Docker Commands

```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down

# Rebuild backend service
docker-compose up -d --build backend

# View logs
docker-compose logs -f backend

# Access database
docker-compose exec postgres psql -U aicademy -d aicademy_db
```

## Security

- All passwords are hashed using bcrypt
- JWT tokens are used for authentication
- Sensitive data is excluded from JSON responses
- Environment variables are used for configuration

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/new-feature`)
3. Commit changes (`git commit -am 'Add new feature'`)
4. Push to branch (`git push origin feature/new-feature`)
5. Create Pull Request

## License

This project is licensed under the MIT License.