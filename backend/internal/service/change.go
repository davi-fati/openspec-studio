package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
)

var ErrChangeNotFound = errors.New("change not found")

var taskCheckboxRe = regexp.MustCompile(`(?m)^\s*-\s*\[( |x|X)\]`)

type ChangeService struct {
	projects SpecProjectLister
}

func NewChangeService(projects SpecProjectLister) *ChangeService {
	return &ChangeService{projects: projects}
}

// ListChanges returns every change across every registered, available
// project - both active (openspec/changes/*) and archived
// (openspec/changes/archive/*) - with lifecycle status computed.
func (s *ChangeService) ListChanges(ctx context.Context) ([]domain.Change, error) {
	projects, err := s.projects.ListProjects(ctx)
	if err != nil {
		return nil, err
	}

	var changes []domain.Change
	for _, p := range projects {
		if !p.Available {
			continue
		}
		found, err := walkProjectChanges(p)
		if err != nil {
			return nil, err
		}
		changes = append(changes, found...)
	}
	return changes, nil
}

// GetChangeDetail resolves one change's full artifact content for the
// Kanban card detail view.
func (s *ChangeService) GetChangeDetail(ctx context.Context, projectID int64, name string) (domain.ChangeDetail, error) {
	p, err := s.projects.GetProject(ctx, projectID)
	if err != nil {
		return domain.ChangeDetail{}, err
	}

	dir, base, err := findChangeDir(p.Path, name)
	if err != nil {
		return domain.ChangeDetail{}, err
	}

	change := changeFromDir(p, dir, base == "archive")

	detail := domain.ChangeDetail{Change: change}
	detail.ProposalMarkdown = readFileIfExists(filepath.Join(dir, "proposal.md"))
	detail.DesignMarkdown = readFileIfExists(filepath.Join(dir, "design.md"))
	detail.TasksMarkdown = readFileIfExists(filepath.Join(dir, "tasks.md"))

	specsDir := filepath.Join(dir, "specs")
	if entries, err := os.ReadDir(specsDir); err == nil {
		detail.Specs = make(map[string]string)
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			specPath := filepath.Join(specsDir, e.Name(), "spec.md")
			if content := readFileIfExists(specPath); content != "" {
				detail.Specs[e.Name()] = content
			}
		}
	}

	return detail, nil
}

func findChangeDir(projectPath, name string) (dir string, location string, err error) {
	active := filepath.Join(projectPath, "openspec", "changes", name)
	if info, statErr := os.Stat(active); statErr == nil && info.IsDir() {
		return active, "active", nil
	}
	archived := filepath.Join(projectPath, "openspec", "changes", "archive", name)
	if info, statErr := os.Stat(archived); statErr == nil && info.IsDir() {
		return archived, "archive", nil
	}
	return "", "", ErrChangeNotFound
}

func walkProjectChanges(p domain.Project) ([]domain.Change, error) {
	var changes []domain.Change

	changesRoot := filepath.Join(p.Path, "openspec", "changes")
	entries, err := os.ReadDir(changesRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	for _, e := range entries {
		if !e.IsDir() || e.Name() == "archive" {
			continue
		}
		changes = append(changes, changeFromDir(p, filepath.Join(changesRoot, e.Name()), false))
	}

	archiveRoot := filepath.Join(changesRoot, "archive")
	archiveEntries, err := os.ReadDir(archiveRoot)
	if err == nil {
		for _, e := range archiveEntries {
			if !e.IsDir() {
				continue
			}
			changes = append(changes, changeFromDir(p, filepath.Join(archiveRoot, e.Name()), true))
		}
	}

	return changes, nil
}

func changeFromDir(p domain.Project, dir string, archived bool) domain.Change {
	name := filepath.Base(dir)
	hasProposal := fileExists(filepath.Join(dir, "proposal.md"))
	hasDesign := fileExists(filepath.Join(dir, "design.md"))
	hasTasks := fileExists(filepath.Join(dir, "tasks.md"))
	hasSpecs := changeHasSpecs(dir)

	total, done := 0, 0
	if hasTasks {
		total, done = countTaskCheckboxes(filepath.Join(dir, "tasks.md"))
	}

	status := domain.ComputeStatus(hasProposal, hasTasks, total, done)
	if archived {
		status = domain.ChangeStatusArchived
	}

	return domain.Change{
		Name:        name,
		ProjectPath: p.Path,
		ProjectName: p.Name,
		Status:      status,
		HasProposal: hasProposal,
		HasDesign:   hasDesign,
		HasSpecs:    hasSpecs,
		HasTasks:    hasTasks,
		TasksTotal:  total,
		TasksDone:   done,
	}
}

// changeHasSpecs reports whether a change has at least one delta spec under
// its own specs/ directory.
func changeHasSpecs(dir string) bool {
	entries, err := os.ReadDir(filepath.Join(dir, "specs"))
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() && fileExists(filepath.Join(dir, "specs", e.Name(), "spec.md")) {
			return true
		}
	}
	return false
}

func countTaskCheckboxes(path string) (total, done int) {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, 0
	}
	matches := taskCheckboxRe.FindAllStringSubmatch(string(content), -1)
	for _, m := range matches {
		total++
		if strings.EqualFold(m[1], "x") {
			done++
		}
	}
	return total, done
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readFileIfExists(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}
