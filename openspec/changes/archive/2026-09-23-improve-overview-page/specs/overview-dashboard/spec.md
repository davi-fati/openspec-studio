## MODIFIED Requirements

### Requirement: Per-project progress card
The system SHALL show one card per registered project summarizing: count of specs, count of changes by Kanban stage (Draft/Todo/In Progress/Done/Archived), an overall applied-vs-planned progress indicator that counts Done and Archived changes as completed work, and a count of specs with a broken cross-project dependency (if any). The progress indicator SHALL be shown in a distinct "complete" color when the project has no changes remaining in Todo or In Progress, and in a distinct "in progress" color otherwise. The card SHALL be operable as a single focusable control reachable and activatable via keyboard, not only by mouse click.

#### Scenario: Project with changes in multiple stages
- **WHEN** a project has changes in Draft, In Progress, and Done
- **THEN** its Overview card shows counts for each stage and a combined progress indicator

#### Scenario: Unavailable project is marked
- **WHEN** a registered project's path is currently inaccessible
- **THEN** its Overview card shows an unavailable state rather than stale or zeroed data

#### Scenario: Progress indicator counts archived changes as completed
- **WHEN** a project's non-draft changes are a mix of Todo, In Progress, Done, and Archived
- **THEN** the progress indicator treats Done and Archived changes as completed work when computing the ratio

#### Scenario: Fully closed project shows a complete-colored indicator
- **WHEN** a project has no changes remaining in Todo or In Progress (every non-draft change is Done or Archived)
- **THEN** its progress indicator is shown in the distinct "complete" color

#### Scenario: Project still in progress shows an in-progress-colored indicator
- **WHEN** a project has at least one change in Todo or In Progress
- **THEN** its progress indicator is shown in the distinct "in progress" color

#### Scenario: Project with a broken spec dependency
- **WHEN** one or more of a project's specs has a cross-project dependency that is no longer available
- **THEN** its Overview card shows a warning badge with the count of specs affected

#### Scenario: Project with no broken dependencies
- **WHEN** none of a project's specs has a broken cross-project dependency
- **THEN** its Overview card shows no warning badge

#### Scenario: Card is keyboard-operable
- **WHEN** the user tabs to a project's Overview card and activates it with the keyboard
- **THEN** the same navigation behavior fires as a mouse click on that card

### Requirement: Aggregate multi-project summary
The system SHALL show an aggregate summary across all registered projects: total open changes, total in-progress changes, and a most-recently-updated list where each entry names the project and its last-opened timestamp and is itself clickable to jump to that project.

#### Scenario: Aggregate reflects all projects
- **WHEN** the user has three registered projects with open changes
- **THEN** the aggregate summary's total open-changes count equals the sum across all three

#### Scenario: Recently-updated entry shows a timestamp and is clickable
- **WHEN** the user views the recently-updated list
- **THEN** each entry shows the project's name and its last-opened timestamp, and clicking an entry navigates to that project's Kanban tab with that project selected

### Requirement: Drill-through navigation
The system SHALL let the user choose, from a project's Overview card, which of that project's Kanban, Specs, or Specflow tabs to navigate to, pre-filtered to that project.

#### Scenario: Click card opens filtered Kanban
- **WHEN** the user activates a project's Overview card and chooses Kanban
- **THEN** the Studio switches to the Kanban tab with that project already selected in the project filter

#### Scenario: Choose Specs from a card
- **WHEN** the user activates a project's Overview card and chooses Specs
- **THEN** the Studio switches to the Specs tab with that project already selected in the project filter

#### Scenario: Choose Specflow from a card
- **WHEN** the user activates a project's Overview card and chooses Specflow
- **THEN** the Studio switches to the Specflow tab with that project already selected in the project filter

## ADDED Requirements

### Requirement: Loading state on initial load
The system SHALL show a loading indicator while the Overview page's initial data request is in flight, instead of a blank page.

#### Scenario: First load shows a loading indicator
- **WHEN** the user opens the Overview tab and the aggregate data has not yet arrived
- **THEN** a loading indicator is shown in place of the summary and project cards

#### Scenario: Loading indicator clears once data arrives
- **WHEN** the aggregate data finishes loading
- **THEN** the loading indicator is replaced by the summary and project cards

### Requirement: Actionable empty state
The system SHALL let the user open or scaffold a project directly from the Overview page's empty state, without needing to use the header controls, when no projects are registered.

#### Scenario: Empty state offers Open and New
- **WHEN** no projects are registered and the user views the Overview page
- **THEN** the empty state shows both an Open and a New action that launch the same in-app directory browser as the header's equivalent controls
