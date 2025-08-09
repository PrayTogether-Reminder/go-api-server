package domain

import (
	"context"
)

// Repository interface for member domain
type Repository interface {
	// Create creates a new member
	Create(ctx context.Context, member *Member) error

	// FindByID finds a member by ID
	FindByID(ctx context.Context, id uint64) (*Member, error)

	// FindByEmail finds a member by email
	FindByEmail(ctx context.Context, email string) (*Member, error)

	// FindMemberProfileByID finds member profile by ID
	FindMemberProfileByID(ctx context.Context, id uint64) (*MemberProfile, error)

	// Update updates a member
	Update(ctx context.Context, member *Member) error

	// Delete deletes a member
	Delete(ctx context.Context, id uint64) error

	// ExistsByEmail checks if email exists
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByID checks if member exists by ID
	ExistsByID(ctx context.Context, id uint64) (bool, error)

	// SearchByName searches members by name (partial match)
	SearchByName(ctx context.Context, name string) ([]*Member, error)
}
