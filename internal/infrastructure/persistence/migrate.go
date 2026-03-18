package persistence

import (
	"database/sql"
	"embed"
	"log"
	"sort"
	"strings"
)

const schemaMigrationsTable = `CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations 在启动时执行未应用的迁移（仅执行 .up.sql），按版本号顺序
func RunMigrations(db *sql.DB) error {
	if _, err := db.Exec(schemaMigrationsTable); err != nil {
		return err
	}
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return err
	}
	var upFiles []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".up.sql") {
			upFiles = append(upFiles, name)
		}
	}
	sort.Strings(upFiles)
	for _, name := range upFiles {
		version := strings.TrimSuffix(name, ".up.sql")
		var applied string
		err := db.QueryRow("SELECT version FROM schema_migrations WHERE version = ?", version).Scan(&applied)
		if err == nil {
			continue
		}
		if err != sql.ErrNoRows {
			return err
		}
		body, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		// SQLite 需要逐个执行语句
		stmts := strings.Split(string(body), ";")
		for _, stmt := range stmts {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := tx.Exec(stmt); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		log.Printf("✅ migration applied: %s", version)
	}
	return nil
}
