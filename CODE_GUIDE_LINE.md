# Go + Gin Clean Architecture 코드 가이드라인

> **대상**: Java/Spring 경험이 있는 팀원이 Go + Gin 코드를 빠르게 이해하도록 돕는 문서  
> **목적**: 현재 프로젝트의 실제 패키지 구성과 흐름을 기준으로 Clean Architecture 원칙을 정리

---

## 1. 현재 프로젝트 레이아웃

이 레포는 “Light Clean Architecture + 도메인별 패키지” 방식을 사용합니다. 각 도메인은 별도의 패키지로 나누되, 파일은 필요 시 하나의 패키지 루트에 모읍니다.

```
internal/
├── auth/        # 인증 도메인 (handler, usecase, service, dto, errors) - Repository는 member에 의존
├── member/      # 회원 도메인 (handler, usecase, service, repository, dto, errors)
├── room/        # 방 도메인 (handler, usecase, service, repository, dto, errors)
├── model/       # 공용 도메인 엔티티 (Member, Room, MemberRoom, BaseEntity 등)
├── router/      # Gin 라우터와 의존성 주입
└── shared/      # 데이터베이스, 에러, HTTP 유틸, 로그, 토큰 등 공용 모듈
```

### 1.1 4계층 아키텍처 원칙

**모든 도메인 패키지(Auth, Member, Room)는 통일된 4계층 구조**를 따릅니다:

```
Handler (HTTP 계층) → UseCase (애플리케이션 계층) → Service (도메인 계층) → Repository (데이터 계층)
```

각 계층의 역할:
- **Handler**: HTTP 요청/응답 처리, DTO 변환, 인증 정보 추출
- **UseCase**: 트랜잭션 경계 관리, 여러 서비스 조율
- **Service**: 비즈니스 규칙 구현, 도메인 로직, Repository 호출
- **Repository**: 순수 데이터베이스 접근, GORM 기반 CRUD

### 1.2 Auth 패키지 구성

`internal/auth` 패키지는 로그인/회원가입 인증을 담당합니다.

| 파일 | 타입 | 역할 |
| --- | --- | --- |
| `handler.go` | `AuthHandler` | Gin handler (`Login`, `Signup`) |
| `usecase.go` | `AuthUseCase` | 인증 비즈니스 로직 조율 |
| `service.go` | `AuthService` | 비밀번호 해싱/비교 유틸리티 |
| `dto.go`, `errors.go` | - | HTTP 전용 DTO와 도메인 에러 |

**특징**: Auth는 별도 Repository가 없으며, 회원 데이터 접근 시 `member.MemberService`를 주입받아 사용합니다.

### 1.3 Member 패키지 구성

`internal/member` 패키지는 회원 도메인의 전체 생명주기를 관리합니다.

| 파일 | 타입 | 역할 |
| --- | --- | --- |
| `handler.go` | `MemberHandler` | Gin handler (`FetchProfile` 등) |
| `usecase.go` | `MemberUseCase` | 애플리케이션 유스케이스 (DB 트랜잭션 경계) |
| `service.go` | `MemberService` | 도메인 서비스 (비즈니스 규칙, Repository 호출) |
| `repository.go` | `MemberRepository` | GORM 기반 DAO |
| `dto.go`, `errors.go` | - | HTTP 전용 DTO와 도메인 에러 |
| `member.go` | - | `internal/model.Member` 를 노출하는 type alias 및 팩토리 |

### 1.4 Room 패키지 구성

`internal/room` 패키지는 기도방 도메인을 관리하며, 회원 검증이 필요한 경우 `member.MemberService`를 주입받습니다.

| 파일 | 타입 | 역할 |
| --- | --- | --- |
| `handler.go` | `RoomHandler` | Gin handler (방 조회, 생성, 나가기 등) |
| `usecase.go` | `RoomUseCase` | 애플리케이션 유스케이스 (DB 트랜잭션 경계) |
| `service.go` | `RoomService` | 도메인 서비스 (비즈니스 규칙, Repository 호출) |
| `repository.go` | `RoomRepository`, `MemberRoomRepository` | GORM 기반 DAO |
| `dto.go`, `errors.go` | - | HTTP 전용 DTO와 도메인 에러 |

