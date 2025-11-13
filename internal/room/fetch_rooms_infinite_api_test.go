package room_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/room"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/testutil"
	"github.com/stretchr/testify/assert"
)

func TestFetchRoomsInfiniteScroll_Success_FirstPage(t *testing.T) {
	// Given: Setup test environment
	roomHandler, db, memberID := setupTestEnvironment(t)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.GET("/api/v1/rooms", roomHandler.GetRoomsByInfiniteScroll)

	// Given: Create 15 test rooms for pagination test
	testutil.CreateTestRooms(t, db, memberID, 15)

	// When: Request first page with after=0
	request := testutil.TestRequest{
		Method: http.MethodGet,
		URL:    "/api/v1/rooms?orderBy=time&after=0&dir=desc",
	}

	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify response
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response room.InfiniteScrollResponse
	testutil.ParseResponse(t, recorder, &response)

	// Should return page size (10) items
	assert.Len(t, response.Rooms, room.InfiniteScrollPageSize)

	// Verify rooms are ordered by joinedTime DESC
	for i := 0; i < len(response.Rooms)-1; i++ {
		assert.True(t, response.Rooms[i].JoinedTime.After(response.Rooms[i+1].JoinedTime) ||
			response.Rooms[i].JoinedTime.Equal(response.Rooms[i+1].JoinedTime),
			"방 조회는 반드시 시간 최신순서 이어야 합니다 ordered by joinedTime DESC")
	}

	// Verify all rooms have member count
	for _, rm := range response.Rooms {
		assert.NotEmpty(t, rm.ID)
		assert.NotEmpty(t, rm.Name)
		assert.GreaterOrEqual(t, rm.MemberCount, 1, "Each room should have at least 1 member")
	}
}

func TestFetchRoomsInfiniteScroll_Success_MultiplePages(t *testing.T) {
	// Given: Setup test environment
	roomHandler, db, memberID := setupTestEnvironment(t)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.GET("/api/v1/rooms", roomHandler.GetRoomsByInfiniteScroll)

	// Given: Create 15 test rooms for pagination test
	testutil.CreateTestRooms(t, db, memberID, 15)

	// When: Request first page with after=0
	firstRequest := testutil.TestRequest{
		Method: http.MethodGet,
		URL:    "/api/v1/rooms?orderBy=time&after=0&dir=desc",
	}

	firstRecorder := testutil.ExecuteRequest(t, router, firstRequest)

	// Then: Verify first page response
	assert.Equal(t, http.StatusOK, firstRecorder.Code)

	var firstResponse room.InfiniteScrollResponse
	testutil.ParseResponse(t, firstRecorder, &firstResponse)
	assert.Len(t, firstResponse.Rooms, 10, "First page should have 10 items")

	// When: Request second page using last item's joinedTime as cursor
	lastJoinedTime := firstResponse.Rooms[len(firstResponse.Rooms)-1].JoinedTime
	afterCursor := url.QueryEscape(lastJoinedTime.Format(time.RFC3339))

	secondRequest := testutil.TestRequest{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/api/v1/rooms?orderBy=time&after=%s&dir=desc", afterCursor),
	}

	secondRecorder := testutil.ExecuteRequest(t, router, secondRequest)

	// Then: Verify second page response
	if !assert.Equal(t, http.StatusOK, secondRecorder.Code) {
		t.Logf("Second request failed. Response body: %s", secondRecorder.Body.String())
		t.Logf("Cursor used: %s", afterCursor)
		return
	}

	var secondResponse room.InfiniteScrollResponse
	testutil.ParseResponse(t, secondRecorder, &secondResponse)

	if len(secondResponse.Rooms) != 5 {
		t.Logf("Expected 5 rooms, got %d", len(secondResponse.Rooms))
		t.Logf("First page last item joinedTime: %s", lastJoinedTime.Format(time.RFC3339))
	}
	assert.Len(t, secondResponse.Rooms, 5, "Second page should have remaining 5 items")

	// Verify no duplicate rooms between pages
	firstPageIDs := make(map[uint32]bool)
	for _, rm := range firstResponse.Rooms {
		firstPageIDs[rm.ID] = true
	}

	for _, rm := range secondResponse.Rooms {
		assert.False(t, firstPageIDs[rm.ID], "Room ID %d should not appear in both pages", rm.ID)
	}

	// Verify second page rooms are older than first page
	if len(secondResponse.Rooms) > 0 {
		lastFromFirstPage := firstResponse.Rooms[len(firstResponse.Rooms)-1].JoinedTime
		firstFromSecondPage := secondResponse.Rooms[0].JoinedTime
		assert.True(t, lastFromFirstPage.After(firstFromSecondPage),
			"Second page rooms should be older than first page")
	}
}

