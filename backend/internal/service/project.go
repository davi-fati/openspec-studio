package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
)

// ProjectRepository is the persistence contract this service depends on.
type ProjectRepository interface {
	Upsert(ctx context.Context, name, path string, openedAt time.Time) (domain.Project, error)
	List(ctx context.Context) ([]domain.Project, error)
	GetByPath(ctx context.Context, path string) (domain.Project, bool, error)
	GetByID(ctx context.Context, id int64) (domain.Project, bool, error)
}

type ProjectService struct {
	repo ProjectRepository
}

func NewProjectService(repo ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

// ErrPathNotFound is returned when a requested project directory does not exist.
var ErrPathNotFound = fmt.Errorf("path does not exist")

// ErrDirNotEmpty is returned when "create new project" targets a non-empty directory.
var ErrDirNotEmpty = fmt.Errorf("directory is not empty")

// ErrProjectNotFound is returned when a project id does not resolve to a registered project.
var ErrProjectNotFound = fmt.Errorf("project not found")

// ErrNoOpenSpecProject is returned when Open targets a directory that has no
// valid openspec/ structure - Open is read-only discovery; it never
// scaffolds. Use CreateProject ("New") to scaffold a fresh project instead.
var ErrNoOpenSpecProject = fmt.Errorf("no valid OpenSpec project at this path")

// OpenProject registers absPath as an active project. It is read-only: the
// directory must already contain a valid openspec/ structure, or it returns
// ErrNoOpenSpecProject rather than scaffolding one (that's CreateProject's job).
func (s *ProjectService) OpenProject(ctx context.Context, absPath string) (domain.Project, error) {
	info, err := os.Stat(absPath)
	if err != nil || !info.IsDir() {
		return domain.Project{}, ErrPathNotFound
	}

	if !hasOpenSpecStructure(absPath) {
		return domain.Project{}, ErrNoOpenSpecProject
	}

	name := filepath.Base(absPath)
	return s.repo.Upsert(ctx, name, absPath, time.Now())
}

// ListProjects returns every registered project, most recently opened first.
func (s *ProjectService) ListProjects(ctx context.Context) ([]domain.Project, error) {
	return s.repo.List(ctx)
}

// GetProject resolves a project by its registry id.
func (s *ProjectService) GetProject(ctx context.Context, id int64) (domain.Project, error) {
	p, ok, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Project{}, err
	}
	if !ok {
		return domain.Project{}, ErrProjectNotFound
	}
	return p, nil
}

// CreateProject scaffolds a brand-new OpenSpec project in absPath, which
// must exist and be empty (aside from dotfiles), and registers it.
func (s *ProjectService) CreateProject(ctx context.Context, absPath string) (domain.Project, error) {
	info, err := os.Stat(absPath)
	if err != nil || !info.IsDir() {
		return domain.Project{}, ErrPathNotFound
	}

	empty, err := isEmptyDir(absPath)
	if err != nil {
		return domain.Project{}, fmt.Errorf("read directory: %w", err)
	}
	if !empty {
		return domain.Project{}, ErrDirNotEmpty
	}

	if err := ensureOpenSpecScaffold(absPath); err != nil {
		return domain.Project{}, fmt.Errorf("scaffold openspec structure: %w", err)
	}

	name := filepath.Base(absPath)
	return s.repo.Upsert(ctx, name, absPath, time.Now())
}

// hasOpenSpecStructure reports whether dir already contains a valid
// openspec/ structure (specs/ and changes/ present). Shared by OpenProject's
// validation and the directory browser's per-entry "is this an existing
// OpenSpec project" flag, so the two can never disagree about what counts.
func hasOpenSpecStructure(dir string) bool {
	root := filepath.Join(dir, "openspec")
	specsInfo, err := os.Stat(filepath.Join(root, "specs"))
	if err != nil || !specsInfo.IsDir() {
		return false
	}
	changesInfo, err := os.Stat(filepath.Join(root, "changes"))
	if err != nil || !changesInfo.IsDir() {
		return false
	}
	return true
}

// isEmptyDir reports whether dir has no entries aside from dotfiles. Shared
// by CreateProject's validation and the directory browser's "is empty" flag.
func isEmptyDir(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), ".") {
			return false, nil
		}
	}
	return true, nil
}

// ensureOpenSpecScaffold creates the minimal openspec/ structure
// (specs/, changes/, config.yaml) in dir if it does not already exist,
// following the same layout `openspec init` produces.
func ensureOpenSpecScaffold(dir string) error {
	root := filepath.Join(dir, "openspec")
	specsDir := filepath.Join(root, "specs")
	changesDir := filepath.Join(root, "changes")
	archiveDir := filepath.Join(changesDir, "archive")
	configPath := filepath.Join(root, "config.yaml")

	for _, d := range []string{specsDir, changesDir, archiveDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		const defaultConfig = "schema: spec-driven\n"
		if err := os.WriteFile(configPath, []byte(defaultConfig), 0o644); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}
