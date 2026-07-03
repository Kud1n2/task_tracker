package mysql

import (
	"context"
	"errors"
	"fmt"
	"task_tracker/internal/domain"
)

func (s *Storage) CreateTeam(ctx context.Context, name string, owner int64) error {
	const fn = "storage.team.CreateTeam"

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%s:%w", fn, err)
	}
	defer tx.Rollback()

	res, err := s.db.Exec(`INSERT INTO teams(name, created_by) VALUES (?, ?)`, name, owner)
	if err != nil {
		return fmt.Errorf("%s:%w", fn, err)
	}

	teamID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("%s:%w", fn, err)
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO team_members(team_id, user_id, role) VALUES(?,?,'owner')`, teamID, owner)
	if err != nil {
		return fmt.Errorf("%s:%w", fn, err)
	}

	return tx.Commit()
}

func (s *Storage) GetTeams(ctx context.Context, userID int64) ([]domain.TeamResponse, error) {
	const fn = "storage.teams.GetTeams"

	rows, err := s.db.QueryContext(ctx, `SELECT t.name, t.created_by FROM teams t JOIN team_members tm ON t.id = tm.team_id WHERE tm.user_id = ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	var teams []domain.TeamResponse

	for rows.Next() {
		var team domain.TeamResponse
		rows.Scan(&team.Team_name, &team.Owner)

		teams = append(teams, team)
	}

	return teams, nil
}

func (s *Storage) AddUser(ctx context.Context, team_id, userID, owner int64) error {
	const fn = "storage.teams.AddUser"

	var role string

	_ = s.db.QueryRowContext(ctx, `SELECT role FROM team_members WHERE team_id = ? AND user_id = ?`, team_id, owner).Scan(&role)

	if role != "owner" {
		return fmt.Errorf("%s:%w", fn, errors.New("Only owner can invite"))
	}

	_, err := s.db.ExecContext(ctx, `INSERT INTO team_members(user_id, team_id, role) VALUES(?,?,'user')`, userID, team_id)
	if err != nil {
		return fmt.Errorf("%s:%w", fn, err)
	}

	return nil
}

func (s *Storage) GetTeamsInfo(ctx context.Context) ([]domain.TeamInfo, error) {
	const fn = "storage.team.GetTeamsInfo"

	rows, err := s.db.QueryContext(ctx, `
	SELECT t.id, t.name,
	COUNT(DISTINCT tm.user_id) AS members_count, 
	COUNT(DISTINCT ta.id) AS tasks_count 
	FROM teams t 
	JOIN team_members tm 
	ON tm.team_id = t.id
	JOIN tasks ta 
	ON t.id = ta.team_id 
	AND ta.status = 'done' 
	AND ta.created_at >= NOW() - INTERVAL 7 DAY 
	GROUP BY t.id, t.name 
	ORDER BY t.name;`)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	var teams_info []domain.TeamInfo

	for rows.Next() {
		var info domain.TeamInfo

		rows.Scan(&info.Team_id, &info.Title, &info.Team_members, &info.Done_tasks_count)
		teams_info = append(teams_info, info)
	}

	return teams_info, nil
}

func (s *Storage) TopUsers(ctx context.Context) ([]domain.TopUsers, error) {
	const fn = "storage.team.TopUsers"

	rows, err := s.db.QueryContext(ctx, `
	SELECT team_name, user_name, tasks_count FROM (
		SELECT t.id AS team_id, t.name AS team_name, u.id AS user_id, u.name AS user_name, COUNT(task.id) AS tasks_count,
		ROW_NUMBER() OVER (
			PARTITION BY t.id
			ORDER BY COUNT(task.id) DESC
		) AS rn
		FROM teams t 
		JOIN tasks task ON t.id = task.team_id 
		JOIN users u ON task.created_by = u.id
		WHERE task.created_at >= NOW() - INTERVAL 1 MONTH
		GROUP BY t.id, t.name, u.id, u.name
	) ranked
	WHERE rn <= 3 ORDER BY team_name, tasks_count DESC;
	`)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	var users []domain.TopUsers

	for rows.Next() {
		var user domain.TopUsers

		rows.Scan(&user.Team_name, &user.User_name, &user.Task_count)
		users = append(users, user)
	}

	return users, nil
}

func (s *Storage) NotInTeam(ctx context.Context) ([]domain.NotInTeam, error) {
	const fn = "storage.team.NotInTeam"

	rows, err := s.db.QueryContext(ctx, `
	SELECT
		t.id,
		t.title,
		t.team_id,
		t.assignee_id
	FROM tasks t
	LEFT JOIN team_members tm
		ON tm.team_id = t.team_id
	AND tm.user_id = t.assignee_id
	WHERE tm.user_id IS NULL 
	AND t.assignee_id <> 0;
	`)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	var users []domain.NotInTeam

	for rows.Next() {
		var user domain.NotInTeam

		rows.Scan(&user.ID, &user.Title, &user.TeamID, &user.UserID)
		users = append(users, user)
	}

	return users, nil
}
