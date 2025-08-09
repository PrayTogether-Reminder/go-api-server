# Go API Server - DDD 프로젝트 구조

## 프로젝트 아키텍처
```
go-api-server/
├── cmd/
│   └── api/
│       └── main.go                 # 애플리케이션 진입점
├── config/
│   └── config.go                   # 설정 관리
├── internal/
│   ├── domain/                     # 도메인 계층 (엔티티, 비즈니스 로직)
│   │   ├── base/
│   │   │   ├── entity.go          # BaseEntity (공통 필드)
│   │   │   └── response.go        # 공통 응답 구조
│   │   ├── auth/
│   │   │   ├── entity.go          # JWT Claims, Principal
│   │   │   ├── repository.go      # 인터페이스
│   │   │   └── service.go         # 도메인 서비스
│   │   ├── member/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── room/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── member_room/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── prayer/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── invitation/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── fcm_token/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   └── notification/
│   │       ├── entity.go
│   │       ├── repository.go
│   │       └── service.go
│   ├── application/                # 애플리케이션 서비스 (유스케이스)
│   │   ├── auth/
│   │   │   └── usecase.go
│   │   ├── member/
│   │   │   └── usecase.go
│   │   ├── room/
│   │   │   └── usecase.go
│   │   ├── prayer/
│   │   │   └── usecase.go
│   │   ├── invitation/
│   │   │   └── usecase.go
│   │   └── notification/
│   │       └── usecase.go
│   ├── infrastructure/             # 인프라 계층
│   │   ├── persistence/
│   │   │   ├── gorm/
│   │   │   │   └── database.go   # DB 연결
│   │   │   └── repository/       # Repository 구현체
│   │   │       ├── member_repository.go
│   │   │       ├── room_repository.go
│   │   │       ├── member_room_repository.go
│   │   │       ├── prayer_repository.go
│   │   │       ├── invitation_repository.go
│   │   │       ├── fcm_token_repository.go
│   │   │       └── notification_repository.go
│   │   ├── cache/
│   │   │   ├── redis/
│   │   │   │   └── redis.go
│   │   │   └── memory/
│   │   │       ├── otp_cache.go
│   │   │       └── refresh_token_cache.go
│   │   ├── email/
│   │   │   └── smtp.go           # 이메일 서비스
│   │   ├── fcm/
│   │   │   └── firebase.go       # FCM 푸시 알림
│   │   └── security/
│   │       ├── jwt.go            # JWT 처리
│   │       └── bcrypt.go         # 패스워드 암호화
│   └── interfaces/                 # 인터페이스 어댑터
│       ├── http/
│       │   ├── router/
│       │   │   └── router.go     # 라우터 설정
│       │   ├── middleware/
│       │   │   ├── auth.go       # 인증 미들웨어
│       │   │   ├── cors.go       # CORS 설정
│       │   │   └── logger.go     # 로깅
│       │   ├── handler/
│       │   │   ├── auth_handler.go
│       │   │   ├── member_handler.go
│       │   │   ├── room_handler.go
│       │   │   ├── prayer_handler.go
│       │   │   ├── invitation_handler.go
│       │   │   ├── fcm_token_handler.go
│       │   │   └── common.go     # 공통 핸들러 함수
│       │   └── dto/               # Request/Response DTO
│       │       ├── auth_dto.go
│       │       ├── member_dto.go
│       │       ├── room_dto.go
│       │       ├── prayer_dto.go
│       │       ├── invitation_dto.go
│       │       └── fcm_token_dto.go
│       └── grpc/                   # gRPC (향후 확장용)
├── pkg/                            # 외부에서 사용 가능한 패키지
│   ├── errors/
│   │   └── errors.go              # 커스텀 에러
│   ├── validator/
│   │   └── validator.go           # 검증 유틸
│   └── utils/
│       ├── pagination.go          # 페이지네이션
│       └── time.go                # 시간 유틸
├── migrations/                     # DB 마이그레이션
│   └── *.sql
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── go.mod
└── go.sum
```

## DDD 계층 구조

### 1. Domain Layer (도메인 계층)
- 비즈니스 로직의 핵심
- Entity, Value Object, Domain Service, Repository Interface 정의
- 외부 의존성 없음

### 2. Application Layer (애플리케이션 계층)
- Use Case 구현
- 도메인 객체들을 조합하여 비즈니스 요구사항 구현
- 트랜잭션 경계 설정

### 3. Infrastructure Layer (인프라 계층)
- Repository 구현체
- 외부 서비스 연동 (DB, Cache, Email, FCM 등)
- 기술적 세부사항 처리

### 4. Interface Layer (인터페이스 계층)
- HTTP Handler, DTO
- 요청/응답 변환
- 인증/인가 처리

## 주요 설계 원칙

1. **의존성 역전 원칙 (DIP)**
   - 도메인은 인프라에 의존하지 않음
   - Repository 인터페이스는 도메인에, 구현체는 인프라에

2. **단일 책임 원칙 (SRP)**
   - 각 계층과 모듈은 하나의 책임만 가짐

3. **개방-폐쇄 원칙 (OCP)**
   - 확장에는 열려있고 수정에는 닫혀있음

4. **클린 아키텍처**
   - 비즈니스 로직이 프레임워크, DB, UI에 독립적

## 도메인 관계도

```
Member ←→ MemberRoom ←→ Room
   ↓         ↓
FcmToken   Prayer (Title/Content/Completion)
   ↓         ↑
Notification ←┘
   ↑
Invitation → Member
   ↓
  Room
```

## 주요 기능

1. **인증/인가**
   - JWT 토큰 기반 인증
   - OTP 이메일 인증
   - Refresh Token

2. **기도방 관리**
   - CRUD 작업
   - 멤버 초대/관리
   - 역할 기반 권한 (OWNER, MEMBER)

3. **기도 관리**
   - 기도 제목/내용 CRUD
   - 기도 완료 처리
   - 무한 스크롤

4. **알림**
   - FCM 푸시 알림
   - 기도 완료 알림

5. **초대**
   - 이메일 초대
   - 초대 수락/거절