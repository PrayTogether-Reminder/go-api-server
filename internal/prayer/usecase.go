package prayer

import (
	"context"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/member"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/room"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/database"
	"gorm.io/gorm"
)

type PrayerUseCase struct {
	db            *gorm.DB
	prayerService *PrayerService
	roomService   *room.RoomService
	memberService member.MemberService
}

func NewPrayerUseCase(db *gorm.DB, prayerService *PrayerService, roomService *room.RoomService, memberService *member.MemberService) *PrayerUseCase {
	return &PrayerUseCase{
		db:            db,
		prayerService: prayerService,
		roomService:   roomService,
		memberService: *memberService,
	}
}

func (u *PrayerUseCase) CreatePrayerTitle(ctx context.Context, memberID int64, request *CreatePrayerTitleRequest) (*CreatePrayerTitleResponse, error) {
	var prayerTitle *model.PrayerTitle
	err := database.WithTransaction(ctx, u.db, func(tx *gorm.DB) error {
		// 방 존재 여부 확인
		prayerRoom, err := u.roomService.GetRoomByID(ctx, tx, request.RoomID)
		if err != nil {
			return err
		}

		// 회원이 방에 속해있는지 검증
		err = u.roomService.ValidateMemberExistInRoom(ctx, tx, memberID, prayerRoom.ID)
		if err != nil {
			return err
		}

		// Service에 위임하여 기도제목 생성 및 저장
		createdTitle, err := u.prayerService.CreatePrayerTitle(ctx, tx, prayerRoom, request.Title)
		if err != nil {
			return err
		}

		prayerTitle = createdTitle
		return nil
	})

	return &CreatePrayerTitleResponse{
		ID:          prayerTitle.ID,
		Title:       prayerTitle.Title,
		CreatedTime: prayerTitle.CreatedAt,
	}, err
}
