package servr

import (
	"context"
	"errors"
	"fmt"
	"gitlab.com/dgb9/todo-api/internal/data"
	"gitlab.com/dgb9/todo-api/internal/servr/dao"
	"log/slog"
	"os"
	"slices"
)

func (h *server) SwitchSeqNo(ctx context.Context, persons []string, ids []string) (data.CompleteItemData, error) {
	res := data.CompleteItemData{}
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return res, err
	}

	defer rollback(tx)

	// must be two attachments
	if len(ids) != 2 {
		return res, errors.New("there must be exactly two attachment ids")
	}

	// get ids
	itemIds, err := dao.GetItemIdsByAttachmentIds(ctx, tx, ids)
	if err != nil {
		return res, err
	}

	if len(itemIds) == 0 {
		return res, errors.New("cannot find the item associated with these two attachments")
	}

	if len(itemIds) > 1 {
		return res, errors.New("the attachments must belong to the same item, not different items")
	}

	// ok, now let's see to whom this item belongs
	owners, err := dao.GetPersonsByItemIds(ctx, tx, itemIds)
	if err != nil {
		return res, err
	}

	for _, owner := range owners {
		if !slices.Contains(persons, owner) {
			return res, errors.New("at least one item does not belong to any of the logged in user")
		}
	}

	// so now it is an item, and let's switch
	err = dao.SwitchAttachmentSeq(ctx, tx, ids)
	if err != nil {
		return res, err
	}

	// now load the iem
	items, err := dao.LoadCompleteItemsByIds(ctx, tx, itemIds)
	if err != nil {
		return res, err
	}

	err = tx.Commit()
	return items[0], err
}

func (h *server) RemoveAttachment(ctx context.Context, persons []string, ids []string) error {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer rollback(tx)

	itemIds, err := dao.GetItemIdsByAttachmentIds(ctx, tx, ids)
	if err != nil {
		return err
	}

	personIds, err := dao.GetPersonsByItemIds(ctx, tx, itemIds)
	if err != nil {
		return err
	}

	for _, person := range personIds {
		if !slices.Contains(persons, person) {
			return errors.New("at least one of the attachments are not owned by the logged in users")
		}
	}

	// this time we do it in loop; we understand that this will take longer, however
	// it will stop short if any issue shows up avoidind the situation that you delete all the records
	// and you cannot delete the associated files
	for _, id := range ids {
		slog.Info(fmt.Sprintf("deleting attachment: %s", id))

		// delete the file
		fileName := h.GetUploadedFileName(id)
		err = os.Remove(fileName)

		if err != nil {
			return err
		}

		// delete record
		err = dao.DeleteAttachmentRecord(ctx, tx, id)
		if err != nil {
			return err
		}

	}

	return tx.Commit()
}

func (h *server) RemoveItems(ctx context.Context, persons []string, ids []string) error {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)

	// get person ids and see if they are logged in
	ownerIds, err := dao.GetPersonsByItemIds(ctx, tx, ids)
	if err != nil {
		return err
	}

	for _, ownerId := range ownerIds {
		if !slices.Contains(persons, ownerId) {
			return errors.New("at least one of the items does not belong to the logged in users")
		}
	}

	// now let's see if there are attachments
	hasAttachments, err := dao.HasAttachmentsByItemIds(ctx, tx, ids)
	if err != nil {
		return err
	}

	if hasAttachments {
		return errors.New("at least one of the items you try to delete has attachments; delete the attachments before you try to delete the item")
	}

	// now let's remove
	err = dao.RemoveItems(ctx, tx, ids)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (h *server) UpdateItem(ctx context.Context, persons []string, input data.UpdateItemData) (data.ItemData, error) {
	res := data.ItemData{}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return res, err
	}

	defer rollback(tx)

	// get the person associated with the item
	owners, err := dao.GetPersonsByItemIds(ctx, tx, []string{input.Item.ItemId})

	if len(owners) == 0 && !input.Adding {
		// the item being edited does not exist, it is error
		return res, errors.New("the item you want to update does not exist")
	}

	if len(owners) > 0 && input.Adding {
		return res, errors.New("the item with this id already exists, you cannot add it again")
	}

	for _, owner := range owners {
		if !slices.Contains(persons, owner) {
			return res, errors.New("item with the provided id is not owned by any of the logged in users")
		}
	}

	// now the person id must be within the logged in users
	personId := input.Item.PersonId
	if !slices.Contains(persons, personId) {
		return res, errors.New("the person you want to assign this item to is not among logged in users")
	}

	err = checkItem(input.Item)
	if err != nil {
		return res, err
	}

	// OK now let's see
	if input.Adding {
		err = dao.AddItem(ctx, tx, input.Item)

		if err != nil {
			return res, err
		}
	} else {
		err = dao.UpdateItem(ctx, tx, input.Item)

		if err != nil {
			return res, err
		}
	}

	// get the item and return it
	items, err := dao.LoadCompleteItemsByIds(ctx, tx, []string{input.Item.ItemId})
	if err != nil {
		return res, err
	}

	if len(items) != 1 {
		return res, errors.New("something wrong happened when trying to load the item")
	}

	res = convertToItem(items[0])

	err = tx.Commit()
	return res, err
}

func convertToItem(in data.CompleteItemData) data.ItemData {
	return data.ItemData{
		ItemId:      in.ItemId,
		PersonId:    in.PersonId,
		Name:        in.Name,
		Description: in.Description,
		Category:    in.Category,
		Flagged:     in.Flagged,
		Added:       in.Added,
		Updated:     in.Updated,
	}
}

func (h *server) GetFlaggedItems(ctx context.Context, persons []string) ([]data.CompleteItemData, error) {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	ids, err := dao.GetFlaggedItemIds(ctx, tx, persons)
	if err != nil {
		return nil, err
	}

	res, err := dao.LoadCompleteItemsByIds(ctx, tx, ids)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (h *server) SearchItems(ctx context.Context, persons []string, search string) ([]data.CompleteItemData, error) {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer rollback(tx)

	ids, err := dao.SearchItemIds(ctx, tx, persons, search)
	if err != nil {
		return nil, err
	}

	res, err := dao.LoadCompleteItemsByIds(ctx, tx, ids)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	if res == nil {
		res = make([]data.CompleteItemData, 0)
	}
	return res, nil
}
