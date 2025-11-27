package prayer_test

import (
	"fmt"
	"net/http"
	"testing"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/http"
	testutil2 "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/testutil"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/prayer"
	"github.com/stretchr/testify/assert"
)

func TestUpdatePrayerContent_Success(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)

	// Given: Create a test room with the member
	testRoom := testutil2.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")

	// Given: Create a prayer title and content
	prayerTitle := testutil2.CreateTestPrayerTitle(t, db, testRoom.ID, "Prayer Title")
	testMember := testutil2.CreateTestMemberWithIndex(t, db, 1)
	prayerContent := testutil2.CreateTestPrayerContent(
		t, db, prayerTitle.ID, memberID, "Writer Name", &testMember.ID, "Member Name", "Original Content",
	)

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.PUT("/api/v1/prayers/:titleId/contents/:contentId", prayerHandler.UpdatePrayerContent)

	// Given: Valid update prayer content request
	request := testutil2.TestRequest{
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/api/v1/prayers/%d/contents/%d", prayerTitle.ID, prayerContent.ID),
		Body: prayer.UpdatePrayerContentRequest{
			ChangedContent: "Updated Content",
		},
	}

	// When: Execute update prayer content request
	recorder := testutil2.ExecuteRequest(t, router, request)

	// Then: Verify response
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response sharedHttp.MessageResponse
	testutil2.ParseResponse(t, recorder, &response)
	assert.Equal(t, "기도 내용을 변경했습니다.", response.Message)

	// Verify prayer content was updated in database
	var updatedPrayerContent model.PrayerContent
	err := db.Table("prayer_content").Where("id = ?", prayerContent.ID).First(&updatedPrayerContent).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Content", updatedPrayerContent.Content)
	assert.Equal(t, prayerTitle.ID, updatedPrayerContent.PrayerTitleID)
}

func TestUpdatePrayerContent_ValidationError_MissingContent(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)

	// Given: Create a test room, prayer title, and content
	testRoom := testutil2.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")
	prayerTitle := testutil2.CreateTestPrayerTitle(t, db, testRoom.ID, "Prayer Title")
	testMember := testutil2.CreateTestMemberWithIndex(t, db, 1)
	prayerContent := testutil2.CreateTestPrayerContent(
		t, db, prayerTitle.ID, memberID, "Writer Name", &testMember.ID, "Member Name", "Original Content",
	)

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.PUT("/api/v1/prayers/:titleId/contents/:contentId", prayerHandler.UpdatePrayerContent)

	testCases := []struct {
		name        string
		requestBody map[string]interface{}
		description string
	}{
		{
			name:        "Missing changedContent field",
			requestBody: map[string]interface{}{},
			description: "Should fail when changedContent is missing",
		},
		{
			name: "Empty changedContent",
			requestBody: map[string]interface{}{
				"changedContent": "",
			},
			description: "Should fail when changedContent is empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When: Execute request with invalid field
			request := testutil2.TestRequest{
				Method: http.MethodPut,
				URL:    fmt.Sprintf("/api/v1/prayers/%d/contents/%d", prayerTitle.ID, prayerContent.ID),
				Body:   tc.requestBody,
			}

			recorder := testutil2.ExecuteRequest(t, router, request)

			// Then: Verify validation error
			assert.Equal(t, http.StatusBadRequest, recorder.Code, tc.description)

			var errorResponse sharedError.ErrorResponse
			testutil2.ParseResponse(t, recorder, &errorResponse)
			assert.NotEmpty(t, errorResponse.Status, tc.description)
			assert.NotEmpty(t, errorResponse.Message, tc.description)
			assert.NotEmpty(t, errorResponse.Code, tc.description)
		})
	}
}

