# 실무 통합 테스트 패턴

## 1. 현재 구현 방식 (Full Stack Integration Test)
```go
// 실제 핸들러와 미들웨어를 모두 등록
router := gin.New()
router.Use(authMiddleware)
v1 := router.Group("/api/v1")
memberHandler.RegisterRoutes(v1)

// HTTP 요청 테스트
w := httptest.NewRecorder()
req := httptest.NewRequest("GET", "/api/v1/members/me", nil)
router.ServeHTTP(w, req)
```

**장점:**
- 실제 환경과 가장 유사
- 라우팅, 미들웨어, 직렬화 모두 테스트
- E2E 테스트에 가까움

**단점:**
- 설정이 복잡
- 테스트 속도가 느림

## 2. Mock을 활용한 통합 테스트
```go
// 외부 의존성은 Mock으로 대체
type MockMemberService struct {
    mock.Mock
}

func (m *MockMemberService) GetMember(ctx context.Context, id uint64) (*Member, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*Member), args.Error(1)
}

// 테스트에서 Mock 사용
mockService := new(MockMemberService)
mockService.On("GetMember", mock.Anything, uint64(1)).Return(testMember, nil)
handler := NewHandler(mockService)
```

**장점:**
- 외부 의존성 제어 가능
- 테스트 속도 빠름
- 에러 케이스 테스트 용이

**단점:**
- Mock 관리 복잡
- 실제 통합 검증 부족

## 3. TestContainers를 활용한 통합 테스트
```go
// Docker 컨테이너로 실제 DB 사용
func SetupTestDB(t *testing.T) *sql.DB {
    ctx := context.Background()
    req := testcontainers.ContainerRequest{
        Image:        "postgres:14",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_PASSWORD": "test",
            "POSTGRES_DB":       "testdb",
        },
        WaitingFor: wait.ForListeningPort("5432/tcp"),
    }
    
    postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    require.NoError(t, err)
    
    // Connect to container DB
    // ...
}
```

**장점:**
- 실제 데이터베이스 사용
- 격리된 테스트 환경
- CI/CD에서도 동일하게 동작

**단점:**
- Docker 필요
- 초기 설정 시간

## 4. 계층별 분리 테스트
```go
// Repository 레이어 테스트
func TestMemberRepository(t *testing.T) {
    db := setupInMemoryDB(t)
    repo := NewMemberRepository(db)
    
    // Repository 메서드만 테스트
    member, err := repo.Create(ctx, &Member{...})
    assert.NoError(t, err)
}

// Service 레이어 테스트  
func TestMemberService(t *testing.T) {
    mockRepo := new(MockRepository)
    service := NewMemberService(mockRepo)
    
    // Service 로직만 테스트
    result, err := service.CreateMember(ctx, ...)
    assert.NoError(t, err)
}

// Handler 레이어 테스트
func TestMemberHandler(t *testing.T) {
    mockService := new(MockService)
    handler := NewMemberHandler(mockService)
    
    // HTTP 핸들링만 테스트
    req := httptest.NewRequest("POST", "/members", body)
    w := httptest.NewRecorder()
    handler.CreateMember(w, req)
}
```

**장점:**
- 각 계층 독립적 테스트
- 빠른 실행 속도
- 명확한 책임 분리

**단점:**
- 통합 이슈 놓칠 수 있음
- 더 많은 테스트 코드

## 5. 실무 권장 조합

### 테스트 피라미드
```
         /\
        /E2E\       <- 적게 (중요 시나리오만)
       /------\
      /Integra-\    <- 적당히 (API 엔드포인트)
     /tion Tests\
    /------------\
   / Unit Tests   \ <- 많이 (비즈니스 로직)
  /________________\
```

### 권장 구조
```
test/
├── unit/           # 단위 테스트
│   ├── domain/
│   └── service/
├── integration/    # 통합 테스트 (현재 구현 방식)
│   ├── api/       
│   └── repository/
└── e2e/           # E2E 테스트
    └── scenarios/
```

## 실무 체크리스트

✅ **현재 구현이 적절한 경우:**
- API 계약 테스트가 중요한 경우
- 팀에서 통합 테스트를 우선시하는 경우
- 마이크로서비스 환경

✅ **개선이 필요한 경우:**
- 테스트가 너무 느린 경우 → Mock 추가
- 외부 서비스 의존성이 많은 경우 → 계층 분리
- CI/CD 환경 구축 → TestContainers 고려

## 결론

현재 구현한 방식은 **실무에서도 널리 사용되는 유효한 패턴**입니다. 
특히 API 서버의 경우 이런 통합 테스트가 매우 중요합니다.

다만 프로젝트 규모가 커지면:
1. 단위 테스트 추가로 테스트 피라미드 균형 맞추기
2. 느린 테스트는 Mock으로 대체
3. CI/CD 파이프라인에 통합
4. 테스트 병렬 실행으로 속도 개선

이런 점진적 개선을 고려하시면 좋습니다.