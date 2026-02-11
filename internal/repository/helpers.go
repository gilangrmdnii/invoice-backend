package repository

import "strings"

// buildInClause returns a SQL "IN (?, ?, ...)" placeholder string and
// a []interface{} slice suitable for use with database/sql query args.
func buildInClause(ids []uint64) (string, []interface{}) {
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	return strings.Join(placeholders, ","), args
}
