package test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	invitationdomain "pray-together/internal/domains/invitation/domain"
	memberdomain "pray-together/internal/domains/member/domain"
	roomdomain "pray-together/internal/domains/room/domain"
)

// InvitationUpdateStatusIntegrationTestSuite tests invitation status update API (matching Java InvitationUpdateStatusIntegrateTest)
type InvitationUpdateStatusIntegrationTestSuite struct {
	IntegrationTestSuite

	memberInviter *memberdomain.Member
	memberInvitee *memberdomain.Member
	room          *roomdomain.Room
	headers       map[string]string
	invitation    *invitationdomain.Invitation
}

// SetupTest runs before each test (matching Java @BeforeEach setup)
func (suite *InvitationUpdateStatusIntegrationTestSuite) SetupTest() {
	// Call parent SetupTest
	suite.IntegrationTestSuite.SetupTest()

	// 초대자 회원 생성
	suite.memberInviter = suite.testUtils.CreateUniqueMember()
	// memberRepository.save(memberInviter) - already saved in CreateUniqueMember

	// 초대받는 회원 생성
	suite.memberInvitee = suite.testUtils.CreateUniqueMember()
	// memberRepository.save(memberInvitee) - already saved in CreateUniqueMember

	// 방 생성
	suite.room = suite.testUtils.CreateUniqueRoom()
	// roomRepository.save(room) - already saved in CreateUniqueRoom

	// 초대자-방 연관관계 생성
	memberRoom := suite.testUtils.CreateUniqueMemberRoomWithMemberAndRoom(suite.memberInviter, suite.room)
	suite.db.Create(memberRoom)

	// 초대장 생성 - Invitation.create(memberInviter, memberInvitee, room)
	suite.invitation = &invitationdomain.Invitation{
		InviterID:   suite.memberInviter.ID,
		InviterName: suite.memberInviter.Name,
		InviteeID:   suite.memberInvitee.ID,
		RoomID:      suite.room.ID,
		Status:      invitationdomain.StatusPending,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour), // 7 days expiry
		BaseEntity: invitationdomain.BaseEntity{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	suite.db.Create(suite.invitation)

	// 생성된 초대장 조회
	var invitations []invitationdomain.Invitation
	suite.db.Find(&invitations)
	suite.invitation = &invitations[0]

	// 테스트용 인증 헤더 설정
	suite.headers = suite.testUtils.CreateAuthHeaderWithMember(suite.memberInvitee)

	// Set up router with invitation routes
}

// TearDownTest runs after each test (matching Java @AfterEach cleanup)
func (suite *InvitationUpdateStatusIntegrationTestSuite) TearDownTest() {
	suite.CleanRepository()
}

// TestUpdateInvitationStatusThenReturn200OK tests valid status updates (matching Java update_invitation_status_then_return_200_ok)
func (suite *InvitationUpdateStatusIntegrationTestSuite) TestUpdateInvitationStatusThenReturn200OK() {
	testCases := []struct {
		name          string
		updatedStatus invitationdomain.InvitationStatus
		koreanName    string
	}{
		{"수락", invitationdomain.StatusAccepted, "수락"},
		{"거절", invitationdomain.StatusRejected, "거절"},
	}

	for _, tc := range testCases {
		// Reset database for each test case
		suite.TearDownTest()
		suite.SetupTest()

		suite.T().Run(tc.name, func(t *testing.T) {
			// Given
			request := InvitationStatusUpdateRequest{
				Status: tc.updatedStatus,
			}

			url := fmt.Sprintf("%s/%d", suite.InvitationsAPIURL, suite.invitation.ID)

			var memberRoomCnt int64
			suite.db.Model(&roomdomain.RoomMember{}).Count(&memberRoomCnt)

			// Debug: Check if invitee is already in room
			var existingMember roomdomain.RoomMember
			checkErr := suite.db.Where("room_id = ? AND member_id = ?", suite.room.ID, suite.memberInvitee.ID).First(&existingMember).Error
			if checkErr == nil {
				t.Logf("WARNING: Invitee %d is already in room %d before accepting invitation", suite.memberInvitee.ID, suite.room.ID)
			}

			// When
			w := suite.PatchJSON(url, request, suite.headers)

			// Then
			suite.AssertStatusCode(w, http.StatusOK, "초대장 상태 변경 API 응답 상태 코드가 200 OK가 아닙니다.")

			// DB에서 초대장 확인
			var updatedInvitation invitationdomain.Invitation
			err := suite.db.First(&updatedInvitation, suite.invitation.ID).Error
			assert.NoError(t, err)

			// 상태 확인
			assert.Equal(t, tc.updatedStatus, updatedInvitation.Status,
				"초대장 상태가 변경되지 않았습니다.")

			// 응답 시간 확인
			assert.NotNil(t, updatedInvitation.ResponseTime,
				"초대장 응답 시간이 설정되지 않았습니다.")

			// 메시지 응답 확인
			var response MessageResponse
			err = suite.UnmarshalResponse(w, &response)
			assert.NoError(t, err)
			assert.NotNil(t, response, "초대장 상태 변경 API 응답 결과가 NULL 입니다.")
			assert.Contains(t, response.Message, tc.koreanName,
				"초대장 상태 변경 API 응답 메시지가 예상과 다릅니다.")

			// 방 참가 인원 수 확인
			var allMemberRoomCount int64
			suite.db.Model(&roomdomain.RoomMember{}).Count(&allMemberRoomCount)

			if tc.updatedStatus == invitationdomain.StatusAccepted {
				// 방 인원 증가
				assert.Equal(t, memberRoomCnt+1, allMemberRoomCount,
					"초대장 수락시 방 참가 인원이 달라져야(1 증가해야) 합니다.")
			} else if tc.updatedStatus == invitationdomain.StatusRejected {
				// 방 인원 유지
				assert.Equal(t, memberRoomCnt, allMemberRoomCount,
					"초대장 수락시 방 참가 인원이 이전과 동일해야 합니다.")
			}
		})
	}
}

// TestUpdateInvitationStatusWithInvalidInputThenReturn400BadRequest tests invalid status updates
func (suite *InvitationUpdateStatusIntegrationTestSuite) TestUpdateInvitationStatusWithInvalidInputThenReturn400BadRequest() {
	testCases := []struct {
		name          string
		invalidStatus interface{}
	}{
		{"status가 빈 문자열", ""},
		{"status가 소문자 accepted", "accepted"},
		{"status가 소문자 rejected", "rejected"},
		{"status가 존재하지 않는 값", "UNKNOWN_STATUS"},
		{"status가 숫자", 123},
		{"status가 불리언", true},
		{"status가 null", nil},
	}

	for _, tc := range testCases {
		// Reset database for each test case
		suite.TearDownTest()
		suite.SetupTest()

		suite.T().Run(tc.name, func(t *testing.T) {
			// Given
			requestMap := map[string]interface{}{
				"status": tc.invalidStatus,
			}

			var memberRoomCnt int64
			suite.db.Model(&roomdomain.RoomMember{}).Count(&memberRoomCnt)

			url := fmt.Sprintf("%s/%d", suite.InvitationsAPIURL, suite.invitation.ID)

			// When
			w := suite.PatchJSON(url, requestMap, suite.headers)

			// Then
			assert.Equal(t, http.StatusBadRequest, w.Code,
				"잘못된 초대장 상태 변경 요청 시 400 Bad Request가 반환되어야 합니다.")

			var response ExceptionResponse
			err := suite.UnmarshalResponse(w, &response)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, response.Status,
				"잘못된 초대장 상태 변경 요청 시 예외 응답에서 400 Bad Request가 반환되어야 합니다.")

			// DB에서 초대장 확인 - 상태가 변경되지 않아야 함
			var unchangedInvitation invitationdomain.Invitation
			err = suite.db.First(&unchangedInvitation, suite.invitation.ID).Error
			assert.NoError(t, err)

			// 상태가 PENDING 유지 확인
			assert.Equal(t, invitationdomain.StatusPending, unchangedInvitation.Status,
				"잘못된 요청에도 초대장 상태가 변경되었습니다.")

			// 응답 시간이 여전히 null인지 확인
			assert.Nil(t, unchangedInvitation.ResponseTime,
				"잘못된 요청에도 초대장 응답 시간이 설정되었습니다.")

			// 아무도 방에 초대되지 않음
			var allMemberRoomCount int64
			suite.db.Model(&roomdomain.RoomMember{}).Count(&allMemberRoomCount)
			assert.Equal(t, memberRoomCnt, allMemberRoomCount)
		})
	}
}

