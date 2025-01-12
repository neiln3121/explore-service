// Package migrations contains the *.sql files for all the migrations
// It also exports an embed.FS to allow iteration through the migration files
package migrations

import (
	"embed"

	migrate "github.com/rubenv/sql-migrate"
)

// FS contains the migration files
//
//go:embed *.sql
var FS embed.FS

// GetMigrationSource returns the migration source.
func GetMigrationSource() *migrate.EmbedFileSystemMigrationSource {
	return &migrate.EmbedFileSystemMigrationSource{
		FileSystem: FS,
		Root:       ".",
	}
}
