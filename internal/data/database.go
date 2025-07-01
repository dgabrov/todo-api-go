package data

import "time"

type Person struct {
	PersonId   string `json:"personId"`
	ProvidedId string `json:"providedId"`
	Name       string `json:"name"`
	Login      string `json:"login"`
}

type Session struct {
	SessionId string    `json:"sessionId"`
	Updated   time.Time `json:"updated"`
	Token     string    `json:"token"`
	Expired   bool      `json:"expired"`
	Persons   []string  `json:"persons"`
}

type TodoItem struct {
	TodoItemId string     `json:"todoItemId"`
	PersonId   string     `json:"personId"`
	Comments   string     `json:"comments"`
	ProjectCd  *string    `json:"projectCd"`
	ContextCd  *string    `json:"contextCd"`
	Priority   int        `json:"priority"`
	Added      time.Time  `json:"added"`
	Due        *time.Time `json:"due"`
	Completed  bool       `json:"completed"`
	Updated    time.Time  `json:"updated"`
}

type Item struct {
	ItemId      string    `json:"itemId"`
	PersonId    string    `json:"personId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Flagged     bool      `json:"flagged"`
	Added       time.Time `json:"added"`
	Updated     time.Time `json:"updated"`
}
