# 🏗️ 도메인별 모듈 구조 (Domain-Driven Module Structure)

## 📁 새로운 프로젝트 구조

```
go-api-server/
├── cmd/
│   └── api/
│       └── main.go              # 애플리케이션 진입점
├── config/
│   └── config.go                # 설정 관리
├── internal/
│   ├── domains/                 # 🌟 도메인별 모듈
│   │   ├── member/              # Member 도메인
│   │   │   ├── domain/          # 도메인 계층
│   │   │   │   ├── entity.go
│   │   │   │   ├── repository.go
│   │   │   │   └── service.go
│   │   │   ├── application/     # 애플리케이션 계층
│   │   │   │   └── usecase.go
│   │   │   ├── infrastructure/  # 인프라 계층
│   │   │   │   └── repository.go
│   │   │   └── interfaces/      # 인터페이스 계층
│   │   │       ├── http/
│   │   │       │   ├── handler.go
│   │   │       │   └── dto.go
│   │   │       └── api.go      # 외부 모듈용 인터페이스
│   │   ├── room/                # Room 도메인 (MemberRoom 포함)
│   │   │   ├── domain/
│   │   │   │   ├── room.go
│   │   │   │   ├── room_member.go  # 👈 MemberRoom은 여기!
│   │   │   │   ├── repository.go
│   │   │   │   └── service.go
│   │   │   ├── application/
│   │   │   ├── infrastructure/
│   │   │   └── interfaces/
│   │   ├── prayer/              # Prayer 도메인
│   │   │   ├── domain/
│   │   │   ├── application/
│   │   │   ├── infrastructure/
│   │   │   └── interfaces/
│   │   ├── auth/                # Auth 도메인
│   │   │   ├── domain/
│   │   │   ├── application/
│   │   │   ├── infrastructure/
│   │   │   └── interfaces/
│   │   ├── invitation/          # Invitation 도메인
│   │   │   ├── domain/
│   │   │   ├── application/
│   │   │   ├── infrastructure/
│   │   │   └── interfaces/
│   │   ├── fcm_token/           # FCM Token 도메인
│   │   │   ├── domain/
│   │   │   ├── application/
│   │   │   ├── infrastructure/
│   │   │   └── interfaces/
│   │   └── notification/        # Notification 도메인
│   │       ├── domain/
│   │       ├── application/
│   │       ├── infrastructure/
│   │       └── interfaces/
│   ├── domain_services/         # 🌟 도메인 간 협업 서비스
│   │   ├── prayer_creation_service.go
│   │   ├── invitation_service.go
│   │   └── notification_service.go
│   ├── shared/                  # 공통 코드
│   │   ├── base/                # BaseEntity 등
│   │   │   ├── entity.go
│   │   │   └── response.go
│   │   ├── errors/              # 공통 에러
│   │   │   └── errors.go
│   │   ├── events/              # 도메인 이벤트
│   │   │   └── event_bus.go
│   │   └── utils/               # 유틸리티
│   │       ├── pagination.go
│   │       └── time.go
│   └── infrastructure/          # 공통 인프라
│       ├── database/
│       │   └── gorm.go
│       ├── cache/
│       │   └── memory/
│       ├── email/
│       │   └── smtp.go
│       └── security/
│           ├── jwt.go
│           └── bcrypt.go
├── pkg/                         # 외부 패키지
├── migrations/                  # DB 마이그레이션
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── go.mod
└── go.sum
```

## 🎯 핵심 변경사항

### 1. 도메인별 완전한 캡슐화
- 각 도메인이 자체 4계층 구조를 가짐
- 도메인 간 직접 참조 금지

### 2. MemberRoom → Room 도메인에 포함
- `room/domain/room_member.go`로 위치
- Room의 Aggregate 일부로 관리

### 3. Domain Services
- 도메인 간 협업이 필요한 비즈니스 로직
- 예: Prayer 생성 시 Member, Room 검증

### 4. 각 도메인의 API 인터페이스
- `interfaces/api.go`: 다른 도메인이 사용할 수 있는 공개 인터페이스
- 구현 디테일 은닉

## 🔄 의존성 규칙

```
1. Domain → 없음 (순수 비즈니스 로직)
2. Application → Domain, Domain Services
3. Infrastructure → Domain
4. Interfaces → Application
5. Domain Services → 여러 Domain의 Repository/Service
6. Main → 모든 모듈 조립
```

## 🚀 장점

1. **높은 응집도**: 관련 코드가 한 곳에
2. **명확한 경계**: 도메인 간 인터페이스로만 통신
3. **독립적 개발**: 팀별로 도메인 담당 가능
4. **테스트 용이**: 도메인별 독립 테스트
5. **마이크로서비스 전환 용이**: 도메인별 분리 가능

## 📝 마이그레이션 순서

1. ✅ 폴더 구조 생성
2. ⏳ Shared 코드 이동
3. ⏳ 각 도메인별 코드 재구성
4. ⏳ Domain Services 구현
5. ⏳ Main 함수 수정