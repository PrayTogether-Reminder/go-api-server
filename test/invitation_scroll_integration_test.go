package test

import (
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	invitationdomain "pray-together/internal/domains/invitation/domain"
	memberdomain "pray-together/internal/domains/member/domain"
)

// InvitationScrollIntegrationTestSuite tests invitation scroll API (matching Java InvitationScrollIntegrateTest)
type InvitationScrollIntegrationTestSuite struct {
	IntegrationTestSuite

	inviteeMember   *memberdomain.Member
	headers         map[string]string
	invitationCount int
}

// SetupTest runs before each test (matching Java @BeforeEach setup)
func (suite *InvitationScrollIntegrationTestSuite) SetupTest() {
	// Call parent SetupTest
	suite.IntegrationTestSuite.SetupTest()

	suite.invitationCount = 5

	// 초대 받을 회원 생성
	suite.inviteeMember = suite.testUtils.CreateUniqueMember()
	// memberRepository.save(inviteeMember) - already saved in CreateUniqueMember
	suite.headers = suite.testUtils.CreateAuthHeaderWithMember(suite.inviteeMember)

	// 초대한 회원들 생성 및 초대장 생성
	for i := 0; i < suite.invitationCount; i++ {
		// 초대자 생성
		inviterMember := suite.testUtils.CreateUniqueMember()
		// memberRepository.save(inviterMember) - already saved in CreateUniqueMember

		// 방 생성
		room := suite.testUtils.CreateUniqueRoom()
		// roomRepository.save(room) - already saved in CreateUniqueRoom

		// 초대장 생성 - Invitation.create(inviterMember, inviteeMember, room)
		invitation := &invitationdomain.Invitation{
			InviterID:   inviterMember.ID,
			InviterName: inviterMember.Name,
			InviteeID:   suite.inviteeMember.ID,
			RoomID:      room.ID,
			Status:      invitationdomain.StatusPending,
			ExpiresAt:   time.Now().Add(7 * 24 * time.Hour), // 7 days expiry
			BaseEntity: invitationdomain.BaseEntity{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}
		suite.db.Create(invitation)

		// Add small delay to ensure different timestamps for sorting test
		time.Sleep(time.Millisecond)
	}

	// Set up router with invitation routes
}

// TearDownTest runs after each test (matching Java @AfterEach cleanup)
func (suite *InvitationScrollIntegrationTestSuite) TearDownTest() {
	suite.CleanRepository()
}

// TestFetchInvitationScrollThenReturn200OK tests fetching invitation scroll (matching Java fetch_invitation_scroll_then_return_200_ok)
func (suite *InvitationScrollIntegrationTestSuite) TestFetchInvitationScrollThenReturn200OK() {
	// Given
	// Request entity is created with headers in the actual request

	// When
	w := suite.GetRequest(suite.InvitationsAPIURL, suite.headers)

	// Then
	suite.AssertStatusCode(w, http.StatusOK, "초대 목록 조회 API 응답 상태 코드가 200 OK가 아닙니다.")

	var response InvitationInfoScrollResponse
	err := suite.UnmarshalResponse(w, &response)
	assert.NoError(suite.T(), err, "응답을 파싱할 수 있어야 합니다")
	assert.NotNil(suite.T(), response, "초대 목록 조회 API 응답 결과가 NULL 입니다.")

	invitations := response.Invitations
	assert.NotNil(suite.T(), invitations, "초대 목록 조회 API 응답 결과 데이터가 NULL 입니다.")
	assert.Equal(suite.T(), suite.invitationCount, len(invitations),
		"초대 목록 조회 API 응답 결과 개수가 기대값과 다릅니다.")

	// Check if sorted by createdTime in ascending order
	isSorted := sort.SliceIsSorted(invitations, func(i, j int) bool {
		return invitations[i].CreatedTime.Before(invitations[j].CreatedTime)
	})
	assert.True(suite.T(), isSorted, "초대 목록이 createdTime 기준으로 오름차순 정렬되지 않았습니다.")

	// 초대 항목 데이터 검증
	for _, invitation := range invitations {
		assert.NotZero(suite.T(), invitation.InvitationID, "초대 ID가 NULL 입니다.")
		assert.NotEmpty(suite.T(), invitation.InviterName, "초대자 이름이 NULL 입니다.")
		assert.NotEmpty(suite.T(), invitation.RoomName, "방 이름이 NULL 입니다.")
		assert.NotEmpty(suite.T(), invitation.RoomDescription, "방 설명이 NULL 입니다.")
		assert.NotZero(suite.T(), invitation.CreatedTime, "초대 생성 시간이 NULL 입니다.")
	}
}

// InvitationInfo represents invitation info (matching Java InvitationInfo)
type InvitationInfo struct {
	InvitationID    uint64    `json:"invitationId"`
	InviterName     string    `json:"inviterName"`
	RoomName        string    `json:"roomName"`
	RoomDescription string    `json:"roomDescription"`
	CreatedTime     time.Time `json:"createdTime"`
}

// InvitationInfoScrollResponse represents invitation scroll response (matching Java InvitationInfoScrollResponse)
type InvitationInfoScrollResponse struct {
	Invitations []InvitationInfo `json:"invitations"`
}

// TestInvitationScrollIntegration runs the invitation scroll integration test suite
func TestInvitationScrollIntegration(t *testing.T) {
	suite.Run(t, new(InvitationScrollIntegrationTestSuite))
}
