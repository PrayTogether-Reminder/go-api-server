# Pray Together API Server (Go)

## Overview
This is a Go implementation of the Pray Together API server, migrated from SpringBoot to Go using Domain-Driven Design (DDD) principles with a domain-based modular architecture.

## Architecture

### Domain-Based Modular Structure
Each domain is a self-contained module with its own 4-layer architecture:

```
internal/domains/
├── member/           # Member management domain
├── room/            # Room and member-room relationship domain
├── prayer/          # Prayer management domain
├── auth/            # Authentication and authorization domain
├── invitation/      # Room invitation domain
├── fcmtoken/        # FCM token management domain
└── notification/    # Notification management domain
```

### Each Domain Contains:
- **domain/**: Core business logic, entities, domain services
- **application/**: Use cases and application services
- **infrastructure/**: Repository implementations, external services
- **interfaces/**: Public APIs and HTTP handlers

## Key Design Principles

1. **Complete Domain Independence**: Each domain has its own BaseEntity and no shared dependencies
2. **Interface-Based Communication**: Domains communicate only through defined public interfaces
3. **Aggregate Design**: MemberRoom is part of Room aggregate (following DDD principles)
4. **Clean Architecture**: Clear separation of concerns across layers

## Technology Stack

- **Framework**: Gin (HTTP framework)
- **ORM**: GORM
- **Database**: PostgreSQL
- **Authentication**: JWT
- **Password Hashing**: bcrypt
- **Email**: SMTP
- **Push Notifications**: FCM
- **Configuration**: Environment variables (.env)

## Getting Started

### Prerequisites
- Go 1.21 or higher
- PostgreSQL
- Redis (optional, for caching)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd go-api-server
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Run database migrations:
```bash
go run cmd/server/main.go migrate
```

5. Start the server:
```bash
go run cmd/server/main.go
```

The server will start on port 8080 by default (configurable via PORT environment variable).

## API Endpoints

### Authentication
- `POST /api/v1/auth/signup` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout
- `POST /api/v1/auth/refresh` - Refresh token
- `POST /api/v1/auth/otp/send` - Send OTP
- `POST /api/v1/auth/otp/verify` - Verify OTP

### Members
- `GET /api/v1/members/me` - Get current user profile
- `GET /api/v1/members/:id` - Get member by ID
- `PUT /api/v1/members/me` - Update profile
- `DELETE /api/v1/members/me` - Delete account
- `GET /api/v1/members/search` - Search members

### Rooms
- `POST /api/v1/rooms` - Create room
- `GET /api/v1/rooms` - Get user's rooms
- `GET /api/v1/rooms/:id` - Get room details
- `PUT /api/v1/rooms/:id` - Update room
- `DELETE /api/v1/rooms/:id` - Delete room
- `POST /api/v1/rooms/:id/join` - Join room
- `POST /api/v1/rooms/:id/leave` - Leave room
- `GET /api/v1/rooms/:id/members` - Get room members
- `PUT /api/v1/rooms/:id/notification` - Update notification settings

### Prayers
- `POST /api/v1/prayers` - Create prayer
- `GET /api/v1/prayers` - Get prayers
- `GET /api/v1/prayers/:id` - Get prayer details
- `PUT /api/v1/prayers/:id` - Update prayer
- `DELETE /api/v1/prayers/:id` - Delete prayer
- `POST /api/v1/prayers/:id/complete` - Mark prayer as completed

### Invitations
- `POST /api/v1/invitations` - Create invitation
- `GET /api/v1/invitations` - Get invitations
- `PUT /api/v1/invitations/:id/accept` - Accept invitation
- `PUT /api/v1/invitations/:id/reject` - Reject invitation

### FCM Tokens
- `POST /api/v1/fcm-tokens` - Register FCM token
- `DELETE /api/v1/fcm-tokens` - Delete FCM token

## Domain Dependencies

```
Prayer → Room (validate access)
Prayer → Member (get member info)
Auth → Member (authenticate)
Invitation → Room, Member
Notification → Member, FCMToken
```

## Development

### Running Tests
```bash
go test ./...
```

### Building for Production
```bash
go build -o bin/server cmd/server/main.go
```

### Docker Support
```bash
docker build -t pray-together-api .
docker run -p 8080:8080 pray-together-api
```

## Migration from SpringBoot

This Go implementation maintains full feature parity with the original SpringBoot application:

- ✅ JWT-based authentication
- ✅ OTP email verification
- ✅ Room management with roles (Owner/Member)
- ✅ Prayer sharing within rooms
- ✅ Invitation system
- ✅ FCM push notifications
- ✅ Soft delete support
- ✅ Audit fields (created_at, updated_at, deleted_at)

## Benefits of Domain-Based Architecture

1. **Microservice Ready**: Each domain can be extracted into a separate service
2. **Team Scalability**: Different teams can own different domains
3. **Clear Boundaries**: Business logic is clearly separated
4. **Testability**: Each domain can be tested in isolation
5. **Maintainability**: Changes in one domain don't affect others

## License
[Your License]

## Contributing
[Contributing Guidelines]