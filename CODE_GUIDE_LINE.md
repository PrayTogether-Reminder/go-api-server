# Go + Gin Clean Architecture 코드 가이드라인

> **대상**: Java Spring Boot 개발자를 위한 Go + Gin 아키텍처 가이드
> **목적**: 코드 리뷰 시 Clean Architecture 준수 여부와 Best Practice 체크

## 🏗️ 프로젝트 아키텍처: Light Clean Architecture + 도메인별 수직 분할

이 프로젝트는 **Light Clean Architecture + 도메인별 수직 분할** 구조를 따릅니다:

```
internal/
├── member/                    # Member 도메인
│   ├── handler/               # Presentation Layer (HTTP Interface)
│   ├── usecase/               # Application Layer (Use Case Logic)
│   ├── domain/                # Domain (Business Logic + GORM Tags)
│   │   └── member.go          # Domain Entity with business methods
│   └── repository/            # Infrastructure Layer (Data Access)
├── room/                      # Room 도메인
│   ├── handler/
│   ├── usecase/
│   ├── domain/
│   └── repository/
└── shared/                    # 공통 인프라
    ├── domain/                # 공유 Base Entity (BaseEntity 등)
    ├── error/
    ├── http/
    ├── validator/
    └── database/
```

**핵심 원칙 (Light Clean - 실용적 접근):**
- ✅ 도메인별 독립적인 모듈 구성 (순환 참조 절대 금지)
- ✅ **Domain Entity는 GORM 태그 포함 가능** (실용적, 변환 로직 불필요)
- ✅ **비즈니스 로직은 Domain Entity 메서드로 구현** (Validate, HashPassword 등)
- ✅ UseCase는 다른 도메인 Repository 참조 가능 (UseCase 간 참조 금지)
- ✅ 의존성 방향: `Handler → UseCase → Repository` (간결한 단방향)
- ✅ 도메인 간 의존성: `member ← room ← prayer` (단방향만 허용)
- ✅ **Repository는 구현체 직접 의존 OK** (Interface 선택적)

---

## 📋 빠른 체크리스트

코드 리뷰 시 다음 항목들을 확인하세요:

### ✅ 아키텍처 체크리스트

- [ ] **올바른 레이어에 위치**하는가?
- [ ] **의존성 방향**이 올바른가? (Handler → UseCase → Repository)
- [ ] **레이어 간 책임 분리**가 명확한가?
- [ ] **순환 참조**가 없는가? (Domain Entity는 다른 도메인 Entity 직접 참조 금지)
- [ ] **Domain Entity에 GORM 태그**가 있는가? (Light Clean에서는 OK)
- [ ] **비즈니스 로직이 Domain Entity 메서드**로 구현되어 있는가?

### ✅ Go Best Practice

- [ ] **에러 처리**를 명시적으로 하는가?
- [ ] **Context 전파**가 올바른가?
- [ ] **nil 체크**를 하는가?
- [ ] **defer** 사용이 적절한가?
- [ ] **gofmt/goimports**를 통과하는가?

### ✅ Gin Best Practice

- [ ] **c.Request.Context()** 사용하는가?
- [ ] **gin.H vs struct** 선택이 적절한가?
- [ ] **HTTP 상태 코드**가 올바른가?
- [ ] **ShouldBindJSON** 에러 처리가 있는가?

---

## 🏗️ 레이어별 아키텍처 가이드

### 1️⃣ Handler Layer (Presentation / ≈ Spring Controller)

#### ✅ 올바른 예시

```go
// internal/member/handler/create.go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "your-project/internal/member/usecase"  // 같은 도메인의 usecase
)

type Handler struct {
    memberUseCase *usecase.MemberUseCase  // 같은 도메인 UseCase
}

func NewHandler(memberUseCase *usecase.MemberUseCase) *Handler {
    return &Handler{
        memberUseCase: memberUseCase,
    }
}

func (h *Handler) Create(c *gin.Context) {
    // 1. Request DTO 파싱
    var req CreateMemberRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }

    // 2. Context 추출
    ctx := c.Request.Context()

    // 3. UseCase 호출 (DTO → Domain Entity 변환)
    member, err := h.memberUseCase.Create(ctx, req.ToDomain())
    if err != nil {
        // 4. 에러 타입에 따른 HTTP 상태 코드 매핑
        switch {
        case errors.Is(err, usecase.ErrEmailAlreadyExists):
            c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        }
        return
    }

    // 5. Response DTO 변환 및 반환
    c.JSON(http.StatusCreated, NewMemberResponse(member))
}
```

