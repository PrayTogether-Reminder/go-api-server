package auth_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/member"
	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithdraw_Success(t *testing.T) {
	// Given: Setup test environment
	authHandler, db := setupTestEnvironmentWithDB(t)

	// Create test member
	testMember := testutil.CreateTestMember(t, db)
	router := testutil.SetupAuthenticatedRouter(testMember.ID)
	router.DELETE("/api/v1/auth/withdraw", authHandler.Withdraw)

	// When: Execute withdraw request
	request := testutil.TestRequest{
		Method: http.MethodDelete,
		URL:    "/api/v1/auth/withdraw",
		Body:   nil,
	}
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify response
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response sharedHttp.MessageResponse
	testutil.ParseResponse(t, recorder, &response)
	assert.Equal(t, "회원 탈퇴를 완료했습니다.\n 함께 기도해 주셔 감사합니다.", response.Message)

	// Verify member is deleted from database
	memberRepo := member.NewMemberRepository()
	memberService := member.NewMemberService(memberRepo)
	exists, err := memberService.ExistsByEmail(context.Background(), db, testMember.Email)
	require.NoError(t, err)
	assert.False(t, exists, "Member should be deleted from database")
}

func TestWithdraw_Unauthenticated(t *testing.T) {
	// Given: Setup test environment without authentication
	authHandler, _ := setupTestEnvironment(t)
	router := testutil.SetupTestRouter() // No authentication
	router.DELETE("/api/v1/auth/withdraw", authHandler.Withdraw)

	// When: Execute withdraw request without auth
	request := testutil.TestRequest{
		Method: http.MethodDelete,
		URL:    "/api/v1/auth/withdraw",
		Body:   nil,
	}
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify unauthorized error
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errorResponse)
	assert.Equal(t, "AUTH-000", errorResponse.Code)
	assert.Equal(t, "로그인을 해주세요.", errorResponse.Message)
}

func TestWithdraw_NonexistentMember(t *testing.T) {
	// Given: Setup test environment with non-existent member ID
	authHandler, _ := setupTestEnvironment(t)

	nonexistentMemberID := int64(99999)
	router := testutil.SetupAuthenticatedRouter(nonexistentMemberID)
	router.DELETE("/api/v1/auth/withdraw", authHandler.Withdraw)

	// When: Execute withdraw request
	request := testutil.TestRequest{
		Method: http.MethodDelete,
		URL:    "/api/v1/auth/withdraw",
		Body:   nil,
	}
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify not found error
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errorResponse)
	assert.Equal(t, "MEMBER-001", errorResponse.Code)
}

func TestWithdraw_AlreadyWithdrawn(t *testing.T) {
	// Given: Setup test environment
	authHandler, db := setupTestEnvironmentWithDB(t)

	// Create and withdraw member
	testMember := testutil.CreateTestMember(t, db)
	router := testutil.SetupAuthenticatedRouter(int64(testMember.ID))
	router.DELETE("/api/v1/auth/withdraw", authHandler.Withdraw)

	// First withdrawal
	firstRequest := testutil.TestRequest{
		Method: http.MethodDelete,
		URL:    "/api/v1/auth/withdraw",
		Body:   nil,
	}
	firstRecorder := testutil.ExecuteRequest(t, router, firstRequest)
	require.Equal(t, http.StatusOK, firstRecorder.Code)

	// When: Try to withdraw again
	secondRequest := testutil.TestRequest{
		Method: http.MethodDelete,
		URL:    "/api/v1/auth/withdraw",
		Body:   nil,
	}
	secondRecorder := testutil.ExecuteRequest(t, router, secondRequest)

	// Then: Verify not found error
	assert.Equal(t, http.StatusNotFound, secondRecorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil.ParseResponse(t, secondRecorder, &errorResponse)
	assert.Equal(t, "MEMBER-001", errorResponse.Code)
}
