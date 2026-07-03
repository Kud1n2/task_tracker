package domain

type TeamRequest struct {
	Name string `json:"name"`
}

type TeamResponse struct {
	Team_name string `json:"team_name"`
	Owner     int64  `json:"owner_id"`
}

type TeamInviteRequest struct {
	User_id int64 `json:"user_id"`
}

type TeamInfo struct {
	Team_id          int64  `json:"team_id"`
	Title            string `json:"title"`
	Team_members     int    `json:"number_of_members"`
	Done_tasks_count int    `json:"done_tasks_count"`
}

type TopUsers struct {
	Team_name  string `json:"team_name"`
	User_name  string `json:"user_name"`
	Task_count int    `json:"task_count"`
}

type NotInTeam struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	TeamID int64  `json:"team_id"`
	UserID int64  `json:"assignee_id"`
}
