package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
)

var (
	ErrSpecNotFound      = errors.New("spec not found")
	ErrSpecExists        = errors.New("spec already exists")
	ErrSelfDependency    = errors.New("a spec cannot depend on itself")
	ErrSpecHasDependents = errors.New("spec has dependents")
)

// SpecProjectLister is the subset of ProjectService a SpecService needs: the
// registered project list, used to walk every project's openspec/specs/ and
// to resolve cross-project dependency links.
type SpecProjectLister interface {
	ListProjects(ctx context.Context) ([]domain.Project, error)
	GetProject(ctx context.Context, id int64) (domain.Project, error)
}

// ProviderInvoker is the subset of *ProviderService generation needs: run a
// prompt against the configured (or explicitly chosen) AI provider.
type ProviderInvoker interface {
	Invoke(ctx context.Context, providerID int64, prompt string) (string, error)
}

type SpecService struct {
	projects  SpecProjectLister
	author    string
	providers ProviderInvoker // optional; nil falls back to the deterministic template
}

func NewSpecService(projects SpecProjectLister, author string) *SpecService {
	return &SpecService{projects: projects, author: author}
}

// SetProviderInvoker wires spec generation to an AI provider (add-ai-assistant-config).
// Without it, GenerateDraft/GenerateDraftWithProvider fall back to the
// deterministic template - generation always works, provider quality is additive.
func (s *SpecService) SetProviderInvoker(p ProviderInvoker) {
	s.providers = p
}

// ListSpecs returns every spec across every registered, available project,
// with dependency links resolved (marking broken links) and DependedBy
// computed by cross-referencing every spec's declared dependencies.
func (s *SpecService) ListSpecs(ctx context.Context) ([]domain.Spec, error) {
	projects, err := s.projects.ListProjects(ctx)
	if err != nil {
		return nil, err
	}

	byProjectPath := make(map[string]domain.Project, len(projects))
	for _, p := range projects {
		byProjectPath[p.Path] = p
	}

	var specs []domain.Spec
	for _, p := range projects {
		if !p.Available {
			continue
		}
		found, err := walkProjectSpecs(p)
		if err != nil {
			return nil, err
		}
		specs = append(specs, found...)
	}

	resolveDependencies(specs, byProjectPath)
	return specs, nil
}

// GetSpec resolves a single spec by project id + capability.
func (s *SpecService) GetSpec(ctx context.Context, projectID int64, capability string) (domain.Spec, error) {
	specs, err := s.ListSpecs(ctx)
	if err != nil {
		return domain.Spec{}, err
	}
	p, err := s.projects.GetProject(ctx, projectID)
	if err != nil {
		return domain.Spec{}, err
	}
	for _, sp := range specs {
		if sp.ProjectPath == p.Path && sp.Capability == capability {
			return sp, nil
		}
	}
	return domain.Spec{}, ErrSpecNotFound
}

// CreateSpec writes a new spec.md + .meta.json for capability in the given project.
func (s *SpecService) CreateSpec(ctx context.Context, projectID int64, capability, purpose string, requirements []domain.SpecRequirement) (domain.Spec, error) {
	p, err := s.projects.GetProject(ctx, projectID)
	if err != nil {
		return domain.Spec{}, err
	}

	dir := filepath.Join(p.Path, "openspec", "specs", capability)
	specPath := filepath.Join(dir, "spec.md")
	if _, err := os.Stat(specPath); err == nil {
		return domain.Spec{}, ErrSpecExists
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return domain.Spec{}, fmt.Errorf("create spec directory: %w", err)
	}

	content := renderSpecMarkdown(purpose, requirements)
	if err := os.WriteFile(specPath, []byte(content), 0o644); err != nil {
		return domain.Spec{}, fmt.Errorf("write spec.md: %w", err)
	}

	meta := domain.SpecMeta{
		Provenance: domain.SpecProvenance{
			ID:        uuid.NewString(),
			CreatedAt: time.Now().UTC(),
			Author:    s.author,
			Project:   p.Name,
		},
	}
	if err := writeSpecMeta(dir, meta); err != nil {
		return domain.Spec{}, fmt.Errorf("write spec metadata: %w", err)
	}

	return s.GetSpec(ctx, projectID, capability)
}

