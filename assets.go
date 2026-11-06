package calendar

import (
	"embed"

	"github.com/mgnsk/webkit/assets"
)

// Migrations are the SQL migrations, in the migrations directory.
//
//go:embed migrations/*.sql
var Migrations embed.FS

// Assets are the static assets.
var Assets = assets.New(assetsFS)
