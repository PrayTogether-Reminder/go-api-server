# Member API Specification

## Overview
Member 도메인은 회원의 프로필 조회, 프로필 수정, 회원 검색 API를 제공합니다. 모든 엔드포인트는 `MemberController`(`src/main/java/site/praytogether/pray_together/domain/member/controller/MemberController.java`)를 통해 노출됩니다.

## API 목록 (총 3개)

### 프로필 관리 (2개)
1. **GET /api/v1/members/me** - 내 프로필 조회
2. **PATCH /api/v1/members/me** - 내 프로필 수정

### 회원 검색 (1개)
3. **GET /api/v1/members/search** - 이름으로 회원 검색

---

## Architecture Layers
- **Handler/Controller**: HTTP 요청 수신, DTO 검증, 적절한 HTTP Status/Body 반환 (`MemberController`).
- **UseCase/Application**: 도메인 서비스 조합 및 트랜잭션 관리 (`MemberApplicationService`).
- **Domain/Service**: 회원 조회/수정 로직 처리 (`MemberService`).
- **Repository/Infra**: 회원 데이터 영속화 및 조회 (`MemberRepository`).

### Common Concepts

#### Authentication Context
- `@PrincipalId` 파라미터 리졸버가 SecurityContext의 `PrayTogetherPrincipal`에서 memberId를 추출합니다.
- Access Token(JWT)은 Authorization 헤더(`Bearer <token>`)로 전달되며, 미인증 시 401 Unauthorized.

#### PhoneNumber Value Object
- 전화번호는 `PhoneNumber` 임베디드 타입으로 관리됩니다.
- 하이픈 포함/미포함 입력을 허용하며, 저장 시 `XXX-XXXX-XXXX` 형식으로 정규화됩니다.
- 검증 규칙: 한국 휴대폰 번호 형식 (010, 011, 016, 017, 018, 019 + 8자리 숫자).
- `getSuffix()` 메서드로 뒤 4자리 추출이 가능합니다.

#### Error Responses
- 공통 유효성 검증 실패: 400 Bad Request (Bean Validation 메시지 그대로 반환).
- `MemberNotFoundException`: 404 Not Found (`MEMBER-001`) - "회원 정보를 찾을 수 없습니다."

---

## API Endpoints

### API 1. 내 프로필 조회 (Get My Profile)

#### HTTP Request
```
GET /api/v1/members/me
```

#### Request Headers
- `Authorization`: Bearer access token (required)

#### Request Body
- 없음

#### Response
**Status Code**: 200 OK

**Body**:
```json
{
  "id": "Long",
  "name": "String",
  "email": "String",
  "phoneNumber": "String (nullable, format: XXX-XXXX-XXXX)"
}
```

**Example**:
```json
{
  "id": 123,
  "name": "홍길동",
  "email": "hong@example.com",
  "phoneNumber": "010-1234-5678"
}
```

#### Business Logic Flow
**UseCase Layer** (`memberApplication.getProfile`):
1. SecurityContext에서 추출된 `memberId`를 입력받습니다.
2. `MemberService.fetchProfileById(memberId)` 호출로 회원 프로필 조회.
   - 존재하지 않으면 `MemberNotFoundException`(404) 발생.
3. `MemberProfile` 도메인 모델을 `MemberProfileResponse` DTO로 변환.
4. 전화번호가 없는 경우 `phoneNumber` 필드는 `null`로 반환됩니다.

#### Authorization Rules
- Access Token 인증이 완료된 본인의 프로필만 조회할 수 있습니다.

---

### API 2. 내 프로필 수정 (Update My Profile)

#### HTTP Request
```
PATCH /api/v1/members/me
```

#### Request Headers
- `Authorization`: Bearer access token (required)
- `Content-Type`: application/json

#### Request Body
```json
{
  "name": "String (optional, 1-10 characters)",
  "phoneNumber": "String (optional, Korean phone format)"
}
```

**Validation Rules:**
- `name`: 선택적, 1자 이상 10자 이하 (`@Size`) - "이름은 1자 이상 10자 이하로 입력해 주세요."
- `phoneNumber`: 선택적, 한국 휴대폰 번호 형식 (`@Pattern`) - "올바른 휴대폰 번호 형식이 아닙니다."
  - 정규식: `^01[016789]-?\\d{4}-?\\d{4}$`
  - 하이픈 포함 여부 무관 (예: `010-1234-5678` 또는 `01012345678`)

**Note**:
- `name`과 `phoneNumber`는 모두 선택적(optional) 필드입니다.
- 수정하고 싶은 필드만 포함하면 됩니다.
- 빈 JSON `{}`을 보내면 아무것도 수정되지 않습니다.

#### Response
**Status Code**: 200 OK

**Body**:
```json
{
  "message": "String"
}
```

