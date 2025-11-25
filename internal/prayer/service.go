package prayer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"gorm.io/gorm"
)

type PrayerService struct {
	prayerRepository *PrayerRepository
}

func NewPrayerService(repo *PrayerRepository) *PrayerService {
	return &PrayerService{prayerRepository: repo}
}

// CreateTitle creates a new prayer title
func (s *PrayerService) CreateTitle(ctx context.Context, tx *gorm.DB, room *model.Room, title string) (*model.PrayerTitle, error) {
	prayerTitle := model.NewPrayerTitle(room, title)

	// Repository로 저장
	if err := s.prayerRepository.Create(ctx, tx, prayerTitle); err != nil {
		return nil, fmt.Errorf("기도제목 생성 실패: title=%s %s %w", title, ErrPrayerTitleCreateFailed, err)
	}

	return prayerTitle, nil
}

// GetTitleById fetches a prayer title by its ID
func (s *PrayerService) GetTitleById(ctx context.Context, db *gorm.DB, titleID int64) (*model.PrayerTitle, error) {
	prayerTitle, err := s.prayerRepository.FindTitleByID(ctx, db, titleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("기도 제목을 찾을 수 없습니다: TitleID=%d %w", titleID, ErrPrayerTitleNotFound)
		}
		return nil, fmt.Errorf("기도 제목 조회 실패: TitleID=%d %w", titleID, err)
	}
	return prayerTitle, nil
}

// GetTitleInRoom fetches a prayer title by ID and validates it belongs to the specified room
// Returns the title if it exists in the room, error otherwise
func (s *PrayerService) GetTitleInRoom(ctx context.Context, db *gorm.DB, titleID int64, roomID int64) (*model.PrayerTitle, error) {
	prayerTitle, err := s.prayerRepository.FindTitleByIDAndRoomID(ctx, db, titleID, roomID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("기도 제목을 찾을 수 없습니다: TitleID=%d, RoomID=%d %w", titleID, roomID, ErrPrayerTitleNotFound)
		}
		return nil, fmt.Errorf("기도 제목 조회 실패: TitleID=%d, RoomID=%d %w", titleID, roomID, err)
	}
	return prayerTitle, nil
}

// GetTitleInfosByRoom retrieves prayer titles in a room with cursor-based pagination
func (s *PrayerService) GetTitleInfosByRoom(ctx context.Context, db *gorm.DB, roomID int64, after string) ([]PrayerTitleInfo, error) {
	cursor := after
	if cursor == "" {
		cursor = DefaultPrayerTitleAfter
	}

	if cursor == DefaultPrayerTitleAfter {
		return s.prayerRepository.FindTitleInfosByRoomIDInitial(ctx, db, roomID, PrayerTitleInfiniteScrollLimit)
	}

	cursorTime, err := parseCursorInstant(cursor)
	if err != nil {
		return nil, fmt.Errorf("기도 제목 커서 파싱 실패: after=%s %w", cursor, ErrPrayerTitleInvalidCursor)
	}

	return s.prayerRepository.FindTitleInfosByRoomIDAfter(ctx, db, roomID, cursorTime, PrayerTitleInfiniteScrollLimit)
}

func parseCursorInstant(value string) (time.Time, error) {
	layouts := []string{time.RFC3339Nano, time.RFC3339}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, value); err == nil {
			return ts, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time: %s", value)
}

// CreateContent creates a new prayer content
func (s *PrayerService) CreateContent(
	ctx context.Context,
	tx *gorm.DB,
	prayerTitle *model.PrayerTitle,
	writer *model.Member,
	request *CreatePrayerContentRequest,
) error {
	prayerContent := model.NewPrayerContent(
		prayerTitle,
		writer,
		request.MemberID,
		request.MemberName,
		request.Content,
	)

	if err := s.prayerRepository.CreateContent(ctx, tx, prayerContent); err != nil {
		return fmt.Errorf("기도 내용 생성 실패: %s %w", ErrPrayerContentCreateFailed, err)
	}

	return nil
}

