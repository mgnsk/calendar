package model

import "github.com/uptrace/bun"

// RegisterModels registers models for bun database.
func RegisterModels(db *bun.DB) {
	// Register many to many model so bun can better recognize m2m relation.
	// This should be done before you use the model for the first time.
	db.RegisterModel(
		(*eventToTag)(nil),
	)
}
