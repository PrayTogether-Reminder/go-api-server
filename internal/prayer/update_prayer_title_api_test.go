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

func TestUpdatePrayerTitle_Success(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)

	// Given: Create a test room with the member
	testRoom := testutil2.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")

	// Given: Create a prayer title
	prayerTitle := testutil2.CreateTestPrayerTitle(t, db, testRoom.ID, "Original Prayer Title")

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.PUT("/api/v1/prayers/:titleId", prayerHandler.UpdatePrayerTitle)

	// Given: Valid update prayer title request
	request := testutil2.TestRequest{
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/api/v1/prayers/%d", prayerTitle.ID),
		Body: prayer.UpdatePrayerTitleRequest{
			ChangedTitle: "Updated Prayer Title",
		},
	}

	// When: Execute update prayer title request
	recorder := testutil2.ExecuteRequest(t, router, request)

	// Then: Verify response
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response sharedHttp.MessageResponse
	testutil2.ParseResponse(t, recorder, &response)
	assert.Equal(t, "기도 제목을 변경했습니다.", response.Message)

	// Verify prayer title was updated in database
	var updatedPrayerTitle model.PrayerTitle
	err := db.Table("prayer_title").Where("id = ?", prayerTitle.ID).First(&updatedPrayerTitle).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Prayer Title", updatedPrayerTitle.Title)
	assert.Equal(t, testRoom.ID, updatedPrayerTitle.RoomID)
}

func TestUpdatePrayerTitle_ValidationError_MissingTitle(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)

	// Given: Create a test room and prayer title
	testRoom := testutil2.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")
	prayerTitle := testutil2.CreateTestPrayerTitle(t, db, testRoom.ID, "Original Prayer Title")

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.PUT("/api/v1/prayers/:titleId", prayerHandler.UpdatePrayerTitle)

	testCases := []struct {
		name        string
		requestBody map[string]interface{}
		description string
	}{
		{
			name:        "Missing changedTitle field",
			requestBody: map[string]interface{}{},
			description: "Should fail when changedTitle is missing",
		},
		{
			name: "Empty changedTitle",
			requestBody: map[string]interface{}{
				"changedTitle": "",
			},
			description: "Should fail when changedTitle is empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When: Execute request with invalid field
			request := testutil2.TestRequest{
				Method: http.MethodPut,
				URL:    fmt.Sprintf("/api/v1/prayers/%d", prayerTitle.ID),
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

func TestUpdatePrayerTitle_ValidationError_TitleTooLong(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)

	// Given: Create a test room and prayer title
	testRoom := testutil2.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")
	prayerTitle := testutil2.CreateTestPrayerTitle(t, db, testRoom.ID, "Original Prayer Title")

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.PUT("/api/v1/prayers/:titleId", prayerHandler.UpdatePrayerTitle)

	// Given: Title longer than 50 characters
	longTitle := "This is a very long prayer title that exceeds the maximum length of fifty characters"
	request := testutil2.TestRequest{
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/api/v1/prayers/%d", prayerTitle.ID),
		Body: prayer.UpdatePrayerTitleRequest{
			ChangedTitle: longTitle,
		},
	}

	// When: Execute request
	recorder := testutil2.ExecuteRequest(t, router, request)

	// Then: Verify validation error
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil2.ParseResponse(t, recorder, &errorResponse)
	assert.NotEmpty(t, errorResponse.Message)
}

func TestUpdatePrayerTitle_ValidationError_InvalidTitleID(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, _, memberID := setupTestEnvironment(t)

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.PUT("/api/v1/prayers/:titleId", prayerHandler.UpdatePrayerTitle)

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
			request := testutil2.TestRequest{
				Method: http.MethodPut,
				URL:    "/api/v1/prayers/" + tc.titleID,
				Body: prayer.UpdatePrayerTitleRequest{
					ChangedTitle: "Updated Prayer Title",
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

func TestUpdatePrayerTitle_TitleNotFound(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, _, memberID := setupTestEnvironment(t)

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.PUT("/api/v1/prayers/:titleId", prayerHandler.UpdatePrayerTitle)

	// Given: Request with non-existent titleId
	nonExistentTitleID := int64(99999)
	request := testutil2.TestRequest{
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/api/v1/prayers/%d", nonExistentTitleID),
		Body: prayer.UpdatePrayerTitleRequest{
			ChangedTitle: "Updated Prayer Title",
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

func TestUpdatePrayerTitle_MemberNotInRoom(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)

	// Given: Create another member and a room owned by that member
	anotherMember := testutil2.CreateTestMemberWithIndex(t, db, 1)
	roomOwnedByAnother := testutil2.CreateTestRoom(t, db, anotherMember.ID, "Another Room", "Test Description")
	prayerTitle := testutil2.CreateTestPrayerTitle(t, db, roomOwnedByAnother.ID, "Original Prayer Title")

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.PUT("/api/v1/prayers/:titleId", prayerHandler.UpdatePrayerTitle)

	// Given: Request to update prayer title in a room the member doesn't belong to
	request := testutil2.TestRequest{
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/api/v1/prayers/%d", prayerTitle.ID),
		Body: prayer.UpdatePrayerTitleRequest{
			ChangedTitle: "Updated Prayer Title",
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
