package model_test

import (
	"testing"

	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/sqlite"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/uptrace/bun"
)

var db *bun.DB

var _ = BeforeEach(func() {
	db = sqlite.NewDB(":memory:").Connect()
	DeferCleanup(db.Close)

	Expect(calendar.MigrateUp(db.DB)).To(Succeed())
	DeferCleanup(func() error {
		return calendar.MigrateDown(db.DB)
	})

	model.RegisterModels(db)
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "model")
}
