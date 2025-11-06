package member

import (
	"context"
	"fmt"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/database"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/logger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type MemberService interface {
	Signup(ctx context.Context, request *SignupRequest) error
}

type memberService struct {
	db               *gorm.DB
	memberRepository MemberRepository
}

func NewMemberService(db *gorm.DB, memberRepository MemberRepository) MemberService {
	return &memberService{
		db:               db,
		memberRepository: memberRepository,
	}
}

func (m *memberService) Signup(ctx context.Context, request *SignupRequest) error {
	log := logger.FromContext(ctx)
	return database.WithTransaction(ctx, m.db, func(tx *gorm.DB) error {
		exists, err := m.memberRepository.IsExist(ctx, tx, request.Email)
		if err != nil {
			log.Error("Failed to check member existence", "error", err)
			return fmt.Errorf("check member existence: %w", err)
		}
		if exists {
			log.Warn("Member already exists", "email", logger.MaskEmail(request.Email))
			return fmt.Errorf("error %w", ErrMemberAlreadyExists)
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Error("Failed to hash password", "error", err)
			return fmt.Errorf("hash password: %w", err)
		}

		member := model.NewMember(request.Name, request.Email, string(hashedPassword))
		if err := m.memberRepository.Create(ctx, tx, member); err != nil {
			log.Error("Failed to create member", "error", err)
			return fmt.Errorf("create member: %w", err)
		}

		log.Info("Member created successfully", "email", logger.MaskEmail(request.Email))
		return nil
	})
}
