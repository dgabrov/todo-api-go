package servr

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gitlab.com/dgb9/todo-api/internal/data"
)

// Basic tests for Servr layer methods that don't require complex mocking
func TestRemoveLogin_EmptyPersonId(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	ctx := context.Background()
	config := data.ServerConfig{}
	srv := GetServr(db, config)

	session := data.Session{
		SessionId: "session-123",
		Persons:   []string{"person-111", "person-222"},
	}

	err := srv.RemoveLogin(ctx, session, "  ")
	if err == nil {
		t.Fatal("expected error for empty person id")
	}

	if err.Error() != "person id not filled out" {
		t.Errorf("expected empty person id error, got %v", err)
	}
}

func TestRemoveLogin_OnlyPerson(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	ctx := context.Background()
	config := data.ServerConfig{}
	srv := GetServr(db, config)

	session := data.Session{
		SessionId: "session-123",
		Persons:   []string{"person-111"},
	}

	err := srv.RemoveLogin(ctx, session, "person-111")
	if err == nil {
		t.Fatal("expected error when removing only person")
	}

	if err.Error() != "you cannot remove the only logged in person in the current session" {
		t.Errorf("expected only person error, got %v", err)
	}
}

func TestRemoveLogin_PersonNotLogged(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	ctx := context.Background()
	config := data.ServerConfig{}
	srv := GetServr(db, config)

	session := data.Session{
		SessionId: "session-123",
		Persons:   []string{"person-111"},
	}

	err := srv.RemoveLogin(ctx, session, "person-999")
	if err == nil {
		t.Fatal("expected error for person not in session")
	}

	if err.Error() != "person with id person-999 not logged in" {
		t.Errorf("expected not in session error, got %v", err)
	}
}

func TestGetUploadedFileName_WithTrailingSlash(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	config := data.ServerConfig{
		StorageFolder: "/data/uploads/",
	}
	srv := GetServr(db, config)

	result := srv.GetUploadedFileName("attach-id-123")
	expected := "/data/uploads/attach-id-123.dat"

	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestGetUploadedFileName_WithoutTrailingSlash(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	config := data.ServerConfig{
		StorageFolder: "/data/uploads",
	}
	srv := GetServr(db, config)

	result := srv.GetUploadedFileName("attach-id-123")
	expected := "/data/uploadsattach-id-123.dat"

	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

// Input validation tests
func TestValidateLoginInput_EmptyLogin(t *testing.T) {
	loginData := data.LoginData{
		Login:    "",
		Password: "password",
	}

	err := checkLoginData(loginData)
	if err == nil {
		t.Fatal("expected error for empty login")
	}

	if err.Error() != "login is mandatory and it seems to be missing" {
		t.Errorf("expected empty login error, got %v", err)
	}
}

func TestValidateLoginInput_WhitespaceLogin(t *testing.T) {
	loginData := data.LoginData{
		Login:    "   \t\n   ",
		Password: "password",
	}

	err := checkLoginData(loginData)
	if err == nil {
		t.Fatal("expected error for whitespace login")
	}

	if err.Error() != "login is mandatory and it seems to be missing" {
		t.Errorf("expected empty login error, got %v", err)
	}
}

func TestValidateLoginInput_EmptyPassword(t *testing.T) {
	loginData := data.LoginData{
		Login:    "testuser",
		Password: "",
	}

	err := checkLoginData(loginData)
	if err == nil {
		t.Fatal("expected error for empty password")
	}

	if err.Error() != "empty passwords are no longer valid, please provide the password" {
		t.Errorf("expected empty password error, got %v", err)
	}
}

func TestValidateLoginInput_WhitespacePassword(t *testing.T) {
	loginData := data.LoginData{
		Login:    "testuser",
		Password: "   ",
	}

	err := checkLoginData(loginData)
	if err == nil {
		t.Fatal("expected error for whitespace password")
	}

	if err.Error() != "empty passwords are no longer valid, please provide the password" {
		t.Errorf("expected empty password error, got %v", err)
	}
}

func TestValidateLoginInput_Valid(t *testing.T) {
	loginData := data.LoginData{
		Login:    "testuser",
		Password: "password",
	}

	err := checkLoginData(loginData)
	if err != nil {
		t.Errorf("expected no error for valid input, got %v", err)
	}
}

// Crypto function tests
func TestGenerateSalt(t *testing.T) {
	salt, err := generateSalt()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(salt) != 16 {
		t.Errorf("expected salt length 16, got %d", len(salt))
	}

	// Two salts should be different (probabilistically)
	salt2, _ := generateSalt()
	if string(salt) == string(salt2) {
		t.Error("expected different salts")
	}
}

func TestEncryptDecryptAES(t *testing.T) {
	salt, _ := generateSalt()
	password := "test-password"
	key := deriveKey(password, salt)

	plaintext := []byte("test data to encrypt")
	encrypted, err := encryptAES(key, plaintext)
	if err != nil {
		t.Fatalf("expected no error on encrypt, got %v", err)
	}

	if string(encrypted) == string(plaintext) {
		t.Error("encrypted data should not equal plaintext")
	}

	decrypted, err := decryptAES(key, encrypted)
	if err != nil {
		t.Fatalf("expected no error on decrypt, got %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("expected decrypted %s, got %s", plaintext, decrypted)
	}
}

func TestEncryptDecryptAES_WrongKey(t *testing.T) {
	salt1, _ := generateSalt()
	salt2, _ := generateSalt()
	password := "test-password"
	wrongPassword := "wrong-password"

	key1 := deriveKey(password, salt1)
	key2 := deriveKey(wrongPassword, salt2)

	plaintext := []byte("test data")
	encrypted, _ := encryptAES(key1, plaintext)

	_, err := decryptAES(key2, encrypted)
	if err == nil {
		t.Fatal("expected error when decrypting with wrong key")
	}
}

func TestDeriveKey_ConsistentOutput(t *testing.T) {
	salt, _ := generateSalt()
	password := "test-password"

	key1 := deriveKey(password, salt)
	key2 := deriveKey(password, salt)

	if string(key1) != string(key2) {
		t.Error("expected consistent key derivation for same inputs")
	}
}

func TestDeriveKey_DifferentPassword(t *testing.T) {
	salt, _ := generateSalt()
	password1 := "password1"
	password2 := "password2"

	key1 := deriveKey(password1, salt)
	key2 := deriveKey(password2, salt)

	if string(key1) == string(key2) {
		t.Error("expected different keys for different passwords")
	}
}
