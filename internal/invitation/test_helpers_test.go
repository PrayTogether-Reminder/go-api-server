package invitation_test

import (
	testing "testing"

	testutil2 "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/testutil"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/invitation"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/member"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/room"
	"gorm.io/gorm"
)

// setupInvitationTestEnvironment wires dependencies needed for invitation handler tests.
func setupInvitationTestEnvironment(t *testing.T) (*invitation.InvitationHandler, *gorm.DB, *model.Member) {
	t.Helper()

	db := testutil2.SetupTestDB(t)
	t.Cleanup(func() {
		testutil2.CleanupTestDB(t, db)
	})

	inviter := testutil2.CreateTestMember(t, db)

	memberRepo := member.NewMemberRepository()
	memberService := member.NewMemberService(memberRepo)
	roomRepo := room.NewRoomRepository()
	memberRoomRepo := room.NewMemberRoomRepository()
	roomService := room.NewRoomService(roomRepo, memberRoomRepo, memberService)

	invitationRepo := invitation.NewInvitationRepository()
	invitationService := invitation.NewInvitationService(invitationRepo)
	invitationUseCase := invitation.NewInvitationUseCase(db, invitationService, roomService, memberService)
	invitationHandler := invitation.NewInvitationHandler(invitationUseCase)

	return invitationHandler, db, inviter
}
