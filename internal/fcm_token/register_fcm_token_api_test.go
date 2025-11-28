package fcm_token_test

import (
	"net/http"
	"testing"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/error"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/testutil"
	fcmtoken "github.com/changhyeonkim/pray-together/go-api-server/internal/fcm_token"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterFcmToken_ReplacesExistingToken(t *testing.T) {
	handler, db, member := setupFcmTokenTestEnvironment(t)

	router := testutil.SetupAuthenticatedRouter(member.ID)
	router.POST("/api/v1/fcm-token", handler.RegisterToken)

	// First registration
	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/fcm-token",
		Body: fcmtoken.RegisterFcmTokenRequest{
			FcmToken: "first-token",
		},
	})
	assert.Equal(t, http.StatusOK, recorder.Code)

	var tokens []model.FcmToken
	require.NoError(t, db.Find(&tokens).Error)
	require.Len(t, tokens, 1)
	assert.Equal(t, member.ID, tokens[0].MemberID)
	assert.Equal(t, "first-token", tokens[0].Token)
	assert.True(t, tokens[0].IsActive)

	// Second registration should replace existing token and trim whitespace
	recorder = testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/fcm-token",
		Body: fcmtoken.RegisterFcmTokenRequest{
			FcmToken: "   second-token   ",
		},
	})
	assert.Equal(t, http.StatusOK, recorder.Code)

	tokens = nil
	require.NoError(t, db.Find(&tokens).Error)
	require.Len(t, tokens, 1)
	assert.Equal(t, "second-token", tokens[0].Token)
}

func TestRegisterFcmToken_ValidationError(t *testing.T) {
	handler, db, member := setupFcmTokenTestEnvironment(t)

	router := testutil.SetupAuthenticatedRouter(member.ID)
	router.POST("/api/v1/fcm-token", handler.RegisterToken)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/fcm-token",
		Body: fcmtoken.RegisterFcmTokenRequest{
			FcmToken: "   ",
		},
	})

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	var errResp sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errResp)
	assert.Equal(t, "값이 비어 있습니다.", errResp.Message)

	var count int64
	require.NoError(t, db.Model(&model.FcmToken{}).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}

func TestRegisterFcmToken_MemberNotFound(t *testing.T) {
	handler, _, member := setupFcmTokenTestEnvironment(t)

	missingMemberID := member.ID + 99999
	router := testutil.SetupAuthenticatedRouter(missingMemberID)
	router.POST("/api/v1/fcm-token", handler.RegisterToken)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/fcm-token",
		Body: fcmtoken.RegisterFcmTokenRequest{
			FcmToken: "token-value",
		},
	})

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	var errResp sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errResp)
	assert.Equal(t, "MEMBER-001", errResp.Code)
}

func TestRegisterFcmToken_Unauthorized(t *testing.T) {
	handler, _, _ := setupFcmTokenTestEnvironment(t)

	router := testutil.SetupTestRouter()
	router.POST("/api/v1/fcm-token", handler.RegisterToken)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/fcm-token",
		Body: fcmtoken.RegisterFcmTokenRequest{
			FcmToken: "token-value",
		},
	})

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}
