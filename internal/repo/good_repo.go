package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-test/internal/model"
	"strings"
)

type GoodRepository interface {
	Create(ctx context.Context, g *model.Good) error
	GetByID(ctx context.Context, id int) (*model.Good, error)
	Update(ctx context.Context, g *model.Good) error
	Delete(ctx context.Context, id int, projectID int) (*model.Good, error)
	List(ctx context.Context, projectID, limit, offset int, sort string) ([]model.Good, int, int, error)
	GetMaxPriority(ctx context.Context, projectID int) (int, error)
	Reprioritize(ctx context.Context, id, projectID, newPriority int) ([]model.Good, error)
}

type goodRepo struct {
	db *sql.DB
}

func NewGoodRepo(db *sql.DB) *goodRepo {
	return &goodRepo{db: db}
}

func (r *goodRepo) Create(ctx context.Context, g *model.Good) error {
	err := r.db.QueryRowContext(ctx,
		`
		INSERT INTO goods (project_id, name, description, priority, removed, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
 	`, g.ProjectID, g.Name, g.Description, g.Priority, g.Removed, g.CreatedAt).Scan(&g.ID)
	if err != nil {
		return fmt.Errorf("failed to insert good: %w", err)
	}

	return nil
}

func (r *goodRepo) GetByID(ctx context.Context, id int) (*model.Good, error) {
	var g model.Good

	err := r.db.QueryRowContext(ctx,
		`
	SELECT id, project_id, name, description, priority, removed, created_at
	FROM goods 
	WHERE id = $1 AND removed = false
	`, id).Scan(&g.ID, &g.ProjectID, &g.Name, &g.Description, &g.Priority, &g.Removed, &g.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("good not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get good by id: %w", err)
	}

	return &g, nil
}

func (r *goodRepo) Update(ctx context.Context, g *model.Good) error {
	var exists int

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = tx.QueryRowContext(ctx,
		`
	SELECT 1 
	FROM goods
	WHERE id = $1 AND removed = false
	FOR UPDATE
	`, g.ID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			tx.Rollback()
			return fmt.Errorf("good not found: %w", err)
		}
		tx.Rollback()
		return fmt.Errorf("failed to update good: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`
		UPDATE goods
		SET project_id = $2, name = $3, description = $4, priority = $5, removed = $6
		WHERE id = $1
		`, g.ID, g.ProjectID, g.Name, g.Description, g.Priority, g.Removed)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update good: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *goodRepo) Delete(ctx context.Context, id int, projectID int) (*model.Good, error) {
	res, err := r.db.ExecContext(ctx, `
	UPDATE goods
	SET removed = true
	WHERE id = $1 AND project_id = $2
	`, id, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete good: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to delete good: %w", err)
	}
	if affected == 0 {
		return nil, fmt.Errorf("good not found")
	}

	var g model.Good
	err = r.db.QueryRowContext(ctx, `
	SELECT id, project_id, name, description, priority, removed, created_at
	FROM goods
	WHERE id = $1 AND project_id = $2
	`, id, projectID).Scan(&g.ID, &g.ProjectID, &g.Name, &g.Description, &g.Priority, &g.Removed, &g.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deleted good: %w", err)
	}

	return &g, nil
}

func (r *goodRepo) List(ctx context.Context, projectID, limit, offset int, sort string) ([]model.Good, int, int, error) {
	var goods []model.Good

	var totalCount, removedCount int

	order := "ASC"
	if strings.ToLower(sort) == "desc" {
		order = "DESC"
	}

	query := fmt.Sprintf(`
	SELECT id, project_id, name, description, priority, removed, created_at
	FROM goods
	WHERE project_id = $1 AND removed = false
	ORDER BY created_at %s
	LIMIT $2 OFFSET $3
	`, order)

	err := r.db.QueryRowContext(ctx, `
	SELECT COUNT(*)
	FROM goods
	WHERE project_id = $1
	`, projectID).Scan(&totalCount)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to count total goods: %w", err)
	}

	err = r.db.QueryRowContext(ctx, `
	SELECT COUNT(*)
	FROM goods
	WHERE project_id = $1 AND removed = true
	`, projectID).Scan(&removedCount)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to count removed goods: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, query, projectID, limit, offset)
	if err != nil {
		return nil, totalCount, removedCount, fmt.Errorf("failed to list goods: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var g model.Good

		err := rows.Scan(&g.ID, &g.ProjectID, &g.Name, &g.Description, &g.Priority, &g.Removed, &g.CreatedAt)
		if err != nil {
			return nil, totalCount, removedCount, fmt.Errorf("failed to scan good: %w", err)
		}
		goods = append(goods, g)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return goods, totalCount, removedCount, nil
}

func (r *goodRepo) GetMaxPriority(ctx context.Context, projectID int) (int, error) {
	var maxPriority sql.NullInt64

	err := r.db.QueryRowContext(ctx, `
	SELECT MAX(priority) FROM goods
	WHERE project_id = $1`, projectID).Scan(&maxPriority)
	if err != nil {
		return 0, fmt.Errorf("failed to get max priority: %w", err)
	}
	if !maxPriority.Valid {
		return 0, nil
	}

	return int(maxPriority.Int64), nil
}

func (r *goodRepo) Reprioritize(ctx context.Context, id, projectID, newPriority int) ([]model.Good, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	var currentPriority int
	err = tx.QueryRowContext(ctx, `
	SELECT priority
	FROM goods
	WHERE id = $1 AND project_id = $2 AND removed = false
	FOR UPDATE
	`, id, projectID).Scan(&currentPriority)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("good not found: %w", err)
		}
		return nil, fmt.Errorf("failed to fetch current priority: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
	UPDATE goods
	SET priority = priority + 1
	WHERE project_id = $1 AND removed = false AND id != $2 AND priority >= $3
	`, projectID, id, newPriority)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to shift priorities: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
	UPDATE goods
	SET priority = $1
	WHERE id = $2
	`, newPriority, id)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update good priority: %w", err)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT id, project_id, name, description, priority, removed, created_at
		FROM goods
		WHERE project_id = $1 AND removed = false AND priority >= $2
		ORDER BY priority
	`, projectID, newPriority)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to fetch updated goods: %w", err)
	}
	defer rows.Close()

	var goods []model.Good
	for rows.Next() {
		var g model.Good
		if err := rows.Scan(&g.ID, &g.ProjectID, &g.Name, &g.Description, &g.Priority, &g.Removed, &g.CreatedAt); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to scan good: %w", err)
		}
		goods = append(goods, g)
	}
	if err := rows.Err(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return goods, nil
}