### 1.5 공용 도메인 엔티티 (`internal/model`)

순환 의존을 방지하기 위해 모든 엔티티(`Member`, `Room`, `MemberRoom`)는 `internal/model` 패키지에 존재합니다. 각 도메인 패키지는 필요한 struct 를 import 하거나, `internal/member/member.go` 처럼 type alias 를 통해 외부 API를 유지합니다.

```
internal/model/
├── base.go          # BaseEntity (CreatedAt, UpdatedAt, ...)
├── member.go        # Member struct + NewMember
├── room.go          # Room struct + NewRoom
└── member_room.go   # MemberRoom struct
```

---

## 2. 의존성 방향

**모든 도메인은 동일한 4계층 구조를 따릅니다:**

```
Handler (HTTP) → UseCase (애플리케이션) → Service (도메인) → Repository (DB) → Model (Entity)
```

각 계층의 책임:
- **Handler** 는 DTO 변환과 HTTP 상태 코드를 담당합니다.
- **UseCase** 는 요청 단위 트랜잭션과 여러 서비스 조합을 담당합니다. (모든 도메인에서 사용)
- **Service** 는 비즈니스 규칙, 중복 검사, 도메인 오류 변환을 담당합니다.
- **Repository** 는 반드시 `context.Context` 와 `*gorm.DB` 를 받아 DB 접근만 수행합니다.
- **Model** 은 GORM 태그가 붙은 순수 구조체입니다. 서로 다른 도메인이 해당 struct 를 공유합니다.

### 도메인 간 의존성 규칙

- Auth → Member, Room → Member 처럼 다른 도메인의 Service가 필요할 때는 **단방향 참조만 허용**합니다.
- UseCase ↔ UseCase, Service ↔ Service 간 **직접 의존은 금지**합니다.
- 도메인 의존 방향: `Auth → Member ← Room` (Member가 가장 낮은 계층)

---

## 3. 빠른 체크리스트

- [ ] 파일이 올바른 패키지에 위치하는가? (`internal/auth`, `internal/member`, `internal/room`, `internal/model`, `internal/shared/...`)
- [ ] **모든 도메인이 4계층 구조**를 따르는가? (Handler → UseCase → Service → Repository)
- [ ] **명명 규칙**을 지키는가? (`{Domain}Handler`, `{Domain}UseCase`, `{Domain}Service`, `{Domain}Repository`)
- [ ] `ctx := c.Request.Context()`를 핸들러에서 받은 뒤 모든 레이어에 전달하는가?
- [ ] Repository 는 `db.WithContext(ctx)` 를 사용하고, 트랜잭션은 UseCase가 시작하는가?
- [ ] 도메인 오류는 `internal/shared/error` 에 등록된 코드로 변환되는가?
- [ ] DTO/Request struct 는 HTTP 레이어(`dto.go`)에 있고, Domain struct 는 `internal/model` 에 있는가?
- [ ] 도메인 간 의존성이 단방향인가? (Auth → Member ← Room)
- [ ] `gofmt`, `go test` 등을 통과하는가?

---

## 4. 레이어별 가이드

### 4.1 Handler (`internal/auth/handler.go`, `internal/member/handler.go`, `internal/room/handler.go`)

- Gin Context에서 memberID, query, body를 추출하고 DTO로 변환합니다.
- `shared/http` 패키지의 `RequireMemberID`, `BindJSON`, `BindURI` 등을 사용합니다.
- UseCase/Service 오류는 `shared/error.ResolveDomainError`로 HTTP 상태 코드에 매핑합니다.
- Handler는 비즈니스 규칙을 포함하지 않으며 Repository를 직접 호출하지 않습니다.
- **모든 Handler는 `{Domain}Handler` 명명 규칙을 따릅니다** (예: `AuthHandler`, `MemberHandler`, `RoomHandler`)

