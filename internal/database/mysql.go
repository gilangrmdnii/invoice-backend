package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gilangrmdnii/invoice-backend/internal/config"
)

func NewMySQL(cfg *config.Config) (*sql.DB, error) {
	var dsn string
	if cfg.DBURL != "" {
		// Railway provides MYSQL_URL like: mysql://user:pass@host:port/db
		// go-sql-driver needs: user:pass@tcp(host:port)/db?parseTime=true
		dsn = cfg.DBURL
		// Handle mysql:// prefix from Railway
		if len(dsn) > 8 && dsn[:8] == "mysql://" {
			dsn = dsn[8:]
			// Convert user:pass@host:port/db to user:pass@tcp(host:port)/db
			for i, c := range dsn {
				if c == '@' {
					rest := dsn[i+1:]
					for j, c2 := range rest {
						if c2 == '/' {
							host := rest[:j]
							dbName := rest[j:]
							dsn = dsn[:i+1] + "tcp(" + host + ")" + dbName
							break
						}
					}
					break
				}
			}
		}
		if !contains(dsn, "parseTime") {
			if contains(dsn, "?") {
				dsn += "&parseTime=true"
			} else {
				dsn += "?parseTime=true"
			}
		}
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBName,
		)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
