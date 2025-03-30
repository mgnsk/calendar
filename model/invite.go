package model

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/pkg/snowflake"
	"github.com/mgnsk/calendar/pkg/sqlite"
	"github.com/mgnsk/calendar/pkg/wreck"
	"github.com/uptrace/bun"
)

// Invite is the invite database model.
type Invite struct {
	Token          uuid.UUID    `bun:"token"`
	ValidUntilUnix int64        `bun:"valid_until_unix"`
	CreatedBy      snowflake.ID `bun:"created_by"`

	bun.BaseModel `bun:"invites"`
}

// InsertInvite inserts a new event to the database.
func InsertInvite(ctx context.Context, db *bun.DB, invite *domain.Invite) error {
	return sqlite.WithErrorChecking(db.NewInsert().Model(&Invite{
		Token:          invite.Token,
		ValidUntilUnix: invite.ValidUntil.Unix(),
		CreatedBy:      invite.CreatedBy,
	}).Exec(ctx))
}

// DeleteInvite deletes an invite.
func DeleteInvite(ctx context.Context, db bun.IDB, token uuid.UUID) error {
	return sqlite.WithErrorChecking(db.NewDelete().Model((*Invite)(nil)).
		Where("token = ?", token).
		Exec(ctx))
}

// GetInvite returns an invite.
func GetInvite(ctx context.Context, db *bun.DB, token uuid.UUID) (*domain.Invite, error) {
	model := &Invite{}

	if err := db.NewSelect().Model(model).
		Where("token = ?", token).
		Scan(ctx); err != nil {
		return nil, sqlite.NormalizeError(err)
	}

	return &domain.Invite{
		Token:      model.Token,
		ValidUntil: time.Unix(model.ValidUntilUnix, 0),
		CreatedBy:  model.CreatedBy,
	}, nil
}

// DeleteExpiredInvites deletes expired invites.
func DeleteExpiredInvites(ctx context.Context, db *bun.DB) error {
	err := sqlite.WithErrorChecking(db.NewDelete().Model((*Invite)(nil)).
		Where("valid_until_unix < ?", time.Now().Unix()).
		Exec(ctx))

	if errors.Is(err, wreck.PreconditionFailed) {
		return nil
	}

	return err
}
