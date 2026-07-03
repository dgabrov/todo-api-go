package servr

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"gitlab.com/dgb9/todo-api/internal/cst"
	"gitlab.com/dgb9/todo-api/internal/data"
)

// TestLogin_Success tests successful login with IDP authentication
func TestLogin_Success(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	// Mock IDP server
	idpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id":"provided-id-123","name":"John Doe","login":"johndoe","rights":["%s"]}`, cst.AccessRightNeeded)
	}))
	defer idpServer.Close()

	config := data.ServerConfig{AuthServerUrl: idpServer.URL}
	srv := GetServr(tdb.DB, config)

	loginData := data.LoginData{Login: "johndoe", Password: "password123"}
	result, err := srv.Login(ctx, loginData)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Name != "John Doe" || result.Login != "johndoe" || result.Token == "" {
		t.Error("login result missing expected fields")
	}

	_, err = srv.GetSessionByToken(ctx, result.Token)
	if err != nil {
		t.Errorf("session not found after login: %v", err)
	}
}

// TestLogin_MissingRights tests login with insufficient permissions
func TestLogin_MissingRights(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)

	idpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"user-id","name":"User","login":"user","rights":["some_other_right"]}`)
	}))
	defer idpServer.Close()

	config := data.ServerConfig{AuthServerUrl: idpServer.URL}
	srv := GetServr(tdb.DB, config)

	_, err := srv.Login(ctx, data.LoginData{Login: "user", Password: "pass"})
	if err == nil || err.Error() != "you don't have the right to connect to this application" {
		t.Errorf("expected right error, got %v", err)
	}
}

// TestAddLogin_Success tests adding second login to session
func TestAddLogin_Success(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	// Create first person and session
	person1ID := uuid.New().String()
	sessionID := uuid.New().String()
	tdb.InsertPerson(ctx, person1ID, "provided-1", "Person 1", "user1")
	tdb.InsertSession(ctx, sessionID, "token123", []string{person1ID})

	// Mock IDP for second login
	idpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id":"provided-2","name":"Person 2","login":"user2","rights":["%s"]}`, cst.AccessRightNeeded)
	}))
	defer idpServer.Close()

	config := data.ServerConfig{AuthServerUrl: idpServer.URL}
	srv := GetServr(tdb.DB, config)

	session := data.Session{SessionId: sessionID, Persons: []string{person1ID}}
	person, err := srv.AddLogin(ctx, session, data.LoginData{Login: "user2", Password: "pass"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if person.Name != "Person 2" || person.Login != "user2" {
		t.Error("added person missing expected fields")
	}
}

// TestAddLogin_AlreadyLoggedIn tests adding user already in session
func TestAddLogin_AlreadyLoggedIn(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	personID := uuid.New().String()
	sessionID := uuid.New().String()
	tdb.InsertPerson(ctx, personID, "provided-id", "User", "username")
	tdb.InsertSession(ctx, sessionID, "token", []string{personID})

	idpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id":"provided-id","name":"User","login":"username","rights":["%s"]}`, cst.AccessRightNeeded)
	}))
	defer idpServer.Close()

	config := data.ServerConfig{AuthServerUrl: idpServer.URL}
	srv := GetServr(tdb.DB, config)

	session := data.Session{SessionId: sessionID, Persons: []string{personID}}
	_, err := srv.AddLogin(ctx, session, data.LoginData{Login: "username", Password: "pass"})

	if err == nil {
		t.Fatal("expected error for already logged in user")
	}
}

// TestGetSessionByToken_Success tests retrieving session by token
func TestGetSessionByToken_Success(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	personID := uuid.New().String()
	sessionID := uuid.New().String()
	token := uuid.New().String()

	tdb.InsertPerson(ctx, personID, "prov-id", "User", "user")
	tdb.InsertSession(ctx, sessionID, token, []string{personID})

	srv := GetServr(tdb.DB, data.ServerConfig{})
	session, err := srv.GetSessionByToken(ctx, token)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if session.SessionId != sessionID || len(session.Persons) != 1 || session.Persons[0] != personID {
		t.Error("session data mismatch")
	}
}

