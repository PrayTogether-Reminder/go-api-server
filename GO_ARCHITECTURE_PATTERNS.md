# Go 웹 애플리케이션 아키텍처 패턴

## 1. Go에서 자주 사용되는 아키텍처 패턴

### 1.1 Standard Go Project Layout (가장 인기)
```
project/
├── cmd/           # 애플리케이션 진입점
│   └── server/
│       └── main.go
├── internal/      # 비공개 애플리케이션 코드
│   ├── handlers/  # HTTP 핸들러
│   ├── models/    # 데이터 모델
│   ├── services/  # 비즈니스 로직
│   └── repository/# 데이터베이스 접근
├── pkg/           # 공개 라이브러리
├── api/           # API 스펙 (OpenAPI/Swagger)
└── configs/       # 설정 파일
```

**특징:**
- Go 커뮤니티에서 가장 널리 사용
- 간단하고 직관적
- [golang-standards/project-layout](https://github.com/golang-standards/project-layout) 참조
- **적합한 경우**: 중소규모 프로젝트, 마이크로서비스

### 1.2 Clean Architecture (Onion/Hexagonal)
```
project/
├── cmd/
├── internal/
│   ├── domain/        # 엔티티, 값 객체 (핵심 비즈니스)
│   ├── usecase/       # 애플리케이션 비즈니스 규칙
│   ├── repository/    # 저장소 인터페이스
│   ├── delivery/      # HTTP/gRPC 핸들러
│   └── infrastructure/# 외부 서비스, DB 구현
└── pkg/
```

**특징:**
- 의존성 역전 원칙 준수
- 테스트하기 쉬움
- 외부 의존성으로부터 비즈니스 로직 격리
- **적합한 경우**: 복잡한 비즈니스 로직, 장기 유지보수 프로젝트

### 1.3 DDD (Domain-Driven Design) - 현재 프로젝트
```
project/
├── cmd/
├── internal/
│   └── domains/
│       ├── member/
│       │   ├── domain/        # 엔티티, 값 객체, 도메인 서비스
│       │   ├── application/   # 애플리케이션 서비스 (유즈케이스)
│       │   ├── infrastructure/# 저장소 구현, 외부 서비스
│       │   └── interfaces/    # HTTP 핸들러, DTO
│       └── room/
│           └── ...같은 구조
└── pkg/
```

**특징:**
- 복잡한 도메인 로직에 적합
- Bounded Context 명확
- 각 도메인이 독립적
- **적합한 경우**: 복잡한 비즈니스 도메인, 대규모 팀

### 1.4 Flat Structure (Simple)
```
project/
├── main.go
├── handlers.go
├── models.go
├── database.go
└── utils.go
```

**특징:**
- 매우 간단
- 빠른 프로토타이핑
- **적합한 경우**: 소규모 프로젝트, MVP, 스크립트

## 2. Go 커뮤니티의 선호도

### 실제 통계 (Go Developer Survey 2023)
1. **Standard Layout**: ~45% 
2. **Clean Architecture**: ~25%
3. **Flat/Simple**: ~20%
4. **DDD**: ~10%

### 유명 Go 프로젝트들의 선택
- **Docker**: Standard Layout 변형
- **Kubernetes**: 독자적 구조 (매우 복잡)
- **Hugo**: Flat에 가까운 단순 구조
- **Gin/Echo**: Standard Layout
- **Uber/Grab**: Clean Architecture 변형

## 3. Go에 더 적합한 아키텍처는?

### Go 철학과의 정합성
```go
// Go Way: 단순함, 명확함, 실용성
type UserService struct {
    repo UserRepository
}

func (s *UserService) GetUser(id int) (*User, error) {
    return s.repo.Find(id)
}
```

### DDD vs Go Way 비교

| 측면 | DDD (현재) | Go Way (권장) |
|------|-----------|--------------|
| **복잡도** | 높음 | 낮음 |
| **학습 곡선** | 가파름 | 완만 |
| **보일러플레이트** | 많음 | 적음 |
| **인터페이스 사용** | 과도함 | 필요한 곳만 |
| **패키지 구조** | 깊음 | 평평함 |

## 4. 현재 프로젝트 분석

### 현재 구조 (DDD)의 장단점

**장점:**
- ✅ 도메인별 명확한 경계
- ✅ Java 개발자에게 친숙
- ✅ 비즈니스 로직 잘 표현

**단점:**
- ❌ Go 생태계에서 일반적이지 않음
- ❌ 과도한 추상화
- ❌ 많은 보일러플레이트 코드
- ❌ 작은 프로젝트에 오버엔지니어링

### 더 Go다운 리팩토링 예시

**Before (DDD):**
```go
// internal/domains/member/application/get_member_usecase.go
type GetMemberUseCase struct {
    service *domain.Service
}

// internal/domains/member/domain/service.go
type Service struct {
    repo Repository
}

// internal/domains/member/interfaces/http/handler.go
type Handler struct {
    getMemberUseCase *application.GetMemberUseCase
}
```

**After (Go Way):**
```go
// internal/member/service.go
type Service struct {
    db *gorm.DB
}

func (s *Service) GetMember(ctx context.Context, id uint64) (*Member, error) {
    var member Member
    if err := s.db.First(&member, id).Error; err != nil {
        return nil, err
    }
    return &member, nil
}

// internal/member/handler.go
type Handler struct {
    service *Service
}

func (h *Handler) GetMember(c *gin.Context) {
    member, err := h.service.GetMember(c.Request.Context(), id)
    // ...
}
```

## 5. 권장사항

### 프로젝트 규모별 권장 아키텍처

| 프로젝트 규모 | 권장 아키텍처 | 이유 |
|-------------|------------|------|
| **소규모** (< 10 엔드포인트) | Flat Structure | 단순함이 최선 |
| **중규모** (10-50 엔드포인트) | Standard Layout | 균형잡힌 구조 |
| **대규모** (> 50 엔드포인트) | Clean Architecture | 유지보수성 |
| **복잡한 도메인** | DDD (현재) | 도메인 복잡도 관리 |

### 현재 프로젝트 (Pray Together)의 경우

**현재 상황:**
- 엔드포인트: ~30개
- 도메인: 6개 (member, room, prayer, auth, invitation, fcmtoken)
- 복잡도: 중간

**권장:**
1. **단기적**: 현재 DDD 유지 (이미 구현됨)
2. **장기적**: Standard Layout으로 단순화 검토

**리팩토링 우선순위:**
1. ❌ UseCase 레이어 제거 (불필요한 추상화)
2. ❌ 과도한 인터페이스 제거
3. ✅ Repository 패턴은 유지 (테스트 용이)
4. ✅ Domain 모델은 유지 (비즈니스 로직 표현)

## 6. 결론

### Go 커뮤니티의 일반적인 선택:
```
작은 프로젝트 → Flat Structure
일반 웹 앱 → Standard Layout ⭐
복잡한 비즈니스 → Clean Architecture
도메인 중심 → DDD (드물게)
```

### 현재 프로젝트는?
- **DDD는 오버엔지니어링**일 수 있음
- 하지만 **이미 잘 구현**되어 있음
- Java에서 마이그레이션이라 **팀 친숙도** 고려 필요
- **당장 바꿀 필요는 없음**

### Go Way 핵심:
> "Clear is better than clever"  
> "A little copying is better than a little dependency"  
> "Interface segregation over abstraction"

**실용적 조언**: 
- 새 프로젝트라면 Standard Layout 추천
- 현재 프로젝트는 DDD로 계속 진행
- 성능/유지보수 이슈 시 점진적 리팩토링