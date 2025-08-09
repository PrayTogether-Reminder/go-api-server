package domain

import (
	"context"
	"time"
)

// Service represents prayer domain service
type Service struct {
	Repository         Repository // Made public for use case access
	validateRoomAccess func(ctx context.Context, roomID, memberID uint64) error
	recordMemberPray   func(ctx context.Context, roomID, memberID uint64) error
}

// ValidateRoomAccess validates if member has access to the room
func (s *Service) ValidateRoomAccess(ctx context.Context, roomID, memberID uint64) error {
	if s.validateRoomAccess != nil {
		return s.validateRoomAccess(ctx, roomID, memberID)
	}
	return nil
}

// NewService creates a new prayer service
func NewService(
	repo Repository,
	validateRoomAccess func(ctx context.Context, roomID, memberID uint64) error,
	recordMemberPray func(ctx context.Context, roomID, memberID uint64) error,
) *Service {
	return &Service{
		Repository:         repo,
		validateRoomAccess: validateRoomAccess,
		recordMemberPray:   recordMemberPray,
	}
}

// CreatePrayer creates a new prayer
func (s *Service) CreatePrayer(ctx context.Context, memberID, roomID uint64, content string, prayerType PrayerType) (*Prayer, error) {
	// Validate room access
	if s.validateRoomAccess != nil {
		if err := s.validateRoomAccess(ctx, roomID, memberID); err != nil {
			return nil, err
		}
	}

	// Create prayer
	prayer, err := NewPrayer(memberID, roomID, content, prayerType)
	if err != nil {
		return nil, err
	}

	// Save to repository
	if err := s.Repository.Create(ctx, prayer); err != nil {
		return nil, err
	}

	// Record member pray activity in room
	if s.recordMemberPray != nil {
		_ = s.recordMemberPray(ctx, roomID, memberID)
	}

	return prayer, nil
}

// GetPrayer gets a prayer by ID
func (s *Service) GetPrayer(ctx context.Context, id uint64, requestorID uint64) (*Prayer, error) {
	prayer, err := s.Repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if prayer == nil {
		return nil, ErrPrayerNotFound
	}

	// Check access permission
	if prayer.Type == PrayerTypePersonal && prayer.MemberID != requestorID {
		return nil, ErrUnauthorizedAccess
	}

	// For shared prayers, validate room access
	if prayer.Type == PrayerTypeShared && s.validateRoomAccess != nil {
		if err := s.validateRoomAccess(ctx, prayer.RoomID, requestorID); err != nil {
			return nil, ErrUnauthorizedAccess
		}
	}

	return prayer, nil
}

// UpdatePrayer updates a prayer
func (s *Service) UpdatePrayer(ctx context.Context, id uint64, updaterID uint64, content string) (*Prayer, error) {
	// Get prayer
	prayer, err := s.Repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if prayer == nil {
		return nil, ErrPrayerNotFound
	}

	// Check permission
	if !prayer.CanBeEditedBy(updaterID) {
		return nil, ErrUnauthorizedAccess
	}

	// Update content
	if err := prayer.UpdateContent(content); err != nil {
		return nil, err
	}

	// Save to repository
	if err := s.Repository.Update(ctx, prayer); err != nil {
		return nil, err
	}

	return prayer, nil
}

// MarkPrayerAsAnswered marks a prayer as answered
func (s *Service) MarkPrayerAsAnswered(ctx context.Context, id uint64, memberID uint64) error {
	// Get prayer
	prayer, err := s.Repository.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if prayer == nil {
		return ErrPrayerNotFound
	}

	// Check permission
	if !prayer.CanBeEditedBy(memberID) {
		return ErrUnauthorizedAccess
	}

	// Mark as answered
	prayer.MarkAsAnswered()

	// Save to repository
	return s.Repository.Update(ctx, prayer)
}

