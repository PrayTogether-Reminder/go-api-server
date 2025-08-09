package test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	memberdomain "pray-together/internal/domains/member/domain"
	prayerdomain "pray-together/internal/domains/prayer/domain"
	roomdomain "pray-together/internal/domains/room/domain"
)

// PrayerDeleteIntegrationTestSuite tests prayer deletion API (matching Java PrayerDeleteIntegrateTest)
type PrayerDeleteIntegrationTestSuite struct {
	IntegrationTestSuite

	member        *memberdomain.Member
	room          *roomdomain.Room
	memberRoom    *roomdomain.RoomMember
	headers       map[string]string
	prayerTitle   *prayerdomain.PrayerTitle
	prayerContent *prayerdomain.PrayerContent
}

// SetupTest runs before each test (matching Java @BeforeEach setup)
func (suite *PrayerDeleteIntegrationTestSuite) SetupTest() {
	// Call parent SetupTest
	suite.IntegrationTestSuite.SetupTest()

	// 회원 생성
	suite.member = suite.testUtils.CreateUniqueMember()
	// memberRepository.save(member) - already saved in CreateUniqueMember

	// 방 생성
	suite.room = suite.testUtils.CreateUniqueRoom()
	// roomRepository.save(room) - already saved in CreateUniqueRoom

	// 방 연관관계 생성
	suite.memberRoom = suite.testUtils.CreateUniqueMemberRoomWithMemberAndRoom(suite.member, suite.room)
	// memberRoomRepository.save(memberRoom) - already saved in CreateUniqueMemberRoomWithMemberAndRoom

	// 인증 헤더 생성
	suite.headers = suite.testUtils.CreateAuthHeaderWithMember(suite.member)

	// 기도 제목 생성 (matching Java: PrayerTitle.create(room, "test-prayer-title"))
	suite.prayerTitle = prayerdomain.NewPrayerTitle(suite.room.ID, suite.member.ID, "test-prayer-title")
	suite.db.Create(suite.prayerTitle)

	// 기도 내용 생성 (matching Java: for loop creating 5 prayer contents)
	for i := 0; i < 5; i++ {
		newMember := suite.testUtils.CreateUniqueMember()
		// memberRepository.save(newMember) - already saved in CreateUniqueMember

		// Create prayer content (matching Java: PrayerContent.create)
		prayerContent := &prayerdomain.PrayerContent{
			PrayerTitleID: suite.prayerTitle.ID,
			AuthorID:      newMember.ID,
			Content:       fmt.Sprintf("test-prayer-content%d", i),
		}

		suite.db.Create(prayerContent)

		// Keep reference to last content
		if i == 4 {
			suite.prayerContent = prayerContent
		}
	}

	// Set up router with prayer routes
}

// TearDownTest runs after each test (matching Java @AfterEach cleanup)
func (suite *PrayerDeleteIntegrationTestSuite) TearDownTest() {
	suite.CleanRepository()
}

// TestDeletePrayerThenReturn200OK tests prayer deletion (matching Java delete_prayer_then_return_200_ok)
func (suite *PrayerDeleteIntegrationTestSuite) TestDeletePrayerThenReturn200OK() {
	// Given
	deleteURL := fmt.Sprintf("%s/%d", suite.PrayersAPIURL, suite.prayerTitle.ID)

	// When
	w := suite.DeleteRequest(deleteURL, suite.headers)

	// Then
	// 삭제 응답 상태 검증
	suite.AssertStatusCode(w, http.StatusOK, "기도 삭제 API 응답 상태 코드가 200 OK가 아닙니다.")

	// 기도 제목이 삭제되었는지 확인
	var prayerTitle prayerdomain.PrayerTitle
	result := suite.db.First(&prayerTitle, suite.prayerTitle.ID)
	assert.Error(suite.T(), result.Error, "기도 제목이 삭제되지 않았습니다.")

	// 연관된 기도 내용이 삭제되었는지 확인
	var remainingContents []prayerdomain.PrayerContent
	suite.db.Find(&remainingContents)
	assert.Empty(suite.T(), remainingContents, "연관된 기도 내용이 삭제되지 않았습니다.")
}