#### ❌ 잘못된 예시

```go
// ❌ 비즈니스 로직이 Handler에 있음
func (h *Handler) Create(c *gin.Context) {
    var req CreateMemberRequest
    c.ShouldBindJSON(&req)

    // ❌ 비즈니스 검증이 Handler에 있음
    if len(req.Password) < 8 {
        c.JSON(400, gin.H{"error": "password too short"})
        return
    }

    // ❌ Repository 직접 호출
    if err := h.memberRepo.Create(req.ToDomain()); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
}
```

#### 🔍 체크포인트

| 항목 | 올바른 방법 | 잘못된 방법 |
|-----|-----------|-----------|
| **의존성** | UseCase만 의존 | Repository 직접 의존 |
| **Context** | `c.Request.Context()` 사용 | Context 무시 |
| **에러 처리** | 에러 타입 구분 + HTTP 상태 매핑 | 모든 에러 500 |
| **검증** | 형식 검증만 (JSON validation) | 비즈니스 검증 포함 |
| **응답** | DTO 변환 후 반환 | Domain Entity 직접 반환 |

#### 🆚 Spring Boot vs Go

| Spring Boot | Go + Gin |
|-------------|----------|
| `@RestController` | `handler` 패키지 |
| `@Autowired` | 생성자 DI |
| `@PostMapping` | `router.POST("/path", handler.Method)` |
| `@RequestBody` | `c.ShouldBindJSON(&req)` |
| `ResponseEntity<T>` | `c.JSON(status, data)` |
| Exception → `@ExceptionHandler` | `err → switch/if → HTTP status` |

---

### 2️⃣ UseCase Layer (Application / ≈ Spring Service)

UseCase는 **Application Logic**을 담당하며, Entity와 Repository를 조합하여 Use Case를 구현합니다.

#### ✅ 올바른 예시 (Light Clean)

```go
// internal/member/usecase/member_usecase.go
package usecase

import (
    "context"
    "fmt"
    "your-project/internal/member/domain"       // Domain
    "your-project/internal/member/repository"   // Repository 구현체 직접 의존
)

type MemberUseCase struct {
    memberRepo *repository.MemberRepository  // 구현체 직접 의존 (실용적)
    // 필요시 다른 도메인 Repository 의존 가능 (UseCase 간 의존 금지)
}

func NewMemberUseCase(memberRepo *repository.MemberRepository) *MemberUseCase {
    return &MemberUseCase{
        memberRepo: memberRepo,
    }
}

func (uc *MemberUseCase) Create(ctx context.Context, member *domain.Member) (*domain.Member, error) {
    // 1. Domain Entity 비즈니스 로직 호출
    if err := member.Validate(); err != nil {
        return nil, fmt.Errorf("회원 검증 실패: %w", err)
    }

    // 2. 중복 체크 (Application Logic)
    exists, err := uc.memberRepo.ExistsByEmail(ctx, db, member.Email)
    if err != nil {
        return nil, fmt.Errorf("이메일 중복 확인 실패: %w", err)
    }
    if exists {
        return nil, ErrEmailAlreadyExists
    }

    // 3. 비밀번호 해싱 (Domain Entity 메서드 호출)
    if err := member.HashPassword(); err != nil {
        return nil, fmt.Errorf("비밀번호 해싱 실패: %w", err)
    }

    // 4. Repository 호출
    if err := uc.memberRepo.Create(ctx, db, member); err != nil {
        return nil, fmt.Errorf("회원 생성 실패: %w", err)
    }

    return member, nil
}

// UseCase 레벨 에러 정의
var (
    ErrEmailAlreadyExists = errors.New("email already exists")
    ErrMemberNotFound     = errors.New("member not found")
)
```

### 3️⃣ Domain Layer (≈ Spring Entity with Business Logic)

Domain은 **비즈니스 로직 + GORM 매핑**을 담당합니다 (Light Clean에서는 실용적으로 통합).

#### ✅ Domain Entity 예시 (Light Clean - GORM 태그 포함)

