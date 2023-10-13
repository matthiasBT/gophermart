package migrations

import (
	"embed"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

//go:embed postgres/*.sql
var embedMigrations embed.FS

func Migrate(db *sqlx.DB) {
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}
	if err := goose.Up(db.DB, "postgres"); err != nil {
		panic(err)
	}
}
