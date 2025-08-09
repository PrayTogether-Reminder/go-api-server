# 빠른 시작 가이드

## 1. 환경 설정

```bash
# .env 파일 생성
cp .env.example .env

# .env 파일을 열어서 데이터베이스 정보 수정
# DB_PASSWORD=your_actual_password
```

## 2. 의존성 설치

```bash
# go mod 의존성 다운로드
make deps
# 또는
go mod download
go mod tidy
```

## 3. 애플리케이션 실행

```bash
# 개발 서버 실행
make run
# 또는
go run cmd/api/main.go
```

## 4. API 테스트

```bash
# Health Check
curl http://localhost:8080/health

# 회원가입
curl -X POST http://localhost:8080/api/v1/members/signup \
  -H "Content-Type: application/json" \
  -d '{
    "name": "홍길동",
    "email": "hong@example.com",
    "password": "password123"
  }'
```

## 다음 단계 - Auth 도메인 구현

이제 Member 도메인이 완성되었으니, 다음은 Auth(인증) 도메인을 구현해야 합니다:

1. **JWT 토큰 관리**
   - Access Token / Refresh Token 생성
   - 토큰 검증 미들웨어

2. **로그인/로그아웃**
   - 로그인 엔드포인트
   - 리프레시 토큰 관리

3. **OTP 기능**
   - 이메일 OTP 발송
   - OTP 검증

## 프로젝트 구조 설명

```
internal/
├── domain/           # 핵심 비즈니스 로직 (외부 의존성 없음)
│   └── member/
│       ├── entity.go      # 도메인 모델
│       ├── repository.go  # 리포지토리 인터페이스
│       └── service.go     # 도메인 서비스
│
├── application/      # 유스케이스 (비즈니스 프로세스)
│   └── member/
│       └── usecase.go     # 애플리케이션 서비스
│
├── infrastructure/   # 외부 시스템 연동
│   └── persistence/
│       ├── gorm/          # DB 연결
│       └── repository/    # 리포지토리 구현체
│
└── interfaces/       # 외부 인터페이스 (REST API)
    └── http/
        └── handler/       # HTTP 핸들러 (컨트롤러)
```

## 마이그레이션 진행 상황

✅ **완료된 부분:**
- DDD 프로젝트 구조 설정
- Member 도메인 구현
- 데이터베이스 연결 설정 (Oracle + GORM)
- 기본 HTTP 핸들러 구조

⏳ **다음 구현 예정:**
- [ ] Auth 도메인 (JWT, OTP)
- [ ] Room 도메인
- [ ] Prayer 도메인
- [ ] Invitation 도메인
- [ ] FCM Token 관리
- [ ] 알림 시스템
- [ ] 캐시 구현 (go-cache)
- [ ] 이메일 서비스
- [ ] Firebase 연동

## 트러블슈팅

### Oracle 연결 문제
Oracle Instant Client가 필요합니다:
```bash
# macOS
brew install oracle-instantclient

# Linux
# Oracle 웹사이트에서 다운로드 후 설치
```

### 포트 충돌
기본 포트는 8080입니다. 변경하려면 `.env` 파일의 `SERVER_PORT`를 수정하세요.