**Example**:
```json
{
  "message": "프로필을 변경했습니다."
}
```

#### Business Logic Flow
**UseCase Layer** (`memberApplication.updateProfile`):
1. SecurityContext에서 추출된 `memberId`와 `UpdateProfileRequest`를 입력받습니다.
2. Request를 `UpdateProfileCommand` 도메인 커맨드로 변환.
   - `phoneNumber` 문자열 → `PhoneNumber` Value Object 변환.
   - 변환 과정에서 전화번호 형식 검증 및 정규화 수행.
3. `MemberService.fetchById(memberId)`로 회원 엔티티 조회.
   - 존재하지 않으면 `MemberNotFoundException`(404) 발생.
4. **부분 업데이트 수행**:
   - `command.getName() != null`이면 `member.updateName(name)` 호출.
   - `command.getPhoneNumber() != null`이면 `member.updatePhoneNumber(phoneNumber)` 호출.
5. JPA Dirty Checking으로 변경사항 자동 저장 (@Transactional).
6. 성공 메시지 반환.

#### Authorization Rules
- Access Token 인증이 완료된 본인의 프로필만 수정할 수 있습니다.

#### Domain Business Rules
- 이름과 전화번호는 독립적으로 수정 가능합니다 (부분 업데이트).
- 전화번호는 하이픈 포함/미포함 입력을 모두 허용하지만, 저장 시 `XXX-XXXX-XXXX` 형식으로 정규화됩니다.
- 이메일과 비밀번호는 이 API로 수정할 수 없습니다.

---

### API 3. 회원 검색 (Search Members)

#### HTTP Request
```
GET /api/v1/members/search
```

#### Query Parameters
- `name`: String (required) - 검색할 회원 이름 (부분 일치)

#### Request Headers
- `Authorization`: Bearer access token (required)

#### Response
**Status Code**: 200 OK

**Body**:
```json
{
  "members": [
    {
      "id": "Long",
      "name": "String",
      "phoneNumberSuffix": "String (nullable, 전화번호 뒤 4자리)"
    }
  ]
}
```

**Example**:
```json
{
  "members": [
    {
      "id": 123,
      "name": "홍길동",
      "phoneNumberSuffix": "5678"
    },
    {
      "id": 456,
      "name": "홍길순",
      "phoneNumberSuffix": null
    }
  ]
}
```

**Note**:
- 전화번호 전체가 아닌 뒤 4자리만 반환됩니다 (개인정보 보호).
- 전화번호를 등록하지 않은 회원의 경우 `phoneNumberSuffix`는 `null`입니다.

#### Business Logic Flow
**UseCase Layer** (`memberApplication.searchMembers`):
1. SecurityContext에서 추출된 `memberId` (인증 확인용)와 검색어 `searchName`을 입력받습니다.
2. 검색어를 `SearchQueryMember` 도메인 모델로 변환.
3. `MemberService.searchByName(queryMember)` 호출.
   - Repository에서 `findByNameContaining(name)` 쿼리 실행 (부분 일치 검색).
   - `SearchResultMember` 리스트 반환 (id, name, phoneNumber).
4. `SearchResultMembers` 도메인 모델을 `SearchMemberResponse` DTO로 변환.
   - `phoneNumber.getSuffix()`를 호출해 뒤 4자리만 추출.
5. 검색 결과 리스트 반환.

#### Authorization Rules
- Access Token 인증이 완료된 회원만 검색할 수 있습니다.
- 본인 외의 회원도 검색 가능합니다 (공개 검색).

#### Search Behavior
- **부분 일치 검색**: 입력한 이름이 포함된 모든 회원을 반환합니다.
  - 예: "홍" 검색 시 "홍길동", "홍길순", "김홍철" 모두 반환.
- **대소문자 구분**: 데이터베이스 설정에 따라 다를 수 있습니다.
- **빈 검색어**: 현재 구현상 빈 문자열도 허용되며, 모든 회원을 반환할 수 있습니다.

---

## DTO Summary

### Request DTOs

#### UpdateProfileRequest
```
{
  name: String (optional, @Size(min=1, max=10))
  phoneNumber: String (optional, @Pattern(regexp="^01[016789]-?\\d{4}-?\\d{4}$"))
}
```

**Validation Messages:**
- name: "이름은 1자 이상 10자 이하로 입력해 주세요."
- phoneNumber: "올바른 휴대폰 번호 형식이 아닙니다."

### Response DTOs

#### MemberProfileResponse
```
{
  id: Long
  name: String
  email: String
  phoneNumber: String (nullable)
}
```

#### SearchMemberResponse
```
{
  members: Array<SearchMemberDto>
}

SearchMemberDto:
{
  id: Long
  name: String
  phoneNumberSuffix: String (nullable, 뒤 4자리)
}
```

#### MessageResponse
```
{
  message: String
}
```

---

