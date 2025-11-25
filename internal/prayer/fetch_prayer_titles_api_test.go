package prayer_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/prayer"
	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/testutil"
	"github.com/stretchr/testify/assert"
)

func TestFetchPrayerTitles_ReturnsLatestTitles(t *testing.T) {
	prayerHandler, db, memberID := setupTestEnvironment(t)
	room := testutil.CreateTestRoom(t, db, memberID, "Pray Room", "Room for prayers")

	titles := testutil.SeedPrayerTitlesWithTime(t, db, room.ID, 3)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.GET("/api/v1/prayers", prayerHandler.FetchTitlesByInfiniteScroll)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/api/v1/prayers?roomId=%d", room.ID),
	})

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response prayer.PrayerTitleInfiniteScrollResponse
	testutil.ParseResponse(t, recorder, &response)
	assert.Len(t, response.PrayerTitles, len(titles))
	assert.Equal(t, []int64{titles[2].ID, titles[1].ID, titles[0].ID}, extractPrayerTitleIDs(response.PrayerTitles))
}

func TestFetchPrayerTitles_UsesCursorForNextPage(t *testing.T) {
	prayerHandler, db, memberID := setupTestEnvironment(t)
	room := testutil.CreateTestRoom(t, db, memberID, "Cursor Room", "Room for pagination")

	total := prayer.PrayerTitleInfiniteScrollLimit + 2
	titles := testutil.SeedPrayerTitlesWithTime(t, db, room.ID, total)
	cursorIndex := total - prayer.PrayerTitleInfiniteScrollLimit
	cursorTimestamp := titles[cursorIndex].CreatedAt.UTC().Format(time.RFC3339Nano)

	expectedIDs := make([]int64, 0, cursorIndex)
	for i := cursorIndex - 1; i >= 0; i-- {
		expectedIDs = append(expectedIDs, titles[i].ID)
	}

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.GET("/api/v1/prayers", prayerHandler.FetchTitlesByInfiniteScroll)

	cursorParam := url.QueryEscape(cursorTimestamp)
	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/api/v1/prayers?roomId=%d&after=%s", room.ID, cursorParam),
	})

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response prayer.PrayerTitleInfiniteScrollResponse
	testutil.ParseResponse(t, recorder, &response)
	assert.Len(t, response.PrayerTitles, cursorIndex)
	assert.Equal(t, expectedIDs, extractPrayerTitleIDs(response.PrayerTitles))
}

