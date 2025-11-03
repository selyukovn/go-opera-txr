
## [0.2.0] - 2025-11-04

*Note: v0.1.0 has been retracted.*

### BREAKING CHANGES
- `TxrImplSql` now requires `deadlockDetectionFn` function to handle driver-specific deadlock error detection 
  (e.g. MySQL deadlock error code = 1213, PostgreSQL = 40P01, etc.).

### FEATURES
- Added tests for `TxrImplSql`

### BUGS FIXED
- `NewTxrImplSql` panics on invalid (nil) arguments `db` or `deadlockDetectionFn`.
- `TxrImplSql.Tx` no longer ignores context cancellation during `fn` execution.

### IMPROVEMENTS
- Downgraded minimal Go version to `1.18`  (no functional dependency on higher versions).
- Updated `TxrImplSql` and `TxrInterface.Tx` documentation

---

## [0.1.0] - 2025-11-03

*Note: This release lacked a CHANGELOG section.*

### FEATURES
- Defined basic abstractions.
- Initial release of `TxrImplSql` with retries on deadlock.
