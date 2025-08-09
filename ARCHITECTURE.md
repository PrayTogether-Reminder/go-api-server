# Pray Together API Server - 전체 아키텍처 문서

## 🎯 프로젝트 개요

**Pray Together**는 기독교 공동체를 위한 기도 공유 및 관리 플랫폼입니다.
- 사용자들이 기도방을 만들고, 서로의 기도 제목을 공유하며, 함께 기도할 수 있는 서비스
- Spring Boot에서 Go로 마이그레이션되었으며, DDD(Domain-Driven Design) 원칙을 따름

## 🏗️ 시스템 아키텍처

### 전체 구조
```
┌──────────────────────────────────────────────────┐
│                   Client Apps                     │
│         (iOS, Android, Web)                       │
└──────────────────────────────────────────────────┘
                          │
                          ▼ HTTP/REST
┌──────────────────────────────────────────────────┐
│                  API Gateway                      │
│              (Nginx/Load Balancer)                │
└──────────────────────────────────────────────────┘
                          │
                          ▼
┌──────────────────────────────────────────────────┐
│              Go API Server (This)                 │
│                                                   │
│  ┌────────────────────────────────────────────┐  │
│  │            Interface Layer                  │  │
│  │         (HTTP Handlers, Routes)            │  │
│  └────────────────────────────────────────────┘  │
│                        │                          │
│  ┌────────────────────────────────────────────┐  │
│  │          Application Layer                  │  │
│  │         (Use Cases, DTOs)                  │  │
│  └────────────────────────────────────────────┘  │
│                        │                          │
│  ┌────────────────────────────────────────────┐  │
│  │            Domain Layer                     │  │
│  │    (Entities, Domain Services, Rules)      │  │
│  └────────────────────────────────────────────┘  │
│                        │                          │
│  ┌────────────────────────────────────────────┐  │
│  │         Infrastructure Layer                │  │
│  │     (DB, Email, FCM, External APIs)        │  │
│  └────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        ▼                 ▼                 ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│  PostgreSQL  │ │   Firebase   │ │     SMTP     │
│   Database   │ │     FCM      │ │    Server    │
└──────────────┘ └──────────────┘ └──────────────┘
```

## 🎨 도메인 모델

### 핵심 도메인
```
┌─────────────────────────────────────────────────────┐
│                    Member (회원)                     │
│  - ID, Name, Email, Phone                          │
│  - Password (hashed)                               │
│  - ProfileImageURL                                 │
│  - Status (ACTIVE, INACTIVE, BLOCKED)              │
└─────────────────────────────────────────────────────┘
              │                    │
              ▼                    ▼
┌─────────────────────┐  ┌─────────────────────────┐
│   Room (기도방)      │  │  Invitation (초대)       │
│  - Name             │  │  - InviterID            │
│  - Description      │  │  - InviteeEmail         │
│  - IsPrivate        │  │  - RoomID               │
│  - PrayTime         │  │  - Status               │
│  - NotificationTime │  │  - ExpiryDate           │
└─────────────────────┘  └─────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────────────┐
│           RoomMember (방 멤버 관계)                  │
│  - MemberID, RoomID                                │
│  - Role (OWNER, MEMBER)                            │
│  - JoinedAt, LastPrayedAt                          │
│  - PrayCount, IsNotification                       │
└─────────────────────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────────────┐
│            Prayer (기도 제목)                        │
│  ┌────────────────────────────────────┐            │
│  │    PrayerTitle (제목)               │            │
│  │  - Title, CreatorID, RoomID        │            │
│  └────────────────────────────────────┘            │
│                    │                                │
│  ┌────────────────────────────────────┐            │
│  │    PrayerContent (내용)             │            │
│  │  - Content, AuthorID                │            │
│  │  - IsCompleted                     │            │
│  └────────────────────────────────────┘            │
└─────────────────────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────────────┐
│         Notification (알림)                          │
│  - Type (PRAYER_COMPLETE, INVITATION, etc)          │
│  - FromMemberID, ToMemberID                        │
│  - Content, IsRead                                 │
└─────────────────────────────────────────────────────┘
```

## 📁 프로젝트 구조

### 도메인 기반 모듈러 구조
```
go-api-server/
├── cmd/
│   └── server/
│       └── main.go              # 애플리케이션 진입점
│
├── config/
│   └── config.go               # 환경설정 관리
│
├── internal/
│   ├── domains/                # 도메인 모듈 (DDD)
│   │   ├── member/             # 회원 도메인
│   │   ├── room/               # 기도방 도메인
│   │   ├── prayer/             # 기도 도메인
│   │   ├── auth/               # 인증 도메인
│   │   ├── invitation/         # 초대 도메인
│   │   ├── fcmtoken/           # FCM 토큰 도메인
│   │   └── notification/       # 알림 도메인
│   │
│   └── common/                 # 공통 모듈
│       ├── errors/             # 에러 처리
│       ├── middleware/         # 미들웨어
│       └── utils/              # 유틸리티
│
├── test/                       # 통합 테스트
│   ├── *_integration_test.go  # 각 기능별 통합 테스트
│   └── test_utils.go          # 테스트 유틸리티
│
└── docs/                       # 문서
    └── api/                    # API 문서
```

