package database

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"
)

// RunMigrations executes all .sql migration files that have not been applied yet.
// It tracks applied migrations in a schema_migrations table.
func RunMigrations(db *sql.DB, migrationFS embed.FS) error {
	// Create schema_migrations table if not exists
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	// Baseline: if schema_migrations is empty but tables already exist,
	// mark old migrations as already applied (they were run manually before auto-migrate).
	var migrationCount int
	_ = db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&migrationCount)
	if migrationCount == 0 {
		// Direct check: try to query the invoices table
		var dummy int
		err := db.QueryRow(`SELECT 1 FROM invoices LIMIT 1`).Scan(&dummy)
		if err == nil || err == sql.ErrNoRows {
			// Table exists — old migrations were already applied manually
			baseline := []string{"000001_init_schema.sql", "000002_invoices.sql", "000003_enhanced_invoices.sql"}
			for _, name := range baseline {
				_, _ = db.Exec(`INSERT IGNORE INTO schema_migrations (version) VALUES (?)`, name)
			}
			log.Println("[migrate] baseline: marked 000001-000003 as already applied")
		} else {
			log.Printf("[migrate] fresh database detected (invoices table check: %v)", err)
		}
	}

	// Get already applied migrations
	applied := make(map[string]bool)
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("query schema_migrations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return err
		}
		applied[v] = true
	}

	// Read and sort migration file names
	entries, err := fs.ReadDir(migrationFS, ".")
	if err != nil {
		return fmt.Errorf("read migration dir: %w", err)
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	// Run pending migrations
	for _, name := range names {
		if applied[name] {
			continue
		}

		content, err := fs.ReadFile(migrationFS, name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		log.Printf("[migrate] applying %s ...", name)

		stmts := splitStatements(string(content))
		for _, stmt := range stmts {
			if _, err := db.Exec(stmt); err != nil {
				return fmt.Errorf("migration %s failed: %w\nStatement: %s", name, err, stmt)
			}
		}

		if _, err := db.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, name); err != nil {
			return fmt.Errorf("record migration %s: %w", name, err)
		}

		log.Printf("[migrate] applied %s", name)
	}

	return nil
}

// splitStatements splits a SQL script by semicolons, ignoring empty and comment-only parts.
func splitStatements(sql string) []string {
	parts := strings.Split(sql, ";")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		// Skip comment-only blocks
		lines := strings.Split(trimmed, "\n")
		hasCode := false
		for _, line := range lines {
			l := strings.TrimSpace(line)
			if l != "" && !strings.HasPrefix(l, "--") {
				hasCode = true
				break
			}
		}
		if hasCode {
			result = append(result, trimmed)
		}
	}
	return result
}
