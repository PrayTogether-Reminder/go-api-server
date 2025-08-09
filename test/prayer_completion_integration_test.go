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

// PrayerCompletionIntegrationTestSuite tests prayer completion API (matching Java PrayerCompletionIntegrateTest)
type PrayerCompletionIntegrationTestSuite struct {
	IntegrationTestSuite

	member              *memberdomain.Member
	headers             map[string]string
	room                *roomdomain.Room
	prayerTitle         *prayerdomain.PrayerTitle
	completionURLFormat string

	// 추가 회원 생성
	additionalMembersCount int
	additionalMembers      []*memberdomain.Member
}

// SetupTest runs before each test (matching Java @BeforeEach setup)
func (suite *PrayerCompletionIntegrationTestSuite) SetupTest() {
	// Call parent SetupTest
	suite.IntegrationTestSuite.SetupTest()

	suite.completionURLFormat = "/%d/completion"
	suite.additionalMembersCount = 3

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

	// 추가 회원 생성 및 방에 참여시키기
	suite.additionalMembers = make([]*memberdomain.Member, suite.additionalMembersCount)
	for i := 0; i < suite.additionalMembersCount; i++ {
		suite.additionalMembers[i] = suite.testUtils.CreateUniqueMember()
		// memberRepository.save(additionalMembers[i]) - already saved in CreateUniqueMember

		additionalMemberRoom := suite.testUtils.CreateUniqueMemberRoomWithMemberAndRoom(
			suite.additionalMembers[i], suite.room)
		suite.db.Create(additionalMemberRoom)
	}

	// 인증 헤더 생성
	suite.headers = suite.testUtils.CreateAuthHeaderWithMember(suite.member)

	// Set up router with prayer routes
}

// TearDownTest runs after each test (matching Java @AfterEach cleanup)
func (suite *PrayerCompletionIntegrationTestSuite) TearDownTest() {
	suite.CleanRepository()
}

// TestCompletePrayerThenCreateNotificationsAndReturn200OK tests prayer completion (matching Java complete_prayer_then_create_notifications_and_return_200_ok)
func (suite *PrayerCompletionIntegrationTestSuite) TestCompletePrayerThenCreateNotificationsAndReturn200OK() {
	// Given
	uri := fmt.Sprintf(suite.PrayersAPIURL+suite.completionURLFormat, suite.prayerTitle.ID)

	request := PrayerCompletionCreateRequest{
		RoomID: suite.room.ID,
	}

	// When
	w := suite.PostJSON(uri, request, suite.headers)

	// Then
	// 응답 검증
	suite.AssertStatusCode(w, http.StatusOK, "기도 완료 처리 API 응답 상태 코드가 200 OK이 아닙니다.")

	var response MessageResponse
	err := suite.UnmarshalResponse(w, &response)
	assert.NoError(suite.T(), err, "응답을 파싱할 수 있어야 합니다")
	assert.NotNil(suite.T(), response, "기도 완료 처리 API 응답 결과가 NULL 입니다.")

	// 기도 완료 엔티티 생성 검증
	var completions []prayerdomain.PrayerCompletion
	suite.db.Find(&completions)
	assert.NotEmpty(suite.T(), completions, "기도 완료 정보가 저장되지 않았습니다.")
	assert.Equal(suite.T(), 1, len(completions), "기도 완료 정보 개수가 예상과 다릅니다.")

	completion := completions[0]
	assert.Equal(suite.T(), suite.member.ID, completion.MemberID,
		"기도 완료 정보의 기도자 ID가 예상과 다릅니다.")
	assert.Equal(suite.T(), suite.prayerTitle.ID, completion.PrayerTitleID,
		"기도 완료 정보의 기도 제목 ID가 예상과 다릅니다.")

	// 알림 생성 검증
	var notifications []PrayerCompletionNotification
	suite.db.Find(&notifications)
	assert.NotEmpty(suite.T(), notifications, "기도 완료 알림이 생성되지 않았습니다.")
	assert.Equal(suite.T(), suite.additionalMembersCount, len(notifications),
		"생성된 알림 개수가 예상과 다릅니다. (알림은 자신을 제외한 다른 멤버들에게만 전송됨)")

	for _, notification := range notifications {
		assert.Equal(suite.T(), suite.member.ID, notification.SenderID,
			"알림의 발신자 ID가 예상과 다릅니다.")
		assert.Equal(suite.T(), suite.prayerTitle.ID, notification.PrayerTitleID,
			"알림의 기도 제목 ID가 예상과 다릅니다.")
		assert.NotEmpty(suite.T(), notification.Message, "알림의 메시지가 NULL입니다.")

		// 알림 메시지 형식 검증
		expectedMessage := fmt.Sprintf(NotificationMessageFormatPrayerCompletion,
			suite.member.Name, suite.prayerTitle.Title)
		assert.Equal(suite.T(), expectedMessage, notification.Message,
			"알림 메시지가 예상 형식과 다릅니다.")
	}
}

// PrayerCompletionCreateRequest represents prayer completion create request (matching Java PrayerCompletionCreateRequest)
type PrayerCompletionCreateRequest struct {
	RoomID uint64 `json:"roomId"`
}

// NotificationMessageFormat constants (matching Java NotificationMessageFormat)
const (
	NotificationMessageFormatPrayerCompletion = "%s님이 %s 기도를 완료했습니다."
)

// TestPrayerCompletionIntegration runs the prayer completion integration test suite
func TestPrayerCompletionIntegration(t *testing.T) {
	suite.Run(t, new(PrayerCompletionIntegrationTestSuite))
}
