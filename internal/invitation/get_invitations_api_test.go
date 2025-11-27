package invitation_test

import (
	"net/http"
	"testing"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/invitation"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetInvitations_ReturnsPendingOnly(t *testing.T) {
	invitationHandler, db, inviter := setupInvitationTestEnvironment(t)
	invitee := testutil.CreateTestMemberWithIndex(t, db, 400)
	room := testutil.CreateTestRoom(t, db, inviter.ID, "Room", "설명")

	pending := &model.Invitation{
		InviterName: "초대한 사람",
		InviteeID:   invitee.ID,
		RoomID:      room.ID,
		Status:      model.InvitationPending,
	}
	require.NoError(t, db.Create(pending).Error)

	accepted := &model.Invitation{
		InviterName: "다른 사람",
		InviteeID:   invitee.ID,
		RoomID:      room.ID,
		Status:      model.InvitationAccepted,
	}
	require.NoError(t, db.Create(accepted).Error)

	router := testutil.SetupAuthenticatedRouter(invitee.ID)
	router.GET("/api/v1/invitations", invitationHandler.GetInvitations)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodGet,
		URL:    "/api/v1/invitations",
	})

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response invitation.InvitationInfoScrollResponse
	testutil.ParseResponse(t, recorder, &response)
	if assert.Len(t, response.Invitations, 1) {
		info := response.Invitations[0]
		assert.Equal(t, pending.ID, info.InvitationID)
		assert.Equal(t, pending.InviterName, info.InviterName)
		assert.Equal(t, room.Name, info.RoomName)
		assert.Equal(t, room.Description, info.RoomDescription)
	}
}

func TestGetInvitations_Unauthorized(t *testing.T) {
	invitationHandler, _, _ := setupInvitationTestEnvironment(t)
	router := testutil.SetupTestRouter()
	router.GET("/api/v1/invitations", invitationHandler.GetInvitations)

	recorder := testutil.ExecuteRequest(t, router, testutil.TestRequest{
		Method: http.MethodGet,
		URL:    "/api/v1/invitations",
	})

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}
