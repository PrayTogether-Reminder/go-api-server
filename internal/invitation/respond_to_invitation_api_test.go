package invitation_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/invitation"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRespondToInvitation_AcceptSuccess(t *testing.T) {
	handler, db, inviter := setupInvitationTestEnvironment(t)
	room := testutil.CreateTestRoom(t, db, inviter.ID, "Room Accept", "방 설명")
	invitee := testutil.CreateTestMemberWithIndex(t, db, 500)

	pending := &model.Invitation{
		InviterName: inviter.Name,
		InviteeID:   invitee.ID,
		RoomID:      room.ID,
		Status:      model.InvitationPending,
	}
	require.NoError(t, db.Create(pending).Error)

	router := testutil.SetupAuthenticatedRouter(invitee.ID)
	router.PATCH("/api/v1/invitations/:invitationId", handler.RespondToInvitation)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodPatch,
		URL:    fmt.Sprintf("/api/v1/invitations/%d", pending.ID),
		Body: invitation.InvitationStatusUpdateRequest{
			Status: string(model.InvitationAccepted),
		},
	})

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response sharedHttp.MessageResponse
	testutil.ParseResponse(t, recorder, &response)
	assert.Equal(t, "기도방 초대를 수락했습니다.", response.Message)

	var updated model.Invitation
	require.NoError(t, db.First(&updated, pending.ID).Error)
	assert.Equal(t, model.InvitationAccepted, updated.Status)
	assert.NotNil(t, updated.ResponseTime)

	var memberRoom model.MemberRoom
	require.NoError(t, db.Where("member_id = ? AND room_id = ?", invitee.ID, room.ID).First(&memberRoom).Error)
	assert.Equal(t, model.RoomRoleMember, memberRoom.Role)
}

func TestRespondToInvitation_RejectSuccess(t *testing.T) {
	handler, db, inviter := setupInvitationTestEnvironment(t)
	room := testutil.CreateTestRoom(t, db, inviter.ID, "Room Reject", "방 설명")
	invitee := testutil.CreateTestMemberWithIndex(t, db, 510)

	pending := &model.Invitation{
		InviterName: inviter.Name,
		InviteeID:   invitee.ID,
		RoomID:      room.ID,
		Status:      model.InvitationPending,
	}
	require.NoError(t, db.Create(pending).Error)

	router := testutil.SetupAuthenticatedRouter(invitee.ID)
	router.PATCH("/api/v1/invitations/:invitationId", handler.RespondToInvitation)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodPatch,
		URL:    fmt.Sprintf("/api/v1/invitations/%d", pending.ID),
		Body: invitation.InvitationStatusUpdateRequest{
			Status: string(model.InvitationRejected),
		},
	})

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response sharedHttp.MessageResponse
	testutil.ParseResponse(t, recorder, &response)
	assert.Equal(t, "기도방 초대를 거절했습니다.", response.Message)

	var updated model.Invitation
	require.NoError(t, db.First(&updated, pending.ID).Error)
	assert.Equal(t, model.InvitationRejected, updated.Status)
	assert.NotNil(t, updated.ResponseTime)

	var memberRoom model.MemberRoom
	err := db.Where("member_id = ? AND room_id = ?", invitee.ID, room.ID).First(&memberRoom).Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestRespondToInvitation_NotFound(t *testing.T) {
	handler, db, _ := setupInvitationTestEnvironment(t)
	invitee := testutil.CreateTestMemberWithIndex(t, db, 520)

	router := testutil.SetupAuthenticatedRouter(invitee.ID)
	router.PATCH("/api/v1/invitations/:invitationId", handler.RespondToInvitation)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodPatch,
		URL:    "/api/v1/invitations/999999",
		Body: invitation.InvitationStatusUpdateRequest{
			Status: string(model.InvitationAccepted),
		},
	})

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errorResponse)
	assert.Equal(t, "INVITATION-001", errorResponse.Code)
}

func TestRespondToInvitation_Unauthorized(t *testing.T) {
	handler, db, inviter := setupInvitationTestEnvironment(t)
	room := testutil.CreateTestRoom(t, db, inviter.ID, "Room Unauthorized", "방 설명")
	invitee := testutil.CreateTestMemberWithIndex(t, db, 530)

	pending := &model.Invitation{
		InviterName: inviter.Name,
		InviteeID:   invitee.ID,
		RoomID:      room.ID,
		Status:      model.InvitationPending,
	}
	require.NoError(t, db.Create(pending).Error)

	router := testutil.SetupTestRouter()
	router.PATCH("/api/v1/invitations/:invitationId", handler.RespondToInvitation)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodPatch,
		URL:    fmt.Sprintf("/api/v1/invitations/%d", pending.ID),
		Body: invitation.InvitationStatusUpdateRequest{
			Status: string(model.InvitationAccepted),
		},
	})

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestRespondToInvitation_AlreadyResponded(t *testing.T) {
	handler, db, inviter := setupInvitationTestEnvironment(t)
	room := testutil.CreateTestRoom(t, db, inviter.ID, "Room Already Responded", "방 설명")
	invitee := testutil.CreateTestMemberWithIndex(t, db, 540)

	now := time.Now()
	accepted := &model.Invitation{
		InviterName:  inviter.Name,
		InviteeID:    invitee.ID,
		RoomID:       room.ID,
		Status:       model.InvitationAccepted,
		ResponseTime: &now,
	}
	require.NoError(t, db.Create(accepted).Error)

	router := testutil.SetupAuthenticatedRouter(invitee.ID)
	router.PATCH("/api/v1/invitations/:invitationId", handler.RespondToInvitation)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodPatch,
		URL:    fmt.Sprintf("/api/v1/invitations/%d", accepted.ID),
		Body: invitation.InvitationStatusUpdateRequest{
			Status: string(model.InvitationAccepted),
		},
	})

	assert.Equal(t, http.StatusConflict, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errorResponse)
	assert.Equal(t, "INVITATION-002", errorResponse.Code)
}
