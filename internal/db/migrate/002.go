package migrate

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 2,
		Up:      dropLegacyChannelColumns,
	})
}

// 002:
// - drop legacy channels.key column
// - drop legacy channels.base_url column
func dropLegacyChannelColumns(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	dialect := db.Dialector.Name()

	// column existence helper
	hasColumn := func(table, column string) (bool, error) {
		if dialect == "sqlite" {
			var name string
			if err := db.Raw("SELECT name FROM pragma_table_info(?) WHERE name = ? LIMIT 1", table, column).
				Scan(&name).Error; err != nil {
				return false, fmt.Errorf("failed to check sqlite column %s.%s: %w", table, column, err)
			}
			return name == column, nil
		}
		return db.Migrator().HasColumn(table, column), nil
	}

	dropColumn := func(table, column string) error {
		var quotedTable strings.Builder
		db.Dialector.QuoteTo(&quotedTable, table)
		var quotedColumn strings.Builder
		db.Dialector.QuoteTo(&quotedColumn, column)

		sql := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", quotedTable.String(), quotedColumn.String())
		return db.Exec(sql).Error
	}

	// Drop 'key'
	hasKey, err := hasColumn("channels", "key")
	if err != nil {
		return err
	}
	if hasKey {
		if err := dropColumn("channels", "key"); err != nil {
			return fmt.Errorf("failed to drop column key: %w", err)
		}
	}

	// Drop 'base_url'
	hasBaseURL, err := hasColumn("channels", "base_url")
	if err != nil {
		return err
	}
	if hasBaseURL {
		if err := dropColumn("channels", "base_url"); err != nil {
			return fmt.Errorf("failed to drop column base_url: %w", err)
		}
	}

	return nil
}
