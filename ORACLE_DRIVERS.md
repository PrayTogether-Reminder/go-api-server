# Oracle 드라이버 정리

## 현재 프로젝트의 Oracle 드라이버 구조

### 1. 계층 구조

```
GORM (ORM Layer)
    ↓
godoes/gorm-oracle (GORM Oracle Dialect)
    ↓
sijms/go-ora/v2 (Pure Go Oracle Driver)
    ↓
Oracle Database (ATP)
```

### 2. 사용 중인 드라이버들

#### 2.1 기본 Oracle 드라이버들

| 드라이버 | 설명 | Oracle Client 필요 | 특징 |
|---------|------|-------------------|------|
| **sijms/go-ora/v2** | Pure Go Oracle driver | ❌ 불필요 | - 순수 Go로 구현<br>- Oracle Client 불필요<br>- Docker/컨테이너 친화적<br>- TNS, 직접 연결 모두 지원 |
| **godror/godror** | CGO 기반 Oracle driver | ✅ 필요 | - Oracle OCI 사용<br>- 높은 성능<br>- Oracle Client 설치 필수 |
| **mattn/go-oci8** | CGO 기반 Oracle driver | ✅ 필요 | - 오래된 드라이버<br>- Oracle Client 필요 |

#### 2.2 GORM용 Oracle 드라이버들

| 드라이버 | 내부 사용 드라이버 | Oracle Client 필요 | 상태 |
|---------|-------------------|-------------------|------|
| **godoes/gorm-oracle** | sijms/go-ora/v2 | ❌ 불필요 | ✅ 현재 사용 중 |
| **cengsin/oracle** | godror/godror | ✅ 필요 | ❌ 호환성 문제로 제거 |

### 3. 현재 설정 상황

#### 3.1 사용 중인 드라이버
```go
import (
    "github.com/godoes/gorm-oracle"  // GORM Oracle dialect
    "gorm.io/gorm"
)

// 내부적으로 sijms/go-ora/v2 사용
dialector = oracle.Open(dsn)
db, err := gorm.Open(dialector, &gorm.Config{})
```

#### 3.2 DSN 형식
```go
// Oracle ATP with SSL (현재 사용 중)
dsn = "oracle://ADMIN:password@adb.ap-chuncheon-1.oraclecloud.com:1522/service?SSL=true"

// TNS Descriptor 직접 사용 (시도했지만 파싱 오류)
dsn = "ADMIN/password@(description=...)"

// Wallet 사용 (시도했지만 ACL 오류)
dsn = "oracle://ADMIN:password@host:port/service?WALLET=path&SSL=true"
```

### 4. 시도한 방법들과 결과

| 시도 | 드라이버 | 결과 | 이유 |
|------|---------|------|------|
| 1. TNS Descriptor 직접 사용 | godoes/gorm-oracle | ❌ 실패 | "missing port in address" 파싱 오류 |
| 2. CengSin/oracle 사용 | cengsin/oracle | ❌ 실패 | GORM 최신 버전과 호환성 문제 |
| 3. Wallet 파일 사용 | godoes/gorm-oracle | ⚠️ 부분 성공 | ACL 필터링으로 거부됨 |
| 4. URL 형식 + SSL | godoes/gorm-oracle | ✅ 성공 | 현재 작동 중 |

### 5. 최종 선택: godoes/gorm-oracle + sijms/go-ora

#### 장점:
- ✅ **Oracle Client 불필요** - Pure Go 구현
- ✅ **Docker/컨테이너 친화적** - 추가 설치 없이 배포 가능
- ✅ **GORM 완벽 지원** - AutoMigrate, 쿼리 빌더 등 모든 기능 사용 가능
- ✅ **Oracle ATP 지원** - SSL 연결 지원

#### 단점:
- ⚠️ TNS Descriptor 직접 파싱 제한적
- ⚠️ Wallet 파일 지원 불완전
- ⚠️ godror보다 성능은 다소 낮을 수 있음

### 6. 연결 코드 예시

```go
// cmd/server/main.go
case "oracle":
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        // Oracle ATP URL 형식 with SSL
        encodedPassword := url.QueryEscape(config.Database.Password)
        dsn = fmt.Sprintf("oracle://%s:%s@adb.ap-chuncheon-1.oraclecloud.com:1522/g0524ab680e3e6c_z5f5ees1n47gddba_high.adb.oraclecloud.com?SSL=true",
            config.Database.User, encodedPassword)
    }
    dialector = oracle.Open(dsn)
    db, err := gorm.Open(dialector, &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
```

### 7. 환경 변수 설정

```env
# .env
DB_TYPE=oracle
DB_USER=ADMIN
DB_PASSWORD=Vldrjtmxkdlf3#

# 또는 직접 DATABASE_URL 설정
DATABASE_URL=oracle://ADMIN:password@adb.ap-chuncheon-1.oraclecloud.com:1522/service?SSL=true
```

### 8. 의존성 정리

```bash
# 필요한 의존성
go get github.com/godoes/gorm-oracle  # GORM Oracle dialect
go get gorm.io/gorm                   # GORM ORM

# 자동으로 포함되는 의존성
# - github.com/sijms/go-ora/v2 (godoes/gorm-oracle이 내부적으로 사용)

# 제거한 의존성
# - github.com/cengsin/oracle (GORM 호환성 문제)
# - github.com/godror/godror (Oracle Client 필요)
```

### 9. 결론

현재 프로젝트는 **godoes/gorm-oracle** + **sijms/go-ora/v2** 조합을 사용하여:
- Oracle Client 설치 없이 Oracle ATP 연결 성공
- GORM의 모든 기능 활용 가능
- 컨테이너 배포 시 추가 설정 불필요
- Pure Go 구현으로 크로스 플랫폼 지원

이 구성은 프로젝트의 요구사항을 충족하며, 배포와 유지보수가 용이한 최적의 선택입니다.