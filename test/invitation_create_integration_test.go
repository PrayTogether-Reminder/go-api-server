package test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	invitationdomain "pray-together/internal/domains/invitation/domain"
	memberdomain "pray-together/internal/domains/member/domain"
	roomdomain "pray-together/internal/domains/room/domain"
)

// InvitationCreateIntegrationTestSuite tests invitation creation API (matching Java InvitationCreateIntegrateTest)
type InvitationCreateIntegrationTestSuite struct {
	IntegrationTestSuite

	memberInviter *memberdomain.Member
	memberInvitee *memberdomain.Member
	room          *roomdomain.Room
	headers       map[string]string
}

// SetupTest runs before each test (matching Java @BeforeEach setup)
func (suite *InvitationCreateIntegrationTestSuite) SetupTest() {
	// Call parent SetupTest
	suite.IntegrationTestSuite.SetupTest()

	// Create inviter member
	suite.memberInviter = suite.testUtils.CreateUniqueMember()
	// memberRepository.save(memberInviter) - already saved in CreateUniqueMember

	// Create invitee member
	suite.memberInvitee = suite.testUtils.CreateUniqueMember()
	// memberRepository.save(memberInvitee) - already saved in CreateUniqueMember

	// Create room
	suite.room = suite.testUtils.CreateUniqueRoom()
	// roomRepository.save(room) - already saved in CreateUniqueRoom

	// Create member-room relationship for inviter
	memberRoom := suite.testUtils.CreateUniqueMemberRoomWithMemberAndRoom(suite.memberInviter, suite.room)
	_ = memberRoom // memberRoomRepository.save(memberRoom) - already saved

	// Set up router with invitation routes
}

// TearDownTest runs after each test (matching Java @AfterEach cleanup)
func (suite *InvitationCreateIntegrationTestSuite) TearDownTest() {
	suite.CleanRepository()
}

// TestInviteMemberToRoom tests invitation creation (matching Java invite_member_to_room_then_return_201_created)
func (suite *InvitationCreateIntegrationTestSuite) TestInviteMemberToRoom() {
	// Given
	suite.headers = suite.testUtils.CreateAuthHeaderWithMember(suite.memberInviter)

	request := InvitationCreateRequest{
		RoomID: suite.room.ID,
		Email:  suite.memberInvitee.Email,
	}

	// When
	w := suite.PostJSON(suite.InvitationsAPIURL, request, suite.headers)

	// Then
	suite.AssertStatusCode(w, http.StatusCreated, "방 초대 API 응답 상태 코드가 201 Created가 아닙니다.")

	// Check saved invitations
	var allInvitations []invitationdomain.Invitation
	suite.db.Find(&allInvitations)
	assert.Equal(suite.T(), 1, len(allInvitations), "저장된 초대장의 개수가 예상과 다릅니다.")

	invitation := allInvitations[0]
	assert.Equal(suite.T(), suite.memberInviter.Name, invitation.InviterName, "초대자 이름이 예상과 다릅니다.")
	assert.Equal(suite.T(), suite.memberInvitee.ID, invitation.InviteeID, "초대받은 사용자 ID가 예상과 다릅니다.")
	assert.Equal(suite.T(), suite.room.ID, invitation.RoomID, "초대장에 연결된 방 ID가 예상과 다릅니다.")
	assert.Nil(suite.T(), invitation.ResponseTime, "초대장 응답 시간은 null이어야 합니다.")
	assert.Equal(suite.T(), invitationdomain.StatusPending, invitation.Status, "초대장 상태가 PENDING이 아닙니다.")
}

// InvitationCreateRequest represents the request to create an invitation (matching Java InvitationCreateRequest)
type InvitationCreateRequest struct {
	RoomID uint64 `json:"roomId" binding:"required"`
	Email  string `json:"email" binding:"required,email"`
}

// TestInvitationCreateIntegration runs the invitation create integration test suite
func TestInvitationCreateIntegration(t *testing.T) {
	suite.Run(t, new(InvitationCreateIntegrationTestSuite))
}