func TestUpdatePrayerContent_ValidationError_InvalidIDs(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, _, memberID := setupTestEnvironment(t)

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.PUT("/api/v1/prayers/:titleId/contents/:contentId", prayerHandler.UpdatePrayerContent)

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
				Method: http.MethodPut,
				URL:    fmt.Sprintf("/api/v1/prayers/%s/contents/%s", tc.titleID, tc.contentID),
				Body: prayer.UpdatePrayerContentRequest{
					ChangedContent: "Updated Content",
				},
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

func TestUpdatePrayerContent_TitleNotFound(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, _, memberID := setupTestEnvironment(t)

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.PUT("/api/v1/prayers/:titleId/contents/:contentId", prayerHandler.UpdatePrayerContent)

	// Given: Request with non-existent titleId
	nonExistentTitleID := int64(99999)
	nonExistentContentID := int64(99999)
	request := testutil2.TestRequest{
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/api/v1/prayers/%d/contents/%d", nonExistentTitleID, nonExistentContentID),
		Body: prayer.UpdatePrayerContentRequest{
			ChangedContent: "Updated Content",
		},
	}

	// When: Execute request
	recorder := testutil2.ExecuteRequest(t, router, request)

	// Then: Verify not found error
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil2.ParseResponse(t, recorder, &errorResponse)
	assert.NotEmpty(t, errorResponse.Message)
}

func TestUpdatePrayerContent_ContentNotFound(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)

	// Given: Create a test room and prayer title (but no content)
	testRoom := testutil2.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")
	prayerTitle := testutil2.CreateTestPrayerTitle(t, db, testRoom.ID, "Prayer Title")

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.PUT("/api/v1/prayers/:titleId/contents/:contentId", prayerHandler.UpdatePrayerContent)

	// Given: Request with non-existent contentId
	nonExistentContentID := int64(99999)
	request := testutil2.TestRequest{
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/api/v1/prayers/%d/contents/%d", prayerTitle.ID, nonExistentContentID),
		Body: prayer.UpdatePrayerContentRequest{
			ChangedContent: "Updated Content",
		},
	}

	// When: Execute request
	recorder := testutil2.ExecuteRequest(t, router, request)

	// Then: Verify not found error
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil2.ParseResponse(t, recorder, &errorResponse)
	assert.NotEmpty(t, errorResponse.Message)
}

func TestUpdatePrayerContent_ContentNotBelongToTitle(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)

	// Given: Create two different prayer titles with contents
	testRoom := testutil2.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")
	prayerTitle1 := testutil2.CreateTestPrayerTitle(t, db, testRoom.ID, "Prayer Title 1")
	prayerTitle2 := testutil2.CreateTestPrayerTitle(t, db, testRoom.ID, "Prayer Title 2")
	testMember := testutil2.CreateTestMemberWithIndex(t, db, 1)

	// Create content under prayerTitle1
	prayerContent1 := testutil2.CreateTestPrayerContent(
		t, db, prayerTitle1.ID, memberID, "Writer Name", &testMember.ID, "Member Name", "Content 1",
	)

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.PUT("/api/v1/prayers/:titleId/contents/:contentId", prayerHandler.UpdatePrayerContent)

	// Given: Request to update content1 with wrong titleId (prayerTitle2)
	request := testutil2.TestRequest{
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/api/v1/prayers/%d/contents/%d", prayerTitle2.ID, prayerContent1.ID),
		Body: prayer.UpdatePrayerContentRequest{
			ChangedContent: "Updated Content",
		},
	}

	// When: Execute request
	recorder := testutil2.ExecuteRequest(t, router, request)

	// Then: Verify not found error (content doesn't belong to title)
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil2.ParseResponse(t, recorder, &errorResponse)
	assert.NotEmpty(t, errorResponse.Message)

	// Verify original content is unchanged
	var unchangedContent model.PrayerContent
	err := db.Table("prayer_content").Where("id = ?", prayerContent1.ID).First(&unchangedContent).Error
	assert.NoError(t, err)
	assert.Equal(t, "Content 1", unchangedContent.Content)
}

func TestUpdatePrayerContent_MemberNotInRoom(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)

	// Given: Create another member and a room owned by that member
	anotherMember := testutil2.CreateTestMemberWithIndex(t, db, 1)
	roomOwnedByAnother := testutil2.CreateTestRoom(t, db, anotherMember.ID, "Another Room", "Test Description")
	prayerTitle := testutil2.CreateTestPrayerTitle(t, db, roomOwnedByAnother.ID, "Prayer Title")
	testMember := testutil2.CreateTestMemberWithIndex(t, db, 2)
	prayerContent := testutil2.CreateTestPrayerContent(
		t, db, prayerTitle.ID, anotherMember.ID, "Writer Name", &testMember.ID, "Member Name", "Original Content",
	)

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.PUT("/api/v1/prayers/:titleId/contents/:contentId", prayerHandler.UpdatePrayerContent)

	// Given: Request to update prayer content in a room the member doesn't belong to
	request := testutil2.TestRequest{
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/api/v1/prayers/%d/contents/%d", prayerTitle.ID, prayerContent.ID),
		Body: prayer.UpdatePrayerContentRequest{
			ChangedContent: "Updated Content",
		},
	}

	// When: Execute request
	recorder := testutil2.ExecuteRequest(t, router, request)

	// Then: Verify not found error (member not in room)
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil2.ParseResponse(t, recorder, &errorResponse)
	assert.NotEmpty(t, errorResponse.Message)
}
