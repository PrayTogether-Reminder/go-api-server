package auth_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/error"
	testutil2 "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/testutil"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogout_Success(t *testing.T) {
	authHandler, db := setupTestEnvironmentWithDB(t)
	testMember := testutil2.CreateTestMember(t, db)
	ctx := context.Background()

	refreshTokenRepo := auth.NewRefreshTokenRepository()
	refreshTokenService := auth.NewRefreshTokenService(refreshTokenRepo)

	expiresAt := time.Now().Add(time.Hour)
	require.NoError(t, refreshTokenService.Save(ctx, db, testMember.ID, "test-refresh-token", expiresAt))

	router := testutil2.SetupAuthenticatedRouter(testMember.ID)
	router.POST("/api/v1/auth/logout", authHandler.Logout)

	recorder := testutil2.ExecuteRequest(t, router, testutil2.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/auth/logout",
	})

	assert.Equal(t, http.StatusNoContent, recorder.Code)
	assert.Equal(t, 0, recorder.Body.Len(), "logout should not return a response body")

	exists, err := refreshTokenRepo.Exists(ctx, db, testMember.ID, "test-refresh-token")
	require.NoError(t, err)
	assert.False(t, exists, "refresh token should be deleted during logout")
}

func TestLogout_Unauthenticated(t *testing.T) {
	authHandler, _ := setupTestEnvironment(t)
	router := testutil2.SetupTestRouter()
	router.POST("/api/v1/auth/logout", authHandler.Logout)

	recorder := testutil2.ExecuteRequest(t, router, testutil2.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/auth/logout",
	})

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil2.ParseResponse(t, recorder, &errorResponse)
	assert.Equal(t, "AUTH-000", errorResponse.Code)
}

func TestLogout_SucceedsWithoutRefreshToken(t *testing.T) {
	authHandler, db := setupTestEnvironmentWithDB(t)
	testMember := testutil2.CreateTestMember(t, db)

	router := testutil2.SetupAuthenticatedRouter(testMember.ID)
	router.POST("/api/v1/auth/logout", authHandler.Logout)

	recorder := testutil2.ExecuteRequest(t, router, testutil2.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/auth/logout",
	})

	assert.Equal(t, http.StatusNoContent, recorder.Code)
}