```go
// internal/member/domain/member.go
package domain

import (
    "errors"
    "regexp"
    "strings"
    "golang.org/x/crypto/bcrypt"
)

// Member Entity (비즈니스 로직 + GORM 태그)
type Member struct {
    ID       int64  `gorm:"primaryKey;autoIncrement"`
    Email    string `gorm:"column:email;type:VARCHAR2(255);uniqueIndex;not null"`
    Name     string `gorm:"column:name;type:VARCHAR2(100);not null"`
    Password string `gorm:"column:password;type:VARCHAR2(255);not null"`

    // BaseEntity 임베딩 (선택적)
    // BaseEntity
}

// TableName - GORM 테이블명 매핑
func (*Member) TableName() string {
    return "member"
}

// Factory 함수 (생성자)
func NewMember(name, email, password string) (*Member, error) {
    member := &Member{
        Name:     strings.TrimSpace(name),
        Email:    strings.TrimSpace(strings.ToLower(email)),
        Password: password,
    }

    if err := member.Validate(); err != nil {
        return nil, err
    }

    return member, nil
}

// Validate - 비즈니스 규칙 검증
func (m *Member) Validate() error {
    if m.Name == "" {
        return errors.New("이름은 필수입니다")
    }
    if !emailRegex.MatchString(m.Email) {
        return errors.New("이메일 형식이 올바르지 않습니다")
    }
    if len(m.Password) < 8 {
        return errors.New("비밀번호는 8자 이상이어야 합니다")
    }
    return nil
}

// HashPassword - 비밀번호 해싱
func (m *Member) HashPassword() error {
    hashedPassword, err := bcrypt.GenerateFromPassword(
        []byte(m.Password),
        bcrypt.DefaultCost,
    )
    if err != nil {
        return err
    }
    m.Password = string(hashedPassword)
    return nil
}

// CheckPassword - 비밀번호 확인
func (m *Member) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(m.Password), []byte(password))
    return err == nil
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
```

**✅ Light Clean의 장점:**
- GORM 태그와 비즈니스 로직이 함께 있어도 OK (실용적)
- Entity 변환 로직 불필요
- 여전히 비즈니스 로직은 Entity에 캡슐화
- 테스트 가능 (GORM 태그는 테스트에 영향 없음)

#### ❌ 잘못된 예시 (UseCase)

```go
// ❌ HTTP 처리가 UseCase에 있음
func (uc *MemberUseCase) Create(c *gin.Context) {  // ❌ gin.Context 사용
    var member domain.Member
    c.ShouldBindJSON(&member)

    uc.memberRepo.Create(&member)

    c.JSON(200, member)  // ❌ HTTP 응답이 UseCase에
}

// ❌ SQL 쿼리가 UseCase에 있음
func (uc *MemberUseCase) GetByEmail(ctx context.Context, email string) (*domain.Member, error) {
    var member domain.Member
    // ❌ 직접 SQL 실행
    uc.db.Where("email = ?", email).First(&member)
    return &member, nil
}

// ❌ Context 무시
func (uc *MemberUseCase) Create(member *domain.Member) error {  // ❌ Context 없음
    return uc.memberRepo.Create(member)  // ❌ Context 전달 안 함
}
```

#### ❌ 잘못된 예시 (Domain)

```go
// ❌ 다른 도메인 Entity 직접 참조
// internal/member/domain/member.go
import "your-project/internal/room/domain"

type Member struct {
    ID    int64
    Rooms []domain.Room  // ❌ 다른 도메인 Entity 참조 (순환 참조 위험)
}

// ❌ 외부 의존성이 Domain에 있음
func (m *Member) Save() error {  // ❌ 저장 로직이 Entity에
    db := getDB()
    return db.Create(m).Error
}
```

#### 🔍 체크포인트

**UseCase Layer:**

| 항목 | 올바른 방법 | 잘못된 방법 |
|-----|-----------|-----------|
| **Context** | 첫 번째 파라미터 `ctx context.Context` | Context 없음 |
| **의존성** | Domain의 Repository Interface | DB 직접 접근, UseCase 간 의존 |
| **에러 처리** | `fmt.Errorf("...: %w", err)` 래핑 | `err` 그대로 반환 |
| **트랜잭션** | UseCase에서 시작/관리 | Repository에서 시작 |
| **비즈니스 로직** | Domain 호출 | UseCase에 비즈니스 로직 집중 |

**Domain Layer:**

| 항목 | 올바른 방법 | 잘못된 방법 |
|-----|-----------|-----------|
| **Entity** | 순수 비즈니스 객체 (GORM 태그 없음) | DB 태그 포함, 외부 의존성 |
| **메서드** | 비즈니스 로직만 (`Validate()`, `HashPassword()`) | 저장, 조회 등 인프라 로직 |
| **의존성** | 다른 도메인 참조 금지 (ID만) | 다른 도메인 Entity 직접 참조 |
| **Interface** | Repository Interface 정의 | 구현체 의존 |

#### 🔗 도메인 간 의존 규칙 (Clean Architecture 핵심)

**✅ 올바른 의존: Room UseCase가 Member Repository를 의존**

