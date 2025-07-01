package controller

import (
	"database/sql"
	"errors"
	"fmt"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"io"
	"net/http"
	"os"
	"slices"
)

type download struct {
	db     *sql.DB
	config data.ServerConfig
}

func downloadController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &download{db, config}
}

func (h *download) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.processDownload(w, r)

	// if err is nil then the file was already sent to the client
	if err != nil {
		writeResponse(w, nil, err)
	}
}

func (h *download) processDownload(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	srv := servr.GetServr(h.db, h.config)
	session, err := getSession(ctx, r, srv)
	if err != nil {
		return err
	}

	persons := session.Persons

	// get attachment id
	id := r.URL.Query().Get("id")
	if len(id) == 0 {
		return errors.New("please provide the attachment id")
	}
	att, err := srv.GetAttachmentById(ctx, id)
	if err != nil {
		return err
	}

	itemId := att.ItemId
	item, err := srv.LoadItem(ctx, itemId)

	if err != nil {
		return err
	}

	person := item.PersonId

	if !slices.Contains(persons, person) {
		return errors.New("the id you mentioned is for an attachment that does not belong to any of the logged in users")
	}

	// ok, all good now, look for the file
	fileName := srv.GetUploadedFileName(id)

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}

	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	// all right, file exists, pipe it back
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", att.FileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	_, err = io.Copy(w, file)

	return err
}
