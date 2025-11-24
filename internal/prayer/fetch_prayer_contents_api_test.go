package prayer_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/prayer"
	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupPrayerContentFetchRouter sets up a router for fetching prayer contents
func setupPrayerContentFetchRouter(memberID int64, handler *prayer.PrayerHandler) *gin.Engine {
	router := testutil.SetupAuthenticatedRouter(memberID)
	router.GET("/api/v1/prayers/:titleId/contents", handler.FetchPrayerContents)
	return router
}

func TestFetchPrayerContents_Success(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)
	testRoom := testutil.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")
	prayerTitle := testutil.CreateTestPrayerTitle(t, db, testRoom.ID, "Please pray for my family")
	router := setupPrayerContentFetchRouter(memberID, prayerHandler)

	// Given: Create test prayer contents
	member1 := testutil.CreateTestMemberWithIndex(t, db, 1)
	member2 := testutil.CreateTestMemberWithIndex(t, db, 2)
	targetMemberID := int64(12345)

	content1 := testutil.CreateTestPrayerContent(
		t, db, prayerTitle.ID, memberID, "Test_User_",
		&targetMemberID, "John Doe", "First prayer content",
	)
	content2 := testutil.CreateTestPrayerContent(
		t, db, prayerTitle.ID, member1.ID, member1.Name,
		nil, "Jane Smith", "Second prayer content",
	)
	content3 := testutil.CreateTestPrayerContent(
		t, db, prayerTitle.ID, member2.ID, member2.Name,
		&targetMemberID, "Bob Johnson", "Third prayer content",
	)

	// When: Execute fetch prayer contents request
	request := testutil.TestRequest{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/api/v1/prayers/%d/contents", prayerTitle.ID),
	}
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify response
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response prayer.PrayerContentResponse
	testutil.ParseResponse(t, recorder, &response)

	// Verify all contents are returned
	assert.Len(t, response.PrayerContents, 3)

	// Verify first content
	assert.Equal(t, content1.ID, response.PrayerContents[0].ID)
	assert.Equal(t, memberID, response.PrayerContents[0].WriterID)
	assert.Equal(t, "Test_User_", response.PrayerContents[0].WriterName)
	assert.Equal(t, &targetMemberID, response.PrayerContents[0].MemberID)
	assert.Equal(t, "John Doe", response.PrayerContents[0].MemberName)
	assert.Equal(t, "First prayer content", response.PrayerContents[0].Content)

	// Verify second content
	assert.Equal(t, content2.ID, response.PrayerContents[1].ID)
	assert.Equal(t, member1.ID, response.PrayerContents[1].WriterID)
	assert.Equal(t, member1.Name, response.PrayerContents[1].WriterName)
	assert.Nil(t, response.PrayerContents[1].MemberID)
	assert.Equal(t, "Jane Smith", response.PrayerContents[1].MemberName)
	assert.Equal(t, "Second prayer content", response.PrayerContents[1].Content)

	// Verify third content
	assert.Equal(t, content3.ID, response.PrayerContents[2].ID)
	assert.Equal(t, member2.ID, response.PrayerContents[2].WriterID)
	assert.Equal(t, member2.Name, response.PrayerContents[2].WriterName)
	assert.Equal(t, &targetMemberID, response.PrayerContents[2].MemberID)
	assert.Equal(t, "Bob Johnson", response.PrayerContents[2].MemberName)
	assert.Equal(t, "Third prayer content", response.PrayerContents[2].Content)
}

func TestFetchPrayerContents_EmptyList(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)
	testRoom := testutil.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")
	prayerTitle := testutil.CreateTestPrayerTitle(t, db, testRoom.ID, "Please pray for my family")
	router := setupPrayerContentFetchRouter(memberID, prayerHandler)

	// Given: No prayer contents created

	// When: Execute fetch prayer contents request
	request := testutil.TestRequest{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/api/v1/prayers/%d/contents", prayerTitle.ID),
	}
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify response returns empty array
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response prayer.PrayerContentResponse
	testutil.ParseResponse(t, recorder, &response)
	assert.NotNil(t, response.PrayerContents)
	assert.Len(t, response.PrayerContents, 0)
}

