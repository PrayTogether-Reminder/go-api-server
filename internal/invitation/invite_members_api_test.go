package invitation_test

import (
	"net/http"
	"testing"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/invitation"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInviteMembers_SuccessSkipsDuplicatesAndExistingPending(t *testing.T) {
	handler, db, inviter := setupInvitationTestEnvironment(t)
	room := testutil.CreateTestRoom(t, db, inviter.ID, "Test Room", "설명")
	inviteeOne := testutil.CreateTestMemberWithIndex(t, db, 100)
	inviteeTwo := testutil.CreateTestMemberWithIndex(t, db, 101)

	// Existing pending invitation for inviteeOne should be ignored
	existing := &model.Invitation{
		InviteeID:   inviteeOne.ID,
		RoomID:      room.ID,
		InviterName: "Existing",
		Status:      model.InvitationPending,
	}
	require.NoError(t, db.Create(existing).Error)

	router := testutil.SetupAuthenticatedRouter(inviter.ID)
	router.POST("/api/v2/invitations", handler.InviteMembers)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v2/invitations",
		Body: invitation.InviteMembersRequest{
			RoomID:    room.ID,
			MemberIDs: []int64{inviteeOne.ID, inviteeTwo.ID, inviteeTwo.ID},
		},
	})

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response sharedHttp.MessageResponse
	testutil.ParseResponse(t, recorder, &response)
	assert.Equal(t, "초대를 완료했습니다.", response.Message)

	var invitations []model.Invitation
	require.NoError(t, db.Where("room_id = ?", room.ID).Find(&invitations).Error)
	assert.Len(t, invitations, 2)

	var countOne int64
	require.NoError(t, db.Model(&model.Invitation{}).
		Where("room_id = ? AND invitee_id = ?", room.ID, inviteeOne.ID).
		Count(&countOne).Error)
	assert.Equal(t, int64(1), countOne)

	var countTwo int64
	require.NoError(t, db.Model(&model.Invitation{}).
		Where("room_id = ? AND invitee_id = ?", room.ID, inviteeTwo.ID).
		Count(&countTwo).Error)
	assert.Equal(t, int64(1), countTwo)
}

func TestInviteMembers_MemberNotFound(t *testing.T) {
	handler, db, inviter := setupInvitationTestEnvironment(t)
	room := testutil.CreateTestRoom(t, db, inviter.ID, "Test Room", "설명")
	invitee := testutil.CreateTestMemberWithIndex(t, db, 200)

	router := testutil.SetupAuthenticatedRouter(inviter.ID)
	router.POST("/api/v2/invitations", handler.InviteMembers)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v2/invitations",
		Body: invitation.InviteMembersRequest{
			RoomID:    room.ID,
			MemberIDs: []int64{invitee.ID, 999999},
		},
	})

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errorResponse)
	assert.Equal(t, "MEMBER-001", errorResponse.Code)
}

func TestInviteMembers_AlreadyMember(t *testing.T) {
	handler, db, inviter := setupInvitationTestEnvironment(t)
	room := testutil.CreateTestRoom(t, db, inviter.ID, "Test Room", "설명")
	members := testutil.AddMembersToRoom(t, db, room.ID, 1)
	existing := members[0]

	router := testutil.SetupAuthenticatedRouter(inviter.ID)
	router.POST("/api/v2/invitations", handler.InviteMembers)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v2/invitations",
		Body: invitation.InviteMembersRequest{
			RoomID:    room.ID,
			MemberIDs: []int64{existing.ID},
		},
	})

	assert.Equal(t, http.StatusConflict, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errorResponse)
	assert.Equal(t, "ROOM-006", errorResponse.Code)
}

func TestInviteMembers_Unauthorized(t *testing.T) {
	handler, db, inviter := setupInvitationTestEnvironment(t)
	room := testutil.CreateTestRoom(t, db, inviter.ID, "Test Room", "설명")
	invitee := testutil.CreateTestMemberWithIndex(t, db, 300)

	router := testutil.SetupTestRouter()
	router.POST("/api/v2/invitations", handler.InviteMembers)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v2/invitations",
		Body: invitation.InviteMembersRequest{
			RoomID:    room.ID,
			MemberIDs: []int64{invitee.ID},
		},
	})

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}
