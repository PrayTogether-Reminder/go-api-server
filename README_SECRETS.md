# 환경변수 및 시크릿 관리 가이드

## 현재 환경변수 관리 방식

### 1. 환경변수 로딩
- **godotenv 라이브러리** 사용 (`.env` 파일에서 자동 로드)
- `config/config.go`에서 중앙 집중식 설정 관리
- 기본값(fallback) 지원

### 2. 필수 환경변수

#### 서버 설정
```bash
SERVER_PORT=8080          # 서버 포트 (기본: 8080)
GIN_MODE=debug           # Gin 모드: debug, release, test (기본: debug)
```

#### 데이터베이스 설정
```bash
DB_TYPE=postgres         # DB 타입: postgres, oracle (기본: postgres)
DB_HOST=localhost        # DB 호스트 (기본: localhost)
DB_PORT=5432            # DB 포트 (기본: 5432)
DB_USER=postgres        # DB 사용자 (기본: postgres)
DB_PASSWORD=            # DB 비밀번호 (필수)
DB_NAME=praytogether    # DB 이름 (기본: praytogether)
DB_CHARSET=UTF8         # 문자셋 (기본: UTF8)
```

#### JWT 설정
```bash
JWT_SECRET=your-secret-key   # JWT 시크릿 키 (프로덕션에서 반드시 변경!)
JWT_ACCESS_EXPIRY=1800       # 액세스 토큰 만료 시간(초) (기본: 30분)
JWT_REFRESH_EXPIRY=604800    # 리프레시 토큰 만료 시간(초) (기본: 7일)
```

#### Firebase 설정 (선택사항)
```bash
FIREBASE_CREDENTIALS=path/to/firebase-credentials.json  # Firebase 인증 파일 경로
```

#### 이메일 설정 (선택사항)
```bash
EMAIL_HOST=smtp.gmail.com    # SMTP 호스트 (기본: smtp.gmail.com)
EMAIL_PORT=587               # SMTP 포트 (기본: 587)
EMAIL_USERNAME=              # 이메일 계정
EMAIL_PASSWORD=              # 이메일 비밀번호/앱 비밀번호
EMAIL_FROM=noreply@site     # 발신자 이메일 (기본: noreply@praytogether.site)
```

## 환경별 설정 방법

### 개발 환경
1. `.env.example` 파일을 `.env`로 복사
2. 필요한 값들을 수정
```bash
cp .env.example .env
```

### 프로덕션 환경
1. **환경변수 직접 설정** (권장)
```bash
export DB_PASSWORD=your_secure_password
export JWT_SECRET=your_secure_jwt_secret
```

2. **Docker/Kubernetes 시크릿 사용**
```yaml
# Kubernetes Secret 예시
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
data:
  DB_PASSWORD: <base64-encoded-password>
  JWT_SECRET: <base64-encoded-secret>
```

3. **CI/CD 파이프라인 시크릿**
- GitHub Actions: Repository Secrets 사용
- GitLab CI: CI/CD Variables 사용

## 보안 체크리스트

### ✅ Git 무시 파일 (.gitignore)
```
.env
.env.local
.env.*.local
*-firebase-adminsdk-*.json
firebase-credentials.json
```

### ✅ 민감 정보 관리
- ❌ 절대 하드코딩하지 않기
- ❌ 실제 비밀번호를 커밋하지 않기
- ✅ `.env.example`에는 더미 값만 포함
- ✅ 프로덕션에서는 강력한 비밀번호 사용

### ✅ Firebase 인증 파일
- JSON 파일은 절대 커밋하지 않음
- 환경변수로 파일 경로만 지정
- 프로덕션: 안전한 저장소나 시크릿 매니저 사용

## 시크릿 로테이션

### JWT Secret 변경
1. 새로운 시크릿 생성
2. 환경변수 업데이트
3. 서버 재시작
4. 기존 토큰은 만료될 때까지 유효

### 데이터베이스 비밀번호 변경
1. DB에서 비밀번호 변경
2. 환경변수 업데이트
3. 서버 재시작

## 트러블슈팅

### 환경변수가 로드되지 않을 때
```go
// main.go에서 확인
if err := godotenv.Load(); err != nil {
    log.Println("No .env file found")
}
```

### 기본값 확인
`config/config.go`의 `getEnv()` 함수에서 기본값 설정 확인

### 디버깅
```bash
# 환경변수 확인
go run cmd/server/main.go
# 로그에서 "No .env file found" 메시지 확인
```

## 추가 보안 도구 추천

1. **HashiCorp Vault** - 시크릿 관리
2. **AWS Secrets Manager** - AWS 환경
3. **Google Secret Manager** - GCP 환경
4. **Azure Key Vault** - Azure 환경
5. **Doppler** - 환경변수 관리 SaaS

## 참고 사항

- 테스트 환경에서는 `.env.test` 파일 사용 가능
- CI/CD에서는 환경변수로 직접 주입
- 컨테이너 환경에서는 시크릿 마운트 활용