```go
// internal/room/usecase/room_usecase.go
package usecase

import (
    "context"
    "your-project/internal/room/repository"
    memberRepo "your-project/internal/member/repository"  // ✅ 다른 도메인 (단방향)
)

type RoomUseCase struct {
    roomRepo   *repository.RoomRepository           // 같은 도메인
    memberRepo *memberRepo.MemberRepository         // ✅ 다른 도메인 Repository (단방향)
    // ❌ memberUseCase는 의존하지 않음!
}

func (uc *RoomUseCase) AddMember(ctx context.Context, roomID, memberID int64) error {
    // 다른 도메인 Repository 사용 가능
    member, err := uc.memberRepo.GetByID(ctx, db, memberID)
    if err != nil {
        return err
    }

    return uc.roomRepo.AddMember(ctx, db, roomID, member.ID)  // ✅ ID만 전달
}
```

**✅ 올바른 Domain Entity 간 참조: ID만 사용**

```go
// internal/room/domain/member_room.go
package domain

type MemberRoom struct {
    MemberID int64  // ✅ ID만 참조 (import 불필요)
    RoomID   int64
    Role     string
}
```

**⚠️ 주의: 순환 참조 절대 금지**
```go
// ❌ 절대 금지
// internal/member/usecase/member_usecase.go
import "your-project/internal/room/usecase"

type MemberUseCase struct {
    roomUseCase *usecase.RoomUseCase  // ❌ UseCase 간 의존
}

// internal/room/usecase/room_usecase.go
import "your-project/internal/member/usecase"

type RoomUseCase struct {
    memberUseCase *usecase.MemberUseCase  // ❌ UseCase 간 의존
}
// → import cycle error!
```

#### 🆚 Spring Boot vs Go

**UseCase Layer:**

| Spring Boot | Go + Gin Clean Architecture |
|-------------|----------|
| `@Service` | `usecase` 패키지 |
| `@Transactional` | 수동 트랜잭션 (`db.Transaction(...)`) |
| Custom Exception | `var ErrXXX = errors.New("...")` |
| `Optional<T>` | `*T, error` 반환 |
| `@Async` | `go func() { ... }()` |
| Service → Service 의존 | ❌ 금지 (Repository만 의존) |

**Domain Layer:**

| Spring Boot | Go + Gin Clean Architecture |
|-------------|----------|
| `@Entity` | Domain Entity (순수 객체, GORM 태그 없음) |
| Entity Validation | `Validate()` 메서드 |
| Domain Service | Domain Service 패턴 (optional) |
| Repository Interface | Domain에 Interface 정의 |

---

### 4️⃣ Repository Layer (Infrastructure / ≈ Spring Repository)

Repository는 **데이터 접근**을 담당하며, Domain Entity를 직접 사용합니다 (Light Clean - 변환 불필요).

#### ✅ 올바른 예시 (Light Clean - Domain Entity 직접 사용)

```go
// internal/member/repository/member_repository.go
package repository

import (
    "context"
    "errors"
    "gorm.io/gorm"
    "your-project/internal/member/domain"  // Domain 직접 사용
)

// Repository 구현체
type MemberRepository struct {
    // DB는 UseCase에서 트랜잭션으로 전달받음
}

func NewMemberRepository() *MemberRepository {
    return &MemberRepository{}
}

// Create - Domain Entity 직접 저장 (변환 불필요)
func (r *MemberRepository) Create(ctx context.Context, db *gorm.DB, member *domain.Member) error {
    return db.WithContext(ctx).Create(member).Error  // ✅ Domain Entity 직접 사용
}

// GetByID - Domain Entity 직접 조회
func (r *MemberRepository) GetByID(ctx context.Context, db *gorm.DB, id int64) (*domain.Member, error) {
    var member domain.Member
    err := db.WithContext(ctx).First(&member, id).Error

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrMemberNotFound
        }
        return nil, err
    }

    return &member, nil
}

// GetByEmail - Domain Entity 직접 조회
func (r *MemberRepository) GetByEmail(ctx context.Context, db *gorm.DB, email string) (*domain.Member, error) {
    var member domain.Member
    err := db.WithContext(ctx).Where("email = ?", email).First(&member).Error

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrMemberNotFound
        }
        return nil, err
    }

    return &member, nil
}

// ExistsByEmail - 존재 여부 확인
func (r *MemberRepository) ExistsByEmail(ctx context.Context, db *gorm.DB, email string) (bool, error) {
    var count int64
    err := db.WithContext(ctx).
        Model(&domain.Member{}).
        Where("email = ?", email).
        Count(&count).Error

    return count > 0, err
}

// Repository 에러
var ErrMemberNotFound = errors.New("member not found")
```

