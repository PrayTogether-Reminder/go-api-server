package domain

import (
	"context"
)

// Service represents member domain service
type Service struct {
	repo Repository
}

// NewService creates a new member service
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateMember creates a new member
func (s *Service) CreateMember(ctx context.Context, email, password, name string) (*Member, error) {
	// Check if email already exists
	exists, err := s.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrMemberAlreadyExist
	}

	// Create new member
	member, err := NewMember(name, email, password)
	if err != nil {
		return nil, err
	}

	// Save to repository
	if err := s.repo.Create(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

// GetMember gets a member by ID
func (s *Service) GetMember(ctx context.Context, id uint64) (*Member, error) {
	member, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if member == nil {
		return nil, ErrMemberNotFound
	}

	return member, nil
}

// GetMemberByID gets a member by ID (alias for GetMember)
func (s *Service) GetMemberByID(ctx context.Context, id uint64) (*Member, error) {
	return s.GetMember(ctx, id)
}

// GetMemberByEmail gets a member by email
func (s *Service) GetMemberByEmail(ctx context.Context, email string) (*Member, error) {
	member, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if member == nil {
		return nil, ErrMemberNotFound
	}

	return member, nil
}

// UpdateMember updates a member
func (s *Service) UpdateMember(ctx context.Context, member *Member) error {
	return s.repo.Update(ctx, member)
}

// DeleteMember deletes a member
func (s *Service) DeleteMember(ctx context.Context, id uint64) error {
	exists, err := s.repo.ExistsByID(ctx, id)
	if err != nil {
		return err
	}

	if !exists {
		return ErrMemberNotFound
	}

	return s.repo.Delete(ctx, id)
}

// ValidateMember validates if a member exists
func (s *Service) ValidateMember(ctx context.Context, id uint64) error {
	exists, err := s.repo.ExistsByID(ctx, id)
	if err != nil {
		return err
	}

	if !exists {
		return ErrMemberNotFound
	}

	return nil
}

// SearchMembersByName searches members by name
func (s *Service) SearchMembersByName(ctx context.Context, name string) ([]*Member, error) {
	if name == "" {
		return nil, ErrInvalidName
	}

	members, err := s.repo.SearchByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return members, nil
}