### 각 도메인의 4계층 구조
```
domains/prayer/
├── domain/                     # 도메인 계층
│   ├── prayer_title.go        # 엔티티
│   ├── prayer_content.go      # 엔티티
│   ├── repository.go          # 리포지토리 인터페이스
│   └── service.go             # 도메인 서비스
│
├── application/                # 애플리케이션 계층
│   ├── create_prayer_usecase.go
│   ├── update_prayer_usecase.go
│   └── dto.go                 # DTO 정의
│
├── infrastructure/             # 인프라 계층
│   └── gorm_repository.go     # GORM 리포지토리 구현
│
└── interfaces/                 # 인터페이스 계층
    ├── http/
    │   └── handler.go         # HTTP 핸들러
    └── api.go                 # 공개 API 인터페이스
```

## 🔄 주요 비즈니스 플로우

### 1. 회원가입 및 로그인
```
1. 회원가입 요청 → 이메일 중복 확인
2. 비밀번호 해시화 (bcrypt)
3. 회원 정보 저장
4. JWT 토큰 발급 (Access + Refresh)
5. FCM 토큰 등록 (푸시 알림용)
```

### 2. 기도방 생성 및 참여
```
1. 기도방 생성 → 생성자는 자동으로 OWNER 역할
2. 초대 링크 생성 또는 이메일 초대
3. 초대 수락 → RoomMember 관계 생성
4. 방 멤버는 기도 제목 공유 가능
```

### 3. 기도 제목 공유
```
1. 기도 제목(Title) 생성
2. 여러 멤버가 기도 내용(Content) 추가 가능
3. 기도 완료 시 알림 발송
4. 통계 업데이트 (기도 횟수 등)
```

### 4. 무한 스크롤 페이지네이션
```
1. 시간 기반 커서 페이지네이션
2. after 파라미터로 다음 페이지 요청
3. 마지막 페이지는 빈 배열 반환
4. 페이지 크기: 20개 (고정)
```

## 🔐 보안 및 인증

### JWT 인증
- Access Token: 30분 유효
- Refresh Token: 7일 유효
- Bearer Token 형식 사용

### 비밀번호 보안
- bcrypt 해싱 (cost factor: 10)
- 최소 8자, 영문+숫자 조합

### API 보안
- CORS 설정
- Rate Limiting (계획)
- SQL Injection 방지 (GORM)

## 📊 데이터베이스

### PostgreSQL 스키마
```sql
-- 주요 테이블
- member (회원)
- room (기도방)
- member_room (멤버-방 관계)
- prayer_title (기도 제목)
- prayer_content (기도 내용)
- invitation (초대)
- notification (알림)
- fcm_token (FCM 토큰)
```

### 인덱스 전략
- 외래키 자동 인덱스
- created_at 정렬용 인덱스
- email, phone 유니크 인덱스

## 🚀 배포 및 운영

### 환경 구성
- 개발: `.env` 파일 사용
- 프로덕션: 환경변수 직접 주입
- Docker 컨테이너화 지원

### 모니터링
- 구조화된 로깅
- 에러 추적
- 성능 메트릭 (계획)

## 📈 확장성

### 마이크로서비스 전환 가능
- 각 도메인이 독립적 모듈
- 인터페이스 기반 통신
- 이벤트 드리븐 아키텍처 준비

### 성능 최적화
- 데이터베이스 커넥션 풀
- 캐싱 전략 (Redis 준비)
- 비동기 처리 (고루틴)

## 🧪 테스트 전략

### 테스트 구조
```
- 통합 테스트: 14개 테스트 스위트
- 서브 테스트: 197개 테스트 케이스
- 테스트 커버리지: 주요 비즈니스 로직 100%
```

### 테스트 데이터
- In-memory SQLite 사용
- 테스트별 격리 (트랜잭션 롤백)
- 픽스처 데이터 제공

## 🔍 API 엔드포인트

### 주요 엔드포인트
```
POST   /api/v1/auth/login
POST   /api/v1/members
GET    /api/v1/members/profile
POST   /api/v1/rooms
GET    /api/v1/rooms (infinite scroll)
POST   /api/v1/prayers
GET    /api/v1/prayers (infinite scroll)
POST   /api/v1/invitations
PUT    /api/v1/invitations/:id/status
```

## 📝 마이그레이션 노트

### Java → Go 전환
- Spring Boot → Gin Framework
- JPA → GORM
- Maven/Gradle → Go Modules
- Junit → Go Testing + Testify

### 주요 변경사항
1. 패키지 구조: 레이어 기반 → 도메인 기반
2. 의존성 주입: Spring DI → 생성자 주입
3. 트랜잭션: @Transactional → 명시적 트랜잭션
4. 검증: Bean Validation → 커스텀 검증

## 🎯 향후 계획

### 단기 (1-3개월)
- [ ] API 문서 자동화 (Swagger)
- [ ] 로깅 시스템 개선
- [ ] 캐싱 레이어 추가

### 장기 (6-12개월)
- [ ] GraphQL 지원
- [ ] 실시간 기능 (WebSocket)
- [ ] 분석 대시보드
- [ ] 다국어 지원

## 📚 참고 자료

- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Best Practices](https://golang.org/doc/effective_go.html)
- [Twelve-Factor App](https://12factor.net/)

---

*Last Updated: 2025-08-09*
*Version: 1.0.0*