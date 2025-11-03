package opera_txr

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------------------------------------------------
// Struct
// ---------------------------------------------------------------------------------------------------------------------

type TxrImplSql struct {
	db                       *sql.DB
	deadlockMaxRetries       uint
	deadlockMinRetryInterval time.Duration
}

// ---------------------------------------------------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------------------------------------------------

func NewTxrImplSql(
	db *sql.DB,
	deadlockMaxRetries uint,
	deadlockMinRetryInterval time.Duration,
) *TxrImplSql {
	return &TxrImplSql{
		db:                       db,
		deadlockMaxRetries:       deadlockMaxRetries,
		deadlockMinRetryInterval: deadlockMinRetryInterval,
	}
}

// ---------------------------------------------------------------------------------------------------------------------
// Actions
// ---------------------------------------------------------------------------------------------------------------------

// Tx runs the provided function fn within a transaction context TxCtx.
//
// Panics if:
//   - ctx is nil (programming error: caller must provide a valid context)
//   - nested calls (makes no sense and likely indicates a design flaw)
//   - fn is nil (programming error: transaction body must be provided)
//   - fn panics
//
// Returns the error returned by fn, or a runtime error if processing fails.
func (t *TxrImplSql) Tx(ctx context.Context, fn func(txCtx *TxCtx) error) error {
	return t.processTx(true, ctx, fn)
}

func (t *TxrImplSql) processTx(
	// todo : perhaps, there should be RO/RW-transactions ???
	isWritable bool,
	ctx context.Context,
	fn func(txCtx *TxCtx) error,
) error {
	if ctx == nil {
		panic(fmt.Errorf("%T : ctx must not be nil", t))
	} else if IsInTxCtx(ctx) {
		panic(fmt.Errorf("%T : nested transactions are not allowed", t))
	}

	if fn == nil {
		panic(fmt.Errorf("%T : fn (transaction body) must not be nil", t))
	}

	// --

	var deadlockRetries uint

	for {
		err := t.tx(isWritable, ctx, fn)

		if err == nil {
			break
		}

		if strings.Contains(strings.ToLower(err.Error()), "deadlock") ||
			strings.Contains(strings.ToLower(err.Error()), "lock wait timeout") ||
			strings.Contains(strings.ToLower(err.Error()), "could not obtain lock") {
			if deadlockRetries == t.deadlockMaxRetries {
				return fmt.Errorf(
					"%T : deadlock retry limit (%d) exceeded. Originally caused by : %w",
					t,
					t.deadlockMaxRetries,
					err,
				)
			}

			deadlockRetries++

			select {
			case <-ctx.Done():
				err = fmt.Errorf(
					"%T : transaction retry #%d (originally caused by: %v) cancelled by context: %w",
					t,
					deadlockRetries,
					err,
					ctx.Err(),
				)
			case <-time.After(t.deadlockMinRetryInterval * time.Duration(deadlockRetries)):
				continue
			}
		}

		return err
	}

	// --

	// Unreachable (loop either returns or continues)
	return nil
}

func (t *TxrImplSql) tx(
	isWritable bool,
	ctx context.Context,
	fn func(txCtx *TxCtx) error,
) error {
	sqlTx, err := t.db.BeginTx(ctx, nil)

	defer func() {
		if sqlTx != nil {
			_ = sqlTx.Rollback()
		}
	}()

	if err != nil {
		return err
	}

	txCtx := WithTxCtx(ctx, sqlTx)

	err = fn(txCtx)

	if err == nil && isWritable {
		err = sqlTx.Commit()
	}

	return err
}