// MarkPrayerAsUnanswered marks a prayer as unanswered
func (s *Service) MarkPrayerAsUnanswered(ctx context.Context, id uint64, memberID uint64) error {
	// Get prayer
	prayer, err := s.Repository.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if prayer == nil {
		return ErrPrayerNotFound
	}

	// Check permission
	if !prayer.CanBeEditedBy(memberID) {
		return ErrUnauthorizedAccess
	}

	// Mark as unanswered
	prayer.MarkAsUnanswered()

	// Save to repository
	return s.Repository.Update(ctx, prayer)
}

// ChangePrayerType changes the type of a prayer
func (s *Service) ChangePrayerType(ctx context.Context, id uint64, memberID uint64, newType PrayerType) error {
	// Get prayer
	prayer, err := s.Repository.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if prayer == nil {
		return ErrPrayerNotFound
	}

	// Check permission
	if !prayer.CanBeEditedBy(memberID) {
		return ErrUnauthorizedAccess
	}

	// Change type
	if err := prayer.SetType(newType); err != nil {
		return err
	}

	// Save to repository
	return s.Repository.Update(ctx, prayer)
}

// DeletePrayer deletes a prayer
func (s *Service) DeletePrayer(ctx context.Context, id uint64, deleterID uint64) error {
	// Get prayer
	prayer, err := s.Repository.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if prayer == nil {
		return ErrPrayerNotFound
	}

	// Check permission
	if !prayer.CanBeEditedBy(deleterID) {
		return ErrUnauthorizedAccess
	}

	// Delete from repository
	return s.Repository.Delete(ctx, id)
}

// GetMemberPrayers gets prayers for a member
func (s *Service) GetMemberPrayers(ctx context.Context, memberID uint64, limit, offset int) ([]*Prayer, error) {
	return s.Repository.FindByMemberID(ctx, memberID, limit, offset)
}

// GetRoomPrayers gets shared prayers for a room
func (s *Service) GetRoomPrayers(ctx context.Context, roomID uint64, requestorID uint64, limit, offset int) ([]*Prayer, error) {
	// Validate room access
	if s.validateRoomAccess != nil {
		if err := s.validateRoomAccess(ctx, roomID, requestorID); err != nil {
			return nil, err
		}
	}

	return s.Repository.FindSharedPrayersByRoom(ctx, roomID, limit, offset)
}

