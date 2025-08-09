package test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	memberdomain "pray-together/internal/domains/member/domain"
	prayerdomain "pray-together/internal/domains/prayer/domain"
	roomdomain "pray-together/internal/domains/room/domain"
)

// PrayerUpdateIntegrationTestSuite tests prayer update API (matching Java PrayerUpdateIntegrateTest)
type PrayerUpdateIntegrationTestSuite struct {
	IntegrationTestSuite

	member         *memberdomain.Member
	room           *roomdomain.Room
	memberRoom     *roomdomain.RoomMember
	headers        map[string]string
	prayerTitle    *prayerdomain.PrayerTitle
	prayerContents []*prayerdomain.PrayerContent

	validMemberID uint64
}

const TEST_CNT = 5

// SetupTest runs before each test (matching Java @BeforeEach setup)
func (suite *PrayerUpdateIntegrationTestSuite) SetupTest() {
	// Call parent SetupTest
	suite.IntegrationTestSuite.SetupTest()

	// 회원 생성
	suite.member = suite.testUtils.CreateUniqueMember()
	suite.validMemberID = suite.member.ID

	// 방 생성
	suite.room = suite.testUtils.CreateUniqueRoom()

	// 방 연관관계 생성
	suite.memberRoom = suite.testUtils.CreateUniqueMemberRoomWithMemberAndRoom(suite.member, suite.room)

	// 인증 헤더 생성
	suite.headers = suite.testUtils.CreateAuthHeaderWithMember(suite.member)

	// 기도 제목 생성 (matching Java: PrayerTitle.create(room, "original-prayer-title"))
	suite.prayerTitle = prayerdomain.NewPrayerTitle(suite.room.ID, suite.member.ID, "original-prayer-title")
	suite.db.Create(suite.prayerTitle)

	// 기도 내용 생성
	suite.prayerContents = make([]*prayerdomain.PrayerContent, 0, TEST_CNT)
	for i := 0; i < TEST_CNT; i++ {
		content := &prayerdomain.PrayerContent{
			PrayerTitleID: suite.prayerTitle.ID,
			AuthorID:      suite.member.ID,
			Content:       fmt.Sprintf("original-prayer-content-%d", i),
		}
		suite.db.Create(content)
		suite.prayerContents = append(suite.prayerContents, content)
	}

	// Set up router with prayer routes
}

// TearDownTest runs after each test (matching Java @AfterEach cleanup)
func (suite *PrayerUpdateIntegrationTestSuite) TearDownTest() {
	suite.CleanRepository()
}

// TestUpdatePrayerThenReturn200OK tests prayer update (matching Java update_prayer_then_return_200_ok)
func (suite *PrayerUpdateIntegrationTestSuite) TestUpdatePrayerThenReturn200OK() {
	// Given
	newTitle := "updated-prayer-title"
	newContent := "updated-prayer-content"

	updateContents := []PrayerUpdateContent{
		{
			ID:         &suite.prayerContents[0].ID,
			MemberID:   &suite.member.ID,
			MemberName: suite.member.Name,
			Content:    newContent,
		},
	}

	requestDTO := PrayerUpdateRequest{
		Title:    newTitle,
		Contents: updateContents,
	}

	updateURL := fmt.Sprintf("%s/%d", suite.PrayersAPIURL, suite.prayerTitle.ID)

	// When
	w := suite.PutJSON(updateURL, requestDTO, suite.headers)

	// Then
	suite.AssertStatusCode(w, http.StatusOK, "기도 변경 API 응답 상태 코드가 200 OK가 아닙니다.")

	// 변경된 기도 제목 확인
	var updatedTitle prayerdomain.PrayerTitle
	suite.db.Preload("Contents").First(&updatedTitle, suite.prayerTitle.ID)
	assert.Equal(suite.T(), newTitle, updatedTitle.Title, "기도 제목이 업데이트되지 않았습니다.")

	// 변경된 기도 내용 확인
	var updatedContents []prayerdomain.PrayerContent
	suite.db.Where("prayer_title_id = ?", suite.prayerTitle.ID).Find(&updatedContents)
	assert.Equal(suite.T(), 1, len(updatedContents), "변경된 기도 내용의 개수가 예상과 다릅니다.")
	assert.Equal(suite.T(), newContent, updatedContents[0].Content, "기도 내용이 업데이트되지 않았습니다.")
}