```go
func (h *MemberHandler) FetchProfile(c *gin.Context) {
    memberID, ok := sharedHttp.RequireMemberID(c)
    if !ok {
        return
    }

    profile, err := h.memberUseCase.GetProfile(c.Request.Context(), memberID)
    if err != nil {
        if resp, ok := sharedError.ResolveDomainError(err); ok {
            sharedHttp.RespondError(c, err, resp)
            return
        }
        sharedHttp.RespondError(c, err, sharedError.InternalServerError)
        return
    }

    c.JSON(http.StatusOK, &FetchProfileResponse{
        ID:          profile.ID,
        Name:        profile.Name,
        Email:       profile.Email,
        PhoneNumber: profile.PhoneNumber,
    })
}
```

### 4.2 UseCase (`internal/auth/usecase.go`, `internal/member/usecase.go`, `internal/room/usecase.go`)

- 애플리케이션 흐름과 트랜잭션을 담당합니다.
- `database.WithTransaction` 또는 `gorm.DB` 를 직접 전달해 여러 Repository 호출을 묶습니다.
- UseCase는 다른 도메인의 UseCase/Handler와 직접 연결하지 않습니다.
- **모든 UseCase는 `{Domain}UseCase` 명명 규칙을 따릅니다** (예: `AuthUseCase`, `MemberUseCase`, `RoomUseCase`)

```go
type MemberUseCase struct {
    db            *gorm.DB
    memberService *MemberService
}

func (u *MemberUseCase) Signup(ctx context.Context, member *model.Member) error {
    return database.WithTransaction(ctx, u.db, func(tx *gorm.DB) error {
        return u.memberService.CreateMember(ctx, tx, member)
    })
}
```

### 4.3 Service (`internal/auth/service.go`, `internal/member/service.go`, `internal/room/service.go`)

- 도메인 규칙을 중심으로 Repository 호출과 오류 변환을 수행합니다.
- `member.ErrMemberNotFound`, `room.ErrRoomNotFound` 등 도메인 오류를 정의하고 `shared/error`에 등록합니다.
- 다른 도메인 리포지토리가 필요하면 인터페이스보다는 실제 구현체를 주입받되, 순환 의존이 생기지 않도록 주의합니다.
- **모든 Service는 `{Domain}Service` 명명 규칙을 따릅니다** (예: `AuthService`, `MemberService`, `RoomService`)

```go
func (s *MemberService) GetByEmail(ctx context.Context, db *gorm.DB, email string) (*model.Member, error) {
    member, err := s.memberRepository.FindByEmail(ctx, db, email)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("회원을 찾을 수 없습니다: %w", ErrMemberNotFound)
        }
        return nil, fmt.Errorf("회원 조회 실패: %w", err)
    }
    return member, nil
}
```

### 4.4 Repository (`internal/member/repository.go`, `internal/room/repository.go`)

- `context.Context` + `*gorm.DB` 를 인자로 받아야 하며, 트랜잭션 경계는 상위 레이어가 관리합니다.
- GORM 모델은 `internal/model` 의 struct를 직접 사용합니다.
- DB 접근 외에 비즈니스 로직을 포함하지 않습니다.

```go
func (m *MemberRepository) FindByID(ctx context.Context, db *gorm.DB, id int64) (*model.Member, error) {
    var member model.Member
    if err := db.WithContext(ctx).Where("id = ?", id).First(&member).Error; err != nil {
        return nil, err
    }
    return &member, nil
}
```

### 4.5 Domain Model (`internal/model`)

- Member, Room, MemberRoom, BaseEntity 등이 존재하며 모든 도메인에서 공유합니다.
- gRPC/REST 등 외부 표현과는 분리된 순수 데이터 구조체입니다.
- 새로운 엔티티를 추가할 때에는 반드시 이 패키지에 정의하고, 필요 시 각 도메인 패키지에서 type alias 로 노출합니다.

```go
type Member struct {
    ID          int64
    Email       string `gorm:"column:email;uniqueIndex:idx_member_email;not null"`
    Name        string `gorm:"column:name;not null"`
    PhoneNumber string `gorm:"column:phone_number;not null"`
    Password    string `gorm:"column:password;not null"`
    model.BaseEntity
}
```

### 4.6 Shared 패키지

