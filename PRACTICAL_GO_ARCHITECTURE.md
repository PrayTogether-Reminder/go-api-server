# 실무적인 Go 아키텍처 재구성 가이드

## 현재 구조 (DDD) vs 제안 구조 (Practical Go)

### 현재 DDD 구조의 문제점
```
internal/domains/member/
├── domain/           # 과도한 추상화
├── application/      # 불필요한 UseCase 레이어
├── infrastructure/   # Repository 구현
└── interfaces/       # HTTP 핸들러
```
- 4개 레이어 = 과도한 복잡도
- 작은 기능도 4개 파일 수정 필요
- Go의 철학과 맞지 않음

## 🎯 실무적인 Go 아키텍처 제안

### 1. Service-Oriented Architecture (추천 ⭐)

```
pray-together/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── auth/
│   │   ├── service.go      # 인증 비즈니스 로직
│   │   ├── handler.go      # HTTP 핸들러
│   │   ├── middleware.go   # JWT 미들웨어
│   │   └── types.go        # Request/Response 타입
│   ├── member/
│   │   ├── service.go      # 회원 비즈니스 로직
│   │   ├── handler.go      # HTTP 핸들러
│   │   ├── model.go        # 데이터 모델
│   │   └── types.go        # DTO
│   ├── room/
│   │   ├── service.go
│   │   ├── handler.go
│   │   ├── model.go
│   │   └── types.go
│   ├── prayer/
│   │   ├── service.go
│   │   ├── handler.go
│   │   ├── model.go
│   │   └── types.go
│   └── shared/
│       ├── database/
│       │   └── connection.go
│       ├── response/
│       │   └── response.go  # 공통 응답 포맷
│       └── validator/
│           └── validator.go
├── pkg/
│   ├── firebase/            # 외부 라이브러리 래퍼
│   └── email/
└── config/
    └── config.go
```

### 2. 실제 코드 예시

#### Before (현재 DDD)
```go
// 회원 조회를 위해 4개 파일 필요

// 1. internal/domains/member/application/get_member_usecase.go
type GetMemberUseCase struct {
    service *domain.Service
}
func (uc *GetMemberUseCase) Execute(ctx context.Context, id uint64) (*dto.MemberResponse, error) {
    member, err := uc.service.GetMemberByID(ctx, id)
    // ...
}

// 2. internal/domains/member/domain/service.go
type Service struct {
    repo Repository
}
func (s *Service) GetMemberByID(ctx context.Context, id uint64) (*Member, error) {
    return s.repo.FindByID(ctx, id)
}

// 3. internal/domains/member/infrastructure/gorm_repository.go
func (r *GormRepository) FindByID(ctx context.Context, id uint64) (*domain.Member, error) {
    var member Member
    err := r.db.First(&member, id).Error
    // ...
}

// 4. internal/domains/member/interfaces/http/handler.go
func (h *Handler) GetMember(c *gin.Context) {
    result, err := h.getMemberUseCase.Execute(c.Request.Context(), id)
    // ...
}
```

#### After (Practical Go)
```go
// 회원 조회를 위해 2개 파일만 필요

// 1. internal/member/service.go
type Service struct {
    db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
    return &Service{db: db}
}

func (s *Service) GetMember(ctx context.Context, id uint64) (*Member, error) {
    var member Member
    if err := s.db.WithContext(ctx).First(&member, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrMemberNotFound
        }
        return nil, err
    }
    return &member, nil
}

func (s *Service) CreateMember(ctx context.Context, req *CreateMemberRequest) (*Member, error) {
    // 검증
    if err := req.Validate(); err != nil {
        return nil, err
    }
    
    // 비즈니스 로직
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }
    
    member := &Member{
        Email:    req.Email,
        Name:     req.Name,
        Password: string(hashedPassword),
    }
    
    // 저장
    if err := s.db.Create(member).Error; err != nil {
        return nil, err
    }
    
    return member, nil
}

// 2. internal/member/handler.go
type Handler struct {
    service *Service
}

func NewHandler(service *Service) *Handler {
    return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
    r.GET("/members/:id", h.GetMember)
    r.POST("/members", h.CreateMember)
    r.PUT("/members/:id", h.UpdateMember)
    r.DELETE("/members/:id", h.DeleteMember)
}

func (h *Handler) GetMember(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(400, gin.H{"error": "invalid id"})
        return
    }
    
    member, err := h.service.GetMember(c.Request.Context(), id)
    if err != nil {
        if errors.Is(err, ErrMemberNotFound) {
            c.JSON(404, gin.H{"error": "member not found"})
            return
        }
        c.JSON(500, gin.H{"error": "internal server error"})
        return
    }
    
    c.JSON(200, ToMemberResponse(member))
}
```

### 3. Room 도메인 예시

```go
// internal/room/service.go
type Service struct {
    db           *gorm.DB
    memberSvc    *member.Service  // 의존성 주입
    notification Notifier         // 인터페이스만 의존
}

func (s *Service) CreateRoom(ctx context.Context, creatorID uint64, req *CreateRoomRequest) (*Room, error) {
    // 트랜잭션 시작
    tx := s.db.Begin()
    defer tx.Rollback()
    
    // 방 생성
    room := &Room{
        Name:        req.Name,
        Description: req.Description,
        CreatorID:   creatorID,
    }
    
    if err := tx.Create(room).Error; err != nil {
        return nil, err
    }
    
    // 생성자를 멤버로 추가
    roomMember := &RoomMember{
        RoomID:   room.ID,
        MemberID: creatorID,
        Role:     "OWNER",
    }
    
    if err := tx.Create(roomMember).Error; err != nil {
        return nil, err
    }
    
    tx.Commit()
    return room, nil
}

func (s *Service) InviteMember(ctx context.Context, roomID, inviterID uint64, email string) error {
    // 권한 확인
    if !s.isMemberOfRoom(ctx, roomID, inviterID) {
        return ErrUnauthorized
    }
    
    // 회원 찾기
    invitee, err := s.memberSvc.GetMemberByEmail(ctx, email)
    if err != nil {
        return err
    }
    
    // 초대 로직
    // ...
    
    // 알림 발송
    s.notification.Send(ctx, invitee.ID, "You have been invited to a room")
    
    return nil
}
```

