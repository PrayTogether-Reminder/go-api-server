package prayer_test

import (
	"fmt"
	"net/http"
	"testing"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/http"
	testutil2 "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/testutil"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestDeletePrayerContent_Success(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)

	// Given: Create a test room with the member
	testRoom := testutil2.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")

	// Given: Create a prayer title
	prayerTitle := testutil2.CreateTestPrayerTitle(t, db, testRoom.ID, "Test Prayer Title")

	// Given: Create a prayer content
	writer := testutil2.CreateTestMemberWithIndex(t, db, 1)
	memberID1 := int64(1)
	prayerContent := testutil2.CreateTestPrayerContent(t, db, prayerTitle.ID, writer.ID, writer.Name, &memberID1, "Member 1", "Test Content")

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.DELETE("/api/v1/prayers/:titleId/contents/:contentId", prayerHandler.DeletePrayerContent)

	// Given: Valid delete prayer content request
	request := testutil2.TestRequest{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/api/v1/prayers/%d/contents/%d", prayerTitle.ID, prayerContent.ID),
	}

	// When: Execute delete prayer content request
	recorder := testutil2.ExecuteRequest(t, router, request)

	// Then: Verify response
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response sharedHttp.MessageResponse
	testutil2.ParseResponse(t, recorder, &response)
	assert.Equal(t, "기도 내용을 삭제했습니다.", response.Message)

	// Then: Verify prayer content was deleted from database
	var deletedContent model.PrayerContent
	err := db.Table("prayer_content").Where("id = ?", prayerContent.ID).First(&deletedContent).Error
	assert.Error(t, err) // Should return error because record was deleted
}

func TestDeletePrayerContent_ContentNotFound(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)

	// Given: Create a test room with the member
	testRoom := testutil2.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")

	// Given: Create a prayer title
	prayerTitle := testutil2.CreateTestPrayerTitle(t, db, testRoom.ID, "Test Prayer Title")

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.DELETE("/api/v1/prayers/:titleId/contents/:contentId", prayerHandler.DeletePrayerContent)

	// Given: Request with non-existent contentId
	nonExistentContentID := int64(99999)
	request := testutil2.TestRequest{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/api/v1/prayers/%d/contents/%d", prayerTitle.ID, nonExistentContentID),
	}

	// When: Execute request
	recorder := testutil2.ExecuteRequest(t, router, request)

	// Then: Verify not found error
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil2.ParseResponse(t, recorder, &errorResponse)
	assert.NotEmpty(t, errorResponse.Message)
}

func TestDeletePrayerContent_TitleNotFound(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, _, memberID := setupTestEnvironment(t)

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.DELETE("/api/v1/prayers/:titleId/contents/:contentId", prayerHandler.DeletePrayerContent)

	// Given: Request with non-existent titleId
	nonExistentTitleID := int64(99999)
	nonExistentContentID := int64(99999)
	request := testutil2.TestRequest{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/api/v1/prayers/%d/contents/%d", nonExistentTitleID, nonExistentContentID),
	}

	// When: Execute request
	recorder := testutil2.ExecuteRequest(t, router, request)

	// Then: Verify not found error
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil2.ParseResponse(t, recorder, &errorResponse)
	assert.NotEmpty(t, errorResponse.Message)
}

func TestDeletePrayerContent_MemberNotInRoom(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)

	// Given: Create another member and a room owned by that member
	anotherMember := testutil2.CreateTestMemberWithIndex(t, db, 1)
	roomOwnedByAnother := testutil2.CreateTestRoom(t, db, anotherMember.ID, "Another Room", "Test Description")
	prayerTitle := testutil2.CreateTestPrayerTitle(t, db, roomOwnedByAnother.ID, "Test Prayer Title")

	memberID1 := int64(1)
	prayerContent := testutil2.CreateTestPrayerContent(t, db, prayerTitle.ID, anotherMember.ID, anotherMember.Name, &memberID1, "Member 1", "Test Content")

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.DELETE("/api/v1/prayers/:titleId/contents/:contentId", prayerHandler.DeletePrayerContent)

	// Given: Request to delete prayer content in a room the member doesn't belong to
	request := testutil2.TestRequest{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/api/v1/prayers/%d/contents/%d", prayerTitle.ID, prayerContent.ID),
	}

	// When: Execute request
	recorder := testutil2.ExecuteRequest(t, router, request)

	// Then: Verify not found error (member not in room)
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil2.ParseResponse(t, recorder, &errorResponse)
	assert.NotEmpty(t, errorResponse.Message)
}

func TestDeletePrayerContent_ValidationError_InvalidIDs(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, _, memberID := setupTestEnvironment(t)

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.DELETE("/api/v1/prayers/:titleId/contents/:contentId", prayerHandler.DeletePrayerContent)

	testCases := []struct {
		name        string
		titleID     string
		contentID   string
		description string
	}{
		{
			name:        "TitleID is zero",
			titleID:     "0",
			contentID:   "1",
			description: "Should fail when titleId is 0",
		},
		{
			name:        "TitleID is negative",
			titleID:     "-1",
			contentID:   "1",
			description: "Should fail when titleId is negative",
		},
		{
			name:        "ContentID is zero",
			titleID:     "1",
			contentID:   "0",
			description: "Should fail when contentId is 0",
		},
		{
			name:        "ContentID is negative",
			titleID:     "1",
			contentID:   "-1",
			description: "Should fail when contentId is negative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given: Request with invalid IDs
			request := testutil2.TestRequest{
				Method: http.MethodDelete,
				URL:    fmt.Sprintf("/api/v1/prayers/%s/contents/%s", tc.titleID, tc.contentID),
			}

			// When: Execute request
			recorder := testutil2.ExecuteRequest(t, router, request)

			// Then: Verify validation error
			assert.Equal(t, http.StatusBadRequest, recorder.Code, tc.description)

			var errorResponse sharedError.ErrorResponse
			testutil2.ParseResponse(t, recorder, &errorResponse)
			assert.NotEmpty(t, errorResponse.Message, tc.description)
		})
	}
}

func TestDeletePrayerContent_ContentNotBelongToTitle(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)

	// Given: Create a test room with the member
	testRoom := testutil2.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")

	// Given: Create two prayer titles
	prayerTitle1 := testutil2.CreateTestPrayerTitle(t, db, testRoom.ID, "Prayer Title 1")
	prayerTitle2 := testutil2.CreateTestPrayerTitle(t, db, testRoom.ID, "Prayer Title 2")

	// Given: Create a prayer content for title 1
	writer := testutil2.CreateTestMemberWithIndex(t, db, 1)
	memberID1 := int64(1)
	prayerContent := testutil2.CreateTestPrayerContent(t, db, prayerTitle1.ID, writer.ID, writer.Name, &memberID1, "Member 1", "Test Content")

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.DELETE("/api/v1/prayers/:titleId/contents/:contentId", prayerHandler.DeletePrayerContent)

	// Given: Request to delete content with wrong titleId (title2 instead of title1)
	request := testutil2.TestRequest{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/api/v1/prayers/%d/contents/%d", prayerTitle2.ID, prayerContent.ID),
	}

	// When: Execute request
	recorder := testutil2.ExecuteRequest(t, router, request)

	// Then: Verify not found error (content doesn't belong to title)
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil2.ParseResponse(t, recorder, &errorResponse)
	assert.NotEmpty(t, errorResponse.Message)
}