// TestUpdatePrayerAddContent tests adding content during update (matching Java update_prayer_add_content_then_return_200_ok)
func (suite *PrayerUpdateIntegrationTestSuite) TestUpdatePrayerAddContent() {
	// Given
	newTitle := "updated-prayer-title"
	existingContent := "existing-content"
	newContent := "new-content"

	// 기존 content 업데이트
	updateContents := []PrayerUpdateContent{
		{
			ID:         &suite.prayerContents[0].ID,
			MemberID:   &suite.member.ID,
			MemberName: suite.member.Name,
			Content:    existingContent,
		},
		// 새로운 content 추가
		{
			ID:         nil, // No ID for new content
			MemberID:   &suite.member.ID,
			MemberName: suite.member.Name,
			Content:    newContent,
		},
	}

	requestDTO := PrayerUpdateRequest{
		Title:    newTitle,
		Contents: updateContents,
	}

	updateURL := fmt.Sprintf("%s/%d", suite.PrayersAPIURL, suite.prayerTitle.ID)

	// When
	w := suite.PutJSON(updateURL, requestDTO, suite.headers)

	// Then
	suite.AssertStatusCode(w, http.StatusOK, "기도 변경 API 응답 상태 코드가 200 OK가 아닙니다.")

	// 변경된 기도 제목 확인
	var updatedTitle prayerdomain.PrayerTitle
	suite.db.First(&updatedTitle, suite.prayerTitle.ID)
	assert.Equal(suite.T(), newTitle, updatedTitle.Title, "기도 제목이 업데이트되지 않았습니다.")

	// 변경된 기도 내용 확인
	var updatedContents []prayerdomain.PrayerContent
	suite.db.Where("prayer_title_id = ?", suite.prayerTitle.ID).Find(&updatedContents)
	assert.Equal(suite.T(), 2, len(updatedContents), "변경된 기도 내용의 개수가 예상과 다릅니다.")

	// 내용 확인
	hasExistingContent := false
	hasNewContent := false
	for _, content := range updatedContents {
		if content.Content == existingContent {
			hasExistingContent = true
		}
		if content.Content == newContent {
			hasNewContent = true
		}
	}

	assert.True(suite.T(), hasExistingContent, "기존 기도 내용이 업데이트되지 않았습니다.")
	assert.True(suite.T(), hasNewContent, "새로운 기도 내용이 추가되지 않았습니다.")
}

// TestUpdatePrayerRemoveContent tests removing content during update (matching Java update_prayer_remove_content_then_return_200_ok)
func (suite *PrayerUpdateIntegrationTestSuite) TestUpdatePrayerRemoveContent() {
	// Given
	newTitle := "updated-prayer-title"

	// 빈 내용 목록으로 업데이트 (기존 내용 삭제)
	requestDTO := PrayerUpdateRequest{
		Title:    newTitle,
		Contents: []PrayerUpdateContent{}, // Empty list to remove all contents
	}

	updateURL := fmt.Sprintf("%s/%d", suite.PrayersAPIURL, suite.prayerTitle.ID)

	// When
	w := suite.PutJSON(updateURL, requestDTO, suite.headers)

	// Then
	suite.AssertStatusCode(w, http.StatusOK, "기도 변경 API 응답 상태 코드가 200 OK가 아닙니다.")

	// 변경된 기도 제목 확인
	var updatedTitle prayerdomain.PrayerTitle
	suite.db.First(&updatedTitle, suite.prayerTitle.ID)
	assert.Equal(suite.T(), newTitle, updatedTitle.Title, "기도 제목이 업데이트되지 않았습니다.")

	// 기도 내용이 삭제되었는지 확인
	var updatedContents []prayerdomain.PrayerContent
	suite.db.Where("prayer_title_id = ?", suite.prayerTitle.ID).Find(&updatedContents)
	assert.Empty(suite.T(), updatedContents, "기도 내용이 삭제되지 않았습니다.")
}

// PrayerUpdateRequest represents the request to update a prayer (matching Java PrayerUpdateRequest)
type PrayerUpdateRequest struct {
	Title    string                `json:"title" binding:"required,min=1,max=50"`
	Contents []PrayerUpdateContent `json:"contents"`
}

// PrayerUpdateContent represents prayer content in update request (matching Java PrayerUpdateContent)
type PrayerUpdateContent struct {
	ID         *uint64 `json:"id,omitempty"`
	MemberID   *uint64 `json:"memberId,omitempty"`
	MemberName string  `json:"memberName" binding:"required"`
	Content    string  `json:"content" binding:"required"`
}

// TestUpdatePrayerWithNonexistentID tests updating prayer with nonexistent ID (matching Java update_prayer_with_nonexistent_id_then_return_404_not_found)
func (suite *PrayerUpdateIntegrationTestSuite) TestUpdatePrayerWithNonexistentID() {
	// Given
	nonExistentID := uint64(999999)

	requestDto := PrayerUpdateRequest{
		Title:    "updated-prayer-title",
		Contents: []PrayerUpdateContent{},
	}

	url := fmt.Sprintf("%s/%d", suite.PrayersAPIURL, nonExistentID)

	// When
	w := suite.PutJSON(url, requestDto, suite.headers)

	// Then
	suite.AssertStatusCode(w, http.StatusNotFound,
		"존재하지 않는 기도 제목 ID로 변경 요청 시 404 Not Found가 반환되어야 합니다.")
}

// TestPrayerUpdateIntegration runs the prayer update integration test suite
func TestPrayerUpdateIntegration(t *testing.T) {
	suite.Run(t, new(PrayerUpdateIntegrationTestSuite))
}
