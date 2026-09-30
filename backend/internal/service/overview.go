package service

import (
	"context"
	"sort"

	"github.com/davidlima/openspec-studio/backend/internal/domain"
)

type OverviewService struct {
	projects SpecProjectLister
	specs    *SpecService
	changes  *ChangeService
}

func NewOverviewService(projects SpecProjectLister, specs *SpecService, changes *ChangeService) *OverviewService {
	return &OverviewService{projects: projects, specs: specs, changes: changes}
}

// GetOverview aggregates spec/change progress across every registered
// project in one pass, so the "recently updated" ordering and aggregate
// totals are always consistent with each other (design.md's single-endpoint
// decision) rather than assembled from N separate client-side fetches. When
// projectID is non-nil, the aggregation runs over just that one project, so
// totals/recentlyUpdated end up scoped without a separate branch.
func (s *OverviewService) GetOverview(ctx context.Context, projectID *int64) (domain.Overview, error) {
	allProjects, err := s.projects.ListProjects(ctx)
	if err != nil {
		return domain.Overview{}, err
	}

	projects := allProjects
	if projectID != nil {
		projects = nil
		for _, p := range allProjects {
			if p.ID == *projectID {
				projects = append(projects, p)
				break
			}
		}
	}

	allSpecs, err := s.specs.ListSpecs(ctx)
	if err != nil {
		return domain.Overview{}, err
	}
	specCountByPath := make(map[string]int)
	brokenByPath := make(map[string]int)
	for _, sp := range allSpecs {
		specCountByPath[sp.ProjectPath]++
		broken := false
		for _, d := range sp.DependsOn {
			if !d.Available {
				broken = true
				break
			}
		}
		if !broken {
			for _, d := range sp.DependedBy {
				if !d.Available {
					broken = true
					break
				}
			}
		}
		if broken {
			brokenByPath[sp.ProjectPath]++
		}
	}

	allChanges, err := s.changes.ListChanges(ctx)
	if err != nil {
		return domain.Overview{}, err
	}
	changesByPath := make(map[string][]domain.Change)
	for _, ch := range allChanges {
		changesByPath[ch.ProjectPath] = append(changesByPath[ch.ProjectPath], ch)
	}

	overview := domain.Overview{
		Projects:        make([]domain.ProjectOverview, 0, len(projects)),
		RecentlyUpdated: make([]domain.RecentProject, 0, len(projects)),
	}
	for _, p := range projects {
		po := domain.ProjectOverview{
			ProjectID:       p.ID,
			ProjectName:     p.Name,
			Available:       p.Available,
			SpecCount:       specCountByPath[p.Path],
			BrokenSpecCount: brokenByPath[p.Path],
			ChangeCounts: map[string]int{
				string(domain.ChangeStatusDraft):      0,
				string(domain.ChangeStatusTodo):        0,
				string(domain.ChangeStatusInProgress):  0,
				string(domain.ChangeStatusDone):        0,
				string(domain.ChangeStatusArchived):    0,
			},
		}

		if p.Available {
			for _, ch := range changesByPath[p.Path] {
				po.ChangeCounts[string(ch.Status)]++
			}
			completed := po.ChangeCounts[string(domain.ChangeStatusDone)] + po.ChangeCounts[string(domain.ChangeStatusArchived)]
			planned := po.ChangeCounts[string(domain.ChangeStatusTodo)] +
				po.ChangeCounts[string(domain.ChangeStatusInProgress)] +
				completed
			if planned > 0 {
				po.ProgressRatio = float64(completed) / float64(planned)
			}

			overview.TotalOpenChanges += po.ChangeCounts[string(domain.ChangeStatusDraft)] +
				po.ChangeCounts[string(domain.ChangeStatusTodo)] +
				po.ChangeCounts[string(domain.ChangeStatusInProgress)]
			overview.TotalInProgress += po.ChangeCounts[string(domain.ChangeStatusInProgress)]
		}

		overview.Projects = append(overview.Projects, po)
	}

	sorted := make([]domain.Project, len(projects))
	copy(sorted, projects)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].LastOpenedAt.After(sorted[j].LastOpenedAt) })
	for _, p := range sorted {
		overview.RecentlyUpdated = append(overview.RecentlyUpdated, domain.RecentProject{
			ProjectID:    p.ID,
			ProjectName:  p.Name,
			LastOpenedAt: p.LastOpenedAt,
		})
	}

	return overview, nil
}
