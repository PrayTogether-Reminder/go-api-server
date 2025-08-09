# 통합 테스트 (Integration Tests)

이 디렉토리는 Java API 서버의 통합 테스트를 Go로 동일하게 구현한 것입니다.

## 구조

Java 프로젝트의 테스트 구조를 그대로 따릅니다:

```
test/
├── integration_test.go          # 기본 통합 테스트 스위트 (IntegrateTest.java와 동일)
├── test_utils.go                # 테스트 유틸리티 (TestUtils.java와 동일)
├── prayer_create_integration_test.go  # Prayer 생성 통합 테스트 (PrayerCreateIntegrateTest.java와 동일)
├── example_test.go              # 예제 테스트
└── README.md                    # 이 파일
```

## 주요 구성 요소

### IntegrationTestSuite
- Java의 `IntegrateTest` 클래스와 동일한 역할
- 모든 통합 테스트의 베이스 클래스
- 데이터베이스 설정 및 정리 담당
- HTTP 요청 헬퍼 메서드 제공

### TestUtils
- Java의 `TestUtils` 클래스와 동일한 역할
- 테스트용 데이터 생성 유틸리티
- JWT 인증 헤더 생성
- 데이터베이스 상태 검증

### Prayer Create Integration Test
- Java의 `PrayerCreateIntegrateTest` 클래스와 동일
- Prayer 생성 API의 모든 테스트 케이스 포함
- 정상 케이스 및 예외 케이스 테스트

## 사용 기술

- **testify/suite**: BDD 스타일 테스트 스위트
- **testify/assert**: 어설션 라이브러리
- **testify/require**: 필수 조건 검증
- **GORM**: ORM (인메모리 SQLite 사용)
- **Gin**: HTTP 웹 프레임워크

## 실행 방법

### 모든 테스트 실행
```bash
go test ./test/... -v
```

### 특정 테스트 스위트 실행
```bash
# Example 테스트만 실행
go test ./test -run TestExampleIntegration -v

# Prayer Create 테스트만 실행
go test ./test -run TestPrayerCreateIntegration -v
```

### 테스트 커버리지
```bash
go test ./test/... -cover
```

## Java 테스트와의 대응

| Java 클래스 | Go 파일 | 설명 |
|-------------|---------|------|
| `IntegrateTest` | `integration_test.go` | 기본 통합 테스트 스위트 |
| `TestUtils` | `test_utils.go` | 테스트 유틸리티 |
| `PrayerCreateIntegrateTest` | `prayer_create_integration_test.go` | Prayer 생성 테스트 |

## 주요 특징

1. **Java와 동일한 테스트 케이스**: 모든 테스트 케이스를 Java에서 그대로 이식
2. **동일한 데이터 정리 순서**: 외래키 제약조건을 고려한 정리 순서
3. **동일한 API URL**: `/api/v1/prayers` 등 Java와 동일한 엔드포인트
4. **동일한 검증 로직**: 상태 코드, 응답 메시지, 데이터베이스 상태 검증

## 테스트 데이터베이스

- **인메모리 SQLite** 사용으로 빠른 테스트 실행
- 각 테스트 전후로 자동 정리
- 실제 데이터베이스와 분리되어 안전

## 예제 사용법

```go
// 기본 테스트 스위트 상속
type MyTestSuite struct {
    IntegrationTestSuite
}

func (suite *MyTestSuite) TestMyAPI() {
    // Given
    member := suite.testUtils.CreateUniqueMember()
    headers := suite.testUtils.CreateAuthHeaderWithMember(member)
    
    requestBody := map[string]interface{}{
        "title": "테스트 제목",
    }
    
    // When
    w := suite.PostJSON("/api/v1/my-endpoint", requestBody, headers)
    
    // Then
    suite.AssertStatusCode(w, http.StatusCreated, "API should return 201")
}

func TestMyIntegration(t *testing.T) {
    suite.Run(t, new(MyTestSuite))
}
```