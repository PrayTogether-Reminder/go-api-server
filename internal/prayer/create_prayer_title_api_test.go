package prayer_test

import (
	"net/http"
	"testing"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/error"
	testutil2 "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/testutil"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/prayer"
	"github.com/stretchr/testify/assert"
)

func TestCreatePrayerTitle_Success(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)

	// Given: Create a test room with the member
	testRoom := testutil2.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.POST("/api/v1/prayers", prayerHandler.CreatePrayerTitle)

	// Given: Valid create prayer title request
	request := testutil2.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/prayers",
		Body: prayer.CreatePrayerTitleRequest{
			RoomID: testRoom.ID,
			Title:  "Please pray for my family",
		},
	}

	// When: Execute create prayer title request
	recorder := testutil2.ExecuteRequest(t, router, request)

	// Then: Verify response
	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response prayer.CreatePrayerTitleResponse
	testutil2.ParseResponse(t, recorder, &response)
	assert.NotEmpty(t, response.ID)
	assert.Equal(t, "Please pray for my family", response.Title)
	assert.NotEmpty(t, response.CreatedTime)

	// Verify prayer title was created in database
	var createdPrayerTitle model.PrayerTitle
	err := db.Table("prayer_title").Where("id = ?", response.ID).First(&createdPrayerTitle).Error
	assert.NoError(t, err)
	assert.Equal(t, testRoom.ID, createdPrayerTitle.RoomID)
	assert.Equal(t, "Please pray for my family", createdPrayerTitle.Title)
}

func TestCreatePrayerTitle_ValidationError_MissingRequiredFields(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)

	// Given: Create a test room
	testRoom := testutil2.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.POST("/api/v1/prayers", prayerHandler.CreatePrayerTitle)

	testCases := []struct {
		name        string
		requestBody map[string]interface{}
		description string
	}{
		{
			name: "Missing title",
			requestBody: map[string]interface{}{
				"roomId": testRoom.ID,
			},
			description: "Should fail when title is missing",
		},
		{
			name: "Empty title",
			requestBody: map[string]interface{}{
				"roomId": testRoom.ID,
				"title":  "",
			},
			description: "Should fail when title is empty",
		},
		{
			name: "Missing roomId",
			requestBody: map[string]interface{}{
				"title": "Please pray for my family",
			},
			description: "Should fail when roomId is missing",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When: Execute request with invalid field
			request := testutil2.TestRequest{
				Method: http.MethodPost,
				URL:    "/api/v1/prayers",
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

func TestCreatePrayerTitle_ValidationError_InvalidRoomID(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, _, memberID := setupTestEnvironment(t)

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.POST("/api/v1/prayers", prayerHandler.CreatePrayerTitle)

	testCases := []struct {
		name        string
		roomID      int64
		description string
	}{
		{
			name:        "RoomID is zero",
			roomID:      0,
			description: "Should fail when roomId is 0",
		},
		{
			name:        "RoomID is negative",
			roomID:      -1,
			description: "Should fail when roomId is negative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given: Request with invalid roomId
			request := testutil2.TestRequest{
				Method: http.MethodPost,
				URL:    "/api/v1/prayers",
				Body: prayer.CreatePrayerTitleRequest{
					RoomID: tc.roomID,
					Title:  "Please pray for my family",
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

func TestCreatePrayerTitle_RoomNotFound(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, _, memberID := setupTestEnvironment(t)

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.POST("/api/v1/prayers", prayerHandler.CreatePrayerTitle)

	// Given: Request with non-existent roomId
	nonExistentRoomID := int64(99999)
	request := testutil2.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/prayers",
		Body: prayer.CreatePrayerTitleRequest{
			RoomID: nonExistentRoomID,
			Title:  "Please pray for my family",
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

func TestCreatePrayerTitle_MemberNotInRoom(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)

	// Given: Create another member and a room owned by that member
	anotherMember := testutil2.CreateTestMemberWithIndex(t, db, 1)
	roomOwnedByAnother := testutil2.CreateTestRoom(t, db, anotherMember.ID, "Another Room", "Test Description")

	router := testutil2.SetupAuthenticatedRouter(memberID)
	router.POST("/api/v1/prayers", prayerHandler.CreatePrayerTitle)

	// Given: Request to create prayer title in a room the member doesn't belong to
	request := testutil2.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/prayers",
		Body: prayer.CreatePrayerTitleRequest{
			RoomID: roomOwnedByAnother.ID,
			Title:  "Please pray for my family",
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
