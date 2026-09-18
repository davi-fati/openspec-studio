## Purpose

Lets a user create, edit, and delete OpenSpec specs and change proposals from within the Studio, with every artifact carrying provenance metadata.

## ADDED Requirements

### Requirement: Create a new project
The system SHALL let the user start a new project by choosing a directory and scaffolding OpenSpec structure in it (specs/, changes/, config.yaml), then register it as the active project.

#### Scenario: New project created from empty directory
- **WHEN** the user chooses "New Project" and selects an empty directory
- **THEN** the Studio scaffolds a valid OpenSpec structure there and opens it as the active project

### Requirement: Generate a spec from a description
The system SHALL let the user provide a short natural-language description and generate a draft spec (capability + requirements + scenarios) conforming to OpenSpec's format, for the user to review and edit before saving.

#### Scenario: Generate then edit before save
- **WHEN** the user submits a description and requests spec generation
- **THEN** the Studio produces a draft spec with at least one requirement and scenario, shown in an editable form, and does not write it to disk until the user confirms save

### Requirement: Edit an existing spec
The system SHALL let the user edit a spec's requirements and scenarios and persist changes back to that spec's `spec.md` file.

#### Scenario: Edit is persisted
- **WHEN** the user edits a requirement's text and saves
- **THEN** the corresponding `openspec/specs/<capability>/spec.md` file on disk reflects the change

### Requirement: Delete a spec
The system SHALL let the user delete a spec/capability, after confirmation, removing it from the project's `openspec/specs/` directory.

#### Scenario: Delete requires confirmation
- **WHEN** the user requests to delete a spec
- **THEN** the Studio asks for confirmation before removing the spec's files

#### Scenario: Delete blocked by active dependents
- **WHEN** the user requests to delete a spec that other specs declare a dependency on
- **THEN** the Studio warns which specs depend on it before allowing the deletion to proceed

### Requirement: Spec provenance metadata
Every spec or change proposal created or generated through the Studio SHALL record an id, created date and time, author, and owning project, and this metadata SHALL be displayed wherever the spec is shown in the UI.

#### Scenario: New spec has provenance
- **WHEN** a spec is created through the Studio (manually or generated)
- **THEN** it is stored with a unique id, an ISO 8601 created timestamp, the current author identity, and the owning project name

#### Scenario: Provenance shown in detail view
- **WHEN** the user opens a spec's detail view
- **THEN** the id, created date/time, author, and project are visible