func TestFetchPrayerTitles_ConsecutiveScrolling(t *testing.T) {
	prayerHandler, db, memberID := setupTestEnvironment(t)
	room := testutil.CreateTestRoom(t, db, memberID, "Scroll Room", "Room for consecutive scrolling")

	// Create enough titles for 3 pages: limit*2 + 5 = 25 titles
	total := prayer.PrayerTitleInfiniteScrollLimit*2 + 5
	titles := testutil.SeedPrayerTitlesWithTime(t, db, room.ID, total)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.GET("/api/v1/prayers", prayerHandler.FetchTitlesByInfiniteScroll)

	// First page request
	recorder1 := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/api/v1/prayers?roomId=%d", room.ID),
	})

	assert.Equal(t, http.StatusOK, recorder1.Code)

	var response1 prayer.PrayerTitleInfiniteScrollResponse
	testutil.ParseResponse(t, recorder1, &response1)
	assert.Len(t, response1.PrayerTitles, prayer.PrayerTitleInfiniteScrollLimit, "First page should return limit items")

	// Verify first page returns latest titles in descending order
	firstPageIDs := extractPrayerTitleIDs(response1.PrayerTitles)
	expectedFirstPageIDs := make([]int64, prayer.PrayerTitleInfiniteScrollLimit)
	for i := 0; i < prayer.PrayerTitleInfiniteScrollLimit; i++ {
		expectedFirstPageIDs[i] = titles[total-1-i].ID
	}
	assert.Equal(t, expectedFirstPageIDs, firstPageIDs)

	// Second page request using last item's CreatedAt as cursor
	lastItemPage1 := response1.PrayerTitles[len(response1.PrayerTitles)-1]
	cursorParam := url.QueryEscape(lastItemPage1.CreatedTime.UTC().Format(time.RFC3339Nano))
	recorder2 := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/api/v1/prayers?roomId=%d&after=%s", room.ID, cursorParam),
	})

	assert.Equal(t, http.StatusOK, recorder2.Code)

	var response2 prayer.PrayerTitleInfiniteScrollResponse
	testutil.ParseResponse(t, recorder2, &response2)
	assert.Len(t, response2.PrayerTitles, prayer.PrayerTitleInfiniteScrollLimit, "Second page should return limit items")

	// Verify second page returns next set of titles
	secondPageIDs := extractPrayerTitleIDs(response2.PrayerTitles)
	expectedSecondPageIDs := make([]int64, prayer.PrayerTitleInfiniteScrollLimit)
	for i := 0; i < prayer.PrayerTitleInfiniteScrollLimit; i++ {
		expectedSecondPageIDs[i] = titles[total-1-prayer.PrayerTitleInfiniteScrollLimit-i].ID
	}
	assert.Equal(t, expectedSecondPageIDs, secondPageIDs)

	// Third page request using last item's CreatedAt as cursor
	lastItemPage2 := response2.PrayerTitles[len(response2.PrayerTitles)-1]
	cursorParam2 := url.QueryEscape(lastItemPage2.CreatedTime.UTC().Format(time.RFC3339Nano))
	recorder3 := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/api/v1/prayers?roomId=%d&after=%s", room.ID, cursorParam2),
	})

	assert.Equal(t, http.StatusOK, recorder3.Code)

	var response3 prayer.PrayerTitleInfiniteScrollResponse
	testutil.ParseResponse(t, recorder3, &response3)
	remainingCount := total - prayer.PrayerTitleInfiniteScrollLimit*2
	assert.Len(t, response3.PrayerTitles, remainingCount, "Third page should return remaining items")

	// Verify third page returns remaining titles
	thirdPageIDs := extractPrayerTitleIDs(response3.PrayerTitles)
	expectedThirdPageIDs := make([]int64, remainingCount)
	for i := 0; i < remainingCount; i++ {
		expectedThirdPageIDs[i] = titles[remainingCount-1-i].ID
	}
	assert.Equal(t, expectedThirdPageIDs, thirdPageIDs)

	// Verify no duplicates across all pages
	allIDs := append(append(firstPageIDs, secondPageIDs...), thirdPageIDs...)
	assert.Len(t, allIDs, total, "Total number of titles should match")
	uniqueIDs := make(map[int64]bool)
	for _, id := range allIDs {
		assert.False(t, uniqueIDs[id], "Duplicate ID found: %d", id)
		uniqueIDs[id] = true
	}
}

func TestFetchPrayerTitles_InvalidCursor(t *testing.T) {
	prayerHandler, db, memberID := setupTestEnvironment(t)
	room := testutil.CreateTestRoom(t, db, memberID, "Invalid Cursor Room", "Room for invalid cursor test")
	testutil.CreateTestPrayerTitle(t, db, room.ID, "Pray for family")

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.GET("/api/v1/prayers", prayerHandler.FetchTitlesByInfiniteScroll)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/api/v1/prayers?roomId=%d&after=invalid", room.ID),
	})

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &response)
	assert.Equal(t, "PRAYER-006", response.Code)
}

func TestFetchPrayerTitles_MemberNotInRoom(t *testing.T) {
	prayerHandler, db, memberID := setupTestEnvironment(t)
	otherMember := testutil.CreateTestMemberWithIndex(t, db, 1)
	otherRoom := testutil.CreateTestRoom(t, db, int64(otherMember.ID), "Another Room", "Not joined")

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.GET("/api/v1/prayers", prayerHandler.FetchTitlesByInfiniteScroll)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/api/v1/prayers?roomId=%d", otherRoom.ID),
	})

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var response sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &response)
	assert.Equal(t, "ROOM-002", response.Code)
}

// extractPrayerTitleIDs extracts IDs from a slice of PrayerTitleInfo
func extractPrayerTitleIDs(infos []prayer.PrayerTitleInfo) []int64 {
	ids := make([]int64, 0, len(infos))
	for _, info := range infos {
		ids = append(ids, info.ID)
	}
	return ids
}
