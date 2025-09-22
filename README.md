# AICademy Backend

Backend API untuk aplikasi AICademy - Platform pembelajaran berbasis AI dengan fitur profiling karir dan rekomendasi role menggunakan Go dengan Clean Architecture pattern.

## 🚀 Tech Stack

- **Framework**: Fiber v2
- **Database**: PostgreSQL 15
- **ORM**: GORM
- **Authentication**: JWT
- **AI Integration**: Google Gemini API
- **Containerization**: Docker & Docker Compose
- **Language**: Go 1.24

## 🏗️ Project Structure

```
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   ├── database.go          # Database configuration & migrations
│   │   └── seeder.go            # Database seeding
│   ├── domain/
│   │   ├── admin/               # Admin domain (planned)
│   │   ├── alumni/              # Alumni domain (planned)
│   │   ├── auth/                # Authentication module
│   │   │   ├── dto.go           # Data Transfer Objects
│   │   │   ├── handler.go       # HTTP handlers
│   │   │   ├── repository.go    # Data access layer
│   │   │   └── service.go       # Business logic
│   │   ├── questionnaire/       # Career profiling questionnaire system
│   │   │   ├── dto.go           # Request/Response DTOs
│   │   │   ├── handler.go       # HTTP handlers
│   │   │   ├── model.go         # Domain models
│   │   │   ├── repository.go    # Data access layer
│   │   │   └── service.go       # Business logic
│   │   ├── student/             # Student domain (planned)
│   │   ├── teacher/             # Teacher domain (planned)
│   │   └── user/
│   │       └── model.go         # User domain models
│   ├── middleware/
│   │   └── auth.go              # Authentication middleware
│   ├── migration/               # Database migrations (planned)
│   ├── services/
│   │   └── ai/                  # AI service integration
│   │       ├── ai.go            # Gemini AI service
│   │       └── types.go         # AI service types
│   └── utils/
│       ├── emails.go            # Email utilities
│       ├── encoding.go          # Encoding utilities
│       ├── jwt.go               # JWT utilities
│       ├── password.go          # Password hashing
│       ├── response.go          # API response utilities
│       ├── token.go             # Token utilities
│       └── validation.go        # Validation utilities
├── docker-compose.yaml          # Docker services configuration
├── Dockerfile                   # Application container
├── go.mod                       # Go module dependencies
├── go.sum                       # Go module checksums
├── .env                         # Environment variables (ignored)
└── .env.example                 # Environment template
```

## ✨ Features

### 🔐 Authentication System
- Multi-role authentication (Student, Teacher, Alumni, Admin)
- JWT-based authentication
- Password reset functionality
- Secure password hashing with bcrypt

### 📋 Questionnaire System
- **AI-Powered Question Generation**: Dynamically generate career profiling questions using Gemini AI
- **Multi-Format Questions**: Support for MCQ, Likert scale, case studies, and text responses
- **Career Profiling**: Comprehensive personality and skill assessment
- **Role Recommendations**: AI-powered career path suggestions
- **Admin Management**: Full CRUD operations for questionnaires and roles

### 🤖 AI Integration
- **Google Gemini API**: Advanced language model for question generation and analysis
- **Intelligent Analysis**: Personality assessment and career matching
- **Customizable Prompts**: Template-based AI interactions
- **Async Processing**: Background AI processing for better performance

### 👥 User Management
- **Multi-Role System**: Students, Teachers, Alumni, Admin roles
- **Profile Management**: Role-specific user profiles
- **Secure Authentication**: JWT tokens with role-based access

## 📋 Prerequisites

- Go 1.24+
- Docker & Docker Compose
- PostgreSQL (if running locally)
- Google Gemini API Key (for AI features)

## 🔧 Environment Variables

Copy `.env.example` to `.env` and configure the following variables:

```bash
# Application Settings
APP_ENV=development
APP_PORT=8000
APP_NAME=AICademy
APP_HOST=0.0.0.0

# Database Configuration
DB_HOST=postgres
DB_USER=postgres
DB_PASSWORD=postgres123
DB_NAME=aicademy
DB_PORT=5432
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET=your_super_secure_jwt_secret_key_here_minimum_32_chars
JWT_EXPIRE_HOURS=24

# AI Service Configuration
AI_PROVIDER=gemini  # or "no_ai" for testing
GEMINI_API_KEY=your_gemini_api_key_here

# Email Configuration (Optional)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
EMAIL_FROM=noreply@aicademy.com
```

**Note**: Never commit your actual `.env` file to version control.

## 🚀 Installation & Setup

### Using Docker (Recommended)

1. **Clone the repository**
```bash
git clone <repository-url>
cd AICademy-backend
```

2. **Create environment file**
```bash
cp .env.example .env
# Edit .env with your actual configuration values
```

3. **Start services with Docker Compose**
```bash
docker-compose up -d
```

4. **Verify services are running**
```bash
docker-compose ps
```

### Local Development

1. **Install dependencies**
```bash
go mod download
```

2. **Setup environment variables**
```bash
cp .env.example .env
# Configure your .env file with actual values
```

3. **Start PostgreSQL database**
```bash
docker-compose up -d postgres
```

4. **Run the application**
```bash
go run cmd/server/main.go
```

## 🌐 Services

| Service | Port | Description |
|---------|------|-------------|
| Backend API | 8000 | Main application |
| PostgreSQL | 5432 | Database |
| pgAdmin | 8080 | Database management |

## 📚 API Endpoints

