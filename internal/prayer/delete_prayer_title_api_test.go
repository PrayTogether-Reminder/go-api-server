package prayer_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/testutil"
	"github.com/stretchr/testify/assert"
)

func TestDeletePrayerTitle_Success(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)

	// Given: Create a test room with the member
	testRoom := testutil.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")

	// Given: Create a prayer title
	prayerTitle := testutil.CreateTestPrayerTitle(t, db, testRoom.ID, "Prayer Title to Delete")

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.DELETE("/api/v1/prayers/:titleId", prayerHandler.DeletePrayerTitle)

	// Given: Valid delete prayer title request
	request := testutil.TestRequest{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/api/v1/prayers/%d", prayerTitle.ID),
	}

	// When: Execute delete prayer title request
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify response
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response sharedHttp.MessageResponse
	testutil.ParseResponse(t, recorder, &response)
	assert.Equal(t, "기도 제목을 삭제했습니다.", response.Message)

	// Then: Verify prayer title was deleted from database
	var deletedPrayerTitle model.PrayerTitle
	err := db.Table("prayer_title").Where("id = ?", prayerTitle.ID).First(&deletedPrayerTitle).Error
	assert.Error(t, err) // Should return error because record was deleted
}

func TestDeletePrayerTitle_TitleNotFound(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, _, memberID := setupTestEnvironment(t)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.DELETE("/api/v1/prayers/:titleId", prayerHandler.DeletePrayerTitle)

	// Given: Request with non-existent titleId
	nonExistentTitleID := int64(99999)
	request := testutil.TestRequest{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/api/v1/prayers/%d", nonExistentTitleID),
	}

	// When: Execute request
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify not found error
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errorResponse)
	assert.NotEmpty(t, errorResponse.Message)
}

func TestDeletePrayerTitle_MemberNotInRoom(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)

	// Given: Create another member and a room owned by that member
	anotherMember := testutil.CreateTestMemberWithIndex(t, db, 1)
	roomOwnedByAnother := testutil.CreateTestRoom(t, db, anotherMember.ID, "Another Room", "Test Description")
	prayerTitle := testutil.CreateTestPrayerTitle(t, db, roomOwnedByAnother.ID, "Prayer Title to Delete")

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.DELETE("/api/v1/prayers/:titleId", prayerHandler.DeletePrayerTitle)

	// Given: Request to delete prayer title in a room the member doesn't belong to
	request := testutil.TestRequest{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/api/v1/prayers/%d", prayerTitle.ID),
	}

	// When: Execute request
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify not found error (member not in room)
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errorResponse)
	assert.NotEmpty(t, errorResponse.Message)
}

func TestDeletePrayerTitle_ValidationError_InvalidTitleID(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, _, memberID := setupTestEnvironment(t)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.DELETE("/api/v1/prayers/:titleId", prayerHandler.DeletePrayerTitle)

	testCases := []struct {
		name        string
		titleID     string
		description string
	}{
		{
			name:        "TitleID is zero",
			titleID:     "0",
			description: "Should fail when titleId is 0",
		},
		{
			name:        "TitleID is negative",
			titleID:     "-1",
			description: "Should fail when titleId is negative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given: Request with invalid titleId
			request := testutil.TestRequest{
				Method: http.MethodDelete,
				URL:    "/api/v1/prayers/" + tc.titleID,
			}

			// When: Execute request
			recorder := testutil.ExecuteRequest(t, router, request)

			// Then: Verify validation error
			assert.Equal(t, http.StatusBadRequest, recorder.Code, tc.description)

			var errorResponse sharedError.ErrorResponse
			testutil.ParseResponse(t, recorder, &errorResponse)
			assert.NotEmpty(t, errorResponse.Message, tc.description)
		})
	}
}

func TestDeletePrayerTitle_CascadeDelete(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)

	// Given: Create a test room with the member
	testRoom := testutil.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")

	// Given: Create a prayer title
	prayerTitle := testutil.CreateTestPrayerTitle(t, db, testRoom.ID, "Prayer Title with Contents")

	// Given: Create prayer contents associated with the prayer title
	writer := testutil.CreateTestMemberWithIndex(t, db, 1)
	memberID1 := int64(1)
	memberID2 := int64(2)
	prayerContent1 := testutil.CreateTestPrayerContent(t, db, prayerTitle.ID, writer.ID, writer.Name, &memberID1, "Member 1", "First Content")
	prayerContent2 := testutil.CreateTestPrayerContent(t, db, prayerTitle.ID, writer.ID, writer.Name, &memberID2, "Member 2", "Second Content")

	// Verify prayer contents exist
	var beforeContents []model.PrayerContent
	err := db.Where("prayer_title_id = ?", prayerTitle.ID).Find(&beforeContents).Error
	assert.NoError(t, err)
	assert.Len(t, beforeContents, 2)
	assert.Equal(t, prayerContent1.ID, beforeContents[0].ID)
	assert.Equal(t, prayerContent2.ID, beforeContents[1].ID)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.DELETE("/api/v1/prayers/:titleId", prayerHandler.DeletePrayerTitle)

	// Given: Valid delete prayer title request
	request := testutil.TestRequest{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/api/v1/prayers/%d", prayerTitle.ID),
	}

	// When: Execute delete prayer title request
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify response
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response sharedHttp.MessageResponse
	testutil.ParseResponse(t, recorder, &response)
	assert.Equal(t, "기도 제목을 삭제했습니다.", response.Message)

	// Then: Verify prayer title was deleted from database
	var deletedPrayerTitle model.PrayerTitle
	err = db.Table("prayer_title").Where("id = ?", prayerTitle.ID).First(&deletedPrayerTitle).Error
	assert.Error(t, err) // Should return error because record was deleted

	// Then: Verify prayer contents were also deleted (CASCADE DELETE)
	var afterContents []model.PrayerContent
	err = db.Where("prayer_title_id = ?", prayerTitle.ID).Find(&afterContents).Error
	assert.NoError(t, err)
	assert.Empty(t, afterContents, "Prayer contents should be deleted by CASCADE DELETE")
}