// TestDeletePrayerWithInvalidID tests deletion with invalid ID (matching Java @ParameterizedTest)
func (suite *PrayerDeleteIntegrationTestSuite) TestDeletePrayerWithInvalidID() {
	// Test cases matching Java provideInvalidPrayerDeleteParameters
	testCases := []struct {
		name        string
		encodedURL  string
		expectedMsg string
	}{
		{
			name:        "음수 ID",
			encodedURL:  url.QueryEscape("-1"),
			expectedMsg: "음수 ID로 기도 삭제 요청 시 400 Bad Request가 반환되어야 합니다.",
		},
		{
			name:        "0 ID",
			encodedURL:  url.QueryEscape("0"),
			expectedMsg: "0 ID로 기도 삭제 요청 시 400 Bad Request가 반환되어야 합니다.",
		},
		{
			name:        "문자열 ID",
			encodedURL:  url.QueryEscape("abc"),
			expectedMsg: "문자열 ID로 기도 삭제 요청 시 400 Bad Request가 반환되어야 합니다.",
		},
		{
			name:        "특수문자 ID",
			encodedURL:  url.QueryEscape("!@#"),
			expectedMsg: "특수문자 ID로 기도 삭제 요청 시 400 Bad Request가 반환되어야 합니다.",
		},
		{
			name:        "소수점 ID",
			encodedURL:  url.QueryEscape("1.5"),
			expectedMsg: "소수점 ID로 기도 삭제 요청 시 400 Bad Request가 반환되어야 합니다.",
		},
		{
			name:        "공백 ID",
			encodedURL:  url.QueryEscape(" "),
			expectedMsg: "공백 ID로 기도 삭제 요청 시 400 Bad Request가 반환되어야 합니다.",
		},
		{
			name:        "null",
			encodedURL:  "null",
			expectedMsg: "null ID로 기도 삭제 요청 시 400 Bad Request가 반환되어야 합니다.",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Given
			deleteURL := fmt.Sprintf("%s/%s", suite.PrayersAPIURL, tc.encodedURL)

			// When
			w := suite.DeleteRequest(deleteURL, suite.headers)

			// Then
			suite.AssertStatusCode(w, http.StatusBadRequest, tc.expectedMsg)

			// ExceptionResponse 검증
			var exceptionResponse map[string]interface{}
			err := suite.UnmarshalResponse(w, &exceptionResponse)
			assert.NoError(suite.T(), err, "응답을 파싱할 수 있어야 합니다")
			assert.NotNil(suite.T(), exceptionResponse, "예외 응답이 있어야 합니다")
		})
	}
}

// TestDeletePrayerByMemberFromDifferentRoom tests deletion by unauthorized member (matching Java delete_prayer_by_member_from_different_room_then_return_404_not_found)
func (suite *PrayerDeleteIntegrationTestSuite) TestDeletePrayerByMemberFromDifferentRoom() {
	// Given
	// 새로운 회원 생성
	anotherMember := suite.testUtils.CreateUniqueMember()
	// memberRepository.save(anotherMember) - already saved in CreateUniqueMember

	// 새로운 회원의 인증 헤더 생성
	anotherHeaders := suite.testUtils.CreateAuthHeaderWithMember(anotherMember)
	deleteURL := fmt.Sprintf("%s/%d", suite.PrayersAPIURL, suite.prayerTitle.ID)

	// When
	w := suite.DeleteRequest(deleteURL, anotherHeaders)

	// Then
	suite.AssertStatusCode(w, http.StatusNotFound, "다른 방의 회원이 기도 삭제 요청 시 404 Not Found가 반환되어야 합니다.")

	// ExceptionResponse 검증
	var exceptionResponse map[string]interface{}
	err := suite.UnmarshalResponse(w, &exceptionResponse)
	assert.NoError(suite.T(), err, "응답을 파싱할 수 있어야 합니다")
	assert.NotNil(suite.T(), exceptionResponse, "예외 응답이 있어야 합니다")
}

// TestPrayerDeleteIntegration runs the prayer delete integration test suite
func TestPrayerDeleteIntegration(t *testing.T) {
	suite.Run(t, new(PrayerDeleteIntegrationTestSuite))
}
