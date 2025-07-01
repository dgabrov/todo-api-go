package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"io"
	"net/http"
	"os"
)

type updateAttachment struct {
	db     *sql.DB
	config data.ServerConfig
}

func updateAttachmentController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &updateAttachment{db, config}
}

func (h *updateAttachment) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	attach, err := h.processUpdateAttachment(r)

	writeResponse(w, attach, err)
}

func (h *updateAttachment) processUpdateAttachment(r *http.Request) (data.AttachmentData, error) {
	res := data.AttachmentData{}
	ctx := r.Context()

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		return res, err
	}

	// one part is called data, the other one is called file
	jsonData := r.FormValue("data")
	var input data.UpdateAttachmentData

	err = json.Unmarshal([]byte(jsonData), &input)
	if err != nil {
		return res, err
	}

	srv := servr.GetServr(h.db, h.config)
	session, err := getSession(ctx, r, srv)
	if err != nil {
		return res, err
	}

	persons := session.Persons

	// try to see if there is file
	fileExistent := true
	fileName := ""

	file, header, err := r.FormFile("file")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			fileExistent = false
		} else {
			return res, err
		}
	}

	if file != nil {
		defer file.Close()
	}

	if fileExistent {
		fileName = header.Filename
	}

	attachment := input.Attachment
	adding := input.Adding

	// check first the attachment does not exist
	attachmentId := attachment.AttachmentId
	exists, err := srv.AttachmentExists(ctx, attachmentId)
	if err != nil {
		return res, err
	}

	if adding && exists {
		return res, errors.New("attachment existent, cannot add on the top of it")
	}

	attachment.FileName = fileName

	// process the file first
	uploadedFileName := srv.GetUploadedFileName(attachmentId)

	if adding {
		if !fileExistent {
			return attachment, errors.New("you cannot add an attachment without file attached")
		}
	}

	if fileExistent {
		// if file exists, error
		_ = os.Remove(uploadedFileName)

		// create the file
		uploaded, err := os.Create(uploadedFileName)
		if err != nil {
			return attachment, err
		}
		defer uploaded.Close()

		_, err = io.Copy(uploaded, file)
		if err != nil {
			return attachment, err
		}
	}

	err = srv.UpdateAttachment(ctx, persons, input.Adding, attachment)
	if err != nil {
		return res, err
	}

	att, err := srv.GetAttachmentById(ctx, attachmentId)

	return att, err
}