// TestUpdateAlreadyAcceptedInvitationStatusThenReturn400BadRequest tests updating already accepted invitation
func (suite *InvitationUpdateStatusIntegrationTestSuite) TestUpdateAlreadyAcceptedInvitationStatusThenReturn400BadRequest() {
	testCases := []struct {
		name          string
		updatedStatus invitationdomain.InvitationStatus
	}{
		{"수락", invitationdomain.StatusAccepted},
		{"거절", invitationdomain.StatusRejected},
	}

	for _, tc := range testCases {
		// Reset database for each test case
		suite.TearDownTest()
		suite.SetupTest()

		suite.T().Run(tc.name, func(t *testing.T) {
			// Given
			// invitation.accept()
			err := suite.invitation.Accept()
			assert.NoError(t, err)
			suite.db.Save(suite.invitation)

			// 이미 응답 시간이 설정되어 있는지 확인
			var acceptedInvitation invitationdomain.Invitation
			err = suite.db.First(&acceptedInvitation, suite.invitation.ID).Error
			assert.NoError(t, err)
			assert.NotNil(t, acceptedInvitation.ResponseTime,
				"초대장 응답 시간이 설정되지 않았습니다.")

			// 다시 상태 변경 요청
			request := InvitationStatusUpdateRequest{
				Status: tc.updatedStatus,
			}

			retryURL := fmt.Sprintf("%s/%d", suite.InvitationsAPIURL, suite.invitation.ID)

			// When
			w := suite.PatchJSON(retryURL, request, suite.headers)

			// Then
			assert.Equal(t, http.StatusBadRequest, w.Code,
				"이미 수락한 초대장 상태 변경 요청 시 400 Bad Request가 반환되어야 합니다.")

			var response ExceptionResponse
			err = suite.UnmarshalResponse(w, &response)
			assert.NoError(t, err)
			assert.NotNil(t, response, "예외 응답이 null입니다.")
			assert.Equal(t, http.StatusBadRequest, response.Status,
				"이미 수락한 초대장 상태 변경 요청 시 예외 응답에서 400 Bad Request가 반환되어야 합니다.")

			// 에러 메시지 확인
			assert.NotEmpty(t, response.Message,
				"이미 응답한 초대장에 대한 오류 메시지 Null 입니다..")

			// DB에서 초대장 확인 - 상태가 변경되지 않아야 함
			var unchangedInvitation invitationdomain.Invitation
			err = suite.db.First(&unchangedInvitation, suite.invitation.ID).Error
			assert.NoError(t, err)

			// 상태가 ACCEPTED 그대로 유지되는지 확인
			assert.Equal(t, invitationdomain.StatusAccepted, unchangedInvitation.Status,
				"이미 수락한 초대장의 상태가 변경되었습니다.")
		})
	}
}

// InvitationStatusUpdateRequest represents invitation status update request (matching Java InvitationStatusUpdateRequest)
type InvitationStatusUpdateRequest struct {
	Status invitationdomain.InvitationStatus `json:"status"`
}

// ExceptionResponse represents exception response (matching Java ExceptionResponse)
type ExceptionResponse struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// TestInvitationUpdateStatusIntegration runs the invitation update status integration test suite
func TestInvitationUpdateStatusIntegration(t *testing.T) {
	suite.Run(t, new(InvitationUpdateStatusIntegrationTestSuite))
}
