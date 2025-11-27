package auth_test

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	testutil2 "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/testutil"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/auth"
	"github.com/stretchr/testify/assert"
)

// TestReissueToken_Success - 토큰 재발급 성공 테스트
func TestReissueToken_Success(t *testing.T) {
	// Given: 테스트 환경 설정
	authHandler, db := setupTestEnvironmentWithDB(t)

	// 테스트 회원 생성
	member := testutil2.CreateTestMember(t, db)
	memberIDStr := strconv.FormatInt(member.ID, 10)

	// 실제 JWT Manager 생성 (토큰 발급/검증을 위해)
	cfg := testutil2.NewTestConfig()
	tokenManager := testutil2.NewRealJWTManager(cfg)

	// Refresh Token 생성 및 저장
	refreshToken, err := tokenManager.GenerateRefreshToken(memberIDStr, member.Email)
	assert.NoError(t, err)

	// Refresh Token 저장
	refreshTokenRepo := auth.NewRefreshTokenRepository()
	refreshTokenService := auth.NewRefreshTokenService(refreshTokenRepo)
	expiresAt, err := tokenManager.ExtractExpiration(refreshToken)
	assert.NoError(t, err)
	err = refreshTokenService.Save(context.Background(), db, member.ID, refreshToken, expiresAt)
	assert.NoError(t, err)

	// When: 토큰 재발급 요청
	router := testutil2.SetupTestRouter()
	router.POST("/api/v1/auth/reissue-token", authHandler.ReissueToken)

	request := testutil2.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/auth/reissue-token",
		Body: map[string]interface{}{
			"refreshToken": refreshToken,
		},
	}

	recorder := testutil2.ExecuteRequest(t, router, request)

	// Then: 응답 검증
	assert.Equal(t, http.StatusOK, recorder.Code, "Expected HTTP 200 OK for successful token reissue")

	var response auth.AuthTokenReissueResponse
	testutil2.ParseResponse(t, recorder, &response)

	assert.NotEmpty(t, response.AccessToken, "Access token should not be empty")
	assert.NotEmpty(t, response.RefreshToken, "Refresh token should not be empty")
	assert.NotEmpty(t, response.RefreshToken, "New refresh token should not be empty")
}

// TestReissueToken_InvalidToken - 유효하지 않은 토큰으로 재발급 시도
func TestReissueToken_InvalidToken(t *testing.T) {
	// Given: 테스트 환경 설정
	authHandler, db := setupTestEnvironmentWithDB(t)

	// 존재하지 않는 memberID로 토큰 생성 (JWT 형식은 맞지만 member가 없음)
	cfg := testutil2.NewTestConfig()
	tokenManager := testutil2.NewRealJWTManager(cfg)
	invalidToken, err := tokenManager.GenerateRefreshToken("99999", "nonexistent@example.com")
	assert.NoError(t, err)

	// When: 존재하지 않는 member의 토큰으로 재발급 요청
	router := testutil2.SetupTestRouter()
	router.POST("/api/v1/auth/reissue-token", authHandler.ReissueToken)

	request := testutil2.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/auth/reissue-token",
		Body: map[string]interface{}{
			"refreshToken": invalidToken,
		},
	}

	recorder := testutil2.ExecuteRequest(t, router, request)

	// Then: 404 Not Found 응답 검증 (member가 존재하지 않음)
	assert.Equal(t, http.StatusNotFound, recorder.Code, "Expected HTTP 404 Not Found for non-existent member")

	_ = db // Suppress unused variable warning
}

// TestReissueToken_TokenNotFound - DB에 저장되지 않은 토큰으로 재발급 시도
func TestReissueToken_TokenNotFound(t *testing.T) {
	// Given: 테스트 환경 설정
	authHandler, db := setupTestEnvironmentWithDB(t)

	// 테스트 회원 생성
	member := testutil2.CreateTestMember(t, db)
	memberIDStr := strconv.FormatInt(member.ID, 10)

	// 실제 JWT Manager 생성
	cfg := testutil2.NewTestConfig()
	tokenManager := testutil2.NewRealJWTManager(cfg)

	// Refresh Token 생성 (하지만 DB에 저장하지 않음)
	refreshToken, err := tokenManager.GenerateRefreshToken(memberIDStr, member.Email)
	assert.NoError(t, err)

	// When: DB에 저장되지 않은 토큰으로 재발급 요청
	router := testutil2.SetupTestRouter()
	router.POST("/api/v1/auth/reissue-token", authHandler.ReissueToken)

	request := testutil2.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/auth/reissue-token",
		Body: map[string]interface{}{
			"refreshToken": refreshToken,
		},
	}

	recorder := testutil2.ExecuteRequest(t, router, request)

	// Then: 401 Unauthorized 응답 검증
	assert.Equal(t, http.StatusUnauthorized, recorder.Code, "Expected HTTP 401 Unauthorized for token not found")
}

