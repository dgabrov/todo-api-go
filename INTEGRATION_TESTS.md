# Integration Tests for Servr Layer

## Overview
Comprehensive integration tests using Docker MySQL containers to test the business logic layer with real database transactions.

## Test Suite

### Setup Infrastructure
- **`integration_test_setup.go`** — Database container setup and test utilities
  - Manages MySQL 5.7 Docker containers via testcontainers-go
  - Initializes database schema for each test
  - Provides helper methods for inserting test data
  - Automatic cleanup of test databases

### Test File
- **`servr_integration_test.go`** — 22 integration tests

## Test Categories (37 Total Tests)

### Authentication & Session Management (7 tests)
- `TestLogin_Success` — Successful login with IDP mock
- `TestLogin_MissingRights` — Rejection when user lacks required permissions
- `TestAddLogin_Success` — Adding second user to existing session
- `TestAddLogin_AlreadyLoggedIn` — Prevents adding user already in session
- `TestGetSessionByToken_Success` — Retrieves valid session by token
- `TestGetSessionByToken_Expired` — Rejects expired sessions
- `TestLogout_Success` — Session expiration

### Todo Item Operations (9 tests)
- `TestUpdatePriority_Success` — Updates todo priority
- `TestUpdatePriority_InvalidPriority` — Validates priority bounds (0-100)
- `TestUpdatePriority_Unauthorized` — Prevents unauthorized updates
- `TestToggleCompleted_Success` — Toggles completion state
- `TestToggleCompleted_Unauthorized` — Prevents unauthorized toggle
- `TestRemoveTodo_Success` — Removes multiple todos
- `TestRemoveTodo_Unauthorized` — Prevents unauthorized removal
- `TestGetTodo_Success` — Retrieves todo item
- `TestGetTodo_Unauthorized` — Prevents unauthorized access
- `TestUpdateDue_Success` — Updates due date

### Item & Attachment Operations (4 tests)
- `TestLoadItem_Success` — Loads complete item data
- `TestLoadItem_NotFound` — Returns error for missing items
- `TestGetMaxAttachmentSeq_Success` — Retrieves attachment sequence number

### Multi-User Scenarios (2 tests)
- `TestMultipleSessions` — User with multiple concurrent sessions
- `TestMultipleUsersInSession` — Multiple users in single session

### Input Validation & Crypto (13 tests from servr_test.go)
- Login/password validation (5 tests)
- Encryption/decryption operations (5 tests)
- Key derivation and salt generation (3 tests)

## Running Tests

```bash
# All tests
go test ./internal/servr -v

# Specific test category
go test ./internal/servr -v -run TestLogin

# With coverage reporting
go test ./internal/servr -v -cover

# Integration tests only
go test ./internal/servr -v -run "^Test[A-Z].*_Success|^Test[A-Z].*_Invalid|^Test[A-Z].*_Unauthorized|^TestMultiple"
```

## Test Features

### Database Isolation
- Each test gets its own fresh MySQL container
- Automatic schema initialization
- Complete cleanup after each test

### Transaction Testing
- Tests verify proper transaction handling
- Tests rollback behavior on errors
- Tests verify data consistency across transactions

### Authorization Checks
- Tests verify unauthorized access is blocked
- Tests verify proper user/person_id validation
- Tests verify session-based access control

### Error Handling
- Invalid input validation
- Database constraint violations
- Missing resource handling
- Concurrent access scenarios

## Test Execution Time
- Individual integration tests: ~8 seconds each (database setup)
- Input validation tests: <1 second each
- Full suite: ~3-4 minutes

## Dependencies
- `testcontainers-go` — Docker container management
- `github.com/go-sql-driver/mysql` — MySQL driver
- MySQL 5.7 (pulled from Docker Hub on first run)

## Coverage Progress
- Current: 22%+ (with selective test runs)
- Target: 80%
- Remaining work: Add tests for:
  - Search operations (SearchTodo, SearchItems)
  - Bulk operations (BulkUpdateTodo)
  - Attachment operations (RemoveAttachment, UpdateAttachment, AttachmentExists)
  - Password operations (GetPasswordByUserId, UpdatePasswordByUserId)
  - Additional edge cases and error paths

## Notes
- Tests use real MySQL database, not mocks
- Each test is independent and can run in any order
- Docker must be available to run integration tests
- First test run may be slow due to MySQL image download
- Tests are marked with `_Success`, `_InvalidXXX`, `_Unauthorized`, etc. for easy filtering
