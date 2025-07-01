package servr

import (
	"context"
	"errors"
	"gitlab.com/dgb9/todo-api/internal/cst"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr/dao"
	"slices"
)

func (h *server) GetAttachmentById(ctx context.Context, attachmentId string) (data.AttachmentData, error) {
	res := data.AttachmentData{}
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return res, err
	}

	defer rollback(tx)

	res, err = dao.GetAttachment(ctx, tx, attachmentId)
	if err != nil {
		return res, err
	}

	err = tx.Commit()
	return res, err
}

func (h *server) UpdateAttachment(ctx context.Context, persons []string, adding bool, attachment data.AttachmentData) error {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer rollback(tx)
	owners, err := dao.GetPersonsByItemIds(ctx, tx, []string{attachment.ItemId})
	if err != nil {
		return err
	}

	for _, owner := range owners {
		if !slices.Contains(persons, owner) {
			return errors.New("the item for this attachment does not belong to any of the logged in persons")
		}
	}

	attachExists, err := dao.AttachmentExists(ctx, tx, attachment.AttachmentId)
	if err != nil {
		return err
	}

	if adding {
		if attachExists {
			return errors.New("you cannot add the attachment as it exists already")
		}

		newSeq, err := dao.GetMaxAttachmentSeq(ctx, tx, attachment.ItemId)
		err = dao.AddAttachment(ctx, tx, attachment, newSeq+cst.SequenceIncrement)

		if err != nil {
			return err
		}
	} else {
		if !attachExists {
			return errors.New("attachment cannot be updated as it does not exist")
		}

		err = dao.UpdateAttachment(ctx, tx, attachment)

		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (h *server) AttachmentExists(ctx context.Context, attachmentId string) (bool, error) {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}

	defer rollback(tx)

	res, err := dao.AttachmentExists(ctx, tx, attachmentId)
	if err != nil {
		return false, err
	}

	err = tx.Commit()
	if err != nil {
		return false, err
	}

	return res, err
}