### 4. 의존성 주입 및 초기화

```go
// cmd/server/main.go
func main() {
    // 설정 로드
    cfg := config.Load()
    
    // 데이터베이스 연결
    db := database.Connect(cfg.Database)
    
    // 서비스 초기화 (의존성 주입)
    memberSvc := member.NewService(db)
    roomSvc := room.NewService(db, memberSvc, notificationSvc)
    prayerSvc := prayer.NewService(db, roomSvc, notificationSvc)
    authSvc := auth.NewService(db, memberSvc, jwtService)
    
    // 핸들러 초기화
    memberHandler := member.NewHandler(memberSvc)
    roomHandler := room.NewHandler(roomSvc)
    prayerHandler := prayer.NewHandler(prayerSvc)
    authHandler := auth.NewHandler(authSvc)
    
    // 라우터 설정
    r := gin.New()
    
    api := r.Group("/api/v1")
    {
        // Public routes
        authHandler.RegisterPublicRoutes(api)
        
        // Protected routes
        protected := api.Group("", auth.JWTMiddleware(jwtService))
        memberHandler.RegisterRoutes(protected)
        roomHandler.RegisterRoutes(protected)
        prayerHandler.RegisterRoutes(protected)
    }
    
    r.Run(":8080")
}
```

### 5. 테스트 구조

```go
// internal/member/service_test.go
func TestMemberService_CreateMember(t *testing.T) {
    // Given
    db := setupTestDB()
    service := NewService(db)
    
    req := &CreateMemberRequest{
        Email:    "test@test.com",
        Name:     "Test User",
        Password: "password123",
    }
    
    // When
    member, err := service.CreateMember(context.Background(), req)
    
    // Then
    assert.NoError(t, err)
    assert.NotNil(t, member)
    assert.Equal(t, "test@test.com", member.Email)
}

// internal/member/handler_test.go  
func TestMemberHandler_GetMember(t *testing.T) {
    // Mock service
    mockService := &MockMemberService{}
    handler := NewHandler(mockService)
    
    // Setup router
    r := gin.New()
    handler.RegisterRoutes(r.Group("/"))
    
    // Test
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/members/1", nil)
    r.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
}
```

## 🔑 핵심 원칙

### 1. 패키지 설계
- **기능별 패키지**: auth, member, room, prayer
- **공통 기능**: shared 패키지
- **외부 의존성**: pkg 패키지

### 2. 레이어 단순화
- **Before**: Domain → Application → Infrastructure → Interface (4 레이어)
- **After**: Service → Handler (2 레이어)

### 3. 의존성 관리
```go
// 인터페이스는 사용하는 곳에 정의
type Notifier interface {
    Send(ctx context.Context, userID uint64, message string) error
}

// 구체 타입 직접 의존도 OK (Go way)
type RoomService struct {
    db        *gorm.DB
    memberSvc *member.Service  // 구체 타입 OK
}
```

### 4. 에러 처리
```go
// 패키지 레벨 에러 정의
var (
    ErrMemberNotFound = errors.New("member not found")
    ErrUnauthorized   = errors.New("unauthorized")
    ErrInvalidInput   = errors.New("invalid input")
)
```

## 📊 비교 분석

| 측면 | 현재 (DDD) | 제안 (Practical Go) | 개선도 |
|------|-----------|-------------------|--------|
| **파일 수** | ~200개 | ~50개 | 75% 감소 |
| **코드 라인** | ~15,000 | ~5,000 | 66% 감소 |
| **복잡도** | 높음 | 낮음 | ⭐⭐⭐ |
| **테스트 용이성** | 좋음 | 좋음 | 동일 |
| **성능** | 보통 | 좋음 | 개선 |
| **Go 관례** | 낮음 | 높음 | ⭐⭐⭐ |

## 🚀 마이그레이션 전략

### Phase 1: 새 기능은 새 구조로
```go
// 새 기능 추가 시
internal/
├── notification/  # 새 구조
│   ├── service.go
│   └── handler.go
└── domains/       # 기존 구조 유지
    └── ...
```

### Phase 2: 점진적 리팩토링
1. UseCase 레이어 제거
2. Domain/Infrastructure 통합
3. Interface 단순화

### Phase 3: 완전 이전
- 모든 도메인을 새 구조로
- 테스트 커버리지 유지

## 결론

### 실무적 선택:
1. **신규 프로젝트**: Service-Oriented 구조 사용
2. **기존 프로젝트**: 점진적 마이그레이션
3. **팀 상황**: Go 경험이 적으면 현재 구조 유지

### Go다운 코드의 특징:
- ✅ 단순하고 직관적
- ✅ 보일러플레이트 최소화
- ✅ 필요한 곳에만 추상화
- ✅ 실용적이고 명확함

> "Simplicity is the ultimate sophistication" - Go Proverbs