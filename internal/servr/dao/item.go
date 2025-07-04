package dao

import (
	"context"
	"database/sql"
	"gitlab.com/dgb9/todo-api/internal/data"
	"strings"
	"time"
)

func LoadCompleteItemsByIds(ctx context.Context, tx *sql.Tx, ids []string) ([]data.CompleteItemData, error) {
	var res []*data.CompleteItemData

	strSql := `select i.item_id,
			   i.person_id,
			   i.item_name,
			   i.description,
			   i.category,
			   i.flag_ind,
			   i.added_dt,
			   i.updated_dt,
			   a.attachment_id,
			   a.item_id,
			   a.description,
			   a.file_name,
			   a.content_type,
			   a.seq_no,
			   a.added_dt,
			   a.updated_dt
		from item i
				 left outer join attachment a on i.item_id = a.item_id
		where i.item_id in (_here_) order by i.item_name, a.seq_no`

	if len(ids) == 0 {
		return nil, nil
	}

	// now, there are items
	query := "?" + strings.Repeat(", ?", len(ids)-1)
	strSql = strings.Replace(strSql, "_here_", query, 1)

	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return nil, err
	}

	var params []any
	for _, id := range ids {
		params = append(params, id)
	}

	rs, err := stat.QueryContext(ctx, params...)
	if err != nil {
		return nil, err
	}

	defer rs.Close()

	itemMap := make(map[string]*data.CompleteItemData)

	var iItemId sql.NullString
	var iPersonId sql.NullString
	var iItemName sql.NullString
	var iDescription sql.NullString
	var iCategory sql.NullString
	var iFlagInd sql.NullString
	var iAddedDt sql.NullTime
	var iUpdatedDt sql.NullTime

	var aAttachmentId sql.NullString
	var aItemId sql.NullString
	var aDescription sql.NullString
	var aFileName sql.NullString
	var aContentType sql.NullString
	var aSeqNo sql.NullInt32
	var aAddedDt sql.NullTime
	var aUpdatedDt sql.NullTime

	for rs.Next() {
		err = rs.Scan(&iItemId, &iPersonId, &iItemName, &iDescription, &iCategory, &iFlagInd,
			&iAddedDt, &iUpdatedDt, &aAttachmentId, &aItemId, &aDescription,
			&aFileName, &aContentType, &aSeqNo, &aAddedDt, &aUpdatedDt)

		if err != nil {
			return nil, err
		}

		var item *data.CompleteItemData
		item, ok := itemMap[iItemId.String]

		if !ok {
			item = &data.CompleteItemData{}
			item.ItemId = iItemId.String
			item.PersonId = iPersonId.String
			item.Name = iItemName.String

			if iDescription.Valid {
				item.Description = iDescription.String
			}

			if iCategory.Valid {
				item.Category = iCategory.String
			}

			item.Flagged = convertString(iFlagInd.String)

			if iAddedDt.Valid {
				item.Added = iAddedDt.Time
			}

			if iUpdatedDt.Valid {
				item.Updated = iUpdatedDt.Time
			}

			itemMap[iItemId.String] = item

			item.Attachments = make([]data.AttachmentData, 0)

			// don't forget to add it to the slice
			res = append(res, item)
		}

		if aAttachmentId.Valid {
			// load attachment and add it as well
			att := data.AttachmentData{}
			att.AttachmentId = aAttachmentId.String
			att.ItemId = aItemId.String

			if aDescription.Valid {
				att.Description = aDescription.String
			}

			if aFileName.Valid {
				att.FileName = aFileName.String
			}

			if aContentType.Valid {
				att.ContentType = &aContentType.String
			}

			if aSeqNo.Valid {
				att.SeqNo = int(aSeqNo.Int32)
			}

			if aAddedDt.Valid {
				att.Added = aAddedDt.Time
			}

			if aUpdatedDt.Valid {
				att.Updated = aUpdatedDt.Time
			}

			// add it
			item.Attachments = append(item.Attachments, att)
		}
	}

	var fitems []data.CompleteItemData
	for _, addr := range res {
		fitems = append(fitems, *addr)
	}

	return fitems, nil
}

func GetItemIdsByAttachmentIds(ctx context.Context, tx *sql.Tx, ids []string) ([]string, error) {
	var res []string

	if len(ids) == 0 {
		return res, nil
	}

	strSql := "select distinct item_id from attachment where attachment_id in (_here_)"
	query := "?" + strings.Repeat(", ?", len(ids)-1)
	strSql = strings.Replace(strSql, "_here_", query, 1)

	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return nil, err
	}

	var id string
	var qr []any
	for _, id := range ids {
		qr = append(qr, id)
	}
	rs, err := stat.QueryContext(ctx, qr...)
	if err != nil {
		return nil, err
	}

	defer rs.Close()

	for rs.Next() {
		err = rs.Scan(&id)
		if err != nil {
			return nil, err
		}

		res = append(res, id)
	}

	return res, nil
}

func GetPersonsByItemIds(ctx context.Context, tx *sql.Tx, ids []string) ([]string, error) {
	var res []string
	if len(ids) == 0 {
		return res, nil
	}

	strSql := "select distinct person_id from item where item_id in (_here_)"
	query := "?" + strings.Repeat(", ?", len(ids)-1)
	strSql = strings.Replace(strSql, "_here_", query, 1)

	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return res, err
	}

	var qr []any
	for _, id := range ids {
		qr = append(qr, id)
	}

	rs, err := stat.QueryContext(ctx, qr...)
	if err != nil {
		return res, nil
	}

	defer rs.Close()
	var id string

	for rs.Next() {
		err = rs.Scan(&id)

		if err != nil {
			return nil, err
		}

		res = append(res, id)
	}

	return res, nil
}

