## Purpose

Lets a user declare that a spec depends on one or more other specs, including specs in other opened projects, and see those links visualized.

## ADDED Requirements

### Requirement: Declare a dependency between specs
The system SHALL let the user link a spec to one or more other specs it depends on, including specs belonging to a different opened project.

#### Scenario: Link within same project
- **WHEN** the user adds a dependency from spec A to spec B in the same project
- **THEN** the link is saved as part of spec A's provenance/dependency metadata

#### Scenario: Link across projects
- **WHEN** the user adds a dependency from a spec in project X to a spec in project Y
- **THEN** the Studio saves the cross-project link and can resolve it as long as project Y remains registered

#### Scenario: Cannot link a spec to itself
- **WHEN** the user attempts to add a spec as its own dependency
- **THEN** the Studio rejects the action

### Requirement: Visualize spec dependencies
The system SHALL display a spec's declared dependencies and dependents in its detail view, and SHALL indicate on Kanban/Specs list views when a spec has unresolved (missing/deleted) dependencies.

#### Scenario: Detail view shows dependency graph
- **WHEN** the user opens a spec's detail view
- **THEN** the Studio lists specs it depends on and specs that depend on it

#### Scenario: Broken dependency is flagged
- **WHEN** a spec's declared dependency no longer exists (deleted or project unregistered)
- **THEN** the Studio visually flags the broken link rather than silently ignoring it
