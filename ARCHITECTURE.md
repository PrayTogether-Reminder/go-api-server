# Go Clean Architecture

## 아키텍처 개요

이 프로젝트는 Clean Architecture 원칙을 Go에 맞게 실용적으로 구현합니다.

### 핵심 원칙
1. **의존성 규칙**: 의존성은 외부에서 내부로만 향함
2. **비즈니스 독립성**: 핵심 비즈니스는 프레임워크, DB, UI와 무관
3. **테스트 용이성**: 각 레이어를 독립적으로 테스트 가능
4. **유연성**: 프레임워크나 DB를 쉽게 교체 가능

## 레이어 구조

```
┌─────────────────────────────────────────────────┐
│              Delivery Layer                      │
│         (HTTP/gRPC Handlers, DTOs, Router)      │
├─────────────────────────────────────────────────┤
│        UseCase Layer (Application Layer)         │
│            (Business Use Cases)                  │
├─────────────────────────────────────────────────┤
│               Domain Layer                       │
│        (Entities, Repository Interface)          │
├─────────────────────────────────────────────────┤
│            Infrastructure Layer                  │
│         (Database, External Services)            │
└─────────────────────────────────────────────────┘

→ 의존성 방향: 외부 → 내부
```

## 디렉토리 구조

```
.
├── cmd/
│   └── server/
│       └── main.go              # 애플리케이션 진입점
├── internal/
│   ├── domain/                  # 도메인 레이어 (핵심 비즈니스)
│   │   ├── entity/              # 비즈니스 엔티티
│   │   │   └── user.go
│   │   ├── repository/          # 리포지토리 인터페이스
│   │   │   └── user.go
│   │   └── service/             # 도메인 서비스 (선택적)
│   ├── usecase/                 # 유스케이스 레이어 (= Application Layer)
│   │   └── user/                # 유저 관련 유스케이스
│   │       ├── create.go
│   │       ├── get.go
│   │       └── update.go
│   ├── delivery/                # Delivery 레이어 (외부 인터페이스)
│   │   ├── http/                # HTTP API
│   │   │   ├── handler/         # HTTP 핸들러
│   │   │   │   └── user.go
│   │   │   ├── middleware/      # HTTP 미들웨어
│   │   │   │   └── auth.go
│   │   │   ├── router/          # 라우터 설정
│   │   │   │   └── router.go
│   │   │   └── dto/             # Request/Response DTO
│   │   │       └── user.go
│   │   └── grpc/                # gRPC API (옵션)
│   │       └── server.go
│   └── infrastructure/          # 인프라 레이어
│       ├── persistence/         # 데이터베이스 구현
│       │   ├── postgres/        # PostgreSQL 구현
│       │   │   └── user.go
│       │   └── memory/          # 인메모리 구현 (테스트용)
│       │       └── user.go
│       ├── database/            # DB 연결 관리
│       │   └── postgres.go
│       └── config/              # 설정 관리
│           └── config.go
├── migrations/                  # DB 마이그레이션
├── Makefile
├── go.mod
└── go.sum
```

## 레이어별 상세 설명

### 1. Domain Layer (내부 핵심)
**위치**: `internal/domain/`  
**책임**: 핵심 비즈니스 규칙과 엔티티  
**의존성**: 없음 (가장 독립적인 레이어)

```go
// internal/domain/entity/user.go
package entity

import (
    "errors"
    "time"
)

var (
    ErrInvalidEmail = errors.New("invalid email")
    ErrWeakPassword = errors.New("password too weak")
)

// User 엔티티 - 외부 의존성 없음
type User struct {
    ID        string
    Email     string
    Name      string
    Password  string // hashed
    CreatedAt time.Time
    UpdatedAt time.Time
}

func NewUser(email, name, password string) (*User, error) {
    // 비즈니스 규칙 검증
    if email == "" || !isValidEmail(email) {
        return nil, ErrInvalidEmail
    }
    
    if len(password) < 8 {
        return nil, ErrWeakPassword
    }
    
    return &User{
        Email:     email,
        Name:      name,
        Password:  password,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }, nil
}

// internal/domain/repository/user.go
package repository

import (
    "context"
    "github.com/changhyeonkim/pray-together/go-api-server/internal/domain/entity"
)

type UserRepository interface {
    Create(ctx context.Context, user *entity.User) error
    GetByID(ctx context.Context, id string) (*entity.User, error)
    GetByEmail(ctx context.Context, email string) (*entity.User, error)
    Update(ctx context.Context, user *entity.User) error
    Delete(ctx context.Context, id string) error
}
```