**✅ Light Clean의 장점:**
- Domain Entity 변환 로직 불필요 (toEntity/toDomain 제거)
- 코드 간결
- GORM이 Domain Entity를 직접 매핑
- 여전히 레이어 분리 명확

#### ❌ 잘못된 예시

```go
// ❌ 비즈니스 로직이 Repository에 있음
func (r *memberRepository) Create(ctx context.Context, db *gorm.DB, member *domain.Member) error {
    // ❌ 비즈니스 검증이 Repository에
    if member.Age < 18 {
        return errors.New("too young")
    }

    // ❌ 비밀번호 해싱이 Repository에
    hashedPassword, _ := bcrypt.GenerateFromPassword(...)
    member.Password = string(hashedPassword)

    entity := r.toEntity(member)
    return db.Create(&entity).Error
}

// ❌ Context 무시
func (r *memberRepository) GetByID(id int64) (*domain.Member, error) {
    var entity memberEntity
    // ❌ WithContext 없음, db 매개변수도 없음
    r.db.First(&entity, id)
    return r.toDomain(&entity), nil
}

// ❌ 트랜잭션 시작
func (r *memberRepository) CreateBoth(ctx context.Context, member *domain.Member, room *domain.Room) error {
    // ❌ Repository에서 트랜잭션 시작 (UseCase에서 해야 함)
    return r.db.Transaction(func(tx *gorm.DB) error {
        r.Create(ctx, tx, member)
        r.Create(ctx, tx, room)
        return nil
    })
}

// ❌ Domain Entity를 직접 GORM에 사용
func (r *memberRepository) Create(ctx context.Context, db *gorm.DB, member *domain.Member) error {
    // ❌ Domain Entity를 DB에 직접 저장 (변환 없음)
    return db.WithContext(ctx).Create(member).Error
}
```

#### 🔍 체크포인트

| 항목 | 올바른 방법 | 잘못된 방법 |
|-----|-----------|-----------|
| **인터페이스** | Domain에 정의, Repository에서 구현 | 인터페이스 없음 |
| **Context** | 모든 메서드에 `ctx`, `db` 전달 | Context 무시 |
| **Entity 변환** | Domain ↔ DB Entity 변환 | Domain Entity 직접 사용 |
| **에러 변환** | DB 에러 → Domain 에러 | DB 에러 그대로 반환 |
| **책임** | 데이터 접근 + 변환만 | 비즈니스 로직 포함 |
| **트랜잭션** | 전달받은 `db` 사용 | Repository에서 시작 |

#### 🆚 Spring Boot vs Go

| Spring Boot | Go + Gin Clean Architecture |
|-------------|----------|
| `extends JpaRepository<T, ID>` | Domain에 Interface 정의, Repository에서 구현 |
| `findById(id)` | `GetByID(ctx, db, id)` |
| `existsByEmail(email)` | `ExistsByEmail(ctx, db, email)` |
| `@Query("SELECT ...")` | GORM 체이닝 |
| `Optional<T>` | `*T, error` |
| Entity → Repository 변환 | Domain Entity ↔ DB Entity 변환 |

---

## 🎯 Go/Gin 특화 Best Practices

### 1. 에러 처리

#### ✅ Go 스타일

```go
// 1. 에러 정의 (package level)
var (
    ErrNotFound      = errors.New("resource not found")
    ErrAlreadyExists = errors.New("resource already exists")
)

// 2. 에러 래핑 (context 추가)
if err != nil {
    return fmt.Errorf("failed to create member: %w", err)  // %w로 원본 에러 보존
}

// 3. 에러 체크
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, ErrNotFound
}

// 4. Custom 에러 (필요시)
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
```

#### ❌ Java 스타일 (안티패턴)

```go
// ❌ Exception 던지기 (Go에서는 panic 사용 지양)
func Create(member *Member) {
    if member == nil {
        panic("member is nil")  // ❌ 일반 에러에 panic 사용
    }
}

// ❌ try-catch 패턴 흉내
func Create() {
    defer func() {  // ❌ 일반 에러 처리를 defer/recover로
        if r := recover(); r != nil {
            log.Println("recovered:", r)
        }
    }()
}
```

### 2. Context 사용

#### ✅ 올바른 Context 사용

