# Go API Server 설치 및 실행 가이드

## 데이터베이스 선택

이 프로젝트는 **PostgreSQL** (권장) 또는 **Oracle**을 지원합니다.

### 옵션 1: PostgreSQL 사용 (권장) ✅

#### 1. PostgreSQL 설치

```bash
# macOS
brew install postgresql
brew services start postgresql

# Ubuntu/Debian
sudo apt-get install postgresql postgresql-contrib
sudo systemctl start postgresql

# 데이터베이스 생성
createdb praytogether
```

#### 2. 환경 설정

```bash
# .env 파일 생성
cp .env.example .env

# .env 파일 편집
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=praytogether
```

### 옵션 2: Oracle 사용

#### 1. Oracle Instant Client 설치

```bash
# macOS
brew tap InstantClientTap/instantclient
brew install instantclient-basic

# Linux
# Oracle 웹사이트에서 다운로드 필요
```

#### 2. 환경 설정

```bash
# .env 파일 편집
DB_TYPE=oracle
DB_HOST=localhost
DB_PORT=1521
DB_USER=SYSTEM
DB_PASSWORD=your_password
DB_NAME=FREEPDB1
```

## 프로젝트 실행

### 1. 의존성 설치

```bash
# Go 모듈 다운로드
go mod download
go mod tidy
```

### 2. 애플리케이션 실행

```bash
# 개발 서버 실행
make run

# 또는 직접 실행
go run cmd/api/main.go
```

### 3. API 테스트

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

## 트러블슈팅

### Oracle 연결 오류

Oracle을 사용할 때 다음과 같은 오류가 발생할 수 있습니다:

```
ORA-12541: TNS:no listener
```

해결 방법:
1. Oracle 리스너가 실행 중인지 확인
2. tnsnames.ora 파일 설정 확인
3. 방화벽 설정 확인

### PostgreSQL 연결 오류

```
FATAL: password authentication failed
```

해결 방법:
1. PostgreSQL 사용자 비밀번호 재설정
```bash
psql -U postgres
ALTER USER postgres PASSWORD 'new_password';
```

2. pg_hba.conf 파일에서 인증 방법 확인

### 포트 충돌

8080 포트가 이미 사용 중인 경우:
```bash
# .env 파일에서 포트 변경
SERVER_PORT=3000
```

## Docker로 실행

### PostgreSQL과 함께 실행

```bash
# docker-compose.yml 파일 생성
cat > docker-compose.yml << EOF
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: praytogether
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - postgres
    environment:
      DB_TYPE: postgres
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: postgres
      DB_NAME: praytogether
    volumes:
      - ./.env:/root/.env

volumes:
  postgres_data:
EOF

# 실행
docker-compose up
```

## 개발 도구

### Air (Hot Reload)

개발 중 자동 재시작을 원한다면:

```bash
# Air 설치
go install github.com/cosmtrek/air@latest

# .air.toml 파일 생성
air init

# Air로 실행
air
```

### 테스트 실행

```bash
# 전체 테스트
make test

# 커버리지 확인
make test-coverage

# 특정 패키지 테스트
go test ./internal/domain/member/...
```

## 다음 단계

프로젝트가 정상적으로 실행되면, 다음 기능들을 구현할 수 있습니다:

1. **Auth 도메인** - JWT 인증, 리프레시 토큰
2. **Room 도메인** - 방 생성, 참여, 탈퇴
3. **Prayer 도메인** - 기도 제목, 내용 관리
4. **Notification** - FCM 푸시 알림

각 도메인은 DDD 구조를 따라 구현됩니다.