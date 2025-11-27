package member_test

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
	"gorm.io/gorm"
)

func setupMemberTestEnvironment(t *testing.T) (*member.MemberHandler, *gorm.DB) {
	t.Helper()

	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestDB(t, db)
	})

	memberRepo := member.NewMemberRepository()
	memberService := member.NewMemberService(memberRepo)
	memberUseCase := member.NewMemberUseCase(db, memberService)
	memberHandler := member.NewMemberHandler(memberUseCase)

	return memberHandler, db
}

func TestUpdateProfile_Success(t *testing.T) {
	memberHandler, db := setupMemberTestEnvironment(t)
	createdMember := testutil.CreateTestMember(t, db)

	router := testutil.SetupAuthenticatedRouter(createdMember.ID)
	router.PATCH("/api/v1/members/me", memberHandler.UpdateProfile)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodPatch,
		URL:    "/api/v1/members/me",
		Body: map[string]string{
			"name":        "새이름",
			"phoneNumber": "01099998888",
		},
	})

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response sharedHttp.MessageResponse
	testutil.ParseResponse(t, recorder, &response)
	assert.Equal(t, "프로필을 변경했습니다.", response.Message)

	memberRepo := member.NewMemberRepository()
	updated, err := memberRepo.FindByID(context.Background(), db, createdMember.ID)
	require.NoError(t, err)
	assert.Equal(t, "새이름", updated.Name)
	assert.Equal(t, "010-9999-8888", updated.PhoneNumber)
}

func TestUpdateProfile_ValidationError(t *testing.T) {
	memberHandler, db := setupMemberTestEnvironment(t)
	createdMember := testutil.CreateTestMember(t, db)

	router := testutil.SetupAuthenticatedRouter(createdMember.ID)
	router.PATCH("/api/v1/members/me", memberHandler.UpdateProfile)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodPatch,
		URL:    "/api/v1/members/me",
		Body: map[string]string{
			"name": "",
		},
	})

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errorResponse)
	assert.Equal(t, "ERROR-001", errorResponse.Code)
}

func TestUpdateProfile_Unauthorized(t *testing.T) {
	memberHandler, _ := setupMemberTestEnvironment(t)
	router := testutil.SetupTestRouter()
	router.PATCH("/api/v1/members/me", memberHandler.UpdateProfile)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodPatch,
		URL:    "/api/v1/members/me",
	})

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestUpdateProfile_NotFound(t *testing.T) {
	memberHandler, _ := setupMemberTestEnvironment(t)
	const nonexistentID int64 = 99999

	router := testutil.SetupAuthenticatedRouter(nonexistentID)
	router.PATCH("/api/v1/members/me", memberHandler.UpdateProfile)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodPatch,
		URL:    "/api/v1/members/me",
		Body: map[string]string{
			"name":        "테스트",
			"phoneNumber": "01012345678",
		},
	})

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errorResponse)
	assert.Equal(t, "MEMBER-001", errorResponse.Code)
}
