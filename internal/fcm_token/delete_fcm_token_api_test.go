package fcm_token_test

import (
	"net/http"
	"testing"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/testutil"
	fcmToken "github.com/changhyeonkim/pray-together/go-api-server/internal/fcm_token"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteFcmToken_Success(t *testing.T) {
	handler, db, member := setupFcmTokenTestEnvironment(t)
	require.NoError(t, db.Create(&model.FcmToken{MemberID: member.ID, Token: "token-to-delete", IsActive: true}).Error)

	router := testutil.SetupAuthenticatedRouter(member.ID)
	router.DELETE("/api/v1/fcm-token", handler.DeleteToken)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodDelete,
		URL:    "/api/v1/fcm-token",
		Body: fcmToken.DeleteFcmTokenRequest{
			FcmToken: "token-to-delete",
		},
	})

	assert.Equal(t, http.StatusOK, recorder.Code)
	var count int64
	require.NoError(t, db.Model(&model.FcmToken{}).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}

func TestDeleteFcmToken_IgnoredForOtherMember(t *testing.T) {
	handler, db, member := setupFcmTokenTestEnvironment(t)
	otherMember := testutil.CreateTestMemberWithIndex(t, db, 999)
	require.NoError(t, db.Create(&model.FcmToken{MemberID: member.ID, Token: "my-token", IsActive: true}).Error)
	require.NoError(t, db.Create(&model.FcmToken{MemberID: otherMember.ID, Token: "my-token", IsActive: true}).Error)

	router := testutil.SetupAuthenticatedRouter(member.ID)
	router.DELETE("/api/v1/fcm-token", handler.DeleteToken)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodDelete,
		URL:    "/api/v1/fcm-token",
		Body:   fcmToken.DeleteFcmTokenRequest{FcmToken: "my-token"},
	})

	assert.Equal(t, http.StatusOK, recorder.Code)
	var tokens []model.FcmToken
	require.NoError(t, db.Order("member_id").Find(&tokens).Error)
	assert.Len(t, tokens, 1)
	assert.Equal(t, otherMember.ID, tokens[0].MemberID)
}

func TestDeleteFcmToken_ValidationError(t *testing.T) {
	handler, _, member := setupFcmTokenTestEnvironment(t)
	router := testutil.SetupAuthenticatedRouter(member.ID)
	router.DELETE("/api/v1/fcm-token", handler.DeleteToken)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodDelete,
		URL:    "/api/v1/fcm-token",
		Body:   fcmToken.DeleteFcmTokenRequest{FcmToken: "   "},
	})

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDeleteFcmToken_Unauthorized(t *testing.T) {
	handler, _, _ := setupFcmTokenTestEnvironment(t)
	router := testutil.SetupTestRouter()
	router.DELETE("/api/v1/fcm-token", handler.DeleteToken)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodDelete,
		URL:    "/api/v1/fcm-token",
		Body:   fcmToken.DeleteFcmTokenRequest{FcmToken: "token"},
	})

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}
