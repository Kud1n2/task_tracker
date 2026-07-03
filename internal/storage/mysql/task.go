package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"task_tracker/internal/domain"
)

func (s *Storage) InsertTask(ctx context.Context, task domain.TaskRequest) error {
	const fn = "storage.task.InsertTask"

	var user_id int64
	err := s.db.QueryRowContext(ctx, `SELECT user_id FROM team_members WHERE team_id = ?`, task.Team_id).Scan(&user_id)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("User not in a team")
	}

	res, err := s.db.ExecContext(ctx, `INSERT INTO tasks(created_by, team_id, title, status) VALUES (?,?,?, 'To Do')`, task.Creator, task.Team_id, task.Title)
	if err != nil {
		return fmt.Errorf("%s:%w", fn, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("%s:%w", fn, err)
	}

	_, err = s.db.ExecContext(ctx, `INSERT INTO task_history(task_id, changed_by) VALUES (?,?)`, id, task.Creator)
	if err != nil {
		return fmt.Errorf("%s:%w", fn, err)
	}

	return nil
}

func (s *Storage) SelectTasks(ctx context.Context, filter domain.TaskFilter) ([]domain.TaskResponse, error) {
	const fn = "storage.task.SelectTasks"

	query := `SELECT id, title, status, team_id, assignee_id FROM tasks WHERE 1=1`

	args := []interface{}{}

	if filter.Team_id != nil {
		query += " AND team_id = ?"
		args = append(args, *filter.Team_id)
	}

	if filter.Assignee != nil {
		query += " AND assignee_id = ?"
		args = append(args, *filter.Assignee)
	}
	if filter.Status != nil {
		query += " AND status = ?"
		args = append(args, *filter.Status)
	}

	offset := (filter.Page - 1) * filter.Limit
	query += ` ORDER BY id LIMIT ? OFFSET ?`
	args = append(args, filter.Limit, offset)

	var tasks []domain.TaskResponse

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}
	for rows.Next() {
		var task domain.TaskResponse

		rows.Scan(&task.Id, &task.Title, &task.Status, &task.Team_id, &task.Assignee)
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (s *Storage) TaskUpdate(ctx context.Context, update domain.TaskUpdate) error {
	const fn = "storage.task.TaskUpdate"

	_, err := s.db.ExecContext(ctx, `UPDATE tasks SET assignee_id = ?, status = ? WHERE id = ?`, update.Assignee, update.Status, update.Id)
	if err != nil {
		return fmt.Errorf("%s:%w", fn, err)
	}

	_, err = s.db.ExecContext(ctx, `INSERT INTO task_history(task_id, changed_by) VALUES (?,?)`, update.Id, update.Assignee)
	if err != nil {
		return fmt.Errorf("%s:%w", fn, err)
	}

	return nil
}

func (s *Storage) SelectHistory(ctx context.Context, task_id int64) ([]domain.TaskHistory, error) {
	const fn = "storage.tasks.SelectHistory"

	rows, err := s.db.QueryContext(ctx, `SELECT changed_by, created_at FROM task_history WHERE task_id = ?`, task_id)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	var history []domain.TaskHistory

	for rows.Next() {
		var row domain.TaskHistory

		rows.Scan(&row.Updated_by, &row.Time)
		history = append(history, row)
	}

	return history, nil
}

func (s *Storage) GetTeamTasks(ctx context.Context, teamID int64) ([]domain.TaskResponse, error) {
	const fn = "storage.team.GetTeamTasks"

	rows, err := s.db.QueryContext(ctx, `SELECT id, title, status, assignee_id FROM tasks WHERE team_id = ?`, teamID)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", fn, err)
	}

	var tasks []domain.TaskResponse
	for rows.Next() {
		var task domain.TaskResponse

		rows.Scan(&task.Id, &task.Title, &task.Status, &task.Assignee)
		tasks = append(tasks, task)
	}

	return tasks, nil
}
