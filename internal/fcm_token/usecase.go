package fcm_token

import (
	"context"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/database"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/member"
	"gorm.io/gorm"
)

// FcmTokenUseCase coordinates validation and transactional behavior for FCM token APIs.
type FcmTokenUseCase struct {
	db              *gorm.DB
	memberService   *member.MemberService
	fcmTokenService *FcmTokenService
}

// NewFcmTokenUseCase constructs a new use case instance.
func NewFcmTokenUseCase(db *gorm.DB, memberService *member.MemberService, fcmTokenService *FcmTokenService) *FcmTokenUseCase {
	return &FcmTokenUseCase{
		db:              db,
		memberService:   memberService,
		fcmTokenService: fcmTokenService,
	}
}

// RegisterFcmToken registers a new token for the authenticated member.
func (u *FcmTokenUseCase) RegisterFcmToken(ctx context.Context, memberID int64, token string) error {
	return database.WithTransaction(ctx, u.db, func(tx *gorm.DB) error {
		if err := u.memberService.ValidateMemberExists(ctx, tx, memberID); err != nil {
			return err
		}

		if err := u.fcmTokenService.RegisterToken(ctx, tx, memberID, token); err != nil {
			return err
		}
		return nil
	})
}

// DeleteFcmToken removes a member's token matching the provided value.
func (u *FcmTokenUseCase) DeleteFcmToken(ctx context.Context, memberID int64, token string) error {
	return database.WithTransaction(ctx, u.db, func(tx *gorm.DB) error {
		if err := u.fcmTokenService.DeleteByMemberID(ctx, tx, token, memberID); err != nil {
			return err
		}
		return nil
	})
}