// GetContentsByTitleID retrieves all prayer contents for a specific prayer title
func (s *PrayerService) GetContentsByTitleID(ctx context.Context, db *gorm.DB, titleID int64) ([]PrayerContentInfo, error) {
	contents, err := s.prayerRepository.FindContentsByTitleID(ctx, db, titleID)
	if err != nil {
		return nil, fmt.Errorf("기도 내용 조회 실패: TitleID=%d %w", titleID, err)
	}

	return contents, nil
}

// UpdateTitle updates a prayer title
func (s *PrayerService) UpdateTitle(ctx context.Context, tx *gorm.DB, titleID int64, changedTitle string) error {
	prayerTitle, err := s.GetTitleById(ctx, tx, titleID)
	if err != nil {
		return err
	}

	prayerTitle.UpdateTitle(changedTitle)

	// Repository를 통해 변경사항 저장
	if err := s.prayerRepository.Update(ctx, tx, prayerTitle); err != nil {
		return fmt.Errorf("기도 제목 업데이트 실패: TitleID=%d %w", titleID, err)
	}

	return nil
}

// GetContentByID fetches a prayer content by its ID
func (s *PrayerService) GetContentByID(ctx context.Context, db *gorm.DB, contentID int64) (*model.PrayerContent, error) {
	prayerContent, err := s.prayerRepository.FindContentByID(ctx, db, contentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("기도 내용을 찾을 수 없습니다: ContentID=%d %w", contentID, ErrPrayerContentNotFound)
		}
		return nil, fmt.Errorf("기도 내용 조회 실패: ContentID=%d %w", contentID, err)
	}
	return prayerContent, nil
}

// ValidateContentExistsInTitle validates that a prayer content exists in a specific prayer title
func (s *PrayerService) ValidateContentExistsInTitle(ctx context.Context, db *gorm.DB, contentID int64, titleID int64) error {
	exists, err := s.prayerRepository.ExistsContentInTitle(ctx, db, contentID, titleID)
	if err != nil {
		return fmt.Errorf("기도 내용 존재 확인 실패: ContentID=%d, TitleID=%d %w", contentID, titleID, err)
	}
	if !exists {
		return fmt.Errorf("기도 내용을 찾을 수 없습니다: ContentID=%d, TitleID=%d %w", contentID, titleID, ErrPrayerContentNotFound)
	}
	return nil
}

// UpdateContent updates a prayer content
func (s *PrayerService) UpdateContent(ctx context.Context, tx *gorm.DB, contentID int64, changedContent string) error {
	prayerContent, err := s.GetContentByID(ctx, tx, contentID)
	if err != nil {
		return err
	}

	prayerContent.UpdateContent(changedContent)

	// Repository를 통해 변경사항 저장
	if err := s.prayerRepository.UpdateContent(ctx, tx, prayerContent); err != nil {
		return fmt.Errorf("기도 내용 업데이트 실패: ContentID=%d %w", contentID, err)
	}

	return nil
}

// DeleteTitle deletes a prayer title
// CASCADE 설정으로 연관된 PrayerContent도 자동 삭제됨
func (s *PrayerService) DeleteTitle(ctx context.Context, tx *gorm.DB, titleID int64) error {
	prayerTitle, err := s.GetTitleById(ctx, tx, titleID)
	if err != nil {
		return err
	}

	if err := s.prayerRepository.Delete(ctx, tx, prayerTitle); err != nil {
		return fmt.Errorf("기도 제목 삭제 실패: TitleID=%d %w", titleID, err)
	}

	return nil
}

// DeleteContent deletes a prayer content
func (s *PrayerService) DeleteContent(ctx context.Context, tx *gorm.DB, contentID int64) error {
	_, err := s.GetContentByID(ctx, tx, contentID)
	if err != nil {
		return err
	}

	// 2. Repository를 통해 삭제
	if err := s.prayerRepository.DeleteContent(ctx, tx, contentID); err != nil {
		return fmt.Errorf("기도 내용 삭제 실패: ContentID=%d %w", contentID, err)
	}

	return nil
}