```go
// Handler에서 추출
func (h *Handler) Create(c *gin.Context) {
    ctx := c.Request.Context()  // ✅ Gin Context에서 추출
    result, err := h.useCase.Create(ctx, data)
}

// UseCase에서 전파
func (uc *MemberUseCase) Create(ctx context.Context, member *domain.Member) (*domain.Member, error) {
    // Context timeout/cancel 체크
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    return uc.memberRepo.Create(ctx, db, member)  // ✅ Repository로 전파
}

// Repository에서 사용
func (r *memberRepository) Create(ctx context.Context, db *gorm.DB, member *domain.Member) error {
    entity := r.toEntity(member)
    return db.WithContext(ctx).Create(&entity).Error  // ✅ DB에 전달
}
```

#### ❌ 잘못된 Context 사용

```go
// ❌ Context 무시
func (uc *MemberUseCase) Create(member *domain.Member) error {
    return uc.memberRepo.Create(member)  // ❌ Context 없음
}

// ❌ background context 남발
func (uc *MemberUseCase) Create(member *domain.Member) error {
    ctx := context.Background()  // ❌ 요청 Context 무시
    return uc.memberRepo.Create(ctx, db, member)
}

// ❌ Gin Context를 UseCase에 전달
func (h *Handler) Create(c *gin.Context) {
    h.useCase.Create(c, data)  // ❌ gin.Context 전달 (c.Request.Context() 사용해야)
}
```

### 3. Nil 체크

#### ✅ 올바른 Nil 체크

```go
func (uc *MemberUseCase) Create(ctx context.Context, member *domain.Member) (*domain.Member, error) {
    // 1. 포인터 nil 체크
    if member == nil {
        return nil, errors.New("member is nil")
    }

    // 2. Repository 호출 후 nil 체크
    result, err := uc.memberRepo.GetByEmail(ctx, db, member.Email)
    if err != nil {
        return nil, err
    }
    if result != nil {  // ✅ nil 체크
        return nil, ErrAlreadyExists
    }

    return member, nil
}
```

### 4. Pointer vs Value

#### 📌 일반 가이드라인

```go
// 구조체가 작고 불변: Value
type Point struct {
    X, Y int
}
func Distance(p Point) float64 { ... }  // ✅ Value

// 구조체가 크거나 수정 필요: Pointer
type Member struct {
    ID    int64
    Email string
    Name  string
    // ... 많은 필드
}
func Update(m *Member) error { ... }  // ✅ Pointer

// Repository 반환
func (r *Repo) GetByID(ctx context.Context, id int64) (*Member, error) {
    // ✅ 포인터 반환 (nil 가능, 수정 가능)
    return &member, nil
}
```

### 5. Defer 활용

#### ✅ Defer 올바른 사용

```go
// 1. 리소스 정리
func ProcessFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()  // ✅ 함수 종료 시 자동 close

    // 파일 처리...
    return nil
}

// 2. Lock 해제
func (s *Service) UpdateSafely() {
    s.mu.Lock()
    defer s.mu.Unlock()  // ✅ 함수 종료 시 자동 unlock

    // Critical section...
}

// 3. 트랜잭션 롤백 (GORM 예시는 자동이지만 수동 시)
func (s *Service) CreateWithTransaction(ctx context.Context) error {
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // 작업...
    return tx.Commit().Error
}
```

---

## 🚨 일반적인 실수 (Anti-patterns)

### 1. ❌ Handler에서 DB 직접 접근

```go
// ❌ 잘못됨
func (h *Handler) Create(c *gin.Context) {
    db := c.MustGet("db").(*gorm.DB)  // ❌
    var member model.Member
    c.ShouldBindJSON(&member)
    db.Create(&member)
    c.JSON(200, member)
}

// ✅ 올바름
func (h *Handler) Create(c *gin.Context) {
    var req CreateRequest
    c.ShouldBindJSON(&req)

    ctx := c.Request.Context()
    member, err := h.memberService.Create(ctx, req.ToModel())
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(201, NewMemberResponse(member))
}
```

### 2. ❌ Context에 값 저장 (DI 대신)

```go
// ❌ 잘못됨 - Middleware에서
func DBMiddleware(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Set("db", db)  // ❌ Context에 DB 저장
        c.Next()
    }
}

// Handler에서
func (h *Handler) Create(c *gin.Context) {
    db := c.MustGet("db").(*gorm.DB)  // ❌ Context에서 DB 꺼내기
}

// ✅ 올바름 - 생성자 DI
type Handler struct {
    memberService *service.MemberService  // ✅ 구조체 필드로 의존성
}

func NewHandler(memberService *service.MemberService) *Handler {
    return &Handler{memberService: memberService}
}
```

