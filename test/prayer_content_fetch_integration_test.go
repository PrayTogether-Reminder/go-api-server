package test

import (
	"fmt"
	"net/http"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	memberdomain "pray-together/internal/domains/member/domain"
	prayerdomain "pray-together/internal/domains/prayer/domain"
	roomdomain "pray-together/internal/domains/room/domain"
)

// PrayerContentFetchIntegrationTestSuite tests prayer content fetch API (matching Java PrayerContentFetchIntegrateTest)
type PrayerContentFetchIntegrationTestSuite struct {
	IntegrationTestSuite

	headers     map[string]string
	member      *memberdomain.Member
	room        *roomdomain.Room
	prayerTitle *prayerdomain.PrayerTitle
	testCnt     int
}

// SetupTest runs before each test (matching Java @BeforeEach setup)
func (suite *PrayerContentFetchIntegrationTestSuite) SetupTest() {
	// Call parent SetupTest
	suite.IntegrationTestSuite.SetupTest()

	suite.testCnt = 5

	// 회원 생성
	suite.member = suite.testUtils.CreateUniqueMember()
	// memberRepository.save(member) - already saved in CreateUniqueMember

	// 방 생성
	suite.room = suite.testUtils.CreateUniqueRoom()
	// roomRepository.save(room) - already saved in CreateUniqueRoom

	// 회원-방 연관관계 생성
	memberRoom := suite.testUtils.CreateUniqueMemberRoomWithMemberAndRoom(suite.member, suite.room)
	suite.db.Create(memberRoom)

	// 기도 제목 생성
	suite.prayerTitle = suite.testUtils.CreateUniquePrayerTitleWithRoom(suite.room)
	// prayerTitleRepository.save(prayerTitle) - already saved in CreateUniquePrayerTitleWithRoom

	// 기도 내용 추가
	prayerContent := &prayerdomain.PrayerContent{
		PrayerTitleID: suite.prayerTitle.ID,
		Content:       "test-prayer-content",
		AuthorID:      suite.member.ID,
	}
	suite.db.Create(prayerContent)

	suite.headers = suite.testUtils.CreateAuthHeaderWithMember(suite.member)

	// Set up router with prayer routes
}

// TearDownTest runs after each test (matching Java @AfterEach cleanup)
func (suite *PrayerContentFetchIntegrationTestSuite) TearDownTest() {
	suite.CleanRepository()
}

// TestFetchPrayerContentsListThenReturn200OK tests fetching prayer contents (matching Java fetch_prayer_contents_list_then_return_200_ok)
func (suite *PrayerContentFetchIntegrationTestSuite) TestFetchPrayerContentsListThenReturn200OK() {
	// Given
	// 회원 및 기도 내용 추가
	for i := 1; i < suite.testCnt; i++ {
		newMember := suite.testUtils.CreateUniqueMember()
		// memberRepository.save(newMember) - already saved in CreateUniqueMember

		prayerContent := &prayerdomain.PrayerContent{
			PrayerTitleID: suite.prayerTitle.ID,
			Content:       fmt.Sprintf("test-prayer-content%c", i+'ㄱ'),
			AuthorID:      newMember.ID,
		}
		suite.db.Create(prayerContent)
	}

	uri := fmt.Sprintf("%s/%d/contents", suite.PrayersAPIURL, suite.prayerTitle.ID)

	// Debug: Check if prayer contents exist in DB
	var dbContents []prayerdomain.PrayerContent
	suite.db.Where("prayer_title_id = ?", suite.prayerTitle.ID).Find(&dbContents)
	suite.T().Logf("Prayer contents in DB: %d", len(dbContents))

	// When
	w := suite.GetRequest(uri, suite.headers)

	// Debug: Log response
	suite.T().Logf("Response body: %s", w.Body.String())

	// Then
	suite.AssertStatusCode(w, http.StatusOK, "기도 내용 목록 조회 API 응답 상태 코드가 200 OK가 아닙니다.")

	var response PrayerContentResponse
	err := suite.UnmarshalResponse(w, &response)
	assert.NoError(suite.T(), err, "응답을 파싱할 수 있어야 합니다")
	assert.NotNil(suite.T(), response, "기도 내용 목록 조회 API 응답 결과가 NULL 입니다.")

	prayerContents := response.PrayerContents
	assert.Equal(suite.T(), suite.testCnt, len(prayerContents),
		"기도 내용 목록 조회 API 응답 결과 데이터 개수가 기대값과 다릅니다.")

	// Check if sorted by memberName in ascending order
	isSorted := sort.SliceIsSorted(prayerContents, func(i, j int) bool {
		return prayerContents[i].MemberName < prayerContents[j].MemberName
	})
	assert.True(suite.T(), isSorted, "기도 내용 목록이 memberName 기준으로 오름차순 정렬되지 않았습니다.")
}

// PrayerContentInfo represents prayer content info (matching Java PrayerContentInfo)
type PrayerContentInfo struct {
	ID         uint64 `json:"id"`
	Content    string `json:"content"`
	MemberID   uint64 `json:"memberId"`
	MemberName string `json:"memberName"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

// PrayerContentResponse represents prayer content response (matching Java PrayerContentResponse)
type PrayerContentResponse struct {
	PrayerContents []PrayerContentInfo `json:"prayerContents"`
}

// TestPrayerContentFetchIntegration runs the prayer content fetch integration test suite
func TestPrayerContentFetchIntegration(t *testing.T) {
	suite.Run(t, new(PrayerContentFetchIntegrationTestSuite))
}