func DeleteAttachmentRecord(ctx context.Context, tx *sql.Tx, id string) error {
	strSql := "delete from attachment where attachment_id = ?"

	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return err
	}

	_, err = stat.ExecContext(ctx, id)
	return err
}

func HasAttachmentsByItemIds(ctx context.Context, tx *sql.Tx, ids []string) (bool, error) {
	if len(ids) == 0 {
		return false, nil
	}

	parts := append([]string{}, "select attachment_id from attachment where item_id in (?", strings.Repeat(", ?", len(ids)-1), ")")
	strSql := strings.Join(parts, " ")
	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return false, err
	}

	params := getStringParams(ids)

	rs, err := stat.QueryContext(ctx, params...)
	if err != nil {
		return false, err
	}

	defer rs.Close()

	return rs.Next(), nil
}

func RemoveItems(ctx context.Context, tx *sql.Tx, ids []string) error {
	idsCount := len(ids)

	if idsCount == 0 {
		return nil
	}

	parts := []string{"delete from item where item_id in (?",
		strings.Repeat(", ?", idsCount-1),
		")"}

	strSql := strings.Join(parts, " ")

	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return err
	}

	params := getStringParams(ids)

	_, err = stat.ExecContext(ctx, params...)

	return err
}

func UpdateItem(ctx context.Context, tx *sql.Tx, item data.ItemData) error {
	tm := time.Now()
	strSql := "update item set person_id = ?, item_name = ?, description = ?, category = ?, flag_ind = ?, updated_dt = ? where item_id = ?"

	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return err
	}

	_, err = stat.ExecContext(ctx, item.PersonId, item.Name, item.Description, item.Category, convertBool(item.Flagged), tm, item.ItemId)

	return err
}

func AddItem(ctx context.Context, tx *sql.Tx, item data.ItemData) error {
	tm := time.Now()

	strSql := "insert into item (item_id, person_id, item_name, description, category, flag_ind, added_dt, updated_dt) values (?, ?, ?, ?, ?, ?, ?, ?)"

	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return err
	}

	_, err = stat.ExecContext(ctx, item.ItemId, item.PersonId, item.Name, item.Description, item.Category, convertBool(item.Flagged), tm, tm)

	return err
}

func GetFlaggedItemIds(ctx context.Context, tx *sql.Tx, persons []string) ([]string, error) {
	if len(persons) == 0 {
		return nil, nil
	}

	parts := append([]string{}, "select item_id from item where flag_ind = 'Y' and person_id in (?", strings.Repeat(", ?", len(persons)-1), ")")
	strSql := strings.Join(parts, "")
	params := getStringParams(persons)

	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return nil, err
	}

	var id string
	rs, err := stat.QueryContext(ctx, params...)
	if err != nil {
		return nil, err
	}

	defer rs.Close()

	var res []string
	for rs.Next() {
		err = rs.Scan(&id)
		if err != nil {
			return nil, err
		}

		res = append(res, id)
	}

	return res, nil
}

func SearchItemIds(ctx context.Context, tx *sql.Tx, persons []string, search string) ([]string, error) {
	var params []any
	var parts []string

	nrPersons := len(persons)
	if nrPersons == 0 {
		// not possible to work without persons
		return nil, nil
	}

	anyPersons := getStringParams(persons)
	params = append(params, anyPersons...)

	parts = append(parts, "select distinct i.item_id from item i left outer join attachment a on i.item_id = a.item_id where i.person_id in (?")
	parts = append(parts, strings.Repeat(", ?", nrPersons-1))
	parts = append(parts, ")")

	// now process the amount of strings
	sr := strings.TrimSpace(search)
	words := strings.Fields(sr)

	var processed []any
	for _, word := range words {
		processed = append(processed, prepSearch(word))
	}

	nrWords := len(processed)

	if nrWords > 0 {
		var qr []string
		for i := 0; i < nrWords; i++ {
			qr = append(qr, "field like ?")
		}

		strSearch := "(" + strings.Join(qr, " and ") + ")"
		strName := strings.ReplaceAll(strSearch, "field", "i.item_name")
		strDescription := strings.ReplaceAll(strSearch, "field", "i.description")
		strAttach := strings.ReplaceAll(strSearch, "field", "a.description")
		strFile := strings.ReplaceAll(strSearch, "field", "a.file_name")

		searches := []string{strName, strDescription, strAttach, strFile}

		quer := strings.Join(searches, " or ")
		parts = append(parts, " and (", quer, ")")

		// add parameters - we have four sets of searches, we cound to four
		for j := 0; j < 4; j++ {
			params = append(params, processed...)
		}
	}

	strSql := strings.Join(parts, " ")
	stat, err := tx.PrepareContext(ctx, strSql)
	if err != nil {
		return nil, err
	}

	rs, err := stat.QueryContext(ctx, params...)
	if err != nil {
		return nil, err
	}

	defer rs.Close()

	var res []string
	var id string
	for rs.Next() {
		err = rs.Scan(&id)

		if err != nil {
			return nil, err
		}

		res = append(res, id)
	}

	if res == nil {
		res = make([]string, 0)
	}
	return res, nil
}