### 2. UseCase Layer (Application Layer)
**위치**: `internal/usecase/`  
**책임**: 비즈니스 플로우 조정  
**의존성**: Domain Layer
**참고**: UseCase Layer = Application Layer (같은 레이어의 다른 이름)

```go
// internal/usecase/user/create.go
package user

import (
    "context"
    "errors"
    
    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
    
    "github.com/changhyeonkim/pray-together/go-api-server/internal/domain/entity"
    "github.com/changhyeonkim/pray-together/go-api-server/internal/domain/repository"
)

type CreateUserUseCase struct {
    userRepo repository.UserRepository
}

func NewCreateUserUseCase(userRepo repository.UserRepository) *CreateUserUseCase {
    return &CreateUserUseCase{
        userRepo: userRepo,
    }
}

// UseCase는 domain entity를 직접 받고 반환
func (uc *CreateUserUseCase) Execute(ctx context.Context, email, name, password string) (*entity.User, error) {
    // 1. 도메인 엔티티 생성
    user, err := entity.NewUser(email, name, password)
    if err != nil {
        return nil, err
    }
    
    // 2. 중복 체크
    existing, _ := uc.userRepo.GetByEmail(ctx, email)
    if existing != nil {
        return nil, errors.New("email already exists")
    }
    
    // 3. 패스워드 해싱
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }
    
    // 4. ID 생성 및 저장
    user.ID = uuid.New().String()
    user.Password = string(hashedPassword)
    
    if err := uc.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    return user, nil
}
```

### 3. Delivery Layer (HTTP Handler)
**위치**: `internal/delivery/`  
**책임**: HTTP 요청/응답 처리  
**의존성**: Application Layer

```go
// internal/delivery/http/handler/user.go
package handler

import (
    "encoding/json"
    "net/http"
    
    "github.com/go-chi/chi/v5"
    
    "github.com/changhyeonkim/pray-together/go-api-server/internal/usecase/user"
)

type UserHandler struct {
    createUserUC *user.CreateUserUseCase
    getUserUC    *user.GetUserUseCase
}

func NewUserHandler(createUserUC *user.CreateUserUseCase, getUserUC *user.GetUserUseCase) *UserHandler {
    return &UserHandler{
        createUserUC: createUserUC,
        getUserUC:    getUserUC,
    }
}

// internal/delivery/http/dto/user.go
// DTO는 Delivery 레이어에 속함 (HTTP Request/Response 변환용)
type CreateUserRequest struct {
    Email    string `json:"email"`
    Name     string `json:"name"`
    Password string `json:"password"`
}

type UserResponse struct {
    ID    string `json:"id"`
    Email string `json:"email"`
    Name  string `json:"name"`
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // UseCase 호출 (primitive types 전달)
    user, err := h.createUserUC.Execute(r.Context(), req.Email, req.Name, req.Password)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Domain Entity를 Response DTO로 변환
    response := UserResponse{
        ID:    user.ID,
        Email: user.Email,
        Name:  user.Name,
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    userID := chi.URLParam(r, "id")
    
    user, err := h.getUserUC.Execute(r.Context(), userID)
    if err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    
    // Domain Entity를 Response DTO로 변환
    response := UserResponse{
        ID:    user.ID,
        Email: user.Email,
        Name:  user.Name,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// internal/delivery/http/router/router.go
package router

import (
    "net/http"
    "time"
    
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    
    "github.com/changhyeonkim/pray-together/go-api-server/internal/delivery/http/handler"
)

func NewRouter(userHandler *handler.UserHandler) http.Handler {
    r := chi.NewRouter()
    
    // 미들웨어
    r.Use(middleware.RequestID)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(60 * time.Second))
    
    // 헬스체크
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })
    
    // API 라우트
    r.Route("/api/v1", func(r chi.Router) {
        r.Route("/users", func(r chi.Router) {
            r.Post("/", userHandler.CreateUser)
            r.Get("/{id}", userHandler.GetUser)
        })
    })
    
    return r
}
```

### 4. Infrastructure Layer (데이터베이스)
**위치**: `internal/infrastructure/`  
**책임**: 외부 시스템과의 연결  
**의존성**: Domain Layer

