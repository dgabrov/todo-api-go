package data

import "time"

type MultipleTodoData struct {
	PersonId  string     `json:"personId"`
	Comments  []string   `json:"comments"`
	ProjectCd string     `json:"projectCd"`
	ContextCd string     `json:"contextCd"`
	Priority  int        `json:"priority"`
	Due       *time.Time `json:"due"`
}
