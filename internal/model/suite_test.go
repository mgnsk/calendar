package model_test

import (
	"testing"

	"github.com/mgnsk/calendar/internal"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/mgnsk/calendar/internal/pkg/sqlite"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/uptrace/bun"
)

var db *bun.DB

var _ = BeforeEach(func() {
	// db = sqlite.NewDB(":memory:").WithDebugLogging().Connect()
	db = sqlite.NewDB(":memory:").Connect()
	DeferCleanup(db.Close)

	Expect(internal.MigrateUp(db.DB)).To(Succeed())

	model.RegisterModels(db)
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "internal/model")
}