### 3. ❌ 모든 에러를 500으로 반환

```go
// ❌ 잘못됨
func (h *Handler) GetByID(c *gin.Context) {
    member, err := h.useCase.GetByID(c.Request.Context(), id)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})  // ❌ 모두 500
        return
    }
}

// ✅ 올바름
func (h *Handler) GetByID(c *gin.Context) {
    member, err := h.useCase.GetByID(c.Request.Context(), id)
    if err != nil {
        switch {
        case errors.Is(err, usecase.ErrNotFound):
            c.JSON(404, gin.H{"error": "Member not found"})
        case errors.Is(err, usecase.ErrInvalidID):
            c.JSON(400, gin.H{"error": "Invalid ID format"})
        default:
            c.JSON(500, gin.H{"error": "Internal server error"})
        }
        return
    }
    c.JSON(200, member)
}
```

### 4. ❌ Panic 남발

```go
// ❌ 잘못됨
func (uc *MemberUseCase) Create(member *domain.Member) {
    if member == nil {
        panic("member is nil")  // ❌ 일반 에러에 panic
    }
}

// ✅ 올바름
func (uc *MemberUseCase) Create(ctx context.Context, member *domain.Member) error {
    if member == nil {
        return errors.New("member is nil")  // ✅ error 반환
    }
    return nil
}

// ✅ Panic은 복구 불가능한 상황에만
func init() {
    if os.Getenv("REQUIRED_ENV") == "" {
        panic("REQUIRED_ENV is not set")  // ✅ 초기화 실패
    }
}
```

### 5. ❌ Domain Entity를 API Response로 직접 사용

```go
// ❌ 잘못됨
func (h *Handler) GetByID(c *gin.Context) {
    member, _ := h.useCase.GetByID(c.Request.Context(), id)
    c.JSON(200, member)  // ❌ Password 같은 민감 정보 노출
}

// ✅ 올바름
type MemberResponse struct {
    ID    int64  `json:"id"`
    Email string `json:"email"`
    Name  string `json:"name"`
    // Password 제외
}

func NewMemberResponse(m *domain.Member) *MemberResponse {
    return &MemberResponse{
        ID:    m.ID,
        Email: m.Email,
        Name:  m.Name,
    }
}

func (h *Handler) GetByID(c *gin.Context) {
    member, _ := h.useCase.GetByID(c.Request.Context(), id)
    c.JSON(200, NewMemberResponse(member))  // ✅ DTO 변환
}
```

---

## 📝 코드 리뷰 체크리스트 (요약)

### 전체 구조 (Clean Architecture + 도메인별 수직 분할)

```
✅ 파일이 올바른 패키지에 위치하는가?
   - Handler → internal/{domain}/handler/      예) internal/member/handler/
   - UseCase → internal/{domain}/usecase/      예) internal/member/usecase/
   - Domain → internal/{domain}/domain/        예) internal/member/domain/
   - Repository → internal/{domain}/repository/ 예) internal/member/repository/
   - Shared → internal/shared/                  (공통 인프라: error, http, validator, database)

✅ 의존성 방향이 올바른가?
   Handler → UseCase → Repository (Light Clean - 간결한 단방향)
   도메인 간: member ← room ← prayer (단방향만 허용, 순환 참조 절대 금지)

✅ UseCase 간 의존이 올바른가?
   - 같은 도메인 Repository: ✅ 허용
   - 다른 도메인 Repository: ✅ 허용 (단방향만)
   - 다른 도메인 UseCase: ❌ 절대 금지
   - 순환 참조: ❌ 절대 금지

✅ Domain Entity 간 참조가 올바른가?
   - ID만 참조: ✅ 허용
   - 다른 도메인 Entity 직접 참조: ❌ 절대 금지

✅ 순환 참조가 없는가?
   import cycle 체크
```

### Handler (Presentation Layer)

```
✅ UseCase만 의존하는가? (Repository 직접 접근 X)
✅ c.Request.Context() 사용하는가?
✅ ShouldBindJSON 에러 처리가 있는가?
✅ 에러 타입별 HTTP 상태 코드 매핑하는가?
✅ Response DTO로 변환하는가? (Domain Entity 직접 반환 X)
✅ 비즈니스 로직이 없는가?
```

### UseCase (Application Layer)

```
✅ 첫 번째 파라미터가 context.Context인가?
✅ Domain의 Repository Interface를 의존하는가?
✅ Domain Entity의 메서드를 호출하는가?
✅ 에러를 fmt.Errorf("...: %w", err)로 래핑하는가?
✅ Application Logic만 있는가? (Domain Logic은 Domain에)
✅ HTTP 관련 코드가 없는가? (gin.Context 사용 X)
✅ SQL 쿼리가 없는가? (Repository 사용)
✅ 다른 도메인 UseCase를 의존하지 않는가?
```

