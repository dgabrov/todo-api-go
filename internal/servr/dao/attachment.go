package dao

import (
	"context"
	"database/sql"
	"fmt"
	"gitlab.com/dgb9/todo-api/internal/data"
	"strings"
	"time"
)

type attachmentSeq struct {
	attachmentId string
	seqNo        int
}

func SwitchAttachmentSeq(ctx context.Context, tx *sql.Tx, ids []string) error {
	strSql := "select attachment_id, seq_no from attachment where attachment_id in (_here_)"
	query := "?" + strings.Repeat(", ?", len(ids)-1)
	strSql = strings.Replace(strSql, "_here_", query, 1)

	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return err
	}

	var qr []any
	for _, id := range ids {
		qr = append(qr, id)
	}

	rs, err := stat.QueryContext(ctx, qr...)
	if err != nil {
		return err
	}

	defer rs.Close()

	var attachments []attachmentSeq
	for rs.Next() {
		attachment := attachmentSeq{}

		err = rs.Scan(&attachment.attachmentId, &attachment.seqNo)
		if err != nil {
			return err
		}

		attachments = append(attachments, attachment)
	}

	// manually close the cursor
	rs.Close()

	// length must be 2
	if len(attachments) != 2 {
		return fmt.Errorf("two attachments required, found only %d", len(attachments))
	}

	err = setAttachmentSeq(ctx, tx, attachments[0].attachmentId, attachments[1].seqNo)
	if err != nil {
		return err
	}

	err = setAttachmentSeq(ctx, tx, attachments[1].attachmentId, attachments[0].seqNo)
	if err != nil {
		return err
	}

	return nil
}

func setAttachmentSeq(ctx context.Context, tx *sql.Tx, attachmentId string, seqNo int) error {
	strSql := "update attachment set seq_no = ? where attachment_id = ?"
	stat, err := tx.PrepareContext(ctx, strSql)

	if err != nil {
		return err
	}

	_, err = stat.ExecContext(ctx, seqNo, attachmentId)

	return err
}

func AttachmentExists(ctx context.Context, tx *sql.Tx, id string) (bool, error) {
	strSql := "select attachment_id from attachment where attachment_id = ?"
	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return false, err
	}

	rs, err := stat.QueryContext(ctx, id)
	if err != nil {
		return false, err
	}

	defer rs.Close()

	res := rs.Next()

	return res, nil
}

func AddAttachment(ctx context.Context, tx *sql.Tx, a data.AttachmentData, newSeq int) error {
	tm := time.Now()

	strSql := "insert into attachment (attachment_id, item_id, description, file_name, content_type, seq_no, added_dt, updated_dt) values (?, ?, ?, ?, ?, ?, ?, ?)"
	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return err
	}

	_, err = stat.ExecContext(ctx, a.AttachmentId, a.ItemId, a.Description, a.FileName, a.ContentType, newSeq, tm, tm)

	return err
}

func UpdateAttachment(ctx context.Context, tx *sql.Tx, attachment data.AttachmentData) error {
	tm := time.Now()
	var params []any
	var parts []string
	parts = append(parts, "update attachment set item_id = ?, description = ?, content_type = ?, seq_no = ?, updated_dt = ?")
	params = append(params, attachment.ItemId, attachment.Description, attachment.ContentType, attachment.SeqNo, tm)

	fileName := attachment.FileName
	trimmedFileName := strings.TrimSpace(fileName)
	if len(trimmedFileName) > 0 {
		parts = append(parts, ", file_name = ?")
		params = append(params, trimmedFileName)
	}
	parts = append(parts, " where attachment_id = ?")
	params = append(params, attachment.AttachmentId)

	strSql := strings.Join(parts, " ")

	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return err
	}

	_, err = stat.ExecContext(ctx, params...)

	return err
}

func GetMaxAttachmentSeq(ctx context.Context, tx *sql.Tx, id string) (int, error) {
	var val *int

	strSql := "select max(seq_no) from attachment where item_id = ?"
	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return 0, err
	}

	rs, err := stat.QueryContext(ctx, id)
	if err != nil {
		return 0, err
	}

	defer rs.Close()

	if rs.Next() {
		err = rs.Scan(&val)

		if err != nil {
			return 0, err
		}
	}

	res := 1

	if val != nil {
		res = *val
	}

	return res, nil
}

func GetAttachment(ctx context.Context, tx *sql.Tx, id string) (data.AttachmentData, error) {
	res := data.AttachmentData{}

	strSql := "select attachment_id, item_id, description, file_name, content_type, seq_no, added_dt, updated_dt from attachment where attachment_id = ?"

	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return res, err
	}

	rs, err := stat.QueryContext(ctx, id)
	if err != nil {
		return res, err
	}

	defer rs.Close()

	if rs.Next() {
		err = rs.Scan(&res.AttachmentId, &res.ItemId, &res.Description, &res.FileName, &res.ContentType, &res.SeqNo, &res.Added, &res.Updated)

		if err != nil {
			return res, err
		}
	}

	return res, nil
}
