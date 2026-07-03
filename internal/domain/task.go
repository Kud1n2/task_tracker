package domain

import "time"

type TaskRequest struct {
	Creator int64  `json:"creator"`
	Team_id int64  `json:"team_id"`
	Title   string `json:"title"`
}

type TaskFilter struct {
	Assignee *int64
	Status   *string
	Team_id  *int64

	Limit int
	Page  int
}

type TaskResponse struct {
	Id       int    `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Team_id  int64  `json:"team_id"`
	Assignee int64  `json:"assignee"`
}

type TaskUpdate struct {
	Id       int64  `json:"id"`
	Status   string `json:"status"`
	Assignee int64  `json:"assignee"`
}
type TaskUpdateRequest struct {
	Status string `json:"status"`
}
type TaskHistory struct {
	Updated_by int64     `json:"updated_by"`
	Time       time.Time `json:"update_time"`
}
