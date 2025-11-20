# 🧪 Go 테스트 완벽 가이드

> 처음 보는 사람도 이해할 수 있도록 작성된 Go 테스트 실행 흐름 설명서

## 📚 목차
1. [Go 테스트 기본 원리](#1-go-테스트-기본-원리)
2. [테스트 파일 인식과 실행](#2-테스트-파일-인식과-실행)
3. [우리 프로젝트의 테스트 구조](#3-우리-프로젝트의-테스트-구조)
4. [단계별 실행 흐름](#4-단계별-실행-흐름)
5. [Mock 객체의 동작 원리](#5-mock-객체의-동작-원리)
6. [실제 HTTP 요청 시뮬레이션](#6-실제-http-요청-시뮬레이션)

---

## 1. Go 테스트 기본 원리

### 1.1 테스트 함수의 규칙

Go는 **컨벤션(Convention)** 기반으로 테스트를 인식합니다.

```go
// ✅ 올바른 테스트 함수
func TestSignup_Success(t *testing.T) {
    // 테스트 코드
}

// ❌ 테스트로 인식 안됨
func testSignup(t *testing.T)        // Test로 시작 안함
func TestSignup(t string)             // *testing.T가 아님
func Signup_Test(t *testing.T)        // Test로 시작 안함
```

**필수 규칙:**
1. 함수명이 `Test`로 시작
2. 파라미터가 정확히 `t *testing.T` 하나
3. 파일명이 `_test.go`로 끝남

### 1.2 `testing.T`란?

```go
type T struct {
    // Go 테스트 런타임이 제공하는 구조체
    // 테스트 실패, 로그, Cleanup 등을 관리
}
```

**주요 메서드:**
- `t.Error()`, `t.Fatal()`: 테스트 실패 표시
- `t.Log()`: 로그 출력
- `t.Cleanup()`: 테스트 종료 시 실행할 함수 등록
- `t.Helper()`: 헬퍼 함수임을 표시 (에러 위치 정확히 표시)
- `t.Run()`: 서브테스트 실행
- `t.Parallel()`: 병렬 실행 허용

---

## 2. 테스트 파일 인식과 실행

### 2.1 `go test` 명령어 실행 시

```bash
go test -v ./internal/auth/... -run TestSignup
```

**Go 테스트 러너가 하는 일:**

```
1. ./internal/auth/ 디렉토리를 스캔
   └─ *_test.go 파일들을 찾음
      └─ handler_test.go ✓

2. 파일을 컴파일하고 로드
   └─ package auth_test로 로드됨

3. Test로 시작하는 함수들을 찾음
   ├─ TestSignup_Success ✓
   ├─ TestSignup_DuplicateEmail ✓
   ├─ TestSignup_ValidationError_MissingRequiredFields ✓
   └─ ... (나머지 테스트들)

4. -run 플래그로 필터링
   └─ "TestSignup"과 매치되는 것만 실행

5. 각 테스트 함수를 순차적으로 실행
   └─ 각각 새로운 testing.T 인스턴스 생성해서 전달
```

### 2.2 패키지 구조

```go
// handler_test.go
package auth_test  // ⚠️ auth가 아니라 auth_test!

import (
    "github.com/.../internal/auth"  // auth 패키지를 import해서 사용
)
```

**왜 `auth_test` 패키지를 사용할까?**
- **블랙박스 테스트**: 외부에서 사용하는 것처럼 테스트
- 내부 구현(unexported)에 의존하지 않음
- 진짜 사용자처럼 public API만 테스트

---

## 3. 우리 프로젝트의 테스트 구조

### 3.1 파일 구조

```
internal/
├── auth/
│   ├── handler.go              # 실제 코드
│   ├── handler_test.go         # 테스트 코드
│   └── service.go
└── shared/
    └── testutil/               # 공통 테스트 유틸리티
        ├── database.go         # DB 헬퍼 (SetupTestDB, CleanupTestDB, TruncateTable)
        ├── router.go           # HTTP 테스트 헬퍼 (SetupTestRouter, SetupAuthenticatedRouter, ExecuteRequest)
        ├── token.go            # Mock 토큰 매니저
        ├── config.go           # 테스트용 설정 (NewTestConfig)
        ├── member.go           # Member 테스트 헬퍼 (CreateTestMember, NewMemberRepository 등)
        └── room.go             # Room 테스트 헬퍼 (CreateTestRoom, CreateTestRooms, AddMembersToRoom)
```

### 3.2 테스트 함수 구조

```go
func TestSignup_Success(t *testing.T) {
    // 1️⃣ Given: 테스트 환경 설정
    authHandler, _ := setupTestEnvironment(t)
    router := testutil.SetupTestRouter()
    router.POST("/api/v1/auth/signup", authHandler.Signup)

    // 2️⃣ Given: 테스트 데이터 준비
    request := testutil.TestRequest{
        Method: http.MethodPost,
        URL:    "/api/v1/auth/signup",
        Body: map[string]string{
            "name":        "Test User",
            "email":       "test@example.com",
            "phoneNumber": "010-1234-5678",
            "password":    "password123",
        },
    }

    // 3️⃣ When: 실제 테스트 실행
    recorder := testutil.ExecuteRequest(t, router, request)

    // 4️⃣ Then: 결과 검증
    assert.Equal(t, http.StatusCreated, recorder.Code)
}
```

---

## 4. 단계별 실행 흐름

### 4.1 `setupTestEnvironment(t)` - 테스트 환경 초기화

```go
func setupTestEnvironment(t *testing.T) (*auth.AuthHandler, *testutil.MockTokenManager) {
    t.Helper()  // ← 이 함수에서 에러나면, 호출한 곳(TestSignup_Success)의 라인 번호 표시

    // 📦 1. In-memory SQLite 데이터베이스 생성
    db := testutil.SetupTestDB(t)

    // 🧹 2. 테스트 종료 시 자동으로 DB 정리하도록 등록
    t.Cleanup(func() {
        testutil.CleanupTestDB(t, db)
    })

    // 🏗️ 3. 의존성 생성
    memberRepo := testutil.NewMemberRepository(db)  // ← db를 파라미터로 전달
    mockTokenManager := testutil.NewMockTokenManager()
    authService := auth.NewAuthService(db, memberRepo, mockTokenManager)
    authHandler := auth.NewAuthHandler(authService)

    return authHandler, mockTokenManager
}
```

**실행 흐름:**

```
TestSignup_Success 시작
  │
  ├─► setupTestEnvironment(t) 호출
  │     │
  │     ├─► testutil.SetupTestDB(t)
  │     │     └─► gorm.Open(sqlite.Open(":memory:"))
  │     │           └─► db.AutoMigrate(&model.MemberRoom{}, &model.Room{}, &model.Member{})
  │     │                 └─► CREATE TABLE member_room (...), CREATE TABLE room (...), CREATE TABLE member (...) 실행
  │     │
  │     ├─► t.Cleanup() 등록
  │     │     └─► [나중에 실행될 함수 예약]
  │     │
  │     ├─► testutil.NewMemberRepository(db) 생성
  │     ├─► testutil.NewMockTokenManager() 생성
  │     ├─► auth.NewAuthService(db, memberRepo, mockTokenManager) 생성
  │     └─► auth.NewAuthHandler(service) 생성
  │
  └─► authHandler 반환
```

### 4.2 `testutil.SetupTestRouter()` - Gin 라우터 생성

```go
func SetupTestRouter() *gin.Engine {
    gin.SetMode(gin.TestMode)  // 로그 최소화

    // 커스텀 validator 등록 (phone 등)
    _ = validator.RegisterAll()

    return gin.New()  // 새로운 Gin 엔진 생성
}
```

**실행 흐름:**

```
testutil.SetupTestRouter() 호출
  │
  ├─► gin.SetMode(gin.TestMode)
  │     └─► Gin이 테스트 모드로 동작 (로그 최소화)
  │
  ├─► validator.RegisterAll()
  │     └─► v.RegisterValidation("phone", ValidatePhone)
  │           └─► Gin의 validator에 커스텀 검증 함수 등록
  │
  └─► gin.New() 반환
        └─► 빈 라우터 (미들웨어 없음)
```

### 4.2.1 `testutil.SetupAuthenticatedRouter()` - 인증된 라우터 생성

```go
// SetupAuthenticatedRouter creates a test router with memberID set in context
// This simulates the RESULT of JWT middleware (memberID in context) without actual token validation.
func SetupAuthenticatedRouter(memberID int64) *gin.Engine {
    router := SetupTestRouter()

    // Simulate the result of JWT middleware: memberID in context as string
    memberIDStr := strconv.FormatInt(int64(memberID), 10)
    router.Use(func(c *gin.Context) {
        c.Set(sharedHttp.MemberIDKey, memberIDStr)
        c.Next()
    })

    return router
}
```

**사용 예시:**

```go
// 인증이 필요한 엔드포인트 테스트
func TestGetMyProfile(t *testing.T) {
    db := testutil.SetupTestDB(t)
    member := testutil.CreateTestMember(t, db)

    // memberID가 context에 설정된 라우터
    router := testutil.SetupAuthenticatedRouter(member.ID)
    router.GET("/api/v1/me", handler.GetMyProfile)

    // JWT 토큰 없이도 인증된 요청처럼 동작
    recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
        Method: http.MethodGet,
        URL:    "/api/v1/me",
    })

    assert.Equal(t, http.StatusOK, recorder.Code)
}
```

**실행 흐름:**

```
testutil.SetupAuthenticatedRouter(memberID) 호출
  │
  ├─► SetupTestRouter() 호출
  │     └─► 기본 라우터 생성
  │
  ├─► router.Use(middleware) 등록
  │     └─► 모든 요청에 대해 context에 memberID 설정
  │           c.Set(sharedHttp.MemberIDKey, "1")
  │
  └─► 인증 미들웨어가 적용된 라우터 반환

실제 요청 시:
  Request → Mock Middleware (memberID 설정) → Handler
  (JWT 검증 없이 바로 memberID가 context에 주입됨)
```

### 4.3 라우트 등록

```go
router.POST("/api/v1/auth/signup", authHandler.Signup)
```

**실행 흐름:**

```
router.POST() 호출
  │
  └─► Gin의 라우팅 테이블에 등록
        └─► [POST /api/v1/auth/signup] → authHandler.Signup 함수
```

### 4.4 `testutil.ExecuteRequest()` - HTTP 요청 시뮬레이션

```go
// TestRequest 구조체 정의
type TestRequest struct {
    Method      string
    URL         string
    Body        interface{}
    AccessToken string // Optional JWT token for authenticated requests
}

func ExecuteRequest(t *testing.T, router *gin.Engine, req TestRequest) *httptest.ResponseRecorder {
    t.Helper()

    // 1️⃣ 요청 Body를 JSON으로 변환
    var bodyReader io.Reader
    if req.Body != nil {
        bodyBytes, _ := json.Marshal(req.Body)
        bodyReader = bytes.NewReader(bodyBytes)
    }

    // 2️⃣ HTTP 요청 객체 생성
    httpReq := httptest.NewRequest(req.Method, req.URL, bodyReader)
    httpReq.Header.Set("Content-Type", "application/json")

    // 3️⃣ JWT 토큰이 있으면 헤더에 추가
    if req.AccessToken != "" {
        httpReq.Header.Set("Authorization", "Bearer "+req.AccessToken)
    }

    // 4️⃣ 응답을 기록할 Recorder 생성
    recorder := httptest.NewRecorder()

    // 5️⃣ 실제 HTTP 요청 실행 (라우터에게 전달)
    router.ServeHTTP(recorder, httpReq)

    return recorder
}
```

**실행 흐름:**

```
testutil.ExecuteRequest() 호출
  │
  ├─► 1. Body를 JSON으로 변환
  │     {"name":"Test User", "email":"test@example.com", ...}
  │
  ├─► 2. httptest.NewRequest() 생성
  │     POST /api/v1/auth/signup
  │     Content-Type: application/json
  │     Body: {"name":"Test User", ...}
  │
  ├─► 3. httptest.NewRecorder() 생성
  │     [빈 응답 기록기]
  │
  └─► 4. router.ServeHTTP(recorder, httpReq)
        │
        ├─► Gin이 라우팅 테이블 검색
        │     └─► POST /api/v1/auth/signup 찾음
        │
        ├─► authHandler.Signup(c) 호출
        │     │
        │     ├─► handler.BindJSON(c, &request)
        │     │     └─► JSON 파싱 & 검증
        │     │
        │     ├─► authService.Signup(ctx, &request)
        │     │     │
        │     │     ├─► 트랜잭션 시작
        │     │     ├─► memberRepository.IsExist()
        │     │     │     └─► SELECT COUNT(*) FROM member WHERE email = ?
        │     │     ├─► bcrypt.GenerateFromPassword()
        │     │     ├─► memberRepository.Create()
        │     │     │     └─► INSERT INTO member (...) VALUES (...)
        │     │     └─► 트랜잭션 커밋
        │     │
        │     └─► c.JSON(201, gin.H{})
        │           └─► recorder에 응답 기록
        │                 ├─ Status: 201
        │                 ├─ Header: Content-Type: application/json
        │                 └─ Body: {}
        │
        └─► recorder 반환
```

### 4.5 결과 검증

```go
assert.Equal(t, http.StatusCreated, recorder.Code)

var response map[string]interface{}
testutil.ParseResponse(t, recorder, &response)
assert.NotNil(t, response)
```

**실행 흐름:**

```
assert.Equal(t, 201, recorder.Code)
  │
  ├─► recorder.Code를 읽음 → 201
  ├─► 201 == 201 비교
  └─► ✅ Pass

testutil.ParseResponse(t, recorder, &response)
  │
  ├─► recorder.Body.Bytes() 읽음 → "{}"
  ├─► json.Unmarshal([]byte("{}"), &response)
  └─► response = map[string]interface{}{}

assert.NotNil(t, response)
  │
  ├─► response가 nil인가? → 아니오
  └─► ✅ Pass
```

---

## 5. Mock 객체의 동작 원리

### 5.1 왜 Mock이 필요한가?

```go
// 실제 AuthService
type AuthService struct {
    db               *gorm.DB
    memberRepository *member.MemberRepository
    tokenManager     token.Manager  // ← Interface!
}
```

**문제:**
- 실제 `JWTManager`는 JWT 토큰을 생성함
- 테스트에서는 실제 토큰이 필요 없음
- 생성 로직보다는 "토큰 생성이 호출되었는가?"만 확인하고 싶음

**해결:**
- Interface를 사용하면 구현체를 바꿀 수 있음
- 프로덕션: `JWTManager` (진짜 구현)
- 테스트: `MockTokenManager` (가짜 구현)

### 5.2 Mock 객체 구조

```go
type MockTokenManager struct {
    // 함수 필드: 외부에서 동작을 주입할 수 있음
    GenerateAccessTokenFunc  func(memberID, email string) (string, error)
    GenerateRefreshTokenFunc func(memberID, email string) (string, error)
    ValidateTokenFunc        func(tokenString string) (*token.Claims, error)
}

func (m *MockTokenManager) GenerateAccessToken(memberID, email string) (string, error) {
    // 외부에서 함수를 주입했으면 그걸 실행
    if m.GenerateAccessTokenFunc != nil {
        return m.GenerateAccessTokenFunc(memberID, email)
    }
    // 아니면 기본 동작: 더미 토큰 반환
    return "mock-access-token", nil
}
```

### 5.3 Mock 사용 예시

```go
// 기본 사용 (우리 테스트)
mockTokenManager := testutil.NewMockTokenManager()
// GenerateAccessToken() 호출 시 → "mock-access-token" 반환

// 커스텀 동작 주입 (필요하면)
mockTokenManager.GenerateAccessTokenFunc = func(memberID, email string) (string, error) {
    return fmt.Sprintf("custom-token-for-%s", email), nil
}
// GenerateAccessToken() 호출 시 → "custom-token-for-test@example.com" 반환
```

### 5.4 Interface 구현 보장

```go
// 컴파일 타임에 체크
var _ token.Manager = (*MockTokenManager)(nil)
```

**동작:**

```go
// Go 컴파일러가 체크
type Manager interface {
    GenerateAccessToken(memberID string, email string) (string, error)
    GenerateRefreshToken(memerID string, email string) (string, error)
    ValidateToken(tokenString string) (*Claims, error)
}

// MockTokenManager가 위 3개 메서드를 모두 구현했는가?
// ✅ Yes → 컴파일 성공
// ❌ No → 컴파일 에러: MockTokenManager does not implement Manager
```

---

## 6. 실제 HTTP 요청 시뮬레이션

### 6.1 httptest 패키지

Go 표준 라이브러리 `net/http/httptest`는 **실제 네트워크 없이** HTTP 서버를 테스트할 수 있게 해줍니다.

```go
// 🔴 실제 서버 (프로덕션)
// 1. 서버 시작: http.ListenAndServe(":8080", router)
// 2. 클라이언트가 네트워크로 요청
// 3. 서버가 응답

// 🟢 테스트 서버 (테스트)
// 1. 메모리에서 요청/응답 객체 생성
// 2. router.ServeHTTP(recorder, request)
// 3. 메모리에서 응답 읽기
```

### 6.2 실제 요청 vs 테스트 요청

```
┌─────────────────────────────────────────────────────────────┐
│ 실제 프로덕션 환경                                              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  클라이언트 (브라우저/앱)                                        │
│       │                                                     │
│       │ HTTP Request                                        │
│       │ POST /api/v1/auth/signup                           │
│       │ Content-Type: application/json                      │
│       │ Body: {"name":"..."}                               │
│       │                                                     │
│       ▼                                                     │
│   네트워크 (TCP/IP)                                          │
│       │                                                     │
│       ▼                                                     │
│  서버 (Go 애플리케이션)                                         │
│       │                                                     │
│       ├─► Gin Router                                        │
│       │     └─► AuthHandler.Signup()                       │
│       │           └─► AuthService.Signup()                 │
│       │                 └─► Database (Oracle)              │
│       │                                                     │
│       └─► HTTP Response                                     │
│              Status: 201 Created                            │
│              Body: {}                                       │
│                                                             │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ 테스트 환경                                                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  테스트 함수 (TestSignup_Success)                             │
│       │                                                     │
│       ├─► httptest.NewRequest()                            │
│       │     └─► 메모리에 가짜 HTTP Request 객체 생성           │
│       │                                                     │
│       ├─► httptest.NewRecorder()                           │
│       │     └─► 메모리에 응답 기록할 객체 생성                  │
│       │                                                     │
│       ├─► router.ServeHTTP(recorder, request)              │
│       │     │                                               │
│       │     ├─► Gin Router (진짜)                           │
│       │     │     └─► AuthHandler.Signup() (진짜)          │
│       │     │           └─► AuthService.Signup() (진짜)    │
│       │     │                 └─► Database (SQLite 메모리)  │
│       │     │                       MockTokenManager (가짜) │
│       │     │                                               │
│       │     └─► recorder에 응답 기록                         │
│       │                                                     │
│       └─► assert.Equal(201, recorder.Code)                 │
│              └─► 검증                                        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 6.3 ResponseRecorder 내부

```go
type ResponseRecorder struct {
    Code      int           // HTTP 상태 코드 (200, 201, 400, ...)
    HeaderMap http.Header   // 응답 헤더들
    Body      *bytes.Buffer // 응답 Body
    Flushed   bool
}

// 사용 예
recorder := httptest.NewRecorder()
router.ServeHTTP(recorder, request)

// 이제 recorder에는 다음이 저장됨:
// recorder.Code = 201
// recorder.HeaderMap["Content-Type"] = ["application/json"]
// recorder.Body.String() = "{}"
```

---

## 7. 전체 실행 흐름 다이어그램

```
┌─────────────────────────────────────────────────────────────────────┐
│ go test -v ./internal/auth/... -run TestSignup_Success              │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│ Go Test Runner                                                       │
│ 1. internal/auth/handler_test.go 파일 찾음                           │
│ 2. TestSignup_Success 함수 발견                                      │
│ 3. testing.T 객체 생성                                               │
│ 4. TestSignup_Success(t) 호출                                        │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│ TestSignup_Success(t *testing.T)                                    │
│                                                                     │
│ 1️⃣ authHandler, _ := setupTestEnvironment(t)                       │
│     │                                                               │
│     ├─► SetupTestDB(t)                                             │
│     │     └─► SQLite :memory: 생성                                 │
│     │           └─► AutoMigrate(&Member{})                         │
│     │                 └─► CREATE TABLE member (...);               │
│     │                                                               │
│     ├─► t.Cleanup() 등록                                           │
│     │                                                               │
│     ├─► NewMemberRepository()                                      │
│     ├─► NewMockTokenManager()                                      │
│     ├─► NewAuthService(db, repo, mockToken)                        │
│     └─► NewAuthHandler(service)                                    │
│                                                                     │
│ 2️⃣ router := testutil.SetupTestRouter()                            │
│     └─► gin.New() + validator 등록                                 │
│                                                                     │
│ 3️⃣ router.POST("/api/v1/auth/signup", authHandler.Signup)          │
│     └─► 라우팅 테이블에 등록                                          │
│                                                                     │
│ 4️⃣ request := testutil.TestRequest{...}                            │
│     └─► 요청 데이터 준비                                              │
│                                                                     │
│ 5️⃣ recorder := testutil.ExecuteRequest(t, router, request)         │
│     │                                                               │
│     ├─► json.Marshal(request.Body)                                │
│     │     └─► {"name":"Test User","email":"test@example.com",...} │
│     │                                                               │
│     ├─► httptest.NewRequest(POST, /api/v1/auth/signup, body)      │
│     │                                                               │
│     ├─► httptest.NewRecorder()                                     │
│     │                                                               │
│     └─► router.ServeHTTP(recorder, request)                        │
│           │                                                         │
│           ├─► [Gin 라우터 동작]                                     │
│           │     └─► POST /api/v1/auth/signup 매칭                  │
│           │                                                         │
│           ├─► authHandler.Signup(c)                                │
│           │     │                                                   │
│           │     ├─► handler.BindJSON(c, &request)                  │
│           │     │     ├─► JSON 파싱                                │
│           │     │     └─► Validator 검증 (required, email, ...)    │
│           │     │                                                   │
│           │     ├─► authService.Signup(ctx, &request)              │
│           │     │     │                                             │
│           │     │     ├─► database.WithTransaction(...)            │
│           │     │     │     │                                       │
│           │     │     │     ├─► memberRepo.IsExist(email)          │
│           │     │     │     │     └─► SELECT COUNT(*) FROM member  │
│           │     │     │     │           WHERE email = ?            │
│           │     │     │     │           → 0 (없음)                 │
│           │     │     │     │                                       │
│           │     │     │     ├─► bcrypt.GenerateFromPassword()      │
│           │     │     │     │     └─► "$2a$10$..."                │
│           │     │     │     │                                       │
│           │     │     │     ├─► memberRepo.Create(member)          │
│           │     │     │     │     └─► INSERT INTO member           │
│           │     │     │     │           (name, email, ...)         │
│           │     │     │     │           VALUES (?, ?, ...)         │
│           │     │     │     │           → ID=1 생성됨              │
│           │     │     │     │                                       │
│           │     │     │     └─► COMMIT                             │
│           │     │     │                                             │
│           │     │     └─► return nil (성공)                        │
│           │     │                                                   │
│           │     └─► c.JSON(201, gin.H{})                           │
│           │           └─► recorder에 기록:                          │
│           │                 Code: 201                              │
│           │                 Body: "{}"                             │
│           │                                                         │
│           └─► return recorder                                      │
│                                                                     │
│ 6️⃣ assert.Equal(t, 201, recorder.Code)                             │
│     └─► 201 == 201 ✅                                               │
│                                                                     │
│ 7️⃣ [테스트 함수 종료]                                                │
│     │                                                               │
│     └─► t.Cleanup() 등록된 함수들 실행                               │
│           └─► CleanupTestDB(db)                                    │
│                 └─► db.Close()                                     │
│                       └─► SQLite 메모리 해제                        │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│ Go Test Runner                                                       │
│ ✅ PASS: TestSignup_Success (0.08s)                                  │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 8. 핵심 포인트 요약

### 8.1 테스트는 어떻게 실행되나?

1. **`go test` 명령어** → Go Test Runner 시작
2. **`*_test.go` 파일 검색** → 테스트 파일 찾기
3. **`Test*` 함수 검색** → 테스트 함수 찾기
4. **각 함수에 `testing.T` 전달** → 테스트 실행
5. **assert 실패 시** → 테스트 실패 마킹
6. **`t.Cleanup()` 실행** → 리소스 정리
7. **결과 리포트 출력** → PASS/FAIL

### 8.2 의존성 주입 (Dependency Injection)

```go
// AuthService는 Interface를 받음 (구현체 X)
type AuthService struct {
    tokenManager token.Manager  // Interface!
}

// 프로덕션
realManager := token.NewJWTManager(cfg)
service := auth.NewAuthService(db, repo, realManager)

// 테스트
mockManager := testutil.NewMockTokenManager()
service := auth.NewAuthService(db, repo, mockManager)
```

**장점:**
- 프로덕션 코드 수정 없이 테스트 가능
- Mock으로 외부 의존성 제거
- 테스트 속도 향상

### 8.3 In-Memory Database

```go
// 프로덕션: Oracle Cloud
dsn := "oracle://user:pass@host:1521/service"

// 테스트: SQLite :memory:
dsn := ":memory:"
```

**장점:**
- 실제 DB 없이 테스트 가능
- 빠름 (메모리에서 동작)
- 테스트 간 격리 (매번 새로 생성)

### 8.4 httptest로 실제 네트워크 제거

```go
// 프로덕션: 실제 HTTP 서버
http.ListenAndServe(":8080", router)

// 테스트: 메모리에서 시뮬레이션
recorder := httptest.NewRecorder()
router.ServeHTTP(recorder, request)
```

**장점:**
- 네트워크 없이 HTTP 테스트
- 빠름
- 포트 충돌 없음

---

## 9. 자주 묻는 질문 (FAQ)

### Q1. `t.Helper()`는 왜 쓰나요?

```go
func setupTestEnvironment(t *testing.T) {
    t.Helper()  // ← 이게 없으면?
    // ...
}
```

**Without `t.Helper()`:**
```
--- FAIL: TestSignup_Success (0.08s)
    handler_test.go:15: setupTestEnvironment: DB connection failed
```

**With `t.Helper()`:**
```
--- FAIL: TestSignup_Success (0.08s)
    handler_test.go:35: DB connection failed  ← 실제 호출한 곳!
```

### Q2. `t.Cleanup()`은 언제 실행되나요?

```go
func TestExample(t *testing.T) {
    fmt.Println("1. 테스트 시작")

    t.Cleanup(func() {
        fmt.Println("4. Cleanup 1")
    })

    fmt.Println("2. 테스트 로직")

    t.Cleanup(func() {
        fmt.Println("3. Cleanup 2")
    })

    fmt.Println("테스트 종료")
}

// 출력 순서:
// 1. 테스트 시작
// 2. 테스트 로직
// 테스트 종료
// 3. Cleanup 2  ← LIFO (Last In First Out)
// 4. Cleanup 1
```

### Q3. Mock은 언제 사용하나요?

**Mock을 쓰는 경우:**
- ✅ 외부 API 호출 (네트워크)
- ✅ 파일 시스템 접근
- ✅ 시간이 오래 걸리는 작업
- ✅ 에러 상황 시뮬레이션

**Mock을 안 쓰는 경우:**
- ❌ 단순 로직 (계산, 문자열 처리 등)
- ❌ 데이터베이스 (In-memory DB 사용)

### Q4. Table-Driven Test는 언제 쓰나요?

**쓰는 경우:**
- ✅ 같은 로직, 다른 입력값
- ✅ Validation 테스트
- ✅ 경계값 테스트

**안 쓰는 경우:**
- ❌ 완전히 다른 시나리오
- ❌ 복잡한 조건 분기

---

## 10. 실습 예제

### 예제 1: 간단한 테스트 작성

```go
// 테스트할 함수
func Add(a, b int) int {
    return a + b
}

// 테스트 코드
func TestAdd(t *testing.T) {
    result := Add(2, 3)

    if result != 5 {
        t.Errorf("Add(2, 3) = %d; want 5", result)
    }
}
```

### 예제 2: Table-Driven Test

```go
func TestAdd_TableDriven(t *testing.T) {
    tests := []struct {
        name string
        a    int
        b    int
        want int
    }{
        {"positive", 2, 3, 5},
        {"negative", -1, -2, -3},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("Add(%d, %d) = %d; want %d",
                    tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

### 예제 3: Cleanup 사용

```go
func TestWithCleanup(t *testing.T) {
    // 리소스 생성
    file, err := os.Create("test.txt")
    if err != nil {
        t.Fatal(err)
    }

    // 테스트 종료 시 자동 삭제
    t.Cleanup(func() {
        os.Remove("test.txt")
    })

    // 테스트 로직
    file.WriteString("test data")
}
```

---

## 11. testutil 헬퍼 함수 상세 가이드

### 11.1 Member 테스트 헬퍼 (testutil/member.go)

#### CreateTestMember - 기본 테스트 회원 생성

```go
func CreateTestMember(t *testing.T, db *gorm.DB) *model.Member {
    t.Helper()
    return CreateTestMemberWithIndex(t, db, 0)
}
```

**사용 예시:**

```go
func TestSomething(t *testing.T) {
    db := testutil.SetupTestDB(t)
    member := testutil.CreateTestMember(t, db)
    // member.Email = "test@example.com"
    // member.Name = "Test_User_"
    // member.PhoneNumber = "010-1234-5678"
    // Password: "password123" (bcrypt 해싱됨)
}
```

#### CreateTestMemberWithIndex - 인덱스로 여러 회원 생성

```go
func CreateTestMemberWithIndex(t *testing.T, db *gorm.DB, index int) *model.Member
```

**사용 예시:**

```go
func TestMultipleMembers(t *testing.T) {
    db := testutil.SetupTestDB(t)

    member1 := testutil.CreateTestMemberWithIndex(t, db, 0)
    // Email: "test@example.com"

    member2 := testutil.CreateTestMemberWithIndex(t, db, 1)
    // Email: "test1@example.com"
    // Name: "Test_User_1"
    // PhoneNumber: "010-1234-5679"

    member3 := testutil.CreateTestMemberWithIndex(t, db, 2)
    // Email: "test2@example.com"
}
```

#### NewMemberRepository, NewMemberService, NewMemberUseCase

```go
// 의존성 생성 헬퍼들
memberRepo := testutil.NewMemberRepository(db)
memberService := testutil.NewMemberService(memberRepo)
memberUseCase := testutil.NewMemberUseCase(db, memberService)
```

### 11.2 Room 테스트 헬퍼 (testutil/room.go)

#### CreateTestRoom - 단일 방 생성

```go
func CreateTestRoom(t *testing.T, db *gorm.DB, memberID int64, name, description string) *model.Room
```

**사용 예시:**

```go
func TestRoomCreation(t *testing.T) {
    db := testutil.SetupTestDB(t)
    member := testutil.CreateTestMember(t, db)

    room := testutil.CreateTestRoom(t, db, member.ID, "My Room", "Test room description")
    // room이 생성되고, member가 OWNER로 자동 추가됨
}
```

#### CreateTestRooms - 페이지네이션 테스트용 여러 방 생성

```go
func CreateTestRooms(t *testing.T, db *gorm.DB, memberID int64, count int)
```

**특징:**
- 방들이 시간 순서대로 생성됨 (커서 기반 페이지네이션 테스트용)
- 각 방의 CreatedAt이 1초씩 차이남

**사용 예시:**

```go
func TestInfiniteScroll(t *testing.T) {
    db := testutil.SetupTestDB(t)
    member := testutil.CreateTestMember(t, db)

    // 15개의 방 생성 (페이지네이션 테스트용)
    testutil.CreateTestRooms(t, db, member.ID, 15)

    // 첫 페이지 조회 (10개)
    // 두 번째 페이지 조회 (5개)
}
```

#### AddMembersToRoom - 방에 여러 회원 추가

```go
func AddMembersToRoom(t *testing.T, db *gorm.DB, roomID int64, count int) []*model.Member
```

**사용 예시:**

```go
func TestRoomMembers(t *testing.T) {
    db := testutil.SetupTestDB(t)
    owner := testutil.CreateTestMember(t, db)
    room := testutil.CreateTestRoom(t, db, owner.ID, "Team Room", "Description")

    // 방에 5명의 멤버 추가 (MEMBER 역할)
    members := testutil.AddMembersToRoom(t, db, room.ID, 5)
    // 총 6명 (owner 1명 + members 5명)

    assert.Len(t, members, 5)
}
```

### 11.3 Database 헬퍼 (testutil/database.go)

#### TruncateTable - 테스트 간 데이터 격리

```go
func TruncateTable(t *testing.T, db *gorm.DB, tableName string)
```

**사용 예시:**

```go
func TestWithIsolation(t *testing.T) {
    db := testutil.SetupTestDB(t)

    // 첫 번째 테스트
    testutil.CreateTestMember(t, db)
    testutil.TruncateTable(t, db, "member")

    // 두 번째 테스트 (깨끗한 상태)
    var count int64
    db.Model(&model.Member{}).Count(&count)
    assert.Equal(t, int64(0), count)
}
```

### 11.4 Config 헬퍼 (testutil/config.go)

#### NewTestConfig - 테스트용 설정 생성

```go
func NewTestConfig() *config.Config
```

**사용 예시:**

```go
func TestWithConfig(t *testing.T) {
    cfg := testutil.NewTestConfig()
    // cfg.JWT.Secret = "test-jwt-secret-key-must-be-at-least-32-characters-long"
    // cfg.App.Env = "test"

    tokenManager := token.NewJWTManager(cfg)
}
```

---

## 12. 참고 자료

- [Go Testing Package 공식 문서](https://pkg.go.dev/testing)
- [Go Wiki: TableDrivenTests](https://go.dev/wiki/TableDrivenTests)
- [httptest Package 문서](https://pkg.go.dev/net/http/httptest)
- [testify/assert 문서](https://pkg.go.dev/github.com/stretchr/testify/assert)

---

이제 Go 테스트의 모든 것을 이해하셨나요? 🎉

추가 질문이나 이해 안 되는 부분이 있으면 언제든 물어보세요!
