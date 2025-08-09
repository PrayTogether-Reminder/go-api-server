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
	prayerdomain "pray-together/internal/domains/prayer/domain"
	roomdomain "pray-together/internal/domains/room/domain"
)

// Constants matching Java
const PRAYER_TITLES_INFINITE_SCROLL_SIZE = 10

// PrayerInfiniteScrollIntegrationTestSuite tests prayer infinite scroll API (matching Java PrayerInfiniteScrollIntegrateTest)
type PrayerInfiniteScrollIntegrationTestSuite struct {
	IntegrationTestSuite

	headers     map[string]string
	member      *memberdomain.Member
	room        *roomdomain.Room
	prayerTitle *prayerdomain.PrayerTitle
	testCnt     int

	after  string
	roomID string
}

// SetupTest runs before each test (matching Java @BeforeEach setup)
func (suite *PrayerInfiniteScrollIntegrationTestSuite) SetupTest() {
	// Call parent SetupTest
	suite.IntegrationTestSuite.SetupTest()

	suite.testCnt = PRAYER_TITLES_INFINITE_SCROLL_SIZE * 3
	suite.after = "after"
	suite.roomID = "roomId"

	suite.member = suite.testUtils.CreateUniqueMember()
	// memberRepository.save(member) - already saved in CreateUniqueMember

	suite.room = suite.testUtils.CreateUniqueRoom()
	// roomRepository.save(room) - already saved in CreateUniqueRoom

	memberRoom := suite.testUtils.CreateUniqueMemberRoomWithMemberAndRoom(suite.member, suite.room)
	suite.db.Create(memberRoom)

	suite.prayerTitle = &prayerdomain.PrayerTitle{
		RoomID: suite.room.ID,
		Title:  "test-title",
	}
	suite.db.Create(suite.prayerTitle)

	for i := 0; i < suite.testCnt; i++ {
		prayerTitle := &prayerdomain.PrayerTitle{
			RoomID: suite.room.ID,
			Title:  fmt.Sprintf("test-title%d", i),
		}
		suite.db.Create(prayerTitle)
		// Add small delay to ensure different timestamps
		time.Sleep(time.Millisecond)
	}

	suite.headers = suite.testUtils.CreateAuthHeaderWithMember(suite.member)

	// Set up router with prayer routes
}

// TearDownTest runs after each test (matching Java @AfterEach cleanup)
func (suite *PrayerInfiniteScrollIntegrationTestSuite) TearDownTest() {
	suite.CleanRepository()
}

// TestFetchPrayerContentsListWithDefaultValuesForDifferentParamsThenReturn200OK tests prayer infinite scroll with various parameters
func (suite *PrayerInfiniteScrollIntegrationTestSuite) TestFetchPrayerContentsListWithDefaultValuesForDifferentParamsThenReturn200OK() {
	testCases := []struct {
		name  string
		after string
	}{
		{"after=0", "0"},
		{"after null", ""},
		{"after 빈값", ""},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Given
			params := url.Values{}
			params.Add("roomId", fmt.Sprintf("%d", suite.room.ID))
			if tc.after != "" {
				params.Add("after", tc.after)
			}
			uri := suite.PrayersAPIURL + "?" + params.Encode()

			// When
			w := suite.GetRequest(uri, suite.headers)

			// Then
			suite.AssertStatusCode(w, http.StatusOK, tc.name+": 기도 내용 목록 무한 스크롤 API 응답 상태 코드가 200 OK가 아닙니다.")

			var response PrayerTitleInfiniteScrollResponse
			err := suite.UnmarshalResponse(w, &response)
			assert.NoError(t, err, "응답을 파싱할 수 있어야 합니다")
			assert.NotNil(t, response, tc.name+": 기도 내용 목록 무한 스크롤 API 응답 결과가 NULL 입니다.")

			titles := response.PrayerTitles
			assert.Equal(t, PRAYER_TITLES_INFINITE_SCROLL_SIZE, len(titles),
				tc.name+": 기도 내용 목록 무한 스크롤 API 응답 결과 데이터가 없습니다.")

			// Check if sorted by createdTime in descending order
			isSorted := sort.SliceIsSorted(titles, func(i, j int) bool {
				return titles[i].CreatedTime.After(titles[j].CreatedTime)
			})
			assert.True(t, isSorted, tc.name+": 기도 내용 목록이 createdTime 기준으로 내림차순 정렬되지 않았습니다.")

			repeatCount := 1
			for len(titles) > 0 {
				// Next given
				lastTitle := titles[len(titles)-1]
				lastAfter := lastTitle.CreatedTime

				params := url.Values{}
				params.Add("roomId", fmt.Sprintf("%d", suite.room.ID))
				params.Add("after", lastAfter.Format(time.RFC3339Nano))
				nextURI := suite.PrayersAPIURL + "?" + params.Encode()

				// Next when
				w = suite.GetRequest(nextURI, suite.headers)

				// Next then
				assert.Equal(t, http.StatusOK, w.Code,
					fmt.Sprintf("%s: %d 번째 요청 응답 코드가 200 OK 아닙니다.", tc.name, repeatCount))

				err = suite.UnmarshalResponse(w, &response)
				assert.NoError(t, err)
				assert.NotNil(t, response, fmt.Sprintf("%s: %d 번째 요청 응답 body가 null입니다.", tc.name, repeatCount))

				titles = response.PrayerTitles
				repeatCount++

				// Prevent infinite loop
				if repeatCount > 10 {
					break
				}
			}

			// 최종 검증
			assert.Empty(t, titles, tc.name+": 마지막 요청 결과가 빈 리스트가 아닙니다.")
		})
	}
}

// PrayerTitleInfo represents prayer title info (matching Java PrayerTitleInfo)
type PrayerTitleInfo struct {
	ID          uint64    `json:"id"`
	Title       string    `json:"title"`
	RoomID      uint64    `json:"roomId"`
	RoomName    string    `json:"roomName"`
	CreatedTime time.Time `json:"createdTime"`
}

// PrayerTitleInfiniteScrollResponse represents prayer title infinite scroll response (matching Java PrayerTitleInfiniteScrollResponse)
type PrayerTitleInfiniteScrollResponse struct {
	PrayerTitles []PrayerTitleInfo `json:"prayerTitles"`
	HasNext      bool              `json:"hasNext"`
}

// TestPrayerInfiniteScrollIntegration runs the prayer infinite scroll integration test suite
func TestPrayerInfiniteScrollIntegration(t *testing.T) {
	suite.Run(t, new(PrayerInfiniteScrollIntegrationTestSuite))
}
