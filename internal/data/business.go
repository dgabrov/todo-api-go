package data

import "time"

type UpdatePriorityData struct {
	TodoItemId string `json:"todoItemId"`
	Priority   int    `json:"priority"`
}

type TodoItemIdData struct {
	TodoItemId string `json:"todoItemId"`
}

type CompletedData struct {
	TodoItemId string `json:"todoItemId"`
	Completed  bool   `json:"completed"`
}

type ItemData struct {
	ItemId      string    `json:"itemId"`
	PersonId    string    `json:"personId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Flagged     bool      `json:"flagged"`
	Added       time.Time `json:"added"`
	Updated     time.Time `json:"updated"`
}

type UpdateItemData struct {
	Item   ItemData `json:"item"`
	Adding bool     `json:"adding"`
}

type AttachmentData struct {
	AttachmentId string    `json:"attachmentId"`
	ItemId       string    `json:"itemId"`
	Description  string    `json:"description"`
	FileName     string    `json:"fileName"`
	ContentType  string    `json:"contentType"`
	SeqNo        int       `json:"seqNo"`
	Added        time.Time `json:"added"`
	Updated      time.Time `json:"updated"`
}

type UpdateAttachmentData struct {
	Adding     bool           `json:"adding"`
	Attachment AttachmentData `json:"attachment"`
}

type CompleteItemData struct {
	ItemId      string           `json:"itemId"`
	PersonId    string           `json:"personId"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Category    string           `json:"category"`
	Flagged     bool             `json:"flagged"`
	Added       time.Time        `json:"added"`
	Updated     time.Time        `json:"updated"`
	Attachments []AttachmentData `json:"attachments"`
}

type IdsData struct {
	Ids []string `json:"ids"`
}

type TodoContainerData struct {
	Todo *TodoItem `json:"todo,omitempty"`
}

type UpdateDueData struct {
	TodoItemId string     `json:"todoItemId"`
	Due        *time.Time `json:"due"`
}

type UpdateTodoItem struct {
	Todo   TodoItem `json:"todo"`
	Adding bool     `json:"adding"`
}

type IntervalDue struct {
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
}

type TodoSearch struct {
	Completed   *bool         `json:"completed"`
	Context     []string      `json:"context"`
	Project     []string      `json:"project"`
	Login       []string      `json:"login"`
	General     []string      `json:"general"`
	DueNull     bool          `json:"dueNull"`
	DueInterval []IntervalDue `json:"dueInterval"`
}

type BulkUpdateData struct {
	TodoIds          []string   `json:"todoIds"`
	SelectedOwner    bool       `json:"selectedOwner"`
	SelectedContext  bool       `json:"selectedContext"`
	SelectedProject  bool       `json:"selectedProject"`
	SelectedDue      bool       `json:"selectedDue"`
	SelectedPriority bool       `json:"selectedPriority"`
	OwnerId          string     `json:"ownerId"`
	Context          string     `json:"context"`
	Project          string     `json:"project"`
	Due              *time.Time `json:"due"`
	Priority         int        `json:"priority"`
}

type SearchData struct {
	Search string `json:"search"`
}

type BulkAttachmentData struct {
	ItemId string `json:"itemId"`
	Name   string `json:"name"`
}