## Business Rules & Authorization

### Authorization Rules
1. **프로필 조회**: 인증된 본인만 자신의 프로필 조회 가능
2. **프로필 수정**: 인증된 본인만 자신의 프로필 수정 가능
3. **회원 검색**: 인증된 모든 회원이 다른 회원 검색 가능

### Data Consistency Rules
1. **이름 변경**: 1자 이상 10자 이하로 제한됩니다.
2. **전화번호 변경**: 한국 휴대폰 번호 형식만 허용되며, 저장 시 `XXX-XXXX-XXXX` 형식으로 정규화됩니다.
3. **이메일 불변**: 이메일은 회원 가입 후 변경할 수 없습니다 (고유 식별자).
4. **부분 업데이트**: 프로필 수정 시 제공된 필드만 업데이트되며, 제공되지 않은 필드는 기존 값을 유지합니다.

### Validation Rules
1. **이름 검증**:
   - 길이: 1자 이상 10자 이하
   - 공백만으로는 구성될 수 없음
2. **전화번호 검증**:
   - 형식: `01[016789]-?\\d{4}-?\\d{4}` (하이픈 선택적)
   - 정규화 후 형식: `XXX-XXXX-XXXX`
   - Value Object 생성 시 자동 검증 및 정규화 수행
3. **검색어 검증**:
   - 현재 빈 문자열도 허용됨 (개선 가능)

---

## Implementation Notes for Go Migration

### Handler Layer (Controller)
- 동일한 엔드포인트/경로를 유지합니다.
- Request validation: struct tags를 사용해 검증합니다.
- JSON 바인딩 및 응답 직렬화를 처리합니다.
- 인증 미들웨어에서 memberId를 추출해 handler에 전달합니다.

### UseCase Layer (Application Service)
- `MemberApplicationService`의 역할을 Go service로 분리합니다.
- 트랜잭션 경계를 관리합니다 (@Transactional).
- Request DTO → Domain Command 변환을 담당합니다.
- 부분 업데이트 로직 (nil 체크)을 유지합니다.

### Domain Layer (Service & Entity)
- **PhoneNumber Value Object**:
  - 전화번호 검증 및 정규화 로직을 구조체 메서드로 구현합니다.
  - `Normalize()`, `Validate()`, `GetSuffix()` 메서드 제공.
- **Member Entity**:
  - `UpdateName()`, `UpdatePhoneNumber()` 메서드로 불변성을 보장합니다.
- **부분 업데이트**: Optional 필드 처리를 위해 포인터 타입 사용을 고려합니다.

### Repository Layer (Infrastructure)
- 프로필 조회를 위한 projection 쿼리를 구현합니다.
- 이름 검색: `LIKE '%name%'` 또는 full-text search를 사용합니다.
- 전화번호 뒤 4자리 추출 로직을 repository 또는 mapper에서 처리합니다.

### Cross-Cutting Concerns
- **인증**: JWT 토큰 검증 및 memberId 추출.
- **로깅**: API 시작/종료 로그 (`[API] 내 정보 조회 시작/종료`).
- **에러 처리**: 일관된 에러 응답 포맷 유지.
- **개인정보 보호**: 검색 시 전화번호 전체가 아닌 뒤 4자리만 노출.

### Testing Considerations
1. **프로필 조회**: 존재하지 않는 memberId로 조회 시 404 응답 확인.
2. **프로필 수정**:
   - 부분 업데이트 검증 (name만, phoneNumber만, 둘 다, 둘 다 없음).
   - 전화번호 형식 검증 (하이픈 포함/미포함 모두 성공).
   - 잘못된 전화번호 형식으로 400 응답 확인.
3. **회원 검색**:
   - 부분 일치 검색이 정상 작동하는지 확인.
   - 전화번호 뒤 4자리만 반환되는지 확인.
   - 전화번호가 없는 회원의 경우 null 처리 확인.
4. **인증**: 토큰 없이 호출 시 401 응답 확인.

---

## Security & Privacy Considerations

1. **개인정보 보호**:
   - 회원 검색 시 전화번호 전체를 노출하지 않고 뒤 4자리만 반환합니다.
   - 이메일은 검색 결과에 포함되지 않습니다.

2. **접근 제어**:
   - 프로필 조회/수정은 본인만 가능합니다.
   - 회원 검색은 인증된 모든 사용자가 가능합니다.

3. **데이터 검증**:
   - 전화번호는 Value Object를 통해 도메인 레벨에서 검증됩니다.
   - Controller 레벨의 Bean Validation과 이중 검증 구조를 가집니다.

4. **Rate Limiting**:
   - 현재 구현되어 있지 않으며, 외부 API Gateway나 미들웨어에서 처리해야 합니다.
   - 특히 회원 검색 API는 남용 방지를 위해 rate limiting 적용을 권장합니다.
