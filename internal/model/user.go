package model

import (
	"context"

	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	"github.com/mgnsk/calendar/internal/pkg/sqlite"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
)

// User is the user database model.
type User struct {
	ID       snowflake.ID `bun:"id,pk"`
	Username string       `bun:"username"`
	Password []byte       `bun:"password"`
	Role     string       `bun:"role"`

	bun.BaseModel `bun:"users"`
}

// InsertUser inserts a user into the database.
func InsertUser(ctx context.Context, db bun.IDB, user *domain.User) error {
	return sqlite.WithErrorChecking(db.NewInsert().Model(&User{
		ID:       user.ID,
		Username: user.Username,
		Password: user.Password,
		Role:     string(user.Role),
	}).Exec(ctx))
}

// DeleteUser deletes a user.
func DeleteUser(ctx context.Context, db bun.IDB, username string) error {
	return sqlite.WithErrorChecking(db.NewDelete().Model((*User)(nil)).
		Where("username = ?", username).
		Exec(ctx))
}

// GetUser returns a user.
func GetUser(ctx context.Context, db bun.IDB, username string) (*domain.User, error) {
	model := &User{}

	if err := db.NewSelect().Model(model).
		Where("username = ?", username).
		Scan(ctx); err != nil {
		return nil, sqlite.NormalizeError(err)
	}

	return &domain.User{
		ID:       model.ID,
		Username: model.Username,
		Password: model.Password,
		Role:     domain.Role(model.Role),
	}, nil
}

// ListUsers lists users.
func ListUsers(ctx context.Context, db bun.IDB) ([]*domain.User, error) {
	model := []*User{}

	if err := db.NewSelect().Model(&model).
		Order("id DESC").
		Scan(ctx); err != nil {
		return nil, sqlite.NormalizeError(err)
	}

	return lo.Map(model, func(user *User, _ int) *domain.User {
		return &domain.User{
			ID:       user.ID,
			Username: user.Username,
			Password: user.Password,
			Role:     domain.Role(user.Role),
		}
	}), nil
}
