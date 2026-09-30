package repository

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"time"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
	"github.com/davidlima/openspec-studio/backend/internal/repository/sqlcgen"
)

// ProjectRepository is the sole entry point for project-registry persistence.
// Services call this; nothing outside this package touches sqlc or SQL directly.
type ProjectRepository struct {
	q sqlcgen.Querier
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{q: Queries(db)}
}

// Upsert registers or re-registers a project, updating its last-opened timestamp.
func (r *ProjectRepository) Upsert(ctx context.Context, name, path string, openedAt time.Time) (domain.Project, error) {
	row, err := r.q.UpsertProject(ctx, sqlcgen.UpsertProjectParams{
		Name:         name,
		Path:         path,
		LastOpenedAt: openedAt,
	})
	if err != nil {
		return domain.Project{}, err
	}
	return toDomain(row), nil
}

// List returns all registered projects, most recently opened first, flagging
// any whose path no longer exists on disk as unavailable.
func (r *ProjectRepository) List(ctx context.Context) ([]domain.Project, error) {
	rows, err := r.q.ListProjects(ctx)
	if err != nil {
		return nil, err
	}
	projects := make([]domain.Project, 0, len(rows))
	for _, row := range rows {
		p := toDomain(row)
		if _, err := os.Stat(p.Path); err != nil {
			p.Available = false
		} else {
			p.Available = true
		}
		projects = append(projects, p)
	}
	return projects, nil
}

// GetByPath returns the registered project at path, or (Project{}, false) if none is registered.
func (r *ProjectRepository) GetByPath(ctx context.Context, path string) (domain.Project, bool, error) {
	row, err := r.q.GetProjectByPath(ctx, path)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Project{}, false, nil
		}
		return domain.Project{}, false, err
	}
	return toDomain(row), true, nil
}

// GetByID returns the registered project with the given registry id, or (Project{}, false) if none is registered.
func (r *ProjectRepository) GetByID(ctx context.Context, id int64) (domain.Project, bool, error) {
	row, err := r.q.GetProjectByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Project{}, false, nil
		}
		return domain.Project{}, false, err
	}
	p := toDomain(row)
	if _, err := os.Stat(p.Path); err != nil {
		p.Available = false
	}
	return p, true, nil
}

func toDomain(row sqlcgen.Project) domain.Project {
	return domain.Project{
		ID:           row.ID,
		Name:         row.Name,
		Path:         row.Path,
		LastOpenedAt: row.LastOpenedAt,
		Available:    true,
	}
}
