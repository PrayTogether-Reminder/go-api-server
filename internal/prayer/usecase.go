package prayer

import (
	"context"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/member"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/room"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/database"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
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
		createdTitle, err := u.prayerService.CreateTitle(ctx, tx, prayerRoom, request.Title)
		if err != nil {
			return err
		}

		prayerTitle = createdTitle
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &CreatePrayerTitleResponse{
		ID:          prayerTitle.ID,
		Title:       prayerTitle.Title,
		CreatedTime: prayerTitle.CreatedAt,
	}, nil
}

// FetchTitlesByInfiniteScroll retrieves prayer titles for infinite scroll
func (u *PrayerUseCase) FetchTitlesByInfiniteScroll(ctx context.Context, memberID int64, request *PrayerTitleInfiniteScrollRequest) (*PrayerTitleInfiniteScrollResponse, error) {
	var titleInfos []PrayerTitleInfo

	err := database.WithTransaction(ctx, u.db, func(tx *gorm.DB) error {
		if err := u.roomService.ValidateMemberExistInRoom(ctx, tx, memberID, request.RoomID); err != nil {
			return err
		}

		titles, err := u.prayerService.GetTitleInfosByRoom(ctx, tx, request.RoomID, request.After)
		if err != nil {
			return err
		}

		titleInfos = titles
		return nil
	})

	if err != nil {
		return nil, err
	}

	if titleInfos == nil {
		titleInfos = make([]PrayerTitleInfo, 0)
	}

	return &PrayerTitleInfiniteScrollResponse{PrayerTitles: titleInfos}, nil
}

// validateMemberExistInRoomByTitleId validates if the member exists in the room associated with the prayer title
func (u *PrayerUseCase) validateMemberExistInRoomByTitleId(ctx context.Context, db *gorm.DB, memberID int64, titleID int64) error {
	// 기도 제목 조회
	prayerTitle, err := u.prayerService.GetTitleById(ctx, db, titleID)
	if err != nil {
		return err
	}

	// 멤버가 해당 방에 속하는지 검증
	return u.roomService.ValidateMemberExistInRoom(ctx, db, memberID, prayerTitle.RoomID)
}

// CreatePrayerContent creates a new prayer content
func (u *PrayerUseCase) CreatePrayerContent(
	ctx context.Context,
	writerID int64,
	titleID int64,
	request *CreatePrayerContentRequest,
) (*sharedHttp.MessageResponse, error) {
	var result *sharedHttp.MessageResponse

	err := database.WithTransaction(ctx, u.db, func(tx *gorm.DB) error {
		// 1. 멤버 권한 검증
		if err := u.validateMemberExistInRoomByTitleId(ctx, tx, writerID, titleID); err != nil {
			return err
		}

		// 2. 기도 제목 조회
		prayerTitle, err := u.prayerService.GetTitleById(ctx, tx, titleID)
		if err != nil {
			return err
		}

		// 3. 작성자(Writer) 조회
		writer, err := u.memberService.GetByID(ctx, tx, writerID)
		if err != nil {
			return err
		}

		// 4. 기도 내용 생성 및 저장
		if err := u.prayerService.CreateContent(ctx, tx, prayerTitle, writer, request); err != nil {
			return err
		}

		result = &sharedHttp.MessageResponse{
			Message: "기도 내용을 생성했습니다.",
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// FetchPrayerContents retrieves all prayer contents for a specific prayer title
func (u *PrayerUseCase) FetchPrayerContents(ctx context.Context, memberID int64, titleID int64) (*PrayerContentResponse, error) {
	var contentInfos []PrayerContentInfo

	err := database.WithTransaction(ctx, u.db, func(tx *gorm.DB) error {
		if err := u.validateMemberExistInRoomByTitleId(ctx, tx, memberID, titleID); err != nil {
			return err
		}

		contents, err := u.prayerService.GetContentsByTitleID(ctx, tx, titleID)
		if err != nil {
			return err
		}

		contentInfos = contents
		return nil
	})

	if err != nil {
		return nil, err
	}

	if contentInfos == nil {
		contentInfos = make([]PrayerContentInfo, 0)
	}

	return &PrayerContentResponse{PrayerContents: contentInfos}, nil
}

// UpdatePrayerTitle updates a prayer title
func (u *PrayerUseCase) UpdatePrayerTitle(
	ctx context.Context,
	memberID int64,
	titleID int64,
	request *UpdatePrayerTitleRequest,
) (*sharedHttp.MessageResponse, error) {
	err := database.WithTransaction(ctx, u.db, func(tx *gorm.DB) error {
		if err := u.validateMemberExistInRoomByTitleId(ctx, tx, memberID, titleID); err != nil {
			return err
		}

		if err := u.prayerService.UpdateTitle(ctx, tx, titleID, request.ChangedTitle); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &sharedHttp.MessageResponse{
		Message: "기도 제목을 변경했습니다.",
	}, nil
}

// UpdatePrayerContent updates a prayer content
func (u *PrayerUseCase) UpdatePrayerContent(
	ctx context.Context,
	memberID int64,
	titleID int64,
	contentID int64,
	request *UpdatePrayerContentRequest,
) (*sharedHttp.MessageResponse, error) {
	err := database.WithTransaction(ctx, u.db, func(tx *gorm.DB) error {
		if err := u.validateMemberExistInRoomByTitleId(ctx, tx, memberID, titleID); err != nil {
			return err
		}

		// 2. 기도 내용 존재 검증 (contentId가 titleId에 속하는지)
		if err := u.prayerService.ValidateContentExistsInTitle(ctx, tx, contentID, titleID); err != nil {
			return err
		}

		// 3. 기도 내용 업데이트
		if err := u.prayerService.UpdateContent(ctx, tx, contentID, request.ChangedContent); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &sharedHttp.MessageResponse{
		Message: "기도 내용을 변경했습니다.",
	}, nil
}
