package interfaces

import (
	"context"
	"pray-together/internal/domains/member/domain"
)

// API represents the public interface for member domain
// Other domains should use this interface instead of accessing internal components
type API interface {
	// GetMemberInfo gets basic member information
	GetMemberInfo(ctx context.Context, memberID uint64) (*domain.MemberInfo, error)

	// ValidateMember validates if a member exists
	ValidateMember(ctx context.Context, memberID uint64) error

	// GetMemberByEmail gets member by email
	GetMemberByEmail(ctx context.Context, email string) (*domain.MemberProfile, error)
}

// Module represents the member module with all its components
type Module struct {
	Service    *domain.Service
	Repository domain.Repository
}

// NewModule creates a new member module
func NewModule(repo domain.Repository) *Module {
	return &Module{
		Service:    domain.NewService(repo),
		Repository: repo,
	}
}

// GetMemberInfo implements API interface
func (m *Module) GetMemberInfo(ctx context.Context, memberID uint64) (*domain.MemberInfo, error) {
	member, err := m.Service.GetMember(ctx, memberID)
	if err != nil {
		return nil, err
	}
	return member.ToInfo(), nil
}

// ValidateMember implements API interface
func (m *Module) ValidateMember(ctx context.Context, memberID uint64) error {
	return m.Service.ValidateMember(ctx, memberID)
}

// GetMemberByEmail implements API interface
func (m *Module) GetMemberByEmail(ctx context.Context, email string) (*domain.MemberProfile, error) {
	member, err := m.Service.GetMemberByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return member.ToProfile(), nil
}
