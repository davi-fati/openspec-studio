package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
)

var ErrContextItemNotFound = errors.New("context item not found")

type ContextService struct {
	projects SpecProjectLister
}

func NewContextService(projects SpecProjectLister) *ContextService {
	return &ContextService{projects: projects}
}

type reviewFrontmatter struct {
	ID        string             `yaml:"id"`
	Title     string             `yaml:"title"`
	Link      *domain.ArtifactRef `yaml:"link,omitempty"`
	CreatedAt time.Time          `yaml:"createdAt"`
}

type debtFrontmatter struct {
	ID        string             `yaml:"id"`
	Title     string             `yaml:"title"`
	Status    domain.DebtStatus  `yaml:"status"`
	Link      *domain.ArtifactRef `yaml:"link,omitempty"`
	CreatedAt time.Time          `yaml:"createdAt"`
}

// --- Reviews ---

func (s *ContextService) ListReviews(ctx context.Context, projectID int64) ([]domain.Review, error) {
	p, err := s.projects.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(p.Path, "openspec", "context", "reviews")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var reviews []domain.Review
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var fm reviewFrontmatter
		body, err := parseFrontmatter(string(content), &fm)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, domain.Review{
			ID:          fm.ID,
			ProjectPath: p.Path,
			ProjectName: p.Name,
			Title:       fm.Title,
			Body:        strings.TrimSpace(body),
			Link:        derefArtifactRef(fm.Link),
			CreatedAt:   fm.CreatedAt,
		})
	}
	sort.Slice(reviews, func(i, j int) bool { return reviews[i].CreatedAt.After(reviews[j].CreatedAt) })
	return reviews, nil
}

func (s *ContextService) CreateReview(ctx context.Context, projectID int64, title, body string, link domain.ArtifactRef) (domain.Review, error) {
	p, err := s.projects.GetProject(ctx, projectID)
	if err != nil {
		return domain.Review{}, err
	}

	review := domain.Review{
		ID:          uuid.NewString(),
		ProjectPath: p.Path,
		ProjectName: p.Name,
		Title:       title,
		Body:        body,
		Link:        link,
		CreatedAt:   time.Now().UTC(),
	}

	dir := filepath.Join(p.Path, "openspec", "context", "reviews")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return domain.Review{}, fmt.Errorf("create reviews directory: %w", err)
	}

	fm := reviewFrontmatter{ID: review.ID, Title: title, Link: refArtifactRef(link), CreatedAt: review.CreatedAt}
	content, err := renderFrontmatter(fm, body)
	if err != nil {
		return domain.Review{}, err
	}
	if err := os.WriteFile(filepath.Join(dir, review.ID+".md"), []byte(content), 0o644); err != nil {
		return domain.Review{}, fmt.Errorf("write review: %w", err)
	}
	return review, nil
}

func (s *ContextService) DeleteReview(ctx context.Context, projectID int64, id string) error {
	p, err := s.projects.GetProject(ctx, projectID)
	if err != nil {
		return err
	}
	path := filepath.Join(p.Path, "openspec", "context", "reviews", id+".md")
	if _, err := os.Stat(path); err != nil {
		return ErrContextItemNotFound
	}
	return os.Remove(path)
}

// --- Tech debt ---

func (s *ContextService) ListDebts(ctx context.Context, projectID int64) ([]domain.DebtItem, error) {
	p, err := s.projects.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(p.Path, "openspec", "context", "debts")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var debts []domain.DebtItem
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var fm debtFrontmatter
		body, err := parseFrontmatter(string(content), &fm)
		if err != nil {
			return nil, err
		}
		debts = append(debts, domain.DebtItem{
			ID:          fm.ID,
			ProjectPath: p.Path,
			ProjectName: p.Name,
			Title:       fm.Title,
			Body:        strings.TrimSpace(body),
			Status:      fm.Status,
			Link:        derefArtifactRef(fm.Link),
			CreatedAt:   fm.CreatedAt,
		})
	}
	sort.Slice(debts, func(i, j int) bool { return debts[i].CreatedAt.After(debts[j].CreatedAt) })
	return debts, nil
}

func (s *ContextService) CreateDebt(ctx context.Context, projectID int64, title, body string, link domain.ArtifactRef) (domain.DebtItem, error) {
	p, err := s.projects.GetProject(ctx, projectID)
	if err != nil {
		return domain.DebtItem{}, err
	}

	debt := domain.DebtItem{
		ID:          uuid.NewString(),
		ProjectPath: p.Path,
		ProjectName: p.Name,
		Title:       title,
		Body:        body,
		Status:      domain.DebtStatusOpen,
		Link:        link,
		CreatedAt:   time.Now().UTC(),
	}

	dir := filepath.Join(p.Path, "openspec", "context", "debts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return domain.DebtItem{}, fmt.Errorf("create debts directory: %w", err)
	}

	if err := s.writeDebt(dir, debt); err != nil {
		return domain.DebtItem{}, err
	}
	return debt, nil
}

func (s *ContextService) UpdateDebtStatus(ctx context.Context, projectID int64, id string, status domain.DebtStatus) (domain.DebtItem, error) {
	p, err := s.projects.GetProject(ctx, projectID)
	if err != nil {
		return domain.DebtItem{}, err
	}
	dir := filepath.Join(p.Path, "openspec", "context", "debts")
	path := filepath.Join(dir, id+".md")
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.DebtItem{}, ErrContextItemNotFound
	}
	var fm debtFrontmatter
	body, err := parseFrontmatter(string(content), &fm)
	if err != nil {
		return domain.DebtItem{}, err
	}
	fm.Status = status

	debt := domain.DebtItem{
		ID:          fm.ID,
		ProjectPath: p.Path,
		ProjectName: p.Name,
		Title:       fm.Title,
		Body:        strings.TrimSpace(body),
		Status:      status,
		Link:        derefArtifactRef(fm.Link),
		CreatedAt:   fm.CreatedAt,
	}
	if err := s.writeDebt(dir, debt); err != nil {
		return domain.DebtItem{}, err
	}
	return debt, nil
}

func (s *ContextService) DeleteDebt(ctx context.Context, projectID int64, id string) error {
	p, err := s.projects.GetProject(ctx, projectID)
	if err != nil {
		return err
	}
	path := filepath.Join(p.Path, "openspec", "context", "debts", id+".md")
	if _, err := os.Stat(path); err != nil {
		return ErrContextItemNotFound
	}
	return os.Remove(path)
}

func (s *ContextService) writeDebt(dir string, debt domain.DebtItem) error {
	fm := debtFrontmatter{ID: debt.ID, Title: debt.Title, Status: debt.Status, Link: refArtifactRef(debt.Link), CreatedAt: debt.CreatedAt}
	content, err := renderFrontmatter(fm, debt.Body)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, debt.ID+".md"), []byte(content), 0o644)
}

func refArtifactRef(r domain.ArtifactRef) *domain.ArtifactRef {
	if r.Kind == "" {
		return nil
	}
	return &r
}

func derefArtifactRef(r *domain.ArtifactRef) domain.ArtifactRef {
	if r == nil {
		return domain.ArtifactRef{}
	}
	return *r
}