```go
// internal/infrastructure/persistence/postgres/user.go
package postgres

import (
    "context"
    "database/sql"
    
    "github.com/changhyeonkim/pray-together/go-api-server/internal/domain/entity"
    "github.com/changhyeonkim/pray-together/go-api-server/internal/domain/repository"
)

type userRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
    query := `
        INSERT INTO users (id, email, name, password, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
    _, err := r.db.ExecContext(ctx, query,
        user.ID, user.Email, user.Name, user.Password,
        user.CreatedAt, user.UpdatedAt,
    )
    return err
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
    var user entity.User
    query := `
        SELECT id, email, name, password, created_at, updated_at
        FROM users WHERE id = $1
    `
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &user.ID, &user.Email, &user.Name, &user.Password,
        &user.CreatedAt, &user.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &user, err
}

// internal/infrastructure/database/postgres.go
package database

import (
    "database/sql"
    "fmt"
    "time"
    
    _ "github.com/lib/pq"
)

type Config struct {
    Host     string
    Port     int
    User     string
    Password string
    DBName   string
    SSLMode  string
}

func NewPostgresDB(cfg Config) (*sql.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
    )
    
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }
    
    // 연결 풀 설정
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    // 연결 테스트
    if err := db.Ping(); err != nil {
        return nil, err
    }
    
    return db, nil
}
```

### 5. Main (의존성 주입)
**위치**: `cmd/server/main.go`  
**책임**: 모든 레이어 조립

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/changhyeonkim/pray-together/go-api-server/internal/usecase/user"
    "github.com/changhyeonkim/pray-together/go-api-server/internal/delivery/http/handler"
    "github.com/changhyeonkim/pray-together/go-api-server/internal/delivery/http/router"
    "github.com/changhyeonkim/pray-together/go-api-server/internal/infrastructure/config"
    "github.com/changhyeonkim/pray-together/go-api-server/internal/infrastructure/database"
    "github.com/changhyeonkim/pray-together/go-api-server/internal/infrastructure/persistence/postgres"
)

func main() {
    // 1. 설정 로드
    cfg := config.Load()
    
    // 2. 데이터베이스 연결
    db, err := database.NewPostgresDB(database.Config{
        Host:     cfg.DBHost,
        Port:     cfg.DBPort,
        User:     cfg.DBUser,
        Password: cfg.DBPassword,
        DBName:   cfg.DBName,
        SSLMode:  cfg.DBSSLMode,
    })
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()
    
    // 3. 의존성 주입
    // Repository
    userRepo := postgres.NewUserRepository(db)
    
    // UseCase
    createUserUC := user.NewCreateUserUseCase(userRepo)
    getUserUC := user.NewGetUserUseCase(userRepo)
    
    // Handler
    userHandler := handler.NewUserHandler(createUserUC, getUserUC)
    
    // Router
    r := router.NewRouter(userHandler)
    
    // 4. HTTP 서버
    srv := &http.Server{
        Addr:         ":" + cfg.Port,
        Handler:      r,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    // Graceful shutdown
    go func() {
        log.Printf("Starting server on port %s", cfg.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("Failed to start server:", err)
        }
    }()
    
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
    
    log.Println("Server exited")
}
```

## 실무 가이드

### 레이어 책임 정리

| 레이어 | 책임 | 의존성 |
|--------|------|--------|
| Domain | 핵심 비즈니스 규칙 | 없음 |
| UseCase (=Application) | 비즈니스 플로우 조정 | Domain |
| Delivery | HTTP/gRPC 처리, DTO | UseCase |
| Infrastructure | DB, 외부 서비스 | Domain |

### 개발 순서
1. Domain 엔티티 정의
2. Repository 인터페이스 정의
3. UseCase 구현
4. Repository 구현 (Infrastructure)
5. Handler 구현 (Delivery)
6. Router 설정
7. Main에서 DI

### 테스트 전략
- **Domain**: 단위 테스트 (의존성 없음)
- **UseCase**: Mock Repository 사용
- **Handler**: Mock UseCase 사용
- **Repository**: 통합 테스트 (실제 DB)

## 체크리스트

### 클린 아키텍처 준수
- [ ] Domain이 외부 의존성이 없는가?
- [ ] UseCase가 HTTP를 모르는가?
- [ ] Repository 인터페이스가 Domain에 있는가?
- [ ] Handler가 비즈니스 로직을 포함하지 않는가?

### 실용성
- [ ] 과도한 추상화를 피했는가?
- [ ] 새 기능 추가가 쉬운가?
- [ ] 테스트하기 쉬운가?
- [ ] 팀원이 이해하기 쉬운가?