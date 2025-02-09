package handler_test

import (
	"testing"
	"time"

	"github.com/mgnsk/calendar/internal"
	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	"github.com/mgnsk/calendar/internal/pkg/sqlite"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/uptrace/bun"
)

var db *bun.DB

var _ = BeforeEach(func() {
	db = sqlite.NewDB(":memory:").Connect()
	DeferCleanup(db.Close)

	Expect(internal.MigrateUp(db.DB)).To(Succeed())
	DeferCleanup(func() error {
		return internal.MigrateDown(db.DB)
	})

	model.RegisterModels(db)
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "internal/api")
}

var (
	event1, event2, event3 *domain.Event
)

func init() {
	baseTime := time.Now().Add(time.Hour)

	event1 = &domain.Event{
		ID:          snowflake.Generate(),
		StartAt:     baseTime.Add(3 * time.Hour),
		EndAt:       time.Time{},
		Title:       "Event 1",
		Description: "Desc 1",
		URL:         "https://event1.testing",
		Tags:        []string{"tag1"},
	}

	event2 = &domain.Event{
		ID:          snowflake.Generate(),
		StartAt:     baseTime.Add(2 * time.Hour),
		EndAt:       time.Time{},
		Title:       "Event 2",
		Description: "Desc 2",
		URL:         "https://event2.testing",
		Tags:        []string{"tag1", "tag2"},
	}

	event3 = &domain.Event{
		ID:          snowflake.Generate(),
		StartAt:     baseTime.Add(1 * time.Hour),
		EndAt:       baseTime.Add(2 * time.Hour),
		Title:       "Event 3",
		Description: "Desc 3",
		URL:         "https://event3.testing",
		Tags:        []string{"tag3"},
	}
}
