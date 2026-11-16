package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const bulkInsertSize = 500

func BulkInsert(ctx context.Context, db *sql.DB, table string, columns []string, rows [][]any) error {
	if len(rows) == 0 {
		return nil
	}

	quotedColumns := make([]string, len(columns))
	for i, column := range columns {
		quotedColumns[i] = quoteIdent(column)
	}
	columnList := strings.Join(quotedColumns, ", ")

	for start := 0; start < len(rows); start += bulkInsertSize {
		end := start + bulkInsertSize
		if end > len(rows) {
			end = len(rows)
		}
		chunk := rows[start:end]

		var query strings.Builder
		args := make([]any, 0, len(chunk)*len(columns))
		query.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES ", table, columnList))
		for i, row := range chunk {
			if i > 0 {
				query.WriteString(", ")
			}
			query.WriteString(rowPlaceholder(len(columns)))
			args = append(args, row...)
		}

		if _, err := db.ExecContext(ctx, query.String(), args...); err != nil {
			return fmt.Errorf("bulk insert into %s: %w", table, err)
		}
	}

	return nil
}

func rowPlaceholder(columnCount int) string {
	args := make([]string, columnCount)
	for i := range args {
		args[i] = "?"
	}
	return "(" + strings.Join(args, ", ") + ")"
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
