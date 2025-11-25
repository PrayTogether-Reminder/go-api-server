package prayer_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/prayer"
	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Helper functions for prayer content tests

// setupPrayerContentRouter sets up a router for prayer content tests
func setupPrayerContentRouter(memberID int64, handler *prayer.PrayerHandler) *gin.Engine {
	router := testutil.SetupAuthenticatedRouter(memberID)
	router.POST("/api/v1/prayers/:titleId/contents", handler.CreatePrayerContent)
	return router
}

func TestCreatePrayerContent_Success(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)
	testRoom := testutil.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")
	prayerTitle := testutil.CreateTestPrayerTitle(t, db, testRoom.ID, "Please pray for my family")
	router := setupPrayerContentRouter(memberID, prayerHandler)

	// Given: Valid create prayer content request
	targetMemberID := int64(12345)
	request := testutil.TestRequest{
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/api/v1/prayers/%d/contents", prayerTitle.ID),
		Body: prayer.CreatePrayerContentRequest{
			MemberID:   &targetMemberID,
			MemberName: "John Doe",
			Content:    "I will pray for you",
		},
	}

	// When: Execute create prayer content request
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify response
	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response sharedHttp.MessageResponse
	testutil.ParseResponse(t, recorder, &response)
	assert.Equal(t, "기도 내용을 생성했습니다.", response.Message)

	// Verify prayer content was created in database
	var createdPrayerContent model.PrayerContent
	err := db.Table("prayer_content").
		Where("prayer_title_id = ? AND writer_id = ?", prayerTitle.ID, memberID).
		First(&createdPrayerContent).Error
	assert.NoError(t, err)
	assert.Equal(t, prayerTitle.ID, createdPrayerContent.PrayerTitleID)
	assert.Equal(t, memberID, createdPrayerContent.WriterID)
	assert.NotEmpty(t, createdPrayerContent.WriterName)
	assert.Equal(t, &targetMemberID, createdPrayerContent.MemberID)
	assert.Equal(t, "John Doe", createdPrayerContent.MemberName)
	assert.Equal(t, "I will pray for you", createdPrayerContent.Content)
}

func TestCreatePrayerContent_Success_WithoutMemberID(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)
	testRoom := testutil.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")
	prayerTitle := testutil.CreateTestPrayerTitle(t, db, testRoom.ID, "Please pray for my family")
	router := setupPrayerContentRouter(memberID, prayerHandler)

	// Given: Request without memberID (optional field)
	request := testutil.TestRequest{
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/api/v1/prayers/%d/contents", prayerTitle.ID),
		Body: prayer.CreatePrayerContentRequest{
			MemberID:   nil,
			MemberName: "John Doe",
			Content:    "I will pray for you",
		},
	}

	// When: Execute create prayer content request
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify response
	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response sharedHttp.MessageResponse
	testutil.ParseResponse(t, recorder, &response)
	assert.Equal(t, "기도 내용을 생성했습니다.", response.Message)

	// Verify prayer content was created with nil memberID
	var createdPrayerContent model.PrayerContent
	err := db.Table("prayer_content").
		Where("prayer_title_id = ? AND writer_id = ?", prayerTitle.ID, memberID).
		First(&createdPrayerContent).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, createdPrayerContent.WriterName)
	assert.Nil(t, createdPrayerContent.MemberID)
	assert.Equal(t, "John Doe", createdPrayerContent.MemberName)
}

func TestCreatePrayerContent_ValidationError_MissingRequiredFields(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)
	testRoom := testutil.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")
	prayerTitle := testutil.CreateTestPrayerTitle(t, db, testRoom.ID, "Please pray for my family")
	router := setupPrayerContentRouter(memberID, prayerHandler)

	testCases := []struct {
		name        string
		requestBody map[string]interface{}
		description string
	}{
		{
			name: "Missing memberName",
			requestBody: map[string]interface{}{
				"content": "I will pray for you",
			},
			description: "Should fail when memberName is missing",
		},
		{
			name: "Empty memberName",
			requestBody: map[string]interface{}{
				"memberName": "",
				"content":    "I will pray for you",
			},
			description: "Should fail when memberName is empty",
		},
		{
			name: "Missing content",
			requestBody: map[string]interface{}{
				"memberName": "John Doe",
			},
			description: "Should fail when content is missing",
		},
		{
			name: "Empty content",
			requestBody: map[string]interface{}{
				"memberName": "John Doe",
				"content":    "",
			},
			description: "Should fail when content is empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When: Execute request with invalid field
			request := testutil.TestRequest{
				Method: http.MethodPost,
				URL:    fmt.Sprintf("/api/v1/prayers/%d/contents", prayerTitle.ID),
				Body:   tc.requestBody,
			}

			recorder := testutil.ExecuteRequest(t, router, request)

			// Then: Verify validation error
			assert.Equal(t, http.StatusBadRequest, recorder.Code, tc.description)

			var errorResponse sharedError.ErrorResponse
			testutil.ParseResponse(t, recorder, &errorResponse)
			assert.NotEmpty(t, errorResponse.Status, tc.description)
			assert.NotEmpty(t, errorResponse.Message, tc.description)
			assert.NotEmpty(t, errorResponse.Code, tc.description)
		})
	}
}

func TestCreatePrayerContent_ValidationError_InvalidTitleID(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, _, memberID := setupTestEnvironment(t)
	router := setupPrayerContentRouter(memberID, prayerHandler)

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
		{
			name:        "TitleID is not a number",
			titleID:     "abc",
			description: "Should fail when titleId is not a number",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given: Request with invalid titleId
			request := testutil.TestRequest{
				Method: http.MethodPost,
				URL:    fmt.Sprintf("/api/v1/prayers/%s/contents", tc.titleID),
				Body: prayer.CreatePrayerContentRequest{
					MemberName: "John Doe",
					Content:    "I will pray for you",
				},
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

func TestCreatePrayerContent_PrayerTitleNotFound(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, _, memberID := setupTestEnvironment(t)
	router := setupPrayerContentRouter(memberID, prayerHandler)

	// Given: Request with non-existent titleId
	nonExistentTitleID := int64(99999)
	request := testutil.TestRequest{
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/api/v1/prayers/%d/contents", nonExistentTitleID),
		Body: prayer.CreatePrayerContentRequest{
			MemberName: "John Doe",
			Content:    "I will pray for you",
		},
	}

	// When: Execute request
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify not found error
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errorResponse)
	assert.NotEmpty(t, errorResponse.Message)
	assert.Equal(t, "PRAYER-001", errorResponse.Code)
}

func TestCreatePrayerContent_MemberNotInRoom(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)
	anotherMember := testutil.CreateTestMemberWithIndex(t, db, 1)
	roomOwnedByAnother := testutil.CreateTestRoom(t, db, anotherMember.ID, "Another Room", "Test Description")
	prayerTitle := testutil.CreateTestPrayerTitle(t, db, roomOwnedByAnother.ID, "Please pray for my family")
	router := setupPrayerContentRouter(memberID, prayerHandler)

	// Given: Request to create prayer content in a room the member doesn't belong to
	request := testutil.TestRequest{
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/api/v1/prayers/%d/contents", prayerTitle.ID),
		Body: prayer.CreatePrayerContentRequest{
			MemberName: "John Doe",
			Content:    "I will pray for you",
		},
	}

	// When: Execute request
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify not found error (member not in room)
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errorResponse)
	assert.NotEmpty(t, errorResponse.Message)
}