// TestGetSessionByToken_Expired tests expired session rejection
func TestGetSessionByToken_Expired(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	personID := uuid.New().String()
	sessionID := uuid.New().String()
	token := uuid.New().String()

	tdb.InsertPerson(ctx, personID, "prov-id", "User", "user")
	tdb.InsertSession(ctx, sessionID, token, []string{personID})

	// Mark session as expired
	tdb.DB.ExecContext(ctx, `update session set expired_ind = 'Y' where session_id = ?`, sessionID)

	srv := GetServr(tdb.DB, data.ServerConfig{})
	_, err := srv.GetSessionByToken(ctx, token)

	if err == nil || err.Error() != "session expired" {
		t.Errorf("expected session expired error, got %v", err)
	}
}

// TestLogout_Success tests session expiration
func TestLogout_Success(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	personID := uuid.New().String()
	sessionID := uuid.New().String()
	token := uuid.New().String()

	tdb.InsertPerson(ctx, personID, "prov-id", "User", "user")
	tdb.InsertSession(ctx, sessionID, token, []string{personID})

	srv := GetServr(tdb.DB, data.ServerConfig{})
	err := srv.Logout(ctx, sessionID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify session is expired
	_, err = srv.GetSessionByToken(ctx, token)
	if err == nil {
		t.Fatal("expected error for expired session")
	}
}

// TestUpdatePriority_Success tests updating todo priority
func TestUpdatePriority_Success(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	personID := uuid.New().String()
	todoID := uuid.New().String()

	tdb.InsertPerson(ctx, personID, "prov-id", "User", "user")
	tdb.InsertTodoItem(ctx, todoID, personID, "test", 10)

	srv := GetServr(tdb.DB, data.ServerConfig{})
	session := data.Session{SessionId: uuid.New().String(), Persons: []string{personID}}

	err := srv.UpdatePriority(ctx, session, 50, todoID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify priority was updated
	var priority int
	tdb.DB.QueryRowContext(ctx, `select priority from todo_item where todo_item_id = ?`, todoID).Scan(&priority)
	if priority != 50 {
		t.Errorf("expected priority 50, got %d", priority)
	}
}

// TestUpdatePriority_InvalidPriority tests priority validation
func TestUpdatePriority_InvalidPriority(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	personID := uuid.New().String()
	todoID := uuid.New().String()

	tdb.InsertPerson(ctx, personID, "prov-id", "User", "user")
	tdb.InsertTodoItem(ctx, todoID, personID, "test", 10)

	srv := GetServr(tdb.DB, data.ServerConfig{})
	session := data.Session{SessionId: uuid.New().String(), Persons: []string{personID}}

	testCases := []int{-1, 101, -100, 1000}
	for _, priority := range testCases {
		err := srv.UpdatePriority(ctx, session, priority, todoID)
		if err == nil {
			t.Errorf("expected error for invalid priority %d", priority)
		}
	}
}

// TestUpdatePriority_Unauthorized tests unauthorized user rejection
func TestUpdatePriority_Unauthorized(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	person1ID := uuid.New().String()
	person2ID := uuid.New().String()
	todoID := uuid.New().String()

	tdb.InsertPerson(ctx, person1ID, "prov-1", "User1", "user1")
	tdb.InsertPerson(ctx, person2ID, "prov-2", "User2", "user2")
	tdb.InsertTodoItem(ctx, todoID, person1ID, "test", 10)

	srv := GetServr(tdb.DB, data.ServerConfig{})
	session := data.Session{SessionId: uuid.New().String(), Persons: []string{person2ID}}

	err := srv.UpdatePriority(ctx, session, 50, todoID)
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

// TestToggleCompleted_Success tests toggling completion state
func TestToggleCompleted_Success(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	personID := uuid.New().String()
	todoID := uuid.New().String()

	tdb.InsertPerson(ctx, personID, "prov-id", "User", "user")
	tdb.InsertTodoItem(ctx, todoID, personID, "test", 10)

	srv := GetServr(tdb.DB, data.ServerConfig{})
	result, err := srv.ToggleCompleted(ctx, []string{personID}, todoID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should be toggled from N to Y
	if !result.Completed {
		t.Error("expected completed to be toggled to true")
	}
}

// TestToggleCompleted_Unauthorized tests unauthorized toggle
func TestToggleCompleted_Unauthorized(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	person1ID := uuid.New().String()
	person2ID := uuid.New().String()
	todoID := uuid.New().String()

	tdb.InsertPerson(ctx, person1ID, "prov-1", "User1", "user1")
	tdb.InsertPerson(ctx, person2ID, "prov-2", "User2", "user2")
	tdb.InsertTodoItem(ctx, todoID, person1ID, "test", 10)

	srv := GetServr(tdb.DB, data.ServerConfig{})
	_, err := srv.ToggleCompleted(ctx, []string{person2ID}, todoID)

	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

// TestRemoveTodo_Success tests removing todos
func TestRemoveTodo_Success(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	personID := uuid.New().String()
	todo1ID := uuid.New().String()
	todo2ID := uuid.New().String()

	tdb.InsertPerson(ctx, personID, "prov-id", "User", "user")
	tdb.InsertTodoItem(ctx, todo1ID, personID, "test1", 10)
	tdb.InsertTodoItem(ctx, todo2ID, personID, "test2", 20)

	srv := GetServr(tdb.DB, data.ServerConfig{})
	err := srv.RemoveTodo(ctx, []string{personID}, []string{todo1ID, todo2ID})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify todos are deleted
	var count int
	tdb.DB.QueryRowContext(ctx, `select count(*) from todo_item where person_id = ?`, personID).Scan(&count)
	if count != 0 {
		t.Errorf("expected todos to be deleted, found %d", count)
	}
}

// TestRemoveTodo_Unauthorized tests unauthorized removal
func TestRemoveTodo_Unauthorized(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	person1ID := uuid.New().String()
	person2ID := uuid.New().String()
	todoID := uuid.New().String()

	tdb.InsertPerson(ctx, person1ID, "prov-1", "User1", "user1")
	tdb.InsertPerson(ctx, person2ID, "prov-2", "User2", "user2")
	tdb.InsertTodoItem(ctx, todoID, person1ID, "test", 10)

	srv := GetServr(tdb.DB, data.ServerConfig{})
	err := srv.RemoveTodo(ctx, []string{person2ID}, []string{todoID})

	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

// TestGetTodo_Success tests retrieving todo item
func TestGetTodo_Success(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	personID := uuid.New().String()
	todoID := uuid.New().String()

	tdb.InsertPerson(ctx, personID, "prov-id", "User", "user")
	tdb.InsertTodoItem(ctx, todoID, personID, "test comments", 25)

	srv := GetServr(tdb.DB, data.ServerConfig{})
	result, err := srv.GetTodo(ctx, []string{personID}, todoID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.TodoItemId != todoID || result.Comments != "test comments" || result.Priority != 25 {
		t.Error("todo data mismatch")
	}
}

// TestGetTodo_Unauthorized tests unauthorized access
func TestGetTodo_Unauthorized(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	person1ID := uuid.New().String()
	person2ID := uuid.New().String()
	todoID := uuid.New().String()

	tdb.InsertPerson(ctx, person1ID, "prov-1", "User1", "user1")
	tdb.InsertPerson(ctx, person2ID, "prov-2", "User2", "user2")
	tdb.InsertTodoItem(ctx, todoID, person1ID, "test", 10)

	srv := GetServr(tdb.DB, data.ServerConfig{})
	_, err := srv.GetTodo(ctx, []string{person2ID}, todoID)

	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

// TestLoadItem_Success tests loading complete item data
func TestLoadItem_Success(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	personID := uuid.New().String()
	itemID := uuid.New().String()

	tdb.InsertPerson(ctx, personID, "prov-id", "User", "user")
	tdb.InsertItem(ctx, itemID, personID, "Test Item")

	srv := GetServr(tdb.DB, data.ServerConfig{})
	result, err := srv.LoadItem(ctx, itemID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.ItemId != itemID {
		t.Error("item id mismatch")
	}
}

// TestLoadItem_NotFound tests missing item
func TestLoadItem_NotFound(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)

	srv := GetServr(tdb.DB, data.ServerConfig{})
	_, err := srv.LoadItem(ctx, "nonexistent-id")

	if err == nil {
		t.Fatal("expected error for missing item")
	}
}

// TestGetMaxAttachmentSeq_Success tests retrieving attachment sequence
func TestGetMaxAttachmentSeq_Success(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	personID := uuid.New().String()
	itemID := uuid.New().String()
	attachmentID := uuid.New().String()

	tdb.InsertPerson(ctx, personID, "prov-id", "User", "user")
	tdb.InsertItem(ctx, itemID, personID, "Test")
	tdb.DB.ExecContext(ctx, `insert into attachment (attachment_id, item_id, seq_no, added_dt, updated_dt)
	                          values (?, ?, ?, now(), now())`, attachmentID, itemID, 42)

	srv := GetServr(tdb.DB, data.ServerConfig{})
	seq, err := srv.GetMaxAttachmentSeq(ctx, itemID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if seq != 42 {
		t.Errorf("expected sequence 42, got %d", seq)
	}
}

// TestMultipleSessions tests user with multiple concurrent sessions
func TestMultipleSessions(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	personID := uuid.New().String()
	session1ID := uuid.New().String()
	session2ID := uuid.New().String()
	token1 := uuid.New().String()
	token2 := uuid.New().String()

	tdb.InsertPerson(ctx, personID, "prov-id", "User", "user")
	tdb.InsertSession(ctx, session1ID, token1, []string{personID})
	tdb.InsertSession(ctx, session2ID, token2, []string{personID})

	srv := GetServr(tdb.DB, data.ServerConfig{})

	// Both sessions should be valid
	_, err1 := srv.GetSessionByToken(ctx, token1)
	_, err2 := srv.GetSessionByToken(ctx, token2)

	if err1 != nil || err2 != nil {
		t.Fatal("both sessions should be valid")
	}

	// Logout one session
	srv.Logout(ctx, session1ID)

	// First session should be expired
	_, err1 = srv.GetSessionByToken(ctx, token1)
	_, err2 = srv.GetSessionByToken(ctx, token2)

	if err1 == nil {
		t.Fatal("first session should be expired")
	}
	if err2 != nil {
		t.Fatal("second session should still be valid")
	}
}

// TestMultipleUsersInSession tests session with multiple users
func TestMultipleUsersInSession(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	person1ID := uuid.New().String()
	person2ID := uuid.New().String()
	sessionID := uuid.New().String()

	tdb.InsertPerson(ctx, person1ID, "prov-1", "User1", "user1")
	tdb.InsertPerson(ctx, person2ID, "prov-2", "User2", "user2")
	tdb.InsertSession(ctx, sessionID, uuid.New().String(), []string{person1ID, person2ID})

	srv := GetServr(tdb.DB, data.ServerConfig{})
	session := data.Session{SessionId: sessionID, Persons: []string{person1ID, person2ID}}

	// Remove person1
	err := srv.RemoveLogin(ctx, session, person1ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Can't remove person2 if they're the only one left
	session.Persons = []string{person2ID}
	err = srv.RemoveLogin(ctx, session, person2ID)
	if err == nil {
		t.Fatal("expected error when removing only person")
	}
}

// TestUpdateDue_Success tests updating todo due date
func TestUpdateDue_Success(t *testing.T) {
	ctx := context.Background()
	tdb := setupTestDB(ctx, t)
	defer tdb.Cleanup(t)
	defer tdb.ClearTables(ctx)

	personID := uuid.New().String()
	todoID := uuid.New().String()

	tdb.InsertPerson(ctx, personID, "prov-id", "User", "user")
	tdb.InsertTodoItem(ctx, todoID, personID, "test", 10)

	srv := GetServr(tdb.DB, data.ServerConfig{})
	dueTime := time.Now().Add(24 * time.Hour)
	input := data.UpdateDueData{TodoItemId: todoID, Due: &dueTime}

	err := srv.UpdateDue(ctx, []string{personID}, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify due date was set
	var due *time.Time
	tdb.DB.QueryRowContext(ctx, `select due_dt from todo_item where todo_item_id = ?`, todoID).Scan(&due)
	if due == nil {
		t.Error("expected due date to be set")
	}
}