- `shared/http`: 요청 파싱, 공통 응답, 멤버 ID 추출
- `shared/error`: 도메인 오류 등록 및 HTTP 매핑
- `shared/database`: DB 연결, 트랜잭션 helper, 마이그레이션
- `shared/logger`, `shared/token`, `shared/testutil` 등

필요한 공용 기능을 먼저 `shared`에 추가하고 각 도메인에서 재사용합니다.

---

## 5. 베스트 프랙티스

1. **에러 처리**: `fmt.Errorf("context: %w", err)`로 감싸고, 도메인 오류는 `shared/error`에 등록합니다.  
2. **Context 전파**: Handler → UseCase → Service → Repository 순으로 Context를 그대로 넘깁니다. `context.Background()`를 새로 만들지 않습니다.  
3. **DTO ↔ Model 분리**: HTTP DTO는 `internal/{domain}/dto.go` 또는 핸들러 파일에 두고, DB 모델은 `internal/model`에 둡니다.  
4. **트랜잭션**: `database.WithTransaction` 또는 `db.Transaction`으로 묶고 Repository에는 `tx`를 넘깁니다.  
5. **의존성 주입**: `internal/router/routes.go`에서 한 번만 생성하여 Gin 라우터에 주입합니다.  
6. **테스트**: `internal/shared/testutil` 의 DB helper, member/room builder를 활용해 단위/통합 테스트를 작성합니다.

---

## 6. 확장 가이드

### 새로운 도메인을 추가하려면?

**모든 도메인은 통일된 4계층 구조를 따라야 합니다:**

1. `internal/<domain>` 패키지를 만들고 다음 파일들을 추가합니다:
   - `handler.go` - `{Domain}Handler` 구조체와 HTTP 핸들러 메서드
   - `usecase.go` - `{Domain}UseCase` 구조체와 비즈니스 로직 조율
   - `service.go` - `{Domain}Service` 구조체와 도메인 서비스
   - `repository.go` - `{Domain}Repository` 구조체와 데이터 접근 (필요시)
   - `dto.go` - HTTP 요청/응답 DTO
   - `errors.go` - 도메인별 에러 정의

2. 필요한 엔티티를 `internal/model`에 정의합니다.

3. `internal/router/routes.go`에서 의존성을 주입하고 라우트를 등록합니다:
   ```go
   // Repository 생성
   domainRepo := domain.NewDomainRepository()

   // Service 생성 (Repository 주입)
   domainService := domain.NewDomainService(domainRepo)

   // UseCase 생성 (DB와 Service 주입)
   domainUseCase := domain.NewDomainUseCase(db, domainService)

   // Handler 생성 (UseCase 주입)
   domainHandler := domain.NewDomainHandler(domainUseCase)
   ```

4. `shared/error`에 도메인 오류 코드를 등록합니다.

5. 테스트는 `internal/<domain>/*_test.go` 또는 `internal/shared/testutil`을 활용합니다.

### 명명 규칙

**모든 도메인 구조체는 `{Domain}{Layer}` 패턴을 따릅니다:**
- Handler: `AuthHandler`, `MemberHandler`, `RoomHandler`
- UseCase: `AuthUseCase`, `MemberUseCase`, `RoomUseCase`
- Service: `AuthService`, `MemberService`, `RoomService`
- Repository: `MemberRepository`, `RoomRepository`

### Repository가 없는 경우 (Auth 패키지 참고)

도메인이 자체 데이터를 저장하지 않고 다른 도메인에 의존하는 경우:
- Repository 파일을 생성하지 않습니다.
- 필요한 다른 도메인의 Service를 UseCase에 주입받아 사용합니다.
- 예: `AuthUseCase`는 `member.MemberService`를 주입받아 회원 데이터에 접근합니다.

---

## 7. 참고 링크

- [Gin 공식 문서](https://gin-gonic.com/docs/)
- [GORM 가이드](https://gorm.io/docs/)
- [Go Context 가이드](https://go.dev/blog/context)

이 문서는 실제 코드와 함께 지속적으로 업데이트됩니다. 구조가 변경되면 반드시 본 문서에 반영해 주세요.
