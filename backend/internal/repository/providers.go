package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
	"github.com/davidlima/openspec-studio/backend/internal/repository/sqlcgen"
)

// ProviderRepository is the sole entry point for AI-provider config
// persistence (non-secret fields in SQLite; the API key itself lives in
// CredentialStore, never here).
type ProviderRepository struct {
	q           sqlcgen.Querier
	credentials *CredentialStore
}

func NewProviderRepository(db *sql.DB, credentials *CredentialStore) *ProviderRepository {
	return &ProviderRepository{q: Queries(db), credentials: credentials}
}

func (r *ProviderRepository) List(ctx context.Context) ([]domain.Provider, error) {
	rows, err := r.q.ListAIProviders(ctx)
	if err != nil {
		return nil, err
	}
	providers := make([]domain.Provider, 0, len(rows))
	for _, row := range rows {
		providers = append(providers, r.toDomain(row))
	}
	return providers, nil
}

func (r *ProviderRepository) Get(ctx context.Context, id int64) (domain.Provider, bool, error) {
	row, err := r.q.GetAIProvider(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Provider{}, false, nil
		}
		return domain.Provider{}, false, err
	}
	return r.toDomain(row), true, nil
}

func (r *ProviderRepository) Create(ctx context.Context, p domain.Provider, apiKey string) (domain.Provider, error) {
	isDefault := int64(0)
	if p.IsDefault {
		isDefault = 1
	}
	row, err := r.q.CreateAIProvider(ctx, sqlcgen.CreateAIProviderParams{
		Name:      p.Name,
		Kind:      string(p.Kind),
		Model:     p.Model,
		Endpoint:  p.Endpoint,
		CliPath:   p.CLIPath,
		CliArgs:   strings.Join(p.CLIArgs, " "),
		IsDefault: isDefault,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return domain.Provider{}, err
	}
	if apiKey != "" {
		if err := r.credentials.Set(row.ID, apiKey); err != nil {
			return domain.Provider{}, err
		}
	}
	if p.IsDefault {
		if err := r.SetDefault(ctx, row.ID); err != nil {
			return domain.Provider{}, err
		}
	}
	return r.toDomain(row), nil
}

func (r *ProviderRepository) Update(ctx context.Context, id int64, p domain.Provider, apiKey *string) (domain.Provider, error) {
	row, err := r.q.UpdateAIProvider(ctx, sqlcgen.UpdateAIProviderParams{
		ID:       id,
		Name:     p.Name,
		Model:    p.Model,
		Endpoint: p.Endpoint,
		CliPath:  p.CLIPath,
		CliArgs: strings.Join(p.CLIArgs, " "),
	})
	if err != nil {
		return domain.Provider{}, err
	}
	if apiKey != nil && *apiKey != "" {
		if err := r.credentials.Set(id, *apiKey); err != nil {
			return domain.Provider{}, err
		}
	}
	return r.toDomain(row), nil
}

func (r *ProviderRepository) Delete(ctx context.Context, id int64) error {
	if err := r.q.DeleteAIProvider(ctx, id); err != nil {
		return err
	}
	return r.credentials.Delete(id)
}

func (r *ProviderRepository) SetDefault(ctx context.Context, id int64) error {
	if err := r.q.ClearDefaultAIProvider(ctx); err != nil {
		return err
	}
	return r.q.SetDefaultAIProvider(ctx, id)
}

func (r *ProviderRepository) GetAPIKey(id int64) (string, error) {
	return r.credentials.Get(id)
}

func (r *ProviderRepository) toDomain(row sqlcgen.AiProvider) domain.Provider {
	var args []string
	if row.CliArgs != "" {
		args = strings.Fields(row.CliArgs)
	}
	_, err := r.credentials.Get(row.ID)
	return domain.Provider{
		ID:        row.ID,
		Name:      row.Name,
		Kind:      domain.ProviderKind(row.Kind),
		Model:     row.Model,
		Endpoint:  row.Endpoint,
		CLIPath:   row.CliPath,
		CLIArgs:   args,
		IsDefault: row.IsDefault != 0,
		CreatedAt: row.CreatedAt,
		HasAPIKey: err == nil,
	}
}