func TestFetchPrayerContents_ValidationError_InvalidTitleID(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, _, memberID := setupTestEnvironment(t)
	router := setupPrayerContentFetchRouter(memberID, prayerHandler)

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
			// When: Execute request with invalid titleId
			request := testutil.TestRequest{
				Method: http.MethodGet,
				URL:    fmt.Sprintf("/api/v1/prayers/%s/contents", tc.titleID),
			}
			recorder := testutil.ExecuteRequest(t, router, request)

			// Then: Verify validation error
			assert.Equal(t, http.StatusBadRequest, recorder.Code, tc.description)

			var errorResponse sharedError.ErrorResponse
			testutil.ParseResponse(t, recorder, &errorResponse)
			assert.NotEmpty(t, errorResponse.Message, tc.description)
		})
	}
}

func TestFetchPrayerContents_PrayerTitleNotFound(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, _, memberID := setupTestEnvironment(t)
	router := setupPrayerContentFetchRouter(memberID, prayerHandler)

	// Given: Request with non-existent titleId
	nonExistentTitleID := int64(99999)
	request := testutil.TestRequest{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/api/v1/prayers/%d/contents", nonExistentTitleID),
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

func TestFetchPrayerContents_MemberNotInRoom(t *testing.T) {
	// Given: Setup test environment with two members
	prayerHandler, db, memberID := setupTestEnvironment(t)
	anotherMember := testutil.CreateTestMemberWithIndex(t, db, 1)
	roomOwnedByAnother := testutil.CreateTestRoom(t, db, anotherMember.ID, "Another Room", "Test Description")
	prayerTitle := testutil.CreateTestPrayerTitle(t, db, roomOwnedByAnother.ID, "Please pray for my family")
	router := setupPrayerContentFetchRouter(memberID, prayerHandler)

	// Given: Create prayer content in the other member's room
	testutil.CreateTestPrayerContent(
		t, db, prayerTitle.ID, anotherMember.ID, anotherMember.Name,
		nil, "John Doe", "Prayer content",
	)

	// When: First member tries to fetch contents from other member's room
	request := testutil.TestRequest{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/api/v1/prayers/%d/contents", prayerTitle.ID),
	}
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify access denied (member not in room)
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errorResponse)
	assert.NotEmpty(t, errorResponse.Message)
}

func TestFetchPrayerContents_OrderedByCreatedTime(t *testing.T) {
	// Given: Setup test environment
	prayerHandler, db, memberID := setupTestEnvironment(t)
	testRoom := testutil.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")
	prayerTitle := testutil.CreateTestPrayerTitle(t, db, testRoom.ID, "Please pray for my family")
	router := setupPrayerContentFetchRouter(memberID, prayerHandler)

	// Given: Create multiple prayer contents
	content1 := testutil.CreateTestPrayerContent(
		t, db, prayerTitle.ID, memberID, "Test_User_",
		nil, "Member 1", "First content",
	)
	content2 := testutil.CreateTestPrayerContent(
		t, db, prayerTitle.ID, memberID, "Test_User_",
		nil, "Member 2", "Second content",
	)
	content3 := testutil.CreateTestPrayerContent(
		t, db, prayerTitle.ID, memberID, "Test_User_",
		nil, "Member 3", "Third content",
	)

	// When: Execute fetch prayer contents request
	request := testutil.TestRequest{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/api/v1/prayers/%d/contents", prayerTitle.ID),
	}
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify contents are ordered by created_time ASC
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response prayer.PrayerContentResponse
	testutil.ParseResponse(t, recorder, &response)

	assert.Len(t, response.PrayerContents, 3)
	assert.Equal(t, content1.ID, response.PrayerContents[0].ID)
	assert.Equal(t, content2.ID, response.PrayerContents[1].ID)
	assert.Equal(t, content3.ID, response.PrayerContents[2].ID)
}