// TestReissueToken_ExpiredToken - 만료된 토큰으로 재발급 시도
func TestReissueToken_ExpiredToken(t *testing.T) {
	// Given: 테스트 환경 설정
	authHandler, db := setupTestEnvironmentWithDB(t)

	// 테스트 회원 생성
	member := testutil2.CreateTestMember(t, db)

	// 만료된 토큰을 DB에 저장
	refreshTokenRepo := auth.NewRefreshTokenRepository()
	refreshTokenService := auth.NewRefreshTokenService(refreshTokenRepo)
	expiredToken := "expired.token.here"
	expiredTime := time.Now().Add(-1 * time.Hour) // 1시간 전에 만료
	err := refreshTokenService.Save(context.Background(), db, member.ID, expiredToken, expiredTime)
	assert.NoError(t, err)

	// When: 만료된 토큰으로 재발급 요청
	router := testutil2.SetupTestRouter()
	router.POST("/api/v1/auth/reissue-token", authHandler.ReissueToken)

	request := testutil2.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/auth/reissue-token",
		Body: map[string]interface{}{
			"refreshToken": expiredToken,
		},
	}

	recorder := testutil2.ExecuteRequest(t, router, request)

	// Then: 401 Unauthorized 응답 검증
	assert.Equal(t, http.StatusUnauthorized, recorder.Code, "Expected HTTP 401 Unauthorized for expired token")
}

// TestReissueToken_TokenMismatch - 다른 회원의 토큰으로 재발급 시도
func TestReissueToken_TokenMismatch(t *testing.T) {
	// Given: 테스트 환경 설정
	authHandler, db := setupTestEnvironmentWithDB(t)

	// 두 명의 회원 생성
	member1 := testutil2.CreateTestMemberWithIndex(t, db, 0)
	member2 := testutil2.CreateTestMemberWithIndex(t, db, 1)

	// 실제 JWT Manager 생성
	cfg := testutil2.NewTestConfig()
	tokenManager := testutil2.NewRealJWTManager(cfg)

	// member1의 토큰 생성
	memberID1Str := strconv.FormatInt(member1.ID, 10)
	refreshToken1, err := tokenManager.GenerateRefreshToken(memberID1Str, member1.Email)
	assert.NoError(t, err)

	// member2의 토큰을 DB에 저장 (다른 토큰)
	memberID2Str := strconv.FormatInt(member2.ID, 10)
	refreshToken2, err := tokenManager.GenerateRefreshToken(memberID2Str, member2.Email)
	assert.NoError(t, err)

	refreshTokenRepo := auth.NewRefreshTokenRepository()
	refreshTokenService := auth.NewRefreshTokenService(refreshTokenRepo)
	expiresAt, err := tokenManager.ExtractExpiration(refreshToken2)
	assert.NoError(t, err)
	err = refreshTokenService.Save(context.Background(), db, member1.ID, refreshToken2, expiresAt) // member1의 ID로 저장
	assert.NoError(t, err)

	// When: member1의 실제 토큰으로 재발급 요청 (하지만 DB에는 다른 토큰이 저장됨)
	router := testutil2.SetupTestRouter()
	router.POST("/api/v1/auth/reissue-token", authHandler.ReissueToken)

	request := testutil2.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/auth/reissue-token",
		Body: map[string]interface{}{
			"refreshToken": refreshToken1, // DB에 저장된 것과 다른 토큰
		},
	}

	recorder := testutil2.ExecuteRequest(t, router, request)

	// Then: 401 Unauthorized 응답 검증
	assert.Equal(t, http.StatusUnauthorized, recorder.Code, "Expected HTTP 401 Unauthorized for token mismatch")
}

// TestReissueToken_MissingRefreshToken - RefreshToken 누락
func TestReissueToken_MissingRefreshToken(t *testing.T) {
	// Given: 테스트 환경 설정
	authHandler, _ := setupTestEnvironmentWithDB(t)

	// When: RefreshToken 없이 요청
	router := testutil2.SetupTestRouter()
	router.POST("/api/v1/auth/reissue-token", authHandler.ReissueToken)

	request := testutil2.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/auth/reissue-token",
		Body:   map[string]interface{}{
			// refreshToken 필드 누락
		},
	}

	recorder := testutil2.ExecuteRequest(t, router, request)

	// Then: 400 Bad Request 응답 검증
	assert.Equal(t, http.StatusBadRequest, recorder.Code, "Expected HTTP 400 Bad Request for missing refresh token")
}
