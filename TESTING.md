# Servr Layer Unit Tests

## Overview
Comprehensive unit tests for the business logic layer (`internal/servr`) have been created covering input validation, crypto operations, and core business logic.

## Test Coverage

### Input Validation (5 tests)
- `TestValidateLoginInput_EmptyLogin` - Validates that empty login is rejected
- `TestValidateLoginInput_WhitespaceLogin` - Validates that whitespace-only login is rejected
- `TestValidateLoginInput_EmptyPassword` - Validates that empty password is rejected
- `TestValidateLoginInput_WhitespacePassword` - Validates that whitespace-only password is rejected
- `TestValidateLoginInput_Valid` - Validates that valid input is accepted

### Crypto Operations (5 tests)
- `TestGenerateSalt` - Verifies salt generation produces unique 16-byte salts
- `TestEncryptDecryptAES` - Validates AES encryption/decryption roundtrip
- `TestEncryptDecryptAES_WrongKey` - Ensures wrong key cannot decrypt data
- `TestDeriveKey_ConsistentOutput` - Validates consistent key derivation for same inputs
- `TestDeriveKey_DifferentPassword` - Validates different passwords produce different keys

### Business Logic (5 tests)
- `TestRemoveLogin_EmptyPersonId` - Validates empty person ID rejection
- `TestRemoveLogin_OnlyPerson` - Prevents removing the only logged-in person
- `TestRemoveLogin_PersonNotLogged` - Validates person must be in session
- `TestGetUploadedFileName_WithTrailingSlash` - Tests file path construction with trailing slash
- `TestGetUploadedFileName_WithoutTrailingSlash` - Tests file path construction without trailing slash

## Running Tests

```bash
# Run all tests
go test ./internal/servr -v

# Run specific test
go test ./internal/servr -run TestValidateLoginInput -v

# Run with coverage
go test ./internal/servr -v -cover
```

## Test Results
All 15 tests pass successfully. Tests are designed to be fast and do not require external dependencies or database mocks.

## Future Enhancements
For database-intensive operations (session management, todo operations, etc.), integration tests with a test database are recommended:
- Use Docker containers (MySQL) for integration tests
- Test transaction behavior and rollback scenarios
- Verify database constraints and data consistency