### 🔐 Authentication
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout
- `POST /api/v1/auth/refresh` - Refresh JWT token
- `POST /api/v1/auth/forgot-password` - Request password reset
- `POST /api/v1/auth/reset-password` - Reset password with token

### 📋 Questionnaire (Student)
- `GET /api/v1/questionnaire/active` - Get active questionnaire
- `POST /api/v1/questionnaire/submit` - Submit questionnaire answers
- `GET /api/v1/questionnaire/result/:responseId` - Get specific result
- `GET /api/v1/questionnaire/result/latest` - Get latest result

### 🛡️ Admin - Questionnaire Management
- `POST /api/v1/admin/questionnaires/generate` - Generate AI questionnaire
- `GET /api/v1/admin/questionnaires/generate/status/:id` - Check generation status
- `GET /api/v1/admin/questionnaires` - List all questionnaires
- `GET /api/v1/admin/questionnaires/:id` - Get questionnaire details
- `PUT /api/v1/admin/questionnaires/:id/activate` - Activate questionnaire
- `PUT /api/v1/admin/questionnaires/deactivate` - Deactivate all
- `DELETE /api/v1/admin/questionnaires/:id` - Delete questionnaire
- `GET /api/v1/admin/questionnaires/:id/responses` - Get questionnaire responses

### 🛡️ Admin - Role Management
- `GET /api/v1/admin/roles` - List all career roles
- `POST /api/v1/admin/roles` - Create new role
- `PUT /api/v1/admin/roles/:id` - Update role
- `DELETE /api/v1/admin/roles/:id` - Delete role

All protected routes require `Authorization: Bearer <token>` header.

## 👥 User Roles

The system supports the following user roles:

- **Student** - Take questionnaires and receive career recommendations
- **Teacher** - Access teaching-related features (planned)
- **Alumni** - Access alumni-specific features (planned)
- **Admin** - Full system management access

## 🗄️ Database

The application uses PostgreSQL with GORM as ORM. Database migrations are handled automatically on application startup.

### Database Models

- **Users & Profiles**: Multi-role user system with specific profiles
- **Questionnaires**: AI-generated career profiling questionnaires
- **Questions**: Multiple question types (MCQ, Likert, Case, Text)
- **Responses**: Student responses with AI analysis
- **Role Recommendations**: Career role definitions for AI matching

### pgAdmin Access
- URL: http://localhost:8080
- Email: admin@aicademy.com
- Password: admin123

## 🧪 Testing

### Quick API Testing

```bash
# Set base URL
export BASE_URL="http://localhost:8000/api/v1"

# Admin login
curl -X POST $BASE_URL/auth/login \
-H "Content-Type: application/json" \
-d '{"email": "admin@aicademy.com", "password": "Admin123!"}'

# Student login
curl -X POST $BASE_URL/auth/login \
-H "Content-Type: application/json" \
-d '{"email": "student1@aicademy.com", "password": "telkom@2025"}'

# Generate questionnaire (Admin)
curl -X POST $BASE_URL/admin/questionnaires/generate \
-H "Authorization: Bearer $ADMIN_TOKEN" \
-H "Content-Type: application/json" \
-d '{
  "name": "Career Profiling Questionnaire",
  "question_count": 5,
  "target_roles": ["Backend Developer", "Frontend Developer"],
  "difficulty_level": "intermediate"
}'

# Get active questionnaire (Student)
curl -X GET $BASE_URL/questionnaire/active \
-H "Authorization: Bearer $STUDENT_TOKEN"
```

### Running Unit Tests
```bash
go test ./...
```

## 🏭 Production

### Building for Production
```bash
go build -o bin/aicademy cmd/server/main.go
```

### Docker Production Build
```bash
docker-compose -f docker-compose.prod.yml up -d
```

## 🐳 Docker Commands

```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down

# Rebuild backend service
docker-compose up -d --build aicademy_api

# View logs
docker-compose logs -f aicademy_api

# Access database
docker-compose exec postgres psql -U postgres -d aicademy

# Fresh start (remove volumes)
docker-compose down -v && docker-compose up -d
```

## 🔒 Security Features

- **Password Security**: bcrypt hashing with salt
- **JWT Authentication**: Secure token-based authentication
- **Role-Based Access**: Endpoint protection by user roles
- **Data Privacy**: Sensitive data excluded from JSON responses
- **Environment Security**: Configuration via environment variables
- **Input Validation**: Comprehensive request validation

## 🤖 AI Integration Details

### Gemini AI Features
- **Dynamic Question Generation**: Create contextual career assessment questions
- **Personality Analysis**: Comprehensive psychological profiling
- **Career Matching**: Intelligent role recommendations based on responses
- **Adaptive Questioning**: Questions tailored to specific career paths

### AI Configuration
```bash
# Use Gemini AI (Recommended)
AI_PROVIDER=gemini
GEMINI_API_KEY=your_api_key

# Disable AI for testing
AI_PROVIDER=no_ai
```

## 🚀 Future Enhancements

- [ ] Real-time notifications
- [ ] Advanced analytics dashboard
- [ ] Learning path recommendations
- [ ] Integration with external learning platforms
- [ ] Mobile app support
- [ ] Multi-language support
- [ ] Advanced reporting system

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/new-feature`)
3. Commit changes (`git commit -am 'Add new feature'`)
4. Push to branch (`git push origin feature/new-feature`)
5. Create Pull Request

## 📄 License

This project is licensed under the MIT License.

## 📞 Support

For support, email admin@aicademy.com or create an issue in the repository.

---