package model

import (
	"context"
	"errors"

	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/pkg/sqlite"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
)

// StopWord is the stopword database model.
type StopWord struct {
	Word  string `bun:"word,pk"`
	Order uint64 `bun:"sort_order"`

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

		model := lo.Map(words, func(word string, idx int) *StopWord {
			return &StopWord{
				Word:  word,
				Order: uint64(idx),
			}
		})

		return sqlite.WithErrorChecking(db.NewInsert().Model(&model).Exec(ctx))
	})
}

// ListStopWords lists all stopwords.
func ListStopWords(ctx context.Context, db bun.IDB) ([]string, error) {
	model := []*StopWord{}

	if err := db.NewSelect().Model(&model).
		Order("sort_order ASC").
		Scan(ctx); err != nil {
		return nil, sqlite.NormalizeError(err)
	}

	return lo.Map(model, func(sw *StopWord, _ int) string {
		return sw.Word
	}), nil
}