### Domain (Domain Layer)

```
✅ Domain Entity가 비즈니스 로직을 포함하는가?
✅ Factory 함수 (NewXXX)가 있는가?
✅ 비즈니스 로직이 Domain Entity 메서드로 구현되어 있는가? (Validate(), HashPassword() 등)
✅ 다른 도메인 Entity를 직접 참조하지 않는가? (ID만 참조)
✅ GORM 태그가 있는가? (Light Clean에서는 OK)
✅ 외부 의존성이 없는가? (HTTP, 외부 서비스 X)
```

### Repository (Infrastructure Layer)

```
✅ 구조체로 Repository를 구현하는가? (Interface는 선택적)
✅ 모든 메서드가 context.Context, *gorm.DB를 받는가?
✅ db.WithContext(ctx) 사용하는가?
✅ Domain Entity를 직접 사용하는가? (Light Clean - 변환 불필요)
✅ DB 에러를 적절히 처리하는가?
✅ 비즈니스 로직이 없는가?
✅ 트랜잭션을 시작하지 않는가? (UseCase에서 시작)
```

### Go Best Practice

```
✅ gofmt/goimports를 통과하는가?
✅ 에러를 명시적으로 처리하는가? (err 무시 X)
✅ nil 체크를 하는가?
✅ defer를 적절히 사용하는가?
✅ panic을 남발하지 않는가?
✅ Context를 전파하는가?
```

---

## 🎓 Spring Boot 개발자를 위한 용어 매핑

| Spring Boot | Go + Gin Clean Architecture | 설명 |
|-------------|----------|------|
| `@RestController` | Handler struct (Presentation) | HTTP 요청 처리 |
| `@Service` | UseCase struct (Application) | Application Logic |
| Domain Service | Domain Service (Domain) | 비즈니스 로직 |
| `@Repository` (Interface) | 선택적 Interface | Repository 인터페이스 (Light Clean에서는 선택적) |
| `@Repository` (Impl) | Repository struct (Infrastructure) | Repository 구현체 |
| `@Entity` | Domain Entity (GORM 태그 포함) | 도메인 엔티티 (Light Clean) |
| `@Autowired` | Constructor DI | 의존성 주입 |
| `@RequestBody` | `ShouldBindJSON(&req)` | Request body 파싱 |
| `ResponseEntity<T>` | `c.JSON(status, data)` | HTTP 응답 |
| `@Transactional` | `db.Transaction(func(tx) {...})` | 트랜잭션 (UseCase에서) |
| `Optional<T>` | `*T, error` | Nullable 타입 |
| Exception | `error` interface | 에러 처리 |
| `throw new XXXException()` | `return errors.New("...")` | 에러 반환 |
| `@ExceptionHandler` | Handler에서 switch/if | 에러 처리 |
| Lombok `@Data` | struct + tags | DTO 정의 |
| `@Valid` | `ShouldBindJSON` + validation | 입력 검증 |
| DDD Entity | Domain Entity | 비즈니스 로직 포함 |
| Aggregate Root | Domain Entity | 도메인 루트 |

---

## 🔗 참고 자료

- **프로젝트 아키텍처**: [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md)
- **Clean Architecture**: https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html
- **Uber Go Style Guide**: https://github.com/uber-go/guide
- **Effective Go**: https://golang.org/doc/effective_go.html
- **GORM 문서**: https://gorm.io/docs/
- **Gin 문서**: https://gin-gonic.com/docs/

---

## 💬 코드 리뷰 요청 템플릿

코드 리뷰를 요청할 때 다음과 같이 물어보세요:

```
@CODE_GUIDE_LINE.md 를 참고해서 내 코드를 리뷰해줘.

[리뷰 받고 싶은 파일 경로]
1. internal/member/handler/create.go
2. internal/member/usecase/member_usecase.go
3. internal/member/domain/member.go
4. internal/member/repository/member_repository.go

[확인하고 싶은 사항]
1. Clean Architecture 원칙을 잘 따르고 있나?
2. 도메인별 수직 분할 구조가 올바른가?
3. 의존성 방향이 올바른가? (의존성 역전 적용)
4. 순환 참조가 없는가?
5. Best Practice를 준수하고 있나?
6. Go 관용적 표현(idiomatic)을 사용하고 있나?
```

---

*Happy Coding! 🚀*
