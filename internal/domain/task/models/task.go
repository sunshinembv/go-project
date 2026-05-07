package models

type Task struct {
	TID         string
	UID         string
	Title       string
	Description string
	Status      Status
}

type TaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      Status `json:"status"`
}
