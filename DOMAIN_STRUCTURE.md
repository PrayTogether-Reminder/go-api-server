# Domain-Based Modular Structure

## Overview
The application has been restructured from a layer-based architecture to a domain-based modular architecture following Domain-Driven Design (DDD) principles.

## Directory Structure

```
internal/
├── domains/           # Domain modules (each domain is self-contained)
│   ├── member/       # Member domain module
│   │   ├── domain/          # Domain layer (entities, domain logic, repository interfaces)
│   │   ├── application/     # Application layer (use cases)
│   │   ├── infrastructure/  # Infrastructure layer (repository implementations)
│   │   └── interfaces/      # Interface layer (public API, HTTP handlers)
│   │
│   ├── room/         # Room domain module (includes MemberRoom as part of aggregate)
│   │   ├── domain/
│   │   │   ├── room.go          # Room entity
│   │   │   ├── room_member.go   # RoomMember entity (part of Room aggregate)
│   │   │   ├── repository.go    # Repository interface
│   │   │   └── service.go       # Domain service
│   │   ├── application/
│   │   ├── infrastructure/
│   │   └── interfaces/
│   │
│   ├── prayer/       # Prayer domain module
│   │   ├── domain/
│   │   ├── application/
│   │   ├── infrastructure/
│   │   └── interfaces/
│   │
│   ├── auth/         # Auth domain module
│   │   ├── domain/
│   │   ├── application/
│   │   ├── infrastructure/
│   │   └── interfaces/
│   │
│   ├── invitation/   # Invitation domain module
│   │   ├── domain/
│   │   ├── application/
│   │   ├── infrastructure/
│   │   └── interfaces/
│   │
│   ├── fcmtoken/     # FCMToken domain module
│   │   ├── domain/
│   │   ├── application/
│   │   ├── infrastructure/
│   │   └── interfaces/
│   │
│   └── notification/ # Notification domain module
│       ├── domain/
│       ├── application/
│       ├── infrastructure/
│       └── interfaces/
│
└── shared/           # Shared resources
    ├── base/        # Base entities, value objects
    ├── errors/      # Common error definitions
    ├── config/      # Configuration
    └── utils/       # Utility functions

```

## Key Design Decisions

### 1. Domain-Based Structure
- Each domain is a self-contained module with its own 4-layer architecture
- Domains expose public interfaces through the `interfaces/api.go` file
- Cross-domain communication happens only through these public interfaces

### 2. MemberRoom Placement
- `MemberRoom` entity is placed within the Room domain as it's part of the Room aggregate
- This follows DDD principles where aggregate boundaries determine entity placement
- Room domain manages all room-related operations including member relationships

### 3. Cross-Domain Communication
- Domains communicate through defined interfaces (API contracts)
- Domain services handle cross-domain business logic
- No direct domain-to-domain dependencies at the application layer

### 4. Benefits of This Structure
- **High Cohesion**: Related functionality is grouped together
- **Low Coupling**: Domains are independent and communicate through interfaces
- **Team Ownership**: Each domain can be owned by a different team
- **Scalability**: Domains can be extracted into microservices if needed
- **Testability**: Each domain can be tested in isolation
- **Clear Boundaries**: Business boundaries are clearly defined

## Domain Responsibilities

### Member Domain
- Member registration and profile management
- Member authentication support
- Member information queries

### Room Domain (includes MemberRoom)
- Room creation and management
- Member-room relationships
- Room access control
- Prayer statistics tracking per member in room

### Prayer Domain
- Prayer creation and management
- Prayer sharing within rooms
- Prayer answering tracking
- Prayer statistics

### Auth Domain
- JWT token management
- Login/logout operations
- OTP verification
- Password reset

### Invitation Domain
- Room invitation management
- Invitation acceptance/rejection
- Invitation expiration handling

### FCMToken Domain
- FCM token registration
- Device management
- Token lifecycle management

### Notification Domain
- Notification creation and delivery
- Notification status tracking
- Read/unread management

## Inter-Domain Dependencies

```
Prayer -> Room (validate room access)
Prayer -> Member (get member info)
Auth -> Member (authenticate user)
Invitation -> Room (validate room)
Invitation -> Member (get member info)
Notification -> Member (target member)
Notification -> FCMToken (get device tokens)
```

## Migration Status
✅ All domains have been successfully migrated to the domain-based modular structure.