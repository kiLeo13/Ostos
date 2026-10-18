package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// WithinTransaction executes operation in one explicit transaction.
func WithinTransaction(ctx context.Context, database *gorm.DB, options *sql.TxOptions, operation func(*gorm.DB) error) (err error) {
	if database == nil || operation == nil {
		return errors.New("database and transaction operation are required")
	}

	transaction := database.WithContext(ctx).Begin(options)
	if transaction.Error != nil {
		return fmt.Errorf("begin transaction: %w", transaction.Error)
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = transaction.Rollback().Error
			panic(recovered)
		}
		if err != nil {
			_ = transaction.Rollback().Error
		}
	}()

	if err = operation(transaction); err != nil {
		return err
	}
	if err = transaction.Commit().Error; err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
