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

	// Baseline: check each old migration and mark as applied if its changes already exist.
	// This handles databases where migrations were run manually before auto-migrate.
	baselineChecks := map[string]string{
		"000001_init_schema.sql":            `SELECT 1 FROM users LIMIT 1`,
		"000002_invoices.sql":               `SELECT 1 FROM invoices LIMIT 1`,
		"000003_enhanced_invoices.sql":      `SELECT invoice_type FROM invoices LIMIT 1`,
		"000006_project_plan_items.sql":     `SELECT 1 FROM project_plan_items LIMIT 1`,
		"000007_expense_approval_proof.sql": `SELECT proof_url FROM expense_approvals LIMIT 1`,
		"000008_invoice_dual_tax.sql":       `SELECT ppn_percentage FROM invoices LIMIT 1`,
	}
	for name, checkSQL := range baselineChecks {
		var alreadyApplied int
		_ = db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, name).Scan(&alreadyApplied)
		if alreadyApplied > 0 {
			continue
		}
		var dummy interface{}
		if err := db.QueryRow(checkSQL).Scan(&dummy); err == nil || err == sql.ErrNoRows {
			_, _ = db.Exec(`INSERT IGNORE INTO schema_migrations (version) VALUES (?)`, name)
			log.Printf("[migrate] baseline: marked %s as already applied", name)
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