// UpdateSpec overwrites a spec's purpose and requirements, preserving its
// existing provenance and dependency metadata.
func (s *SpecService) UpdateSpec(ctx context.Context, projectID int64, capability, purpose string, requirements []domain.SpecRequirement) (domain.Spec, error) {
	p, err := s.projects.GetProject(ctx, projectID)
	if err != nil {
		return domain.Spec{}, err
	}

	dir := filepath.Join(p.Path, "openspec", "specs", capability)
	specPath := filepath.Join(dir, "spec.md")
	if _, err := os.Stat(specPath); err != nil {
		return domain.Spec{}, ErrSpecNotFound
	}

	content := renderSpecMarkdown(purpose, requirements)
	if err := os.WriteFile(specPath, []byte(content), 0o644); err != nil {
		return domain.Spec{}, fmt.Errorf("write spec.md: %w", err)
	}

	return s.GetSpec(ctx, projectID, capability)
}

// DeleteSpec removes a spec's directory. If other specs depend on it and
// force is false, it returns ErrSpecHasDependents along with the dependent
// specs rather than deleting anything.
func (s *SpecService) DeleteSpec(ctx context.Context, projectID int64, capability string, force bool) ([]domain.Spec, error) {
	p, err := s.projects.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	target, err := s.GetSpec(ctx, projectID, capability)
	if err != nil {
		return nil, err
	}

	if !force && len(target.DependedBy) > 0 {
		specs, err := s.ListSpecs(ctx)
		if err != nil {
			return nil, err
		}
		var dependents []domain.Spec
		for _, sp := range specs {
			for _, dep := range target.DependedBy {
				if sp.ProjectPath == dep.ProjectPath && sp.Capability == dep.Capability {
					dependents = append(dependents, sp)
				}
			}
		}
		return dependents, ErrSpecHasDependents
	}

	dir := filepath.Join(p.Path, "openspec", "specs", capability)
	if err := os.RemoveAll(dir); err != nil {
		return nil, fmt.Errorf("delete spec directory: %w", err)
	}
	return nil, nil
}

// AddDependency links (projectID, capability) -> (targetProjectID, targetCapability).
func (s *SpecService) AddDependency(ctx context.Context, projectID int64, capability string, targetProjectID int64, targetCapability string) error {
	p, err := s.projects.GetProject(ctx, projectID)
	if err != nil {
		return err
	}
	targetProject, err := s.projects.GetProject(ctx, targetProjectID)
	if err != nil {
		return err
	}

	if p.Path == targetProject.Path && capability == targetCapability {
		return ErrSelfDependency
	}

	dir := filepath.Join(p.Path, "openspec", "specs", capability)
	meta, err := readSpecMeta(dir)
	if err != nil {
		return err
	}

	for _, dep := range meta.DependsOn {
		if dep.ProjectPath == targetProject.Path && dep.Capability == targetCapability {
			return nil // already linked
		}
	}
	meta.DependsOn = append(meta.DependsOn, domain.SpecDependencyRef{
		ProjectPath: targetProject.Path,
		Capability:  targetCapability,
	})
	return writeSpecMeta(dir, meta)
}

// RemoveDependency removes a previously declared link.
func (s *SpecService) RemoveDependency(ctx context.Context, projectID int64, capability string, targetProjectID int64, targetCapability string) error {
	p, err := s.projects.GetProject(ctx, projectID)
	if err != nil {
		return err
	}
	targetProject, err := s.projects.GetProject(ctx, targetProjectID)
	if err != nil {
		return err
	}

	dir := filepath.Join(p.Path, "openspec", "specs", capability)
	meta, err := readSpecMeta(dir)
	if err != nil {
		return err
	}

	filtered := meta.DependsOn[:0]
	for _, dep := range meta.DependsOn {
		if dep.ProjectPath == targetProject.Path && dep.Capability == targetCapability {
			continue
		}
		filtered = append(filtered, dep)
	}
	meta.DependsOn = filtered
	return writeSpecMeta(dir, meta)
}

// GenerateDraft produces a draft spec from a short description, without
// writing anything to disk. If a provider is wired (SetProviderInvoker) it
// asks the provider to draft the spec in OpenSpec's own format and parses
// the result; on no provider, a failed invocation, or output that doesn't
// parse into at least one requirement, it falls back to the deterministic
// template - generation always succeeds, provider quality is additive.
func (s *SpecService) GenerateDraft(ctx context.Context, description string, providerID int64) domain.Spec {
	if s.providers != nil {
		if draft, ok := s.generateDraftViaProvider(ctx, description, providerID); ok {
			return draft
		}
	}
	return s.generateDraftTemplate(description)
}

