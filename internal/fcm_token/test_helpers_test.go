package fcm_token_test

import (
	"testing"

	testutil "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/testutil"
	fcmtoken "github.com/changhyeonkim/pray-together/go-api-server/internal/fcm_token"
	memberpkg "github.com/changhyeonkim/pray-together/go-api-server/internal/member"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"gorm.io/gorm"
)

// setupFcmTokenTestEnvironment wires dependencies for FCM token handler tests.
func setupFcmTokenTestEnvironment(t *testing.T) (*fcmtoken.Handler, *gorm.DB, *model.Member) {
	t.Helper()

	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestDB(t, db)
	})

	member := testutil.CreateTestMember(t, db)

	memberRepo := memberpkg.NewMemberRepository()
	memberService := memberpkg.NewMemberService(memberRepo)

	fcmRepo := fcmtoken.NewFcmTokenRepository()
	fcmService := fcmtoken.NewFcmTokenService(fcmRepo)
	useCase := fcmtoken.NewFcmTokenUseCase(db, memberService, fcmService)

	handler := fcmtoken.NewHandler(useCase)

	return handler, db, member
}
