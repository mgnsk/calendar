package model

import (
	"context"
	"errors"

	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/pkg/snowflake"
	"github.com/mgnsk/calendar/pkg/sqlite"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
)

// StopWord is the stopword database model.
type StopWord struct {
	ID   snowflake.ID `bun:"id,pk"`
	Word string       `bun:"word"`

	bun.BaseModel `bun:"stopwords"`
}

// SetStopWords sets stop words in the database.
func SetStopWords(ctx context.Context, db bun.IDB, words domain.StopWordList) error {
	return db.RunInTx(ctx, nil, func(ctx context.Context, db bun.Tx) error {
		// Delete all stop words.
		if err := sqlite.WithErrorChecking(
			db.NewTruncateTable().Model((*StopWord)(nil)).Exec(ctx),
		); err != nil && !errors.Is(err, calendar.PreconditionFailed) {
			return err
		}

		if len(words) == 0 {
			return nil
		}

		model := lo.Map(words, func(word domain.StopWord, _ int) *StopWord {
			return &StopWord{
				ID:   snowflake.Generate(),
				Word: word.Word,
			}
		})

		return sqlite.WithErrorChecking(db.NewInsert().Model(&model).Ignore().Exec(ctx))
	})
}

// ListStopWords lists all stopwords.
func ListStopWords(ctx context.Context, db bun.IDB) ([]*domain.StopWord, error) {
	model := []*StopWord{}

	if err := db.NewSelect().Model(&model).
		Order("id ASC").
		Scan(ctx); err != nil {
		return nil, sqlite.NormalizeError(err)
	}

	return lo.Map(model, func(sw *StopWord, _ int) *domain.StopWord {
		return &domain.StopWord{
			Word: sw.Word,
		}
	}), nil
}
