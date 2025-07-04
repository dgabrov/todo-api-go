package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"gitlab.com/dgb9/todo-api/internal/cst"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr"
	"io"
	"net/http"
	"os"
	"time"
)

type addBulkAttachment struct {
	db     *sql.DB
	config data.ServerConfig
}

func addBulkAttachmentController(db *sql.DB, config data.ServerConfig) http.Handler {
	return &addBulkAttachment{db, config}
}

func (h *addBulkAttachment) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	attachments, err := h.processAddBulkAttachment(r)

	writeResponse(w, attachments, err)
}

func (h *addBulkAttachment) processAddBulkAttachment(r *http.Request) ([]data.AttachmentData, error) {
	ctx := r.Context()

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		return nil, err // cannot parse the multipart form
	}

	// get the formdata and proceed with it
	dt := r.FormValue("data")
	var input data.BulkAttachmentData

	err = json.Unmarshal([]byte(dt), &input)
	if err != nil {
		return nil, err
	}

	itemId := input.ItemId
	name := input.Name

	srv := servr.GetServr(h.db, h.config)
	session, err := getSession(ctx, r, srv)
	if err != nil {
		return nil, err
	}

	persons := session.Persons

	// now it works, so will process data
	fileHeaders := r.MultipartForm.File["file"]
	if len(fileHeaders) == 0 {
		return nil, errors.New("really, no files, please select some file to add")
	}

	attachments := make([]data.AttachmentData, 0)
	for _, fileHeader := range fileHeaders {

		newSeq, err := srv.GetMaxAttachmentSeq(ctx, itemId)
		if err != nil {
			return attachments, err
		}

		newSeq += cst.SequenceIncrement

		fileName := fileHeader.Filename
		tm := time.Now()

		// build the file attachment
		newId := uuid.New().String()
		contentType := ""
		att := data.AttachmentData{
			AttachmentId: newId,
			ItemId:       itemId,
			Description:  name,
			FileName:     fileName,
			ContentType:  &contentType,
			SeqNo:        newSeq,
			Added:        tm,
			Updated:      tm,
		}

		err = srv.UpdateAttachment(ctx, persons, true, att)
		if err != nil {
			return attachments, err
		}

		loadedAtt, err := srv.GetAttachmentById(ctx, att.AttachmentId)
		if err != nil {
			return attachments, err
		}

		// process file
		file, err := fileHeader.Open()
		if err != nil {
			return attachments, err
		}

		defer file.Close()

		attachmentId := loadedAtt.AttachmentId
		uploadedFileName := srv.GetUploadedFileName(attachmentId)

		output, err := os.Create(uploadedFileName)
		if err != nil {
			return attachments, err
		}
		defer output.Close()

		_, err = io.Copy(output, file)
		if err != nil {
			return attachments, err
		}

		// close them for good measure, even if you put defer, because they are used in a loop
		_ = file.Close()
		_ = output.Close()

		// add to result
		attachments = append(attachments, loadedAtt)
	}

	return attachments, nil
}
