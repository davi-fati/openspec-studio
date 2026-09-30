package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
	"github.com/davidlima/openspec-studio/backend/internal/repository/sqlcgen"
)

// SpecflowRepository is the sole entry point for Specflow (schedule + items) persistence.
type SpecflowRepository struct {
	q sqlcgen.Querier
}

func NewSpecflowRepository(db *sql.DB) *SpecflowRepository {
	return &SpecflowRepository{q: Queries(db)}
}

func (r *SpecflowRepository) List(ctx context.Context) ([]domain.Flow, error) {
	rows, err := r.q.ListSpecflows(ctx)
	if err != nil {
		return nil, err
	}
	flows := make([]domain.Flow, 0, len(rows))
	for _, row := range rows {
		items, err := r.listItems(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		f := toDomainFlow(row)
		f.Items = items
		flows = append(flows, f)
	}
	return flows, nil
}

func (r *SpecflowRepository) Get(ctx context.Context, id int64) (domain.Flow, bool, error) {
	row, err := r.q.GetSpecflow(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Flow{}, false, nil
		}
		return domain.Flow{}, false, err
	}
	items, err := r.listItems(ctx, id)
	if err != nil {
		return domain.Flow{}, false, err
	}
	f := toDomainFlow(row)
	f.Items = items
	return f, true, nil
}

// DueFlows returns pending flows whose scheduled time has arrived.
// DueFlows returns pending, unmissed flows scheduled at or before now.
// now is normalized to UTC here (not just by callers) because SQLite has no
// native datetime type: modernc.org/sqlite compares timestamps as text, so
// a value carrying a local offset would sort incorrectly against stored
// UTC timestamps.
func (r *SpecflowRepository) DueFlows(ctx context.Context, now time.Time) ([]domain.Flow, error) {
	rows, err := r.q.ListDueSpecflows(ctx, now.UTC())
	if err != nil {
		return nil, err
	}
	flows := make([]domain.Flow, 0, len(rows))
	for _, row := range rows {
		items, err := r.listItems(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		f := toDomainFlow(row)
		f.Items = items
		flows = append(flows, f)
	}
	return flows, nil
}

func (r *SpecflowRepository) Create(ctx context.Context, projectID, providerID int64, scheduledAt time.Time, changeNames []string) (domain.Flow, error) {
	row, err := r.q.CreateSpecflow(ctx, sqlcgen.CreateSpecflowParams{
		ProjectID:   projectID,
		ProviderID:  providerID,
		ScheduledAt: scheduledAt.UTC(),
		CreatedAt:   time.Now().UTC(),
	})
	if err != nil {
		return domain.Flow{}, err
	}
	for i, name := range changeNames {
		if _, err := r.q.CreateSpecflowItem(ctx, sqlcgen.CreateSpecflowItemParams{
			FlowID:     row.ID,
			Position:   int64(i),
			ChangeName: name,
		}); err != nil {
			return domain.Flow{}, err
		}
	}
	f, _, err := r.Get(ctx, row.ID)
	return f, err
}

// Reorder sets each item's position to its index in orderedItemIDs.
func (r *SpecflowRepository) Reorder(ctx context.Context, flowID int64, orderedItemIDs []int64) error {
	for i, itemID := range orderedItemIDs {
		if err := r.q.UpdateSpecflowItemPosition(ctx, sqlcgen.UpdateSpecflowItemPositionParams{
			Position: int64(i),
			ID:       itemID,
			FlowID:   flowID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *SpecflowRepository) Delete(ctx context.Context, id int64) error {
	if err := r.q.DeleteSpecflowItems(ctx, id); err != nil {
		return err
	}
	return r.q.DeleteSpecflow(ctx, id)
}

func (r *SpecflowRepository) SetStatus(ctx context.Context, id int64, status domain.FlowStatus) error {
	return r.q.UpdateSpecflowStatus(ctx, sqlcgen.UpdateSpecflowStatusParams{Status: string(status), ID: id})
}

func (r *SpecflowRepository) MarkMissed(ctx context.Context, id int64) error {
	return r.q.MarkSpecflowMissed(ctx, id)
}

func (r *SpecflowRepository) SetItemStatus(ctx context.Context, itemID int64, status domain.FlowItemStatus, log string, startedAt, finishedAt *time.Time) error {
	var started, finished sql.NullTime
	if startedAt != nil {
		started = sql.NullTime{Time: *startedAt, Valid: true}
	}
	if finishedAt != nil {
		finished = sql.NullTime{Time: *finishedAt, Valid: true}
	}
	return r.q.UpdateSpecflowItemStatus(ctx, sqlcgen.UpdateSpecflowItemStatusParams{
		ID:         itemID,
		Status:     string(status),
		Log:        log,
		StartedAt:  started,
		FinishedAt: finished,
	})
}

func (r *SpecflowRepository) listItems(ctx context.Context, flowID int64) ([]domain.FlowItem, error) {
	rows, err := r.q.ListSpecflowItems(ctx, flowID)
	if err != nil {
		return nil, err
	}
	items := make([]domain.FlowItem, 0, len(rows))
	for _, row := range rows {
		item := domain.FlowItem{
			ID:         row.ID,
			Position:   int(row.Position),
			ChangeName: row.ChangeName,
			Status:     domain.FlowItemStatus(row.Status),
			Log:        row.Log,
		}
		if row.StartedAt.Valid {
			t := row.StartedAt.Time
			item.StartedAt = &t
		}
		if row.FinishedAt.Valid {
			t := row.FinishedAt.Time
			item.FinishedAt = &t
		}
		items = append(items, item)
	}
	return items, nil
}

func toDomainFlow(row sqlcgen.Specflow) domain.Flow {
	return domain.Flow{
		ID:          row.ID,
		ProjectID:   row.ProjectID,
		ProviderID:  row.ProviderID,
		ScheduledAt: row.ScheduledAt,
		Status:      domain.FlowStatus(row.Status),
		Missed:      row.Missed != 0,
		CreatedAt:   row.CreatedAt,
	}
}
