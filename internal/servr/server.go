package servr

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr/dao"
)

type Servr interface {
	GetSessionByToken(context.Context, string) (data.Session, error)
	Login(context.Context, data.LoginData) (data.TokenPersonData, error)
	AddLogin(ctx context.Context, session data.Session, loginData data.LoginData) (data.Person, error)
	RemoveLogin(ctx context.Context, session data.Session, personId string) error
	UpdatePriority(ctx context.Context, session data.Session, priority int, todoItemId string) error
	ToggleCompleted(ctx context.Context, persons []string, todoItemId string) (data.CompletedData, error)
	Logout(ctx context.Context, sessionId string) error
	SwitchSeqNo(ctx context.Context, persons []string, ids []string) (data.CompleteItemData, error)
	RemoveTodo(ctx context.Context, persons []string, ids []string) error
	GetTodo(ctx context.Context, persons []string, todoItemId string) (data.TodoItem, error)
	UpdateDue(ctx context.Context, persons []string, input data.UpdateDueData) error
	UpdateTodo(ctx context.Context, persons []string, input data.UpdateTodoItem) error
	RemoveAttachment(ctx context.Context, persons []string, ids []string) error
	SearchTodo(ctx context.Context, persons []string, input data.TodoSearch) ([]data.TodoItem, error)
	MultipleTodo(ctx context.Context, persons []string, input data.MultipleTodoData) ([]data.TodoItem, error)
	BulkUpdateTodo(ctx context.Context, persons []string, input data.BulkUpdateData) ([]data.TodoItem, error)
	RemoveItems(ctx context.Context, persons []string, ids []string) error
	UpdateItem(ctx context.Context, persons []string, input data.UpdateItemData) (data.ItemData, error)
	UpdateAttachment(ctx context.Context, persons []string, adding bool, attachment data.AttachmentData) error
	GetAttachmentById(ctx context.Context, attachmentId string) (data.AttachmentData, error)
	AttachmentExists(ctx context.Context, attachmentId string) (bool, error)
	GetFlaggedItems(ctx context.Context, persons []string) ([]data.CompleteItemData, error)
	SearchItems(ctx context.Context, persons []string, search string) ([]data.CompleteItemData, error)
	GetMaxAttachmentSeq(ctx context.Context, id string) (int, error)
	GetUploadedFileName(id string) string
	LoadItem(ctx context.Context, itemId string) (data.CompleteItemData, error)
	GetPasswordByUserId(ctx context.Context, id string, password string) (data.PasswordData, error)
}

type server struct {
	db     *sql.DB
	config data.ServerConfig
}

func getPasswordData(ctx context.Context, tx *sql.Tx, personId string) (salt []byte, payload []byte, err error) {
	saltStr, payloadStr, err := dao.GetPersonPasswordData(ctx, tx, personId)
	if err != nil || saltStr == "" {
		return nil, nil, err
	}
	salt, err = base64.StdEncoding.DecodeString(saltStr)
	if err != nil {
		return nil, nil, err
	}
	if payloadStr != "" {
		payload, err = base64.StdEncoding.DecodeString(payloadStr)
		if err != nil {
			return nil, nil, err
		}
	}
	return salt, payload, nil
}

func updatePasswordData(ctx context.Context, tx *sql.Tx, personId string, salt []byte, payload []byte) error {
	return dao.UpdatePersonPasswordData(
		ctx, tx, personId,
		base64.StdEncoding.EncodeToString(salt),
		base64.StdEncoding.EncodeToString(payload),
	)
}

func (h *server) GetPasswordByUserId(ctx context.Context, personId string, password string) (data.PasswordData, error) {
	res := data.PasswordData{Passwords: make([]data.SecretData, 0)}

	tx, err := begin(ctx, h.db)
	if err != nil {
		return res, err
	}
	defer rollback(tx)

	salt, encryptedPayload, err := getPasswordData(ctx, tx, personId)
	if err != nil {
		return res, err
	}

	if len(salt) == 0 {
		return res, tx.Commit()
	}

	if len(encryptedPayload) == 0 {
		return res, errors.New("payload is empty but salt is not — data is inconsistent")
	}

	key := deriveKey(password, salt)
	plaintext, err := decryptAES(key, encryptedPayload)
	if err != nil {
		return res, errors.New("the provided password is not correct")
	}

	err = json.Unmarshal(plaintext, &res)
	if err != nil {
		return res, err
	}

	return res, tx.Commit()
}

func (h *server) LoadItem(ctx context.Context, itemId string) (data.CompleteItemData, error) {
	res := data.CompleteItemData{}
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return res, err
	}

	defer rollback(tx)

	itemList, err := dao.LoadCompleteItemsByIds(ctx, tx, []string{itemId})
	if err != nil {
		return res, err
	}

	if len(itemList) == 0 {
		return res, fmt.Errorf("cannot find item with id %s", itemId)
	}

	err = tx.Commit()
	if err != nil {
		return res, err
	}

	res = itemList[0]
	return res, nil
}

func (h *server) GetMaxAttachmentSeq(ctx context.Context, itemId string) (int, error) {
	res := 0
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return res, err
	}

	defer rollback(tx)

	res, err = dao.GetMaxAttachmentSeq(ctx, tx, itemId)
	if err != nil {
		return res, err
	}

	err = tx.Commit()

	return res, err
}

func GetServr(db *sql.DB, config data.ServerConfig) Servr {
	return &server{db, config}
}

func begin(ctx context.Context, db *sql.DB) (*sql.Tx, error) {
	tx, err := db.BeginTx(ctx, nil)

	return tx, err
}

func rollback(tx *sql.Tx) {
	_ = tx.Rollback()
}

func (h *server) GetSessionByToken(ctx context.Context, token string) (data.Session, error) {
	res := data.Session{}
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return res, err
	}

	defer rollback(tx)

	res, err = dao.GetSessionByToken(ctx, tx, token)
	if err != nil {
		return res, err
	}

	// if the session is expired, return error
	if res.Expired {
		return res, errors.New("session expired")
	}

	err = tx.Commit()

	return res, err
}

func (h *server) GetUploadedFileName(attachmentId string) string {
	folder := h.config.StorageFolder
	return fmt.Sprintf("%s%s.dat", folder, attachmentId)
}