func TestFetchRoomsInfiniteScroll_Success_EmptyResult(t *testing.T) {
	// Given: Setup test environment
	roomHandler, db, memberID := setupTestEnvironment(t)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.GET("/api/v1/rooms", roomHandler.GetRoomsByInfiniteScroll)

	// Given: Create only 5 test rooms
	testutil.CreateTestRooms(t, db, memberID, 5)

	// When: Request first page
	firstRequest := testutil.TestRequest{
		Method: http.MethodGet,
		URL:    "/api/v1/rooms?orderBy=time&after=0&dir=desc",
	}

	firstRecorder := testutil.ExecuteRequest(t, router, firstRequest)
	assert.Equal(t, http.StatusOK, firstRecorder.Code)

	var firstResponse room.InfiniteScrollResponse
	testutil.ParseResponse(t, firstRecorder, &firstResponse)
	assert.Len(t, firstResponse.Rooms, 5)

	// When: Request second page (should be empty)
	lastJoinedTime := firstResponse.Rooms[len(firstResponse.Rooms)-1].JoinedTime
	afterCursor := url.QueryEscape(lastJoinedTime.Format(time.RFC3339))

	secondRequest := testutil.TestRequest{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/api/v1/rooms?orderBy=time&after=%s&dir=desc", afterCursor),
	}

	secondRecorder := testutil.ExecuteRequest(t, router, secondRequest)

	// Then: Verify empty response
	assert.Equal(t, http.StatusOK, secondRecorder.Code)

	var secondResponse room.InfiniteScrollResponse
	testutil.ParseResponse(t, secondRecorder, &secondResponse)
	assert.Empty(t, secondResponse.Rooms, "Second page should be empty")
}

func TestFetchRoomsInfiniteScroll_Success_NoRooms(t *testing.T) {
	// Given: Setup test environment with no rooms created
	roomHandler, _, memberID := setupTestEnvironment(t)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.GET("/api/v1/rooms", roomHandler.GetRoomsByInfiniteScroll)

	// When: Request first page
	request := testutil.TestRequest{
		Method: http.MethodGet,
		URL:    "/api/v1/rooms?orderBy=time&after=0&dir=desc",
	}

	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify empty response
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response room.InfiniteScrollResponse
	testutil.ParseResponse(t, recorder, &response)
	assert.Empty(t, response.Rooms, "Should return empty array when no rooms exist")
}

func TestFetchRoomsInfiniteScroll_ValidationError_InvalidAfter(t *testing.T) {
	// Given: Setup test environment
	roomHandler, _, memberID := setupTestEnvironment(t)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.GET("/api/v1/rooms", roomHandler.GetRoomsByInfiniteScroll)

	// When: Request with invalid after format
	request := testutil.TestRequest{
		Method: http.MethodGet,
		URL:    "/api/v1/rooms?orderBy=time&after=invalid-time-format&dir=desc",
	}

	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify error response
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestFetchRoomsInfiniteScroll_Success_DefaultParameters(t *testing.T) {
	// Given: Setup test environment
	roomHandler, db, memberID := setupTestEnvironment(t)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.GET("/api/v1/rooms", roomHandler.GetRoomsByInfiniteScroll)

	// Given: Create test rooms
	testutil.CreateTestRooms(t, db, memberID, 5)

	// When: Request without query parameters (should use defaults)
	request := testutil.TestRequest{
		Method: http.MethodGet,
		URL:    "/api/v1/rooms",
	}

	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify response with default parameters
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response room.InfiniteScrollResponse
	testutil.ParseResponse(t, recorder, &response)
	assert.Len(t, response.Rooms, 5)
}

// Helper function to create multiple test rooms with member relationships
