package test

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	memberdomain "pray-together/internal/domains/member/domain"
	roomdomain "pray-together/internal/domains/room/domain"
)

// RoomInfiniteScrollIntegrationTestSuite tests room infinite scroll API (matching Java RoomInfiniteScrollIntegrateTest)
type RoomInfiniteScrollIntegrationTestSuite struct {
	IntegrationTestSuite

	member  *memberdomain.Member
	headers map[string]string

	orderBy string
	after   string
	dir     string

	orderByTime string
	dirDesc     string
}

// SetupTest runs before each test (matching Java @BeforeEach setup)
func (suite *RoomInfiniteScrollIntegrationTestSuite) SetupTest() {
	// Call parent SetupTest
	suite.IntegrationTestSuite.SetupTest()

	suite.orderBy = "orderBy"
	suite.after = "after"
	suite.dir = "dir"
	suite.orderByTime = "time"
	suite.dirDesc = "desc"

	// member1 생성
	suite.member = suite.testUtils.CreateUniqueMember()
	// memberRepository.save(member) - already saved in CreateUniqueMember

	// room1 생성
	var allRoom []*roomdomain.Room
	for i := 0; i < 30; i++ {
		testRoom := &roomdomain.Room{
			RoomName:              fmt.Sprintf("test%d", i+1),
			IsPrivate:             false,
			PrayStartTime:         "00:00",
			PrayEndTime:           "23:59",
			NotificationStartTime: "00:00",
			NotificationEndTime:   "23:59",
		}
		suite.db.Create(testRoom)
		allRoom = append(allRoom, testRoom)
		// Add small delay to ensure different timestamps
		time.Sleep(time.Millisecond)
	}

	// Reload rooms to get IDs
	suite.db.Find(&allRoom)

	// member1는 홀수 ID 방 추가
	for i := 0; i < 30; i++ {
		room := allRoom[i]
		if room.ID%2 == 0 {
			continue
		}
		memberRoom := &roomdomain.RoomMember{
			MemberID:       suite.member.ID,
			RoomID:         room.ID,
			Role:           roomdomain.RoleMember,
			IsNotification: true,
		}
		suite.db.Create(memberRoom)
	}

	// member2 생성
	member2 := suite.testUtils.CreateUniqueMember()
	// memberRepository.save(member2) - already saved in CreateUniqueMember

	// member2는 짝수 ID 방 추가
	for i := 0; i < 30; i++ {
		room := allRoom[i]
		if room.ID%2 == 1 {
			continue
		}
		memberRoom := &roomdomain.RoomMember{
			MemberID:       member2.ID,
			RoomID:         room.ID,
			Role:           roomdomain.RoleMember,
			IsNotification: true,
		}
		suite.db.Create(memberRoom)
	}

	// Set up router with room routes
}

// TearDownTest runs after each test (matching Java @AfterEach cleanup)
func (suite *RoomInfiniteScrollIntegrationTestSuite) TearDownTest() {
	suite.CleanRepository()
}

// TestFetchRoomsListWithDefaultValuesForDifferentParamsThenReturn200OK tests room infinite scroll with various parameters
func (suite *RoomInfiniteScrollIntegrationTestSuite) TestFetchRoomsListWithDefaultValuesForDifferentParamsThenReturn200OK() {
	testCases := []struct {
		name    string
		orderBy string
		after   string
		dir     string
	}{
		{"기본값", "time", "0", "desc"},
		{"orderBy null", "", "0", "desc"},
		{"orderBy 빈값", "", "0", "desc"},
		{"after null", "time", "", "desc"},
		{"after 빈값", "time", "", "desc"},
		{"dir null", "time", "0", ""},
		{"dir 빈값", "time", "0", ""},
		{"모든 값 null", "", "", ""},
		{"모든 값 빈값", "", "", ""},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Given
			suite.headers = suite.testUtils.CreateAuthHeaderWithMember(suite.member)

			params := url.Values{}
			if tc.orderBy != "" {
				params.Add("orderBy", tc.orderBy)
			}
			if tc.after != "" {
				params.Add("after", tc.after)
			}
			if tc.dir != "" {
				params.Add("dir", tc.dir)
			}
			uri := suite.RoomsAPIURL
			if len(params) > 0 {
				uri = uri + "?" + params.Encode()
			}

			// When
			w := suite.GetRequest(uri, suite.headers)

			// Then
			assert.Equal(t, http.StatusOK, w.Code,
				"방 목록 무한 스크롤 API 응답 상태 코드가 200 OK가 아닙니다.")

			var response RoomInfiniteScrollResponse
			err := suite.UnmarshalResponse(w, &response)
			assert.NoError(t, err)
			assert.NotNil(t, response, "방 목록 무한 스크롤 API 응답 결과가 NULL 입니다.")

			rooms := response.Rooms
			assert.Greater(t, len(rooms), 0, "방 목록 무한 스크롤 API 응답 결과 데이터가 없습니다.")

			// Check if sorted by joinedTime in descending order
			isSorted := sort.SliceIsSorted(rooms, func(i, j int) bool {
				return rooms[i].JoinedTime.After(rooms[j].JoinedTime)
			})
			assert.True(t, isSorted, "방 목록이 joinedTime 기준으로 내림차순 정렬되지 않았습니다.")

			// Check all room IDs are odd
			for _, room := range rooms {
				assert.Equal(t, uint64(1), room.ID%2, "모든 방의 ID가 홀수여야 합니다.")
			}
		})
	}
}