func (s *SpecService) generateDraftViaProvider(ctx context.Context, description string, providerID int64) (domain.Spec, bool) {
	prompt := "Draft an OpenSpec capability spec for: " + description + "\n\n" +
		"Respond ONLY in this exact format:\n" +
		"## Purpose\n<one or two sentences>\n\n## Requirements\n\n" +
		"### Requirement: <name>\nThe system SHALL <behavior>.\n\n#### Scenario: <name>\n- **WHEN** <condition>\n- **THEN** <outcome>\n"

	out, err := s.providers.Invoke(ctx, providerID, prompt)
	if err != nil {
		return domain.Spec{}, false
	}

	parsed := parseSpecMarkdown(out)
	if len(parsed.requirements) == 0 {
		return domain.Spec{}, false
	}

	slug := slugify(description)
	if slug == "" {
		slug = "new-capability"
	}
	return domain.Spec{
		Capability:   slug,
		Purpose:      parsed.purpose,
		Requirements: parsed.requirements,
	}, true
}

func (s *SpecService) generateDraftTemplate(description string) domain.Spec {
	slug := slugify(description)
	if slug == "" {
		slug = "new-capability"
	}
	return domain.Spec{
		Capability: slug,
		Purpose:    strings.TrimSpace(description),
		Requirements: []domain.SpecRequirement{
			{
				Name: "TODO: name this requirement",
				Body: "The system SHALL " + strings.TrimSuffix(strings.TrimSpace(description), ".") + ".\n\n#### Scenario: TODO: name this scenario\n- **WHEN** TODO\n- **THEN** TODO",
			},
		},
	}
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > 60 {
		out = out[:60]
	}
	return out
}

// --- filesystem helpers ---

func walkProjectSpecs(p domain.Project) ([]domain.Spec, error) {
	root := filepath.Join(p.Path, "openspec", "specs")
	var specs []domain.Spec

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() || d.Name() != "spec.md" {
			return nil
		}

		rel, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		capability := filepath.ToSlash(rel)

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		parsed := parseSpecMarkdown(string(content))

		meta, err := readSpecMeta(filepath.Dir(path))
		if err != nil {
			return err
		}

		specs = append(specs, domain.Spec{
			Capability:   capability,
			ProjectPath:  p.Path,
			ProjectName:  p.Name,
			Purpose:      parsed.purpose,
			Requirements: parsed.requirements,
			Meta:         meta,
		})
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return specs, nil
}

func resolveDependencies(specs []domain.Spec, byProjectPath map[string]domain.Project) {
	exists := make(map[string]bool, len(specs))
	for _, sp := range specs {
		exists[sp.ProjectPath+"\x00"+sp.Capability] = true
	}

	for i := range specs {
		sp := &specs[i]
		for _, dep := range sp.Meta.DependsOn {
			key := dep.ProjectPath + "\x00" + dep.Capability
			_, projectRegistered := byProjectPath[dep.ProjectPath]
			available := projectRegistered && exists[key]
			sp.DependsOn = append(sp.DependsOn, domain.SpecDependencyResolved{
				ProjectPath: dep.ProjectPath,
				ProjectName: byProjectPath[dep.ProjectPath].Name,
				Capability:  dep.Capability,
				Available:   available,
			})
		}
	}

	for i := range specs {
		sp := &specs[i]
		for _, other := range specs {
			for _, dep := range other.Meta.DependsOn {
				if dep.ProjectPath == sp.ProjectPath && dep.Capability == sp.Capability {
					sp.DependedBy = append(sp.DependedBy, domain.SpecDependencyResolved{
						ProjectPath: other.ProjectPath,
						ProjectName: other.ProjectName,
						Capability:  other.Capability,
						Available:   true,
					})
				}
			}
		}
	}
}

func readSpecMeta(dir string) (domain.SpecMeta, error) {
	path := filepath.Join(dir, ".meta.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.SpecMeta{}, nil
		}
		return domain.SpecMeta{}, err
	}
	var meta domain.SpecMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return domain.SpecMeta{}, fmt.Errorf("parse .meta.json: %w", err)
	}
	return meta, nil
}

func writeSpecMeta(dir string, meta domain.SpecMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ".meta.json"), data, 0o644)
}
