package domain

import "time"

// SpecProvenance records who created a spec through the Studio, when, and for
// which project - displayed wherever the spec is shown.
type SpecProvenance struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Author    string    `json:"author"`
	Project   string    `json:"project"`
}

// SpecDependencyRef is a declared link from one spec to another, possibly in
// a different registered project.
type SpecDependencyRef struct {
	ProjectPath string `json:"projectPath"`
	Capability  string `json:"capability"`
}

// SpecMeta is the sidecar metadata (provenance + dependencies) persisted
// alongside a spec.md as .meta.json, kept out of the human-readable spec
// content so it never interferes with OpenSpec's own parsing of spec.md.
type SpecMeta struct {
	Provenance SpecProvenance      `json:"provenance"`
	DependsOn  []SpecDependencyRef `json:"dependsOn"`
}

// SpecRequirement is one "### Requirement: <name>" block, kept as raw
// markdown (its description plus all of its scenarios) so edits round-trip
// through spec.md without needing a full markdown AST.
type SpecRequirement struct {
	Name string `json:"name"`
	Body string `json:"body"`
}

// SpecDependencyResolved is a dependency link resolved against the current
// project registry: Available is false when the target project is no
// longer registered or the target spec no longer exists.
type SpecDependencyResolved struct {
	ProjectPath string `json:"projectPath"`
	ProjectName string `json:"projectName"`
	Capability  string `json:"capability"`
	Available   bool   `json:"available"`
}

// Spec is a single OpenSpec capability spec as tracked by the Studio.
type Spec struct {
	Capability   string                    `json:"capability"`
	ProjectPath  string                    `json:"projectPath"`
	ProjectName  string                    `json:"projectName"`
	Purpose      string                    `json:"purpose"`
	Requirements []SpecRequirement         `json:"requirements"`
	Meta         SpecMeta                  `json:"meta"`
	DependsOn    []SpecDependencyResolved  `json:"dependsOn"`
	DependedBy   []SpecDependencyResolved  `json:"dependedBy"`
}

// ID is a stable identifier for this spec, opaque to the frontend, used to
// address it in the REST API without a project path in the URL.
func (s Spec) ID(projectID int64) string {
	return EncodeSpecID(projectID, s.Capability)
}