// GetMemberRoomPrayers gets prayers for a member in a specific room
func (s *Service) GetMemberRoomPrayers(ctx context.Context, memberID, roomID uint64, requestorID uint64, limit, offset int) ([]*Prayer, error) {
	// If requesting own prayers
	if memberID == requestorID {
		return s.Repository.FindByMemberAndRoom(ctx, memberID, roomID, limit, offset)
	}

	// If requesting someone else's prayers, validate room access and only return shared prayers
	if s.validateRoomAccess != nil {
		if err := s.validateRoomAccess(ctx, roomID, requestorID); err != nil {
			return nil, err
		}
	}

	// Get all prayers and filter for shared ones
	prayers, err := s.Repository.FindByMemberAndRoom(ctx, memberID, roomID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Filter for shared prayers only
	sharedPrayers := make([]*Prayer, 0)
	for _, prayer := range prayers {
		if prayer.Type == PrayerTypeShared {
			sharedPrayers = append(sharedPrayers, prayer)
		}
	}

	return sharedPrayers, nil
}

// GetAnsweredPrayers gets answered prayers for a member
func (s *Service) GetAnsweredPrayers(ctx context.Context, memberID uint64, limit, offset int) ([]*Prayer, error) {
	return s.Repository.FindAnsweredPrayers(ctx, memberID, limit, offset)
}

// GetUnansweredPrayers gets unanswered prayers for a member
func (s *Service) GetUnansweredPrayers(ctx context.Context, memberID uint64, limit, offset int) ([]*Prayer, error) {
	return s.Repository.FindUnansweredPrayers(ctx, memberID, limit, offset)
}

// GetPrayersByDateRange gets prayers for a member in a date range
func (s *Service) GetPrayersByDateRange(ctx context.Context, memberID uint64, startDate, endDate time.Time) ([]*Prayer, error) {
	return s.Repository.FindPrayersByDateRange(ctx, memberID, startDate, endDate)
}

// GetPrayerStatistics gets prayer statistics for a member
func (s *Service) GetPrayerStatistics(ctx context.Context, memberID uint64) (*PrayerStatistics, error) {
	total, err := s.Repository.CountByMemberID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	answered, err := s.Repository.CountAnsweredByMemberID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	return &PrayerStatistics{
		TotalPrayers:      total,
		AnsweredPrayers:   answered,
		UnansweredPrayers: total - answered,
		AnswerRate:        float64(answered) / float64(total) * 100,
	}, nil
}

// PrayerStatistics represents prayer statistics
type PrayerStatistics struct {
	TotalPrayers      int     `json:"totalPrayers"`
	AnsweredPrayers   int     `json:"answeredPrayers"`
	UnansweredPrayers int     `json:"unansweredPrayers"`
	AnswerRate        float64 `json:"answerRate"`
}

// GetPrayerTitle gets a prayer title by ID
func (s *Service) GetPrayerTitle(ctx context.Context, id uint64) (*PrayerTitle, error) {
	return s.Repository.FindPrayerTitleByID(ctx, id)
}

// GetPrayerTitleWithContents gets a prayer title with its contents
func (s *Service) GetPrayerTitleWithContents(ctx context.Context, id uint64) (*PrayerTitle, error) {
	return s.Repository.FindPrayerTitleWithContents(ctx, id)
}

// CreatePrayerTitle creates a new prayer title
func (s *Service) CreatePrayerTitle(ctx context.Context, roomID, memberID uint64, title string, contents []string) (*PrayerTitle, error) {
	// Validate room access
	if s.validateRoomAccess != nil {
		if err := s.validateRoomAccess(ctx, roomID, memberID); err != nil {
			return nil, err
		}
	}

	// Create prayer title
	prayerTitle := NewPrayerTitle(roomID, memberID, title)
	if err := s.Repository.CreatePrayerTitle(ctx, prayerTitle); err != nil {
		return nil, err
	}

	// Create prayer contents
	if len(contents) > 0 {
		prayerContents := make([]*PrayerContent, len(contents))
		for i, content := range contents {
			prayerContents[i] = NewPrayerContent(prayerTitle.ID, memberID, content)
		}
		if err := s.Repository.CreatePrayerContents(ctx, prayerContents); err != nil {
			return nil, err
		}
		// Convert to []PrayerContent
		contents := make([]PrayerContent, len(prayerContents))
		for i, pc := range prayerContents {
			contents[i] = *pc
		}
		prayerTitle.Contents = contents
	}

	// Record member pray activity
	if s.recordMemberPray != nil {
		_ = s.recordMemberPray(ctx, roomID, memberID)
	}

	return prayerTitle, nil
}

// UpdatePrayerTitle updates a prayer title and its contents
func (s *Service) UpdatePrayerTitle(ctx context.Context, id, memberID uint64, title string, contents []string) (*PrayerTitle, error) {
	// Get existing prayer title
	prayerTitle, err := s.Repository.FindPrayerTitleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check permission
	if prayerTitle.CreatorID != memberID {
		return nil, ErrUnauthorizedAccess
	}

	// Update title
	prayerTitle.Title = title
	if err := s.Repository.UpdatePrayerTitle(ctx, prayerTitle); err != nil {
		return nil, err
	}

	// Delete existing contents and create new ones
	if err := s.Repository.DeletePrayerContentsByTitleID(ctx, id); err != nil {
		return nil, err
	}

	if len(contents) > 0 {
		prayerContents := make([]*PrayerContent, len(contents))
		for i, content := range contents {
			prayerContents[i] = NewPrayerContent(id, memberID, content)
		}
		if err := s.Repository.CreatePrayerContents(ctx, prayerContents); err != nil {
			return nil, err
		}
		// Convert to []PrayerContent
		contents := make([]PrayerContent, len(prayerContents))
		for i, pc := range prayerContents {
			contents[i] = *pc
		}
		prayerTitle.Contents = contents
	}

	return prayerTitle, nil
}

// DeletePrayerTitle deletes a prayer title
func (s *Service) DeletePrayerTitle(ctx context.Context, id, memberID uint64) error {
	// Get prayer title
	prayerTitle, err := s.Repository.FindPrayerTitleByID(ctx, id)
	if err != nil {
		return err
	}

	if prayerTitle == nil {
		return ErrPrayerNotFound
	}

	// Validate member is in the room
	if s.validateRoomAccess != nil {
		if err := s.validateRoomAccess(ctx, prayerTitle.RoomID, memberID); err != nil {
			return ErrPrayerNotFound // Return 404 for consistency with Java
		}
	}

	// Check permission (must be creator)
	if prayerTitle.CreatorID != memberID {
		return ErrUnauthorizedAccess
	}

	// Delete contents first
	if err := s.Repository.DeletePrayerContentsByTitleID(ctx, id); err != nil {
		return err
	}

	// Delete title
	return s.Repository.DeletePrayerTitle(ctx, id)
}

// GetRoomPrayerTitles gets prayer titles for a room
func (s *Service) GetRoomPrayerTitles(ctx context.Context, roomID, memberID uint64, after uint64, limit int) ([]*PrayerTitle, error) {
	// Validate room access
	if s.validateRoomAccess != nil {
		if err := s.validateRoomAccess(ctx, roomID, memberID); err != nil {
			return nil, err
		}
	}

	return s.Repository.FindPrayerTitlesByRoomID(ctx, roomID, after, limit)
}

// GetRoomPrayerTitleInfos gets prayer title infos for a room with time-based cursor (matching Java)
func (s *Service) GetRoomPrayerTitleInfos(ctx context.Context, roomID uint64, after string, limit int) ([]*PrayerTitleInfo, error) {
	// Java default value for after is "" which means get first page
	if after == "" {
		// Get first page
		return s.Repository.FindFirstPrayerTitleInfosByRoomID(ctx, roomID, limit)
	}

	// Parse time cursor and get next page
	afterTime, err := time.Parse(time.RFC3339, after)
	if err != nil {
		// If invalid time format, get first page
		return s.Repository.FindFirstPrayerTitleInfosByRoomID(ctx, roomID, limit)
	}

	return s.Repository.FindPrayerTitleInfosByRoomIDAfterTime(ctx, roomID, afterTime, limit)
}

// CompletePrayer marks a prayer as completed by a member
func (s *Service) CompletePrayer(ctx context.Context, memberID, prayerTitleID uint64) error {
	// Check if already completed
	exists, err := s.Repository.ExistsPrayerCompletionByMemberAndTitle(ctx, memberID, prayerTitleID)
	if err != nil {
		return err
	}
	if exists {
		return ErrAlreadyCompleted
	}

	// Create completion record
	completion := NewPrayerCompletion(prayerTitleID, memberID)
	return s.Repository.CreatePrayerCompletion(ctx, completion)
}

// HasMemberCompletedPrayer checks if a member has completed a prayer
func (s *Service) HasMemberCompletedPrayer(ctx context.Context, memberID, prayerTitleID uint64) (bool, error) {
	return s.Repository.ExistsPrayerCompletionByMemberAndTitle(ctx, memberID, prayerTitleID)
}

// GetPrayerCompletions gets all completions for a prayer title
func (s *Service) GetPrayerCompletions(ctx context.Context, prayerTitleID uint64) ([]*PrayerCompletion, error) {
	return s.Repository.FindPrayerCompletionsByTitleID(ctx, prayerTitleID)
}