// TestFetchRoomsListWithSequentialRequestsTimeDescAndEmptyFinalResponse tests sequential room fetching
func (suite *RoomInfiniteScrollIntegrationTestSuite) TestFetchRoomsListWithSequentialRequestsTimeDescAndEmptyFinalResponse() {
	// Given
	suite.headers = suite.testUtils.CreateAuthHeaderWithMember(suite.member)

	params := url.Values{}
	params.Add("orderBy", suite.orderByTime)
	params.Add("after", "0")
	params.Add("dir", suite.dirDesc)
	uri := suite.RoomsAPIURL + "?" + params.Encode()

	// 첫 번째 요청
	w := suite.GetRequest(uri, suite.headers)

	assert.Equal(suite.T(), http.StatusOK, w.Code,
		"첫 번째 요청: 방 목록 무한 스크롤 API 응답 상태 코드가 200 OK가 아닙니다.")

	var response RoomInfiniteScrollResponse
	err := suite.UnmarshalResponse(w, &response)
	assert.NoError(suite.T(), err)

	rooms := response.Rooms

	// 모든 방을 가져올 때까지 반복 요청
	requestCount := 0
	for len(rooms) > 0 {
		// 현재 응답 검증
		isSorted := sort.SliceIsSorted(rooms, func(i, j int) bool {
			return rooms[i].JoinedTime.After(rooms[j].JoinedTime)
		})
		assert.True(suite.T(), isSorted, "방 목록이 joinedTime 기준으로 내림차순 정렬되지 않았습니다.")

		// Check all room IDs are odd
		for _, room := range rooms {
			assert.Equal(suite.T(), uint64(1), room.ID%2, "모든 방의 ID가 홀수여야 합니다.")
		}

		// 다음 요청 준비
		lastRoom := rooms[len(rooms)-1]
		params := url.Values{}
		params.Add("orderBy", suite.orderByTime)
		params.Add("after", lastRoom.JoinedTime.Format(time.RFC3339Nano))
		params.Add("dir", suite.dirDesc)
		uri = suite.RoomsAPIURL + "?" + params.Encode()

		// 다음 요청 수행
		w = suite.GetRequest(uri, suite.headers)

		assert.Equal(suite.T(), http.StatusOK, w.Code,
			"연속 요청: 방 목록 무한 스크롤 API 응답 상태 코드가 200 OK가 아닙니다.")

		err = suite.UnmarshalResponse(w, &response)
		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), response, "연속 요청: 방 목록 무한 스크롤 API 응답 결과가 NULL 입니다.")

		rooms = response.Rooms

		// Prevent infinite loop
		requestCount++
		if requestCount > 10 {
			break
		}
	}

	assert.Empty(suite.T(), rooms, "마지막 응답: 빈 컬렉션이어야 합니다.")
}

// RoomInfo represents room info (matching Java RoomInfo)
type RoomInfo struct {
	ID             uint64    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	MemberCount    int       `json:"memberCount"`
	JoinedTime     time.Time `json:"joinedTime"`
	Role           string    `json:"role"`
	IsNotification bool      `json:"isNotification"`
}

// RoomInfiniteScrollResponse represents room infinite scroll response (matching Java RoomInfiniteScrollResponse)
type RoomInfiniteScrollResponse struct {
	Rooms   []RoomInfo `json:"rooms"`
	HasNext bool       `json:"hasNext"`
}

// TestRoomInfiniteScrollIntegration runs the room infinite scroll integration test suite
func TestRoomInfiniteScrollIntegration(t *testing.T) {
	suite.Run(t, new(RoomInfiniteScrollIntegrationTestSuite))
}